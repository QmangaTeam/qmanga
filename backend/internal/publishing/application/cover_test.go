package application

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"qmanga/internal/publishing/domain"
)

type coverRepository struct {
	Repository
	allow bool
	saved CoverAsset
}

func (repository *coverRepository) CanUploadCover(context.Context, string, string) error {
	if !repository.allow {
		return domain.ErrForbidden
	}
	return nil
}
func (repository *coverRepository) SaveCover(_ context.Context, _ string, _ string, cover CoverAsset) error {
	repository.saved = cover
	return nil
}

type coverStorage struct {
	Storage
	stored string
	fail   bool
}

func (storage *coverStorage) PutDerivative(_ context.Context, key string, content io.Reader, size int64, mime string) error {
	if storage.fail {
		return errors.New("storage unavailable")
	}
	body, _ := io.ReadAll(content)
	if int64(len(body)) != size || mime != "image/webp" {
		return errors.New("bad cover body")
	}
	storage.stored = key
	return nil
}

type coverProcessor struct {
	called bool
	closed bool
}

func (processor *coverProcessor) ProcessCover(context.Context, io.Reader) (ProcessedCover, error) {
	processor.called = true
	return ProcessedCover{CoverAsset: CoverAsset{Mime: "image/webp", Checksum: "checksum", Width: 10, Height: 20, Bytes: 5}, Content: bytes.NewReader([]byte("cover")), Close: func() { processor.closed = true }}, nil
}

func TestUploadCoverChecksAccessBeforeProcessingAndSavesOnlyStoredAsset(t *testing.T) {
	repository := &coverRepository{}
	storage := &coverStorage{}
	processor := &coverProcessor{}
	service := New(repository, storage, processor, func() string { return "cover-id" })
	if _, err := service.UploadCover(context.Background(), "other", "project", bytes.NewReader([]byte("source"))); !errors.Is(err, domain.ErrForbidden) || processor.called {
		t.Fatal("unauthorized cover was processed", err)
	}
	repository.allow = true
	storage.fail = true
	if _, err := service.UploadCover(context.Background(), "owner", "project", bytes.NewReader([]byte("source"))); err == nil || repository.saved.ID != "" || !processor.closed {
		t.Fatal("failed storage changed visible cover", err)
	}
	processor.closed = false
	storage.fail = false
	id, err := service.UploadCover(context.Background(), "owner", "project", bytes.NewReader([]byte("source")))
	if err != nil || id != "cover-id" || storage.stored != "cover-id" || repository.saved.ID != "cover-id" || !processor.closed {
		t.Fatal("independent cover not saved", id, err)
	}
}
