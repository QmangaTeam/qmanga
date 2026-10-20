package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	metadataSource "qmanga/internal/catalog/adapters/metadatasource"
	catalogPostgres "qmanga/internal/catalog/adapters/postgres"
	catalogApp "qmanga/internal/catalog/application"
	communityPostgres "qmanga/internal/community/adapters/postgres"
	communityApp "qmanga/internal/community/application"
	avatarimage "qmanga/internal/identity/adapters/avatarimage"
	identityCrypto "qmanga/internal/identity/adapters/crypto"
	googleoidc "qmanga/internal/identity/adapters/googleoidc"
	identityOCIEmail "qmanga/internal/identity/adapters/ociemail"
	identityPostgres "qmanga/internal/identity/adapters/postgres"
	identitySMTP "qmanga/internal/identity/adapters/smtp"
	identityApp "qmanga/internal/identity/application"
	metricsPostgres "qmanga/internal/metrics/adapters/postgres"
	metricsApplication "qmanga/internal/metrics/application"
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
	Avatars        *identityApp.AvatarService
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
	if appConfig.TrustedProxyHost != "" {
		if proxyError := transport.TrustProxyHost(httpRouter, appConfig.TrustedProxyHost); proxyError != nil {
			postgresClient.Close()
			return nil, fmt.Errorf("configuring trusted gateway: %w", proxyError)
		}
	}
	var mailer identityApp.Mailer
	if appConfig.MailTransport == "oci" {
		ociMailer, emailError := identityOCIEmail.New(appConfig.OCIEmailCompartmentID, appConfig.SMTPFrom)
		if emailError != nil {
			postgresClient.Close()
			return nil, fmt.Errorf("initializing email delivery: %w", emailError)
		}
		mailer = ociMailer
	} else {
		smtpHost, _, _ := net.SplitHostPort(appConfig.SMTPAddress)
		mailer = identitySMTP.New(identitySMTP.Config{Address: appConfig.SMTPAddress, Host: smtpHost, From: appConfig.SMTPFrom, Username: appConfig.SMTPUsername, Password: appConfig.SMTPPassword, RequireTLS: appConfig.Environment == "production"})
	}
	identityService := identityApp.New(identityPostgres.New(postgresClient.Pool()), identityCrypto.New(), mailer, appConfig.AppOrigin)
	identityHandler := transport.NewIdentityHandler(identityService, appConfig.AppOrigin, appConfig.Environment == "production")
	identityHandler.Register(httpRouter.Group("/api/v1"))
	teamService := teamApp.New(teamPostgres.New(postgresClient.Pool()), identityCrypto.New(), mailer, appConfig.AppOrigin)
	avatarService := identityApp.NewAvatars(identityPostgres.New(postgresClient.Pool()), storageClient, avatarimage.New(), teamService, uuid.NewString)
	var googleService *identityApp.GoogleService
	if appConfig.GoogleClientID != "" {
		googleService = identityApp.NewGoogle(identityPostgres.New(postgresClient.Pool()), identityCrypto.New(), googleoidc.New(appConfig.GoogleClientID, appConfig.GoogleClientSecret, appConfig.AppOrigin+"/api/v1/auth/google/callback"), identityService, googleoidc.NewAvatarImporter(avatarService))
	}
	transport.NewGoogleHandler(googleService, identityHandler).Register(httpRouter.Group("/api/v1"))
	transport.NewTeamHandler(teamService).Register(httpRouter.Group("/api/v1"), identityHandler)
	transport.NewAvatarHandler(avatarService, identityHandler).Register(httpRouter)
	transport.NewMetricsHandler(metricsApplication.New(metricsPostgres.New(postgresClient.Pool()))).Register(httpRouter.Group("/api/v1"), identityHandler)
	catalogService := catalogApp.New(catalogPostgres.New(postgresClient.Pool()), uuid.NewString)
	transport.NewMetadataImportsHandler(catalogApp.NewMetadataImports(catalogPostgres.New(postgresClient.Pool()), map[string]catalogApp.MetadataSource{"shikimori": metadataSource.New("shikimori"), "anilist": metadataSource.New("anilist")}, titleArtwork{imagemagick.New()})).Register(httpRouter.Group("/api/v1"), identityHandler)
	transport.NewCatalogHandler(catalogService).Register(httpRouter.Group("/api/v1"), identityHandler)
	transport.NewTitleChangesHandler(catalogApp.NewTitleChanges(catalogPostgres.New(postgresClient.Pool()), storageClient, titleArtwork{imagemagick.New()}, uuid.NewString)).Register(httpRouter.Group("/api/v1"), identityHandler)
	publicationRepository := publicationPostgres.New(postgresClient.Pool())
	publicationService := publicationApp.New(publicationRepository, storageClient, imagemagick.New(), uuid.NewString)
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
		Avatars:        avatarService,
		Notifications:  notifications,
	}

	return appInstance, nil
}

func (appInstance *Application) Close() {
	if appInstance != nil && appInstance.PostgresClient != nil {
		appInstance.PostgresClient.Close()
	}
}
