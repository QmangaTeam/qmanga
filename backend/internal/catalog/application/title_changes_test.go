package application

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"qmanga/internal/catalog/domain"
	"strings"
	"testing"
)

type changeRepoFake struct {
	TitleChangeRepository
	current            TitleChangeContext
	deny, fail, exists bool
	created            bool
	assets             []TitleChangeAsset
}

func (repo *changeRepoFake) ChangeContext(context.Context, string, string) (TitleChangeContext, error) {
	if repo.deny {
		return TitleChangeContext{}, domain.ErrForbidden
	}
	return repo.current, nil
}
func (repo *changeRepoFake) CreateChange(_ context.Context, _ string, _ string, id string, _ domain.TitleChangeInput, assets []TitleChangeAsset) (TitleChange, error) {
	if repo.fail {
		return TitleChange{}, errors.New("unknown commit result")
	}
	repo.created = true
	repo.assets = assets
	return TitleChange{ID: id, Status: "pending"}, nil
}
func (repo *changeRepoFake) ChangeExists(context.Context, string) (bool, error) {
	return repo.exists, nil
}

type changeStoreFake struct {
	fail         bool
	put, deleted []string
}

func (store *changeStoreFake) PutDerivative(_ context.Context, key string, _ io.Reader, _ int64, _ string) error {
	store.put = append(store.put, key)
	if store.fail {
		return errors.New("unavailable")
	}
	return nil
}
func (store *changeStoreFake) DeleteDerivative(_ context.Context, key string) error {
	store.deleted = append(store.deleted, key)
	return nil
}
func (*changeStoreFake) OpenDerivative(context.Context, string, int64, int64) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("source")), nil
}

type changeProcessorFake struct{ calls, closed int }

func (processor *changeProcessorFake) ProcessChangeImage(context.Context, io.Reader, string) (ProcessedChangeImage, error) {
	processor.calls++
	return ProcessedChangeImage{Asset: TitleChangeAsset{Mime: "image/webp", Checksum: strings.Repeat("a", 64), Width: 10, Height: 20, Bytes: 4}, Content: bytes.NewReader([]byte("webp")), Close: func() { processor.closed++ }}, nil
}
func changeFixture() (*changeRepoFake, *changeStoreFake, *changeProcessorFake, *TitleChanges, domain.TitleChangeInput) {
	data := domain.ProposedTitleData{Name: "Before", Kind: "manga", OriginalStatus: "ongoing", TranslationStatus: "ongoing", Direction: "ltr"}
	repo := &changeRepoFake{current: TitleChangeContext{ProjectID: "project", TitleID: "title", Base: domain.Metadata{Slug: "title"}, Snapshot: domain.ChangeSnapshot{Data: data}}}
	store := &changeStoreFake{}
	processor := &changeProcessorFake{}
	counter := 0
	service := NewTitleChanges(repo, store, processor, func() string { counter++; return fmt.Sprintf("id-%d", counter) })
	input := domain.TitleChangeInput{BaseVersion: repo.current.Snapshot.Version(), Data: data, Reason: "Verified source", CoverAction: "keep", BannerAction: "keep"}
	input.Data.Name = "After"
	return repo, store, processor, service, input
}
func TestTitleChangeChecksAccessAndVersionBeforeProcessing(t *testing.T) {
	repo, store, processor, service, input := changeFixture()
	repo.deny = true
	if _, err := service.Submit(context.Background(), "other", "project", input, []ChangeImageSource{{Purpose: "evidence", Reader: strings.NewReader("source")}}); !errors.Is(err, domain.ErrForbidden) || processor.calls != 0 || len(store.put) != 0 {
		t.Fatal("unauthorized bytes processed", err)
	}
	repo.deny = false
	input.BaseVersion = strings.Repeat("0", 64)
	if _, err := service.Submit(context.Background(), "owner", "project", input, []ChangeImageSource{{Purpose: "evidence", Reader: strings.NewReader("source")}}); !errors.Is(err, domain.ErrConflict) || processor.calls != 0 {
		t.Fatal("stale bytes processed", err)
	}
}
func TestTitleChangeStoresOnlyProcessedAssetsAndKeepsUnknownCommitObjects(t *testing.T) {
	repo, store, processor, service, input := changeFixture()
	request, err := service.Submit(context.Background(), "owner", "project", input, []ChangeImageSource{{Purpose: "evidence", Reader: strings.NewReader("source")}})
	if err != nil || request.Status != "pending" || !repo.created || len(repo.assets) != 1 || repo.assets[0].Purpose != "evidence" || processor.closed != 1 || len(store.deleted) != 0 {
		t.Fatal(request, err, repo.assets)
	}
	repo, store, processor, service, input = changeFixture()
	store.fail = true
	if _, err = service.Submit(context.Background(), "owner", "project", input, []ChangeImageSource{{Purpose: "evidence", Reader: strings.NewReader("source")}}); err == nil || repo.created || len(store.deleted) != 1 || processor.closed != 1 {
		t.Fatal("failed upload created a request", err)
	}
	repo, store, _, service, input = changeFixture()
	repo.fail = true
	repo.exists = true
	if _, err = service.Submit(context.Background(), "owner", "project", input, []ChangeImageSource{{Purpose: "evidence", Reader: strings.NewReader("source")}}); err == nil || len(store.deleted) != 0 {
		t.Fatal("possibly committed image was deleted", err)
	}
}
