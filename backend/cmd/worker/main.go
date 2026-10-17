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
	for ctx.Err() == nil {
		worked, failure := app.Worker.Once(ctx)
		if failure != nil {
			log.Printf("image job failure: %T", failure)
		}
		if !worked || failure != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}
}
