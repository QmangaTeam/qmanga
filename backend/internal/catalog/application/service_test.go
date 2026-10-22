package application

import (
	"context"
	"errors"
	"fmt"
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
	if _, failure := service.Chapters(context.Background(), "title", 0, 0, ChapterFilter{Query: "  ӘЛЕМ ", Sort: "desc", Team: "dala"}); failure != nil {
		testingTarget.Fatal(failure)
	}
	if repository.filter.Query != "әлем" || repository.filter.Sort != "desc" || repository.filter.Limit != 50 || repository.filter.Team != "dala" {
		testingTarget.Fatalf("incorrect chapter search defaults: %+v", repository.filter)
	}
	for _, filter := range []ChapterFilter{{Query: strings.Repeat("ә", 81)}, {Sort: "raw sql"}, {Team: "bad/sl ug"}} {
		if _, failure := service.Chapters(context.Background(), "title", 0, 50, filter); !errors.Is(failure, domain.ErrInvalid) {
			testingTarget.Fatalf("invalid chapter query accepted: %+v", filter)
		}
	}
}

func (repository *filterRepository) List(ctx context.Context, filter Filter) (Page, error) {
	repository.filter = filter
	return Page{Items: []Summary{}}, nil
}

func thirteenTags() []string {
	tags := make([]string, 13)
	for index := range tags {
		tags[index] = fmt.Sprintf("tag-%d", index)
	}
	return tags
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
	for _, filter := range []Filter{{Query: strings.Repeat("ә", 81)}, {Sort: "raw-sql"}, {Kinds: []string{"novel"}}, {OriginalStatuses: []string{"secret"}}, {TranslationStatuses: []string{"secret"}}, {Genres: []string{"not a slug"}}, {Genres: thirteenTags()}} {
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

func TestCatalogueFacetFilters(testingTarget *testing.T) {
	repository := &filterRepository{}
	service := New(repository, func() string { return "new-id" })
	valid := Filter{
		Kinds:               []string{"manga", "manhwa", "manga"},
		AgeRatings:          []int{0, 12, 0},
		OriginalStatuses:    []string{"ongoing", "completed", "ongoing"},
		TranslationStatuses: []string{"paused"},
		Formats:             []string{"Сборник", "Вебтун", "Сборник"},
		Sources:             []string{"Манга"},
		Countries:           []string{"Япония", "Корея", "Япония"},
		ChaptersMin:         20,
		ChaptersMax:         200,
		YearMin:             2000,
		YearMax:             2026,
	}
	if _, failure := service.List(context.Background(), valid); failure != nil {
		testingTarget.Fatal(failure)
	}
	if len(repository.filter.Kinds) != 2 || len(repository.filter.AgeRatings) != 2 || len(repository.filter.OriginalStatuses) != 2 || len(repository.filter.Formats) != 2 || len(repository.filter.Countries) != 2 {
		testingTarget.Fatalf("repeated facet values not folded: %+v", repository.filter)
	}
	if repository.filter.ChaptersMin != 20 || repository.filter.ChaptersMax != 200 || repository.filter.YearMin != 2000 || repository.filter.YearMax != 2026 {
		testingTarget.Fatalf("range filter lost: %+v", repository.filter)
	}
	for _, filter := range []Filter{
		{Kinds: []string{"novel"}},
		{AgeRatings: []int{14}},
		{OriginalStatuses: []string{"secret"}},
		{TranslationStatuses: []string{"secret"}},
		{Formats: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}},
		{Formats: []string{strings.Repeat("ф", 161)}},
		{ChaptersMin: -1},
		{ChaptersMin: 10, ChaptersMax: 5},
		{YearMin: 500},
		{YearMin: 2020, YearMax: 2010},
		{YearMax: 2300},
	} {
		if _, failure := service.List(context.Background(), filter); !errors.Is(failure, domain.ErrInvalid) {
			testingTarget.Fatalf("invalid facet filter accepted: %+v", filter)
		}
	}
	if _, failure := service.List(context.Background(), Filter{Genres: []string{"horror", "", "  "}, Kinds: []string{"manga", ""}}); failure != nil {
		testingTarget.Fatal(failure)
	}
	if len(repository.filter.Genres) != 1 || len(repository.filter.Kinds) != 1 {
		testingTarget.Fatalf("empty values not dropped: %+v", repository.filter)
	}
}
