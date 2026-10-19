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

type chapterRepository struct {
	Repository
	filter ChapterFilter
}

func (repository *chapterRepository) Chapters(ctx context.Context, slug string, filter ChapterFilter) (ChapterPage, error) {
	repository.filter = filter
	return ChapterPage{}, nil
}

func TestChapterSearchValidation(testingTarget *testing.T) {
	repository := &chapterRepository{}
	service := New(repository, func() string { return "unused" })
	if _, failure := service.Chapters(context.Background(), "title", 0, 0, ChapterFilter{Query: "  ӘЛЕМ ", Sort: "desc"}); failure != nil {
		testingTarget.Fatal(failure)
	}
	if repository.filter.Query != "әлем" || repository.filter.Sort != "desc" || repository.filter.Limit != 50 {
		testingTarget.Fatalf("incorrect chapter search defaults: %+v", repository.filter)
	}
	for _, filter := range []ChapterFilter{{Query: strings.Repeat("ә", 81)}, {Sort: "raw sql"}} {
		if _, failure := service.Chapters(context.Background(), "title", 0, 50, filter); !errors.Is(failure, domain.ErrInvalid) {
			testingTarget.Fatalf("invalid chapter query accepted: %+v", filter)
		}
	}
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
	for _, filter := range []Filter{{Query: strings.Repeat("ә", 81)}, {Sort: "raw-sql"}, {Kind: "novel"}, {OriginalStatus: "secret"}, {TranslationStatus: "secret"}, {Genres: []string{"not a slug"}}, {Genres: make([]string, 13)}} {
		if _, failure := service.List(context.Background(), filter); !errors.Is(failure, domain.ErrInvalid) {
			testingTarget.Fatalf("invalid filter accepted: %+v", filter)
		}
	}
	if _, failure := service.List(context.Background(), Filter{Genres: []string{"horror", "horror", "magic"}}); failure != nil {
		testingTarget.Fatal(failure)
	}
	if len(repository.filter.Genres) != 2 || repository.filter.Genres[0] != "horror" || repository.filter.Genres[1] != "magic" {
		testingTarget.Fatalf("repeated tag not folded: %+v", repository.filter.Genres)
	}
}
