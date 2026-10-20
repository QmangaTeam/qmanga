package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"qmanga/internal/bootstrap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	app, failure := bootstrap.Initialize(ctx)
	if failure != nil {
		log.Fatal(failure)
	}
	defer app.Close()
	// Retry durable avatar cleanup even when no new upload or image job arrives.
	avatarTicker := time.NewTicker(time.Minute)
	defer avatarTicker.Stop()
	go func() {
		drain := func() {
			if err := app.Avatars.DrainAvatarGarbage(ctx); err != nil && ctx.Err() == nil {
				log.Printf("avatar cleanup failure: %T", err)
			}
		}
		drain()
		for {
			select {
			case <-ctx.Done():
				return
			case <-avatarTicker.C:
				drain()
			}
		}
	}()
	for ctx.Err() == nil {
		notified, notificationError := app.Notifications.Deliver(ctx)
		if notificationError != nil {
			log.Printf("notification delivery failure: %T", notificationError)
		}
		worked, failure := app.Worker.Once(ctx)
		if failure != nil {
			log.Printf("image job failure: %T", failure)
		}
		if (!worked && !notified) || failure != nil || notificationError != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}
}
