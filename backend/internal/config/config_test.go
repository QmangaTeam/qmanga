package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoad_ProductionRejectsMissingSecrets(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_ENV", "production")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load() to fail in production with missing secrets, got nil error")
	}

	expectedSubstrings := []string{
		"DATABASE_URL",
		"S3_SECRET_KEY",
		"SESSION_SECRET",
		"CSRF_SECRET",
	}

	errStr := err.Error()
	for _, substr := range expectedSubstrings {
		if !strings.Contains(errStr, substr) {
			t.Errorf("expected error to mention missing %s, got: %s", substr, errStr)
		}
	}
}

func TestLoad_ProductionRejectsShortSecrets(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_ENV", "production")
	os.Setenv("DATABASE_URL", "postgres://user:strong_password@postgres:5432/qmanga?sslmode=disable")
	os.Setenv("S3_ENDPOINT", "minio:9000")
	os.Setenv("S3_ACCESS_KEY", "prod-access-key")
	os.Setenv("S3_SECRET_KEY", "prod-secret-key-super-secure")
	os.Setenv("SESSION_SECRET", "short")
	os.Setenv("CSRF_SECRET", "short")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load() to fail with short secrets, got nil")
	}
	if !strings.Contains(err.Error(), "must be at least 32 characters") {
		t.Errorf("expected error about secret length, got: %v", err)
	}
}

func TestLoad_ValidConfiguration(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_ENV", "production")
	os.Setenv("PORT", "8080")
	os.Setenv("DATABASE_URL", "postgres://user:strong_password@postgres:5432/qmanga?sslmode=disable")
	os.Setenv("S3_ENDPOINT", "minio:9000")
	os.Setenv("S3_ACCESS_KEY", "prod-access-key")
	os.Setenv("S3_SECRET_KEY", "prod-secret-key-super-secure")
	os.Setenv("S3_BUCKET_MEDIA", "qmanga-media")
	os.Setenv("S3_BUCKET_SOURCE", "qmanga-source")
	os.Setenv("SESSION_SECRET", "12345678901234567890123456789012")
	os.Setenv("CSRF_SECRET", "09876543210987654321098765432109")
	os.Setenv("APP_ORIGIN", "https://qmanga.kz")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error loading valid config: %v", err)
	}

	if cfg.Environment != "production" {
		t.Errorf("expected Environment 'production', got %q", cfg.Environment)
	}
	if cfg.HTTPPort != "8080" {
		t.Errorf("expected HTTPPort '8080', got %q", cfg.HTTPPort)
	}
	if cfg.S3BucketMedia != "qmanga-media" {
		t.Errorf("expected S3BucketMedia 'qmanga-media', got %q", cfg.S3BucketMedia)
	}
}
