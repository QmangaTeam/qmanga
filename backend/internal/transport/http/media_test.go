package http

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qmanga/internal/publishing/application"
	"qmanga/internal/publishing/domain"
)

type mediaRepository struct {
	application.Repository
	revoked bool
}

func (repository *mediaRepository) Asset(ctx context.Context, actor, id string) (application.Asset, error) {
	if repository.revoked {
		return application.Asset{}, domain.ErrNotFound
	}
	return application.Asset{Variant: application.Variant{ID: id, ObjectKey: "private-storage-key", Mime: "image/webp", Bytes: 10, Checksum: strings.Repeat("a", 64)}, Public: true, ModifiedAt: time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)}, nil
}

type mediaStorage struct{ application.Storage }

func (storage mediaStorage) OpenDerivative(ctx context.Context, key string, offset, length int64) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("0123456789"[offset : offset+length])), nil
}
func TestMediaRangeAndRevocationBefore304(testingTarget *testing.T) {
	repository := &mediaRepository{}
	router := NewRouter(nil)
	NewPublishingHandler(application.New(repository, mediaStorage{}, nil, func() string { return "id" })).Register(router, NewIdentityHandler(nil, "http://localhost:8080", false))
	send := func(method, rangeValue, etag string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, "/api/v1/media/asset", nil)
		request.Header.Set("Range", rangeValue)
		request.Header.Set("If-None-Match", etag)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		return recorder
	}
	response := send("GET", "bytes=2-5", "")
	if response.Code != 206 || response.Body.String() != "2345" || response.Header().Get("Content-Range") != "bytes 2-5/10" {
		testingTarget.Fatal("range failed", response.Code, response.Body.String())
	}
	if response = send("HEAD", "", ""); response.Code != 200 || response.Body.Len() != 0 || response.Header().Get("Content-Length") != "10" {
		testingTarget.Fatal("HEAD failed")
	}
	if send("GET", "bytes=0-1,3-4", "").Code != 416 {
		testingTarget.Fatal("multiple ranges allowed")
	}
	if send("GET", "bytes=10-", "").Code != 416 {
		testingTarget.Fatal("invalid range allowed")
	}
	etag := "\"" + strings.Repeat("a", 64) + "\""
	if send("GET", "", etag).Code != 304 {
		testingTarget.Fatal("conditional GET failed")
	}
	repository.revoked = true
	if send("GET", "", etag).Code != 404 {
		testingTarget.Fatal("revoked media returned304")
	}
}
