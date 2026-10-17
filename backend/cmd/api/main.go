package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"qmanga/internal/config"
	"qmanga/internal/httpapi"
	"qmanga/internal/storage"
	"qmanga/internal/store"
)

func main() {
	log.Println("Starting Qmanga API backend...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Fatal config error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Printf("Connecting to database at %s (environment: %s)...", maskURL(cfg.DatabaseURL), cfg.Environment)
	st, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Fatal database error: %v", err)
	}
	defer st.Close()

	log.Println("Running startup database migrations...")
	if err := st.Migrate(ctx); err != nil {
		log.Fatalf("Fatal database migration error: %v", err)
	}
	log.Println("Database migrations applied successfully.")

	log.Printf("Connecting to S3 storage at %s...", cfg.S3Endpoint)
	stor, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("Fatal storage error: %v", err)
	}

	if err := stor.EnsureBuckets(ctx); err != nil {
		log.Printf("Warning: ensuring storage buckets: %v", err)
	} else {
		log.Printf("S3 storage buckets verified: %s, %s", cfg.S3BucketMedia, cfg.S3BucketSource)
	}

	handler := httpapi.New(cfg, st, stor)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("API server listening on port %s", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	sig := <-shutdownChan
	log.Printf("Received signal %v, initiating graceful shutdown...", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during graceful server shutdown: %v", err)
	}

	log.Println("Qmanga API backend stopped cleanly.")
}

func maskURL(raw string) string {
	if len(raw) < 10 {
		return "***"
	}
	return raw[:8] + "...masked..."
}
