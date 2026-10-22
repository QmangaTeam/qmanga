package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"qmanga/internal/bootstrap"
	notificationPostgres "qmanga/internal/notifications/adapters/postgres"
	"qmanga/internal/notifications/adapters/telegram"
	notificationApp "qmanga/internal/notifications/application"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	app, failure := bootstrap.Initialize(ctx)
	if failure != nil {
		log.Fatal(failure)
	}
	defer app.Close()
	// Telegram is optional and configured only on the production worker.
	token, chat := os.Getenv("TELEGRAM_BOT_TOKEN"), os.Getenv("TELEGRAM_CHAT_ID")
	if token != "" || chat != "" {
		sender, err := telegram.New(token, chat, app.Config.AppOrigin)
		if err != nil {
			log.Fatal("invalid Telegram configuration")
		}
		notifier := notificationApp.NewModerationNotifier(notificationPostgres.New(app.PostgresClient.Pool()), sender)
		log.Print("Telegram moderation notifier enabled")
		go func() {
			for ctx.Err() == nil {
				worked, err := notifier.Deliver(ctx)
				if err != nil && ctx.Err() == nil {
					log.Print("Telegram moderation delivery failed; queued for retry")
				}
				pause := time.Second
				if !worked || err != nil {
					pause = 5 * time.Second
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(pause):
				}
			}
		}()
	}
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
		digested, digestError := app.Digest.Deliver(ctx)
		if digestError != nil {
			log.Printf("weekly digest delivery failed: %T", digestError)
		}
		if notificationError != nil {
			log.Printf("notification delivery failure: %T", notificationError)
		}
		worked, failure := app.Worker.Once(ctx)
		if failure != nil {
			log.Printf("image job failure: %T", failure)
		}
		if (!worked && !notified && !digested) || failure != nil || notificationError != nil || digestError != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}
}
