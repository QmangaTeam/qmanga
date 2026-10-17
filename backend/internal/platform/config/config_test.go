package config

import (
	"strings"
	"testing"
)

func TestOriginMustBeAnHTTPOrigin(testingTarget *testing.T) {
	for _, origin := range []string{"ftp://example.com", "http://example.com/path", "http://name:password@example.com", "http://example.com?query=1", "http://example.com#fragment"} {
		testingTarget.Run(origin, func(testingTarget *testing.T) {
			testingTarget.Setenv("APP_ENV", "test")
			testingTarget.Setenv("APP_ORIGIN", origin)
			_, loadError := Load()
			if loadError == nil || !strings.Contains(loadError.Error(), "APP_ORIGIN") {
				testingTarget.Fatalf("invalid origin accepted: %s", origin)
			}
		})
	}
}

func TestLoad_RequiresExplicitValidEnvironment(testingTarget *testing.T) {
	testingTarget.Setenv("APP_ENV", "staging")
	_, err := Load()
	if err == nil {
		testingTarget.Fatal("expected error for invalid APP_ENV, got nil")
	}
	if !strings.Contains(err.Error(), "APP_ENV must be explicitly set") {
		testingTarget.Fatalf("expected error mentioning APP_ENV, got %v", err)
	}
}

func TestLoad_RequiresCredentialsInDevelopment(testingTarget *testing.T) {
	testingTarget.Setenv("APP_ENV", "development")
	testingTarget.Setenv("PORT", "8080")
	testingTarget.Setenv("DATABASE_URL", "")

	_, err := Load()
	if err == nil {
		testingTarget.Fatal("expected error for missing DATABASE_URL in development, got nil")
	}
	if !strings.Contains(err.Error(), "missing required DATABASE_URL") {
		testingTarget.Fatalf("expected error mentioning missing DATABASE_URL, got %v", err)
	}
}

func TestLoad_RejectsInvalidPort(testingTarget *testing.T) {
	testingTarget.Setenv("APP_ENV", "development")
	testingTarget.Setenv("PORT", "not-a-port")
	testingTarget.Setenv("DATABASE_URL", "postgres://user:password@localhost:5432/qmanga?sslmode=disable")
	testingTarget.Setenv("S3_ENDPOINT", "localhost:9000")
	testingTarget.Setenv("S3_ACCESS_KEY", "access-key")
	testingTarget.Setenv("S3_SECRET_KEY", "secret-key")
	testingTarget.Setenv("SESSION_SECRET", "12345678901234567890123456789012")
	testingTarget.Setenv("CSRF_SECRET", "09876543210987654321098765432109")

	_, err := Load()
	if err == nil {
		testingTarget.Fatal("expected error for invalid PORT, got nil")
	}
	if !strings.Contains(err.Error(), "PORT must be a valid integer") {
		testingTarget.Fatalf("expected error mentioning PORT, got %v", err)
	}
}

func TestLoad_RejectsInvalidS3UseSSL(testingTarget *testing.T) {
	testingTarget.Setenv("APP_ENV", "development")
	testingTarget.Setenv("PORT", "8080")
	testingTarget.Setenv("S3_USE_SSL", "not-a-boolean")
	testingTarget.Setenv("DATABASE_URL", "postgres://user:password@localhost:5432/qmanga?sslmode=disable")
	testingTarget.Setenv("S3_ENDPOINT", "localhost:9000")
	testingTarget.Setenv("S3_ACCESS_KEY", "access-key")
	testingTarget.Setenv("S3_SECRET_KEY", "secret-key")
	testingTarget.Setenv("SESSION_SECRET", "12345678901234567890123456789012")
	testingTarget.Setenv("CSRF_SECRET", "09876543210987654321098765432109")

	_, err := Load()
	if err == nil {
		testingTarget.Fatal("expected error for invalid S3_USE_SSL, got nil")
	}
	if !strings.Contains(err.Error(), "S3_USE_SSL must be a valid boolean") {
		testingTarget.Fatalf("expected error mentioning S3_USE_SSL, got %v", err)
	}
}

func TestLoad_RejectsInsecureOriginInProduction(testingTarget *testing.T) {
	testingTarget.Setenv("APP_ENV", "production")
	testingTarget.Setenv("PORT", "8080")
	testingTarget.Setenv("DATABASE_URL", "postgres://user:password@localhost:5432/qmanga?sslmode=disable")
	testingTarget.Setenv("S3_ENDPOINT", "localhost:9000")
	testingTarget.Setenv("S3_ACCESS_KEY", "access-key")
	testingTarget.Setenv("S3_SECRET_KEY", "secret-key")
	testingTarget.Setenv("SESSION_SECRET", "12345678901234567890123456789012")
	testingTarget.Setenv("CSRF_SECRET", "09876543210987654321098765432109")
	testingTarget.Setenv("APP_ORIGIN", "http://insecure-domain.kz")

	_, err := Load()
	if err == nil {
		testingTarget.Fatal("expected error for HTTP origin in production, got nil")
	}
	if !strings.Contains(err.Error(), "must use HTTPS in production") {
		testingTarget.Fatalf("expected error mentioning HTTPS, got %v", err)
	}
}

func TestLoad_RejectsPlaceholders(testingTarget *testing.T) {
	testingTarget.Setenv("APP_ENV", "development")
	testingTarget.Setenv("PORT", "8080")
	testingTarget.Setenv("DATABASE_URL", "postgres://qmanga:replace_with_password@localhost:5432/qmanga")
	testingTarget.Setenv("S3_ENDPOINT", "localhost:9000")
	testingTarget.Setenv("S3_ACCESS_KEY", "access-key")
	testingTarget.Setenv("S3_SECRET_KEY", "secret-key")
	testingTarget.Setenv("SESSION_SECRET", "12345678901234567890123456789012")
	testingTarget.Setenv("CSRF_SECRET", "09876543210987654321098765432109")

	_, err := Load()
	if err == nil {
		testingTarget.Fatal("expected error for placeholder database password, got nil")
	}
	if !strings.Contains(err.Error(), "placeholder password") {
		testingTarget.Fatalf("expected error mentioning placeholder, got %v", err)
	}
}

func TestLoad_ValidConfigurationLoadsCorrectly(testingTarget *testing.T) {
	testingTarget.Setenv("SMTP_ADDRESS", "smtp.example.com:587")
	testingTarget.Setenv("SMTP_FROM", "no-reply@example.com")
	testingTarget.Setenv("APP_ENV", "production")
	testingTarget.Setenv("PORT", "8080")
	testingTarget.Setenv("DATABASE_URL", "postgres://qmanga:supersecretpass@postgres:5432/qmanga?sslmode=disable")
	testingTarget.Setenv("S3_ENDPOINT", "minio:9000")
	testingTarget.Setenv("S3_ACCESS_KEY", "valid-access-key")
	testingTarget.Setenv("S3_SECRET_KEY", "valid-secret-key-123")
	testingTarget.Setenv("S3_BUCKET_MEDIA", "custom-media")
	testingTarget.Setenv("S3_BUCKET_SOURCE", "custom-source")
	testingTarget.Setenv("S3_USE_SSL", "true")
	testingTarget.Setenv("SESSION_SECRET", "12345678901234567890123456789012")
	testingTarget.Setenv("CSRF_SECRET", "09876543210987654321098765432109")
	testingTarget.Setenv("APP_ORIGIN", "https://qmanga.kz")

	configInstance, err := Load()
	if err != nil {
		testingTarget.Fatalf("unexpected error: %v", err)
	}
	if configInstance.Environment != "production" {
		testingTarget.Errorf("expected production, got %s", configInstance.Environment)
	}
	if !configInstance.S3UseSSL {
		testingTarget.Errorf("expected S3UseSSL true, got false")
	}
	if configInstance.S3BucketMedia != "custom-media" {
		testingTarget.Errorf("expected custom-media, got %s", configInstance.S3BucketMedia)
	}
}
