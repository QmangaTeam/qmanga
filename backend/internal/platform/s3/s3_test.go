package s3

import (
	"context"
	"strings"
	"testing"
)

func TestClient_NilClientPingReturnsError(testingTarget *testing.T) {
	var nilClient *Client
	pingError := nilClient.Ping(context.Background())
	if pingError == nil {
		testingTarget.Fatal("expected error pinging nil s3 client, got nil")
	}
	if !strings.Contains(pingError.Error(), "not initialized") {
		testingTarget.Fatalf("expected error mentioning not initialized, got %v", pingError)
	}
}

func TestClient_NilMinioClientPingReturnsError(testingTarget *testing.T) {
	clientInstance := &Client{minioClient: nil}
	pingError := clientInstance.Ping(context.Background())
	if pingError == nil {
		testingTarget.Fatal("expected error pinging client with nil minio client, got nil")
	}
}

func TestClient_NilClientEnsureBucketsReturnsError(testingTarget *testing.T) {
	var nilClient *Client
	ensureError := nilClient.EnsureBuckets(context.Background())
	if ensureError == nil {
		testingTarget.Fatal("expected error ensuring buckets with nil s3 client, got nil")
	}
}

func TestNew_EmptyEndpointReturnsError(testingTarget *testing.T) {
	_, initError := New("", "key", "secret", "media", "source", false)
	if initError == nil {
		testingTarget.Fatal("expected error initializing client with empty endpoint, got nil")
	}
}
