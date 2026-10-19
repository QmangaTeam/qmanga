package domain

import (
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

type ProjectRequestInput struct {
	TeamID, TitleID, Name, OriginalName, Slug, Kind, Description, SourceURL, Note string
}
type ProjectRequest struct {
	ID string
	ProjectRequestInput
	TeamName, TitleName, Status, Resolution, ProjectID, CoverID string
	CreatedAt, UpdatedAt                                        time.Time
}

func NewProjectRequest(id string, input ProjectRequestInput) (ProjectRequest, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Note = strings.TrimSpace(input.Note)
	input.SourceURL = strings.TrimSpace(input.SourceURL)
	if id == "" || input.TeamID == "" || len(input.TeamID) > 36 || len(input.TitleID) > 36 || utf8.RuneCountInString(input.Note) < 8 || utf8.RuneCountInString(input.Note) > 2000 || len(input.SourceURL) > 2000 {
		return ProjectRequest{}, ErrInvalid
	}
	if input.SourceURL != "" {
		parsed, failure := url.Parse(input.SourceURL)
		if failure != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.User != nil {
			return ProjectRequest{}, ErrInvalid
		}
	}
	if input.TitleID == "" {
		if input.Slug == "" {
			input.Slug = id
		}
		if _, failure := NewTitle(id, Metadata{Slug: input.Slug, Name: input.Name, OriginalName: input.OriginalName, Kind: input.Kind, Description: input.Description, OriginalStatus: "ongoing", TranslationStatus: "ongoing", Direction: "ltr"}); failure != nil {
			return ProjectRequest{}, failure
		}
	} else {
		input.Name = ""
		input.OriginalName = ""
		input.Slug = ""
		input.Kind = "manga"
		input.Description = ""
	}
	return ProjectRequest{ID: id, ProjectRequestInput: input, Status: "pending"}, nil
}
func ValidRequestDecision(decision, reason string) bool {
	return utf8.RuneCountInString(reason) <= 2000 && (decision == "approve" || decision == "reject" && utf8.RuneCountInString(strings.TrimSpace(reason)) >= 8)
}
