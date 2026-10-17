package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"qmanga/internal/bootstrap"
)

func main() {
	log.Println("Starting Qmanga API backend...")

	appContext := context.Background()
	appInstance, initError := bootstrap.Initialize(appContext)
	if initError != nil {
		log.Fatalf("Fatal initialization error: %v", initError)
	}
	defer appInstance.Close()

	shutdownChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownChannel, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("API server listening on port %s", appInstance.Config.HTTPPort)
		serveError := appInstance.Server.ListenAndServe()
		if serveError != nil && !errors.Is(serveError, http.ErrServerClosed) {
			log.Fatalf("HTTP server failure: %v", serveError)
		}
	}()

	receivedSignal := <-shutdownChannel
	log.Printf("Received signal %v, initiating graceful shutdown...", receivedSignal)

	shutdownContext, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	shutdownError := appInstance.Server.Shutdown(shutdownContext)
	if shutdownError != nil {
		log.Printf("Error during graceful server shutdown: %v", shutdownError)
	}

	log.Println("Qmanga API backend stopped cleanly.")
}
