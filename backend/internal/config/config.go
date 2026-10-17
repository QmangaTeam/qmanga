package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Environment    string
	HTTPPort       string
	DatabaseURL    string
	S3Endpoint     string
	S3AccessKey    string
	S3SecretKey    string
	S3BucketMedia  string
	S3BucketSource string
	S3UseSSL       bool
	SessionSecret  string
	CSRFSecret     string
	AppOrigin      string
}

func Load() (Config, error) {
	environment := getEnv("APP_ENV", "development")
	port := getEnv("PORT", "8080")
	dbURL := os.Getenv("DATABASE_URL")
	s3Endpoint := getEnv("S3_ENDPOINT", "minio:9000")
	s3AccessKey := os.Getenv("S3_ACCESS_KEY")
	s3SecretKey := os.Getenv("S3_SECRET_KEY")
	s3BucketMedia := getEnv("S3_BUCKET_MEDIA", "qmanga-media")
	s3BucketSource := getEnv("S3_BUCKET_SOURCE", "qmanga-source")
	s3UseSSLStr := getEnv("S3_USE_SSL", "false")
	sessionSecret := os.Getenv("SESSION_SECRET")
	csrfSecret := os.Getenv("CSRF_SECRET")
	appOrigin := getEnv("APP_ORIGIN", "http://localhost:8080")

	useSSL, _ := strconv.ParseBool(s3UseSSLStr)

	cfg := Config{
		Environment:    environment,
		HTTPPort:       port,
		DatabaseURL:    dbURL,
		S3Endpoint:     s3Endpoint,
		S3AccessKey:    s3AccessKey,
		S3SecretKey:    s3SecretKey,
		S3BucketMedia:  s3BucketMedia,
		S3BucketSource: s3BucketSource,
		S3UseSSL:       useSSL,
		SessionSecret:  sessionSecret,
		CSRFSecret:     csrfSecret,
		AppOrigin:      appOrigin,
	}

	var validationErrors []string

	if environment == "production" {
		if strings.TrimSpace(dbURL) == "" {
			validationErrors = append(validationErrors, "missing required DATABASE_URL")
		}
		if strings.TrimSpace(s3AccessKey) == "" {
			validationErrors = append(validationErrors, "missing required S3_ACCESS_KEY")
		}
		if strings.TrimSpace(s3SecretKey) == "" {
			validationErrors = append(validationErrors, "missing required S3_SECRET_KEY")
		}
		if len(sessionSecret) < 32 {
			validationErrors = append(validationErrors, "SESSION_SECRET must be at least 32 characters")
		}
		if len(csrfSecret) < 32 {
			validationErrors = append(validationErrors, "CSRF_SECRET must be at least 32 characters")
		}
	} else {
		if strings.TrimSpace(dbURL) == "" {
			dbURL = "postgres://qmanga:qmanga@localhost:5432/qmanga?sslmode=disable"
			cfg.DatabaseURL = dbURL
		}
	}

	if len(validationErrors) > 0 {
		return cfg, fmt.Errorf("configuration validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
