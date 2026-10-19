package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// Proposals deliberately exclude catalogue URL, visibility and editorial feature switches.
type ProposedTitleData struct {
	Name              string      `json:"name"`
	OriginalName      string      `json:"original_name"`
	Aliases           []string    `json:"aliases"`
	Authors           []string    `json:"authors"`
	Kind              string      `json:"kind"`
	OriginalStatus    string      `json:"original_status"`
	TranslationStatus string      `json:"translation_status"`
	Description       string      `json:"description"`
	Direction         string      `json:"direction"`
	AgeRating         int         `json:"age_rating"`
	Publication       Publication `json:"publication"`
	GenreIDs          []string    `json:"genre_ids"`
	CategoryIDs       []string    `json:"category_ids"`
}

func (data ProposedTitleData) Apply(id string, base Metadata) (Title, error) {
	if !validChangeIDs(data.GenreIDs) || !validChangeIDs(data.CategoryIDs) {
		return Title{}, ErrInvalid
	}
	base.Name = data.Name
	base.OriginalName = data.OriginalName
	base.Aliases = data.Aliases
	base.Authors = data.Authors
	base.Kind = data.Kind
	base.OriginalStatus = data.OriginalStatus
	base.TranslationStatus = data.TranslationStatus
	base.Description = data.Description
	base.Direction = data.Direction
	base.AgeRating = &data.AgeRating
	base.Publication = data.Publication
	base.GenreIDs = data.GenreIDs
	return NewTitle(id, base)
}

var changeIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,36}$`)

func validChangeIDs(ids []string) bool {
	if len(ids) > 12 {
		return false
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if !changeIDPattern.MatchString(id) || seen[id] {
			return false
		}
		seen[id] = true
	}
	return true
}
func nonnil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
func (data ProposedTitleData) canonical() ProposedTitleData {
	data.Aliases = nonnil(data.Aliases)
	data.Authors = nonnil(data.Authors)
	data.GenreIDs = append([]string{}, data.GenreIDs...)
	sort.Strings(data.GenreIDs)
	data.CategoryIDs = append([]string{}, data.CategoryIDs...)
	sort.Strings(data.CategoryIDs)
	data.Publication.Publishers = nonnil(data.Publication.Publishers)
	data.Publication.Formats = nonnil(data.Publication.Formats)
	if data.Publication.Links == nil {
		data.Publication.Links = []ExternalLink{}
	}
	return data
}

type ChangeSnapshot struct {
	Data     ProposedTitleData `json:"data"`
	CoverID  string            `json:"cover_id"`
	BannerID string            `json:"banner_id"`
}

func (snapshot ChangeSnapshot) Version() string {
	snapshot.Data = snapshot.Data.canonical()
	encoded, _ := json.Marshal(snapshot)
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:])
}

type TitleChangeInput struct {
	BaseVersion  string            `json:"base_version"`
	Data         ProposedTitleData `json:"data"`
	Reason       string            `json:"reason"`
	Links        []string          `json:"links"`
	CoverAction  string            `json:"cover_action"`
	BannerAction string            `json:"banner_action"`
	CoverPageID  string            `json:"cover_page_id"`
}

func (input *TitleChangeInput) Validate(evidenceImages int, coverFile, bannerFile bool) error {
	input.Reason = strings.TrimSpace(input.Reason)
	version, err := hex.DecodeString(input.BaseVersion)
	if err != nil || len(version) != 32 || utf8.RuneCountInString(input.Reason) < 3 || utf8.RuneCountInString(input.Reason) > 2000 || len(input.Links) > 10 || evidenceImages < 0 || evidenceImages > 5 || len(input.Links)+evidenceImages == 0 {
		return ErrInvalid
	}
	for index, link := range input.Links {
		link = strings.TrimSpace(link)
		parsed, err := url.Parse(link)
		if len(link) > 2000 || err != nil || parsed.Hostname() == "" || parsed.User != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
			return ErrInvalid
		}
		input.Links[index] = link
	}
	if input.CoverAction == "" {
		input.CoverAction = "keep"
	}
	if input.BannerAction == "" {
		input.BannerAction = "keep"
	}
	if !oneOf(input.CoverAction, "keep", "replace", "remove") || !oneOf(input.BannerAction, "keep", "replace", "remove") {
		return ErrInvalid
	}
	if input.CoverPageID != "" && !changeIDPattern.MatchString(input.CoverPageID) {
		return ErrInvalid
	}
	coverSource := coverFile || input.CoverPageID != ""
	if (coverFile && input.CoverPageID != "") || (input.CoverAction == "replace") != coverSource || (input.BannerAction == "replace") != bannerFile {
		return ErrInvalid
	}
	return nil
}
