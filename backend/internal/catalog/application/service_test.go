package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"qmanga/internal/catalog/domain"
)

type filterRepository struct {
	Repository
	filter Filter
}

func (repository *filterRepository) List(ctx context.Context, filter Filter) (Page, error) {
	repository.filter = filter
	return Page{Items: []Summary{}}, nil
}

func TestCatalogueFilterValidation(testingTarget *testing.T) {
	repository := &filterRepository{}
	service := New(repository, func() string { return "new-id" })
	if _, failure := service.List(context.Background(), Filter{Query: "  ӘЛЕМ  "}); failure != nil {
		testingTarget.Fatal(failure)
	}
	if repository.filter.Limit != 24 || repository.filter.Query != "әлем" || repository.filter.Sort != "updated" {
		testingTarget.Fatalf("incorrect defaults: %+v", repository.filter)
	}
	for _, filter := range []Filter{{Query: strings.Repeat("ә", 81)}, {Sort: "raw-sql"}, {Kind: "novel"}, {OriginalStatus: "secret"}, {TranslationStatus: "secret"}} {
		if _, failure := service.List(context.Background(), filter); !errors.Is(failure, domain.ErrInvalid) {
			testingTarget.Fatalf("invalid filter accepted: %+v", filter)
		}
	}
}
