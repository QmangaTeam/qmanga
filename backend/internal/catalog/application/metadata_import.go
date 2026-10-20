package application

import (
	"context"
	"errors"
	"io"
	"qmanga/internal/catalog/domain"
	"strings"
	"unicode/utf8"
)

var ErrSourceUnavailable = errors.New("metadata source unavailable")
var ErrSourceLimited = errors.New("metadata source rate limited")

type ExternalTag struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
}
type ImportTitle struct {
	Source         string            `json:"source"`
	ID             string            `json:"id"`
	URL            string            `json:"url"`
	Name           string            `json:"name"`
	OriginalName   string            `json:"original_name"`
	Aliases        []string          `json:"aliases"`
	Authors        []string          `json:"authors"`
	Description    string            `json:"description"`
	Kind           string            `json:"kind"`
	OriginalStatus string            `json:"original_status"`
	Year           int               `json:"year"`
	Country        string            `json:"country"`
	CoverURL       string            `json:"cover_url"`
	Tags           []ExternalTag     `json:"tags"`
	GenreIDs       []string          `json:"genre_ids"`
	CategoryIDs    []string          `json:"category_ids"`
	Candidates     []ImportCandidate `json:"candidates"`
}
type ImportCandidate struct {
	Source  string   `json:"source"`
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Status  string   `json:"status"`
	Matches []string `json:"matches"`
}
type MetadataSource interface {
	Search(context.Context, string) ([]ImportTitle, error)
	Detail(context.Context, string) (ImportTitle, error)
}
type ImportRepository interface {
	AuthorizeImport(context.Context, string, string) error
	MatchImportTags(context.Context, string, []ExternalTag) ([]string, []string, []ImportCandidate, error)
	RecordImport(context.Context, string, string, ImportTitle) error
	ImportCandidates(context.Context, string) ([]ImportCandidate, error)
	ResolveImportCandidate(context.Context, string, string, string, string, string) error
}
type MetadataImports struct {
	repository ImportRepository
	sources    map[string]MetadataSource
	processor  ChangeImageProcessor
}

func NewMetadataImports(repository ImportRepository, sources map[string]MetadataSource, processor ...ChangeImageProcessor) *MetadataImports {
	s := &MetadataImports{repository: repository, sources: sources}
	if len(processor) > 0 {
		s.processor = processor[0]
	}
	return s
}
func (s *MetadataImports) Search(ctx context.Context, actor, project, source, query string) ([]ImportTitle, error) {
	if err := s.repository.AuthorizeImport(ctx, actor, project); err != nil {
		return nil, err
	}
	query = strings.TrimSpace(query)
	if utf8.RuneCountInString(query) < 2 || utf8.RuneCountInString(query) > 160 {
		return nil, domain.ErrInvalid
	}
	adapter, ok := s.sources[source]
	if !ok {
		return nil, domain.ErrInvalid
	}
	return adapter.Search(ctx, query)
}
func (s *MetadataImports) Select(ctx context.Context, actor, project, source, id string) (ImportTitle, error) {
	var item ImportTitle
	if err := s.repository.AuthorizeImport(ctx, actor, project); err != nil {
		return item, err
	}
	adapter, ok := s.sources[source]
	if !ok || len(id) > 20 || id == "" || strings.IndexFunc(id, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
		return item, domain.ErrInvalid
	}
	item, err := adapter.Detail(ctx, id)
	if err != nil {
		return item, err
	}
	item.GenreIDs, item.CategoryIDs, item.Candidates, err = s.repository.MatchImportTags(ctx, source, item.Tags)
	if err != nil {
		return item, err
	}
	err = s.repository.RecordImport(ctx, actor, project, item)
	return item, err
}
func (s *MetadataImports) Candidates(ctx context.Context, actor string) ([]ImportCandidate, error) {
	return s.repository.ImportCandidates(ctx, actor)
}
func (s *MetadataImports) Resolve(ctx context.Context, actor, source, id, kind, target string) error {
	if source != "shikimori" && source != "anilist" || id == "" || len(id) > 160 || kind != "genres" && kind != "categories" || !domain.ValidReferences([]string{target}, 1) {
		return domain.ErrInvalid
	}
	return s.repository.ResolveImportCandidate(ctx, actor, source, id, kind, target)
}

// Cover creates a private, reprocessed draft file. Existing artwork submission/approval remains authoritative.
func (s *MetadataImports) Cover(ctx context.Context, actor, project, source, id string) (ProcessedChangeImage, error) {
	var empty ProcessedChangeImage
	if err := s.repository.AuthorizeImport(ctx, actor, project); err != nil {
		return empty, err
	}
	adapter, ok := s.sources[source]
	if !ok || id == "" || len(id) > 20 || strings.IndexFunc(id, func(r rune) bool { return r < '0' || r > '9' }) >= 0 || s.processor == nil {
		return empty, domain.ErrInvalid
	}
	images, ok := adapter.(interface {
		OpenCover(context.Context, string) (io.ReadCloser, error)
	})
	if !ok {
		return empty, ErrSourceUnavailable
	}
	reader, err := images.OpenCover(ctx, id)
	if err != nil {
		return empty, err
	}
	defer reader.Close()
	return s.processor.ProcessChangeImage(ctx, reader, "cover")
}
