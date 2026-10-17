package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"

	healthHttp "qmanga/internal/health/adapters/http"
	healthApp "qmanga/internal/health/application"
	"qmanga/internal/platform/config"
	"qmanga/internal/platform/httpserver"
	"qmanga/internal/platform/postgres"
	"qmanga/internal/platform/s3"
)

type Application struct {
	Config         config.Config
	Server         *http.Server
	PostgresClient *postgres.Client
}

func Initialize(parentContext context.Context) (*Application, error) {
	appConfig, configError := config.Load()
	if configError != nil {
		return nil, fmt.Errorf("loading configuration: %w", configError)
	}

	startupContext, cancelStartup := context.WithTimeout(parentContext, 15*time.Second)
	defer cancelStartup()

	postgresClient, postgresError := postgres.New(startupContext, appConfig.DatabaseURL)
	if postgresError != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", postgresError)
	}

	migrationError := postgresClient.Migrate(startupContext)
	if migrationError != nil {
		postgresClient.Close()
		return nil, fmt.Errorf("running startup migrations: %w", migrationError)
	}

	storageClient, storageError := s3.New(
		appConfig.S3Endpoint,
		appConfig.S3AccessKey,
		appConfig.S3SecretKey,
		appConfig.S3BucketMedia,
		appConfig.S3BucketSource,
		appConfig.S3UseSSL,
	)
	if storageError != nil {
		postgresClient.Close()
		return nil, fmt.Errorf("initializing s3 storage: %w", storageError)
	}

	bucketError := storageClient.EnsureBuckets(startupContext)
	if bucketError != nil {
		postgresClient.Close()
		return nil, fmt.Errorf("ensuring required s3 storage buckets: %w", bucketError)
	}

	healthUseCase := healthApp.NewCheckHealthUseCase("0.1.0", postgresClient, storageClient)
	healthHandler := healthHttp.NewHandler(healthUseCase)
	mediaHandler := httpserver.NewMediaHandler()
	httpRouter := httpserver.NewRouter(healthHandler, mediaHandler)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%s", appConfig.HTTPPort),
		Handler:           httpRouter,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	appInstance := &Application{
		Config:         appConfig,
		Server:         httpServer,
		PostgresClient: postgresClient,
	}

	return appInstance, nil
}

func (appInstance *Application) Close() {
	if appInstance != nil && appInstance.PostgresClient != nil {
		appInstance.PostgresClient.Close()
	}
}
