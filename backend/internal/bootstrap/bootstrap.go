package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	catalogPostgres "qmanga/internal/catalog/adapters/postgres"
	catalogApp "qmanga/internal/catalog/application"
	communityPostgres "qmanga/internal/community/adapters/postgres"
	communityApp "qmanga/internal/community/application"
	identityCrypto "qmanga/internal/identity/adapters/crypto"
	identityPostgres "qmanga/internal/identity/adapters/postgres"
	identitySMTP "qmanga/internal/identity/adapters/smtp"
	identityApp "qmanga/internal/identity/application"
	notificationPostgres "qmanga/internal/notifications/adapters/postgres"
	notificationApp "qmanga/internal/notifications/application"
	"qmanga/internal/platform/config"
	"qmanga/internal/platform/postgres"
	"qmanga/internal/platform/readiness"
	"qmanga/internal/platform/s3"
	"qmanga/internal/publishing/adapters/imagemagick"
	publicationPostgres "qmanga/internal/publishing/adapters/postgres"
	publicationApp "qmanga/internal/publishing/application"
	readingPostgres "qmanga/internal/reading/adapters/postgres"
	readingApp "qmanga/internal/reading/application"
	teamPostgres "qmanga/internal/teams/adapters/postgres"
	teamApp "qmanga/internal/teams/application"
	transport "qmanga/internal/transport/http"
)

type Application struct {
	Config         config.Config
	Server         *http.Server
	PostgresClient *postgres.Client
	Publishing     *publicationApp.Service
	Worker         *publicationApp.Worker
	Catalogue      *catalogApp.Service
	Teams          *teamApp.Service
	Notifications  *notificationApp.Service
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

	readinessCheck := readiness.NewCheckHealthUseCase("0.1.0", postgresClient, storageClient)
	httpRouter := transport.NewRouter(readinessCheck)
	smtpHost, _, _ := net.SplitHostPort(appConfig.SMTPAddress)
	mailer := identitySMTP.New(identitySMTP.Config{Address: appConfig.SMTPAddress, Host: smtpHost, From: appConfig.SMTPFrom, Username: appConfig.SMTPUsername, Password: appConfig.SMTPPassword, RequireTLS: appConfig.Environment == "production"})
	identityService := identityApp.New(identityPostgres.New(postgresClient.Pool()), identityCrypto.New(), mailer, appConfig.AppOrigin)
	identityHandler := transport.NewIdentityHandler(identityService, appConfig.AppOrigin, appConfig.Environment == "production")
	identityHandler.Register(httpRouter.Group("/api/v1"))
	teamService := teamApp.New(teamPostgres.New(postgresClient.Pool()), identityCrypto.New(), mailer, appConfig.AppOrigin)
	transport.NewTeamHandler(teamService).Register(httpRouter.Group("/api/v1"), identityHandler)
	catalogService := catalogApp.New(catalogPostgres.New(postgresClient.Pool()), uuid.NewString)
	transport.NewCatalogHandler(catalogService).Register(httpRouter.Group("/api/v1"), identityHandler)
	publicationRepository := publicationPostgres.New(postgresClient.Pool())
	publicationService := publicationApp.New(publicationRepository, storageClient, uuid.NewString)
	transport.NewPublishingHandler(publicationService).Register(httpRouter, identityHandler)
	transport.NewReadingHandler(readingApp.New(readingPostgres.New(postgresClient.Pool()))).Register(httpRouter.Group("/api/v1"), identityHandler)
	transport.NewCommunityHandler(communityApp.New(communityPostgres.New(postgresClient.Pool()), uuid.NewString)).Register(httpRouter.Group("/api/v1"), identityHandler)
	notifications := notificationApp.New(notificationPostgres.New(postgresClient.Pool()))
	transport.NewNotificationHandler(notifications).Register(httpRouter.Group("/api/v1"), identityHandler)

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
		Publishing:     publicationService,
		Worker:         publicationApp.NewWorker(publicationRepository, storageClient, imagemagick.New(), uuid.NewString),
		Catalogue:      catalogService,
		Teams:          teamService,
		Notifications:  notifications,
	}

	return appInstance, nil
}

func (appInstance *Application) Close() {
	if appInstance != nil && appInstance.PostgresClient != nil {
		appInstance.PostgresClient.Close()
	}
}
