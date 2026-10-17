package domain

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalid      = errors.New("invalid catalogue data")
	ErrForbidden    = errors.New("catalogue access denied")
	ErrNotFound     = errors.New("title not found")
	ErrConflict     = errors.New("catalogue conflict")
	ErrRequestLimit = errors.New("project request limit")
)

type Metadata struct {
	Listed                                                                                    bool
	Slug, Name, OriginalName, Kind, OriginalStatus, TranslationStatus, Description, Direction string
	Aliases, GenreIDs, Authors                                                                []string
	Featured                                                                                  bool
	Demo                                                                                      bool
}
type Title struct {
	ID string
	Metadata
}

var slugPattern = regexp.MustCompile("^[a-z0-9]+(-[a-z0-9]+)*$")

func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}
func NewTitle(id string, input Metadata) (Title, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.TrimSpace(input.Slug)
	if id == "" || !slugPattern.MatchString(input.Slug) || len(input.Slug) > 128 || utf8.RuneCountInString(input.Name) < 1 || utf8.RuneCountInString(input.Name) > 200 || utf8.RuneCountInString(input.Description) > 10000 || utf8.RuneCountInString(input.OriginalName) > 200 || len(input.Aliases) > 12 || len(input.GenreIDs) > 12 {
		return Title{}, ErrInvalid
	}
	if !oneOf(input.Kind, "manga", "manhwa", "manhua") || !oneOf(input.OriginalStatus, "ongoing", "completed", "hiatus", "cancelled") || !oneOf(input.TranslationStatus, "ongoing", "completed", "paused", "stopped") || !oneOf(input.Direction, "ltr", "rtl") {
		return Title{}, ErrInvalid
	}
	if len(input.Authors) > 12 {
		return Title{}, ErrInvalid
	}
	if input.Authors == nil {
		input.Authors = []string{}
	}
	for _, author := range input.Authors {
		if strings.TrimSpace(author) == "" || utf8.RuneCountInString(author) > 160 || strings.ContainsAny(author, "\r\n") {
			return Title{}, ErrInvalid
		}
	}
	if input.Aliases == nil {
		input.Aliases = []string{}
	}
	for _, alias := range input.Aliases {
		if utf8.RuneCountInString(alias) > 200 || strings.ContainsAny(alias, "\r\n") {
			return Title{}, ErrInvalid
		}
	}
	return Title{ID: id, Metadata: input}, nil
}
func ValidateChapter(label string, order int) error {
	if strings.TrimSpace(label) == "" || utf8.RuneCountInString(label) > 64 || strings.ContainsAny(label, "\r\n\t") || order < 1 || order > 1000000 {
		return ErrInvalid
	}
	return nil
}
func ValidGenre(name, slug string) bool {
	return utf8.RuneCountInString(strings.TrimSpace(name)) > 0 && utf8.RuneCountInString(name) <= 80 && slugPattern.MatchString(slug) && len(slug) <= 64
}
