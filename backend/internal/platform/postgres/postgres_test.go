package postgres

import (
	"context"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMigrationLockConcurrencyAndCancellation(testingTarget *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" { testingTarget.Skip("TEST_DATABASE_URL required for PostgreSQL integration") }
	testContext, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	base, openError := New(testContext, databaseURL)
	if openError != nil { testingTarget.Fatal(openError) }
	defer base.Close()
	schemaName := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, createError := base.pool.Exec(testContext, "CREATE SCHEMA "+schemaName); createError != nil { testingTarget.Fatal(createError) }
	defer base.pool.Exec(context.Background(), "DROP SCHEMA "+schemaName+" CASCADE")
	parsed, _ := url.Parse(databaseURL)
	parameters := parsed.Query()
	parameters.Set("search_path", schemaName)
	parsed.RawQuery = parameters.Encode()
	isolated, isolatedError := New(testContext, parsed.String())
	if isolatedError != nil { testingTarget.Fatal(isolatedError) }
	defer isolated.Close()
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() { defer workers.Done(); if migrateError := isolated.Migrate(testContext); migrateError != nil { testingTarget.Error(migrateError) } }()
	}
	workers.Wait()
	var applied int
	if queryError := isolated.pool.QueryRow(testContext, "SELECT count(*) FROM schema_migrations").Scan(&applied); queryError != nil || applied == 0 { testingTarget.Fatalf("migrations not applied: %d %v", applied, queryError) }
	blocker, acquireError := base.pool.Acquire(testContext)
	if acquireError != nil { testingTarget.Fatal(acquireError) }
	defer blocker.Release()
	if _, lockError := blocker.Exec(testContext, "SELECT pg_advisory_lock(847291)"); lockError != nil { testingTarget.Fatal(lockError) }
	defer blocker.Exec(context.Background(), "SELECT pg_advisory_unlock(847291)")
	cancelContext, cancelWaiting := context.WithTimeout(testContext, 100*time.Millisecond)
	defer cancelWaiting()
	if migrateError := isolated.Migrate(cancelContext); migrateError == nil { testingTarget.Fatal("migration proceeded without the lock") }
}

func TestClient_NilClientPingReturnsError(testingTarget *testing.T) {
	var nilClient *Client
	pingError := nilClient.Ping(context.Background())
	if pingError == nil {
		testingTarget.Fatal("expected error pinging nil postgres client, got nil")
	}
	if !strings.Contains(pingError.Error(), "not initialized") {
		testingTarget.Fatalf("expected error mentioning not initialized, got %v", pingError)
	}
}

func TestClient_NilPoolPingReturnsError(testingTarget *testing.T) {
	clientInstance := &Client{pool: nil}
	pingError := clientInstance.Ping(context.Background())
	if pingError == nil {
		testingTarget.Fatal("expected error pinging client with nil pool, got nil")
	}
}

func TestClient_NilClientMigrateReturnsError(testingTarget *testing.T) {
	var nilClient *Client
	migrateError := nilClient.Migrate(context.Background())
	if migrateError == nil {
		testingTarget.Fatal("expected error migrating nil postgres client, got nil")
	}
}

func TestClient_CancelledContextMigrationFailsClosed(testingTarget *testing.T) {
	clientInstance := &Client{pool: nil}
	cancelledContext, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()

	migrateError := clientInstance.Migrate(cancelledContext)
	if migrateError == nil {
		testingTarget.Fatal("expected error migrating with cancelled context, got nil")
	}
}
