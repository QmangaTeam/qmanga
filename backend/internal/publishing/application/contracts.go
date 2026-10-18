package application

import (
	"context"
	"io"
	"qmanga/internal/publishing/domain"
	"time"
)

type Page struct {
	OriginalName, ThumbnailID                                               string
	ID, RevisionID, Status, Format, SourceKey, Checksum, Failure, UploadKey string
	Position, Width, Height                                                 int
	Bytes                                                                   int64
}
type StudioRevision struct {
	CoverID, HumanReviewer                                                                       string
	HumanReviewedAt                                                                              *time.Time
	ReleaseStatus, PreparationMethod                                                             string
	HumanReviewed                                                                                bool
	Source, Credits, ConfirmedBy                                                                 string
	RightsConfirmed                                                                              bool
	ID, ReleaseID, ProjectID, ChapterID, TitleID, TitleName, ChapterLabel, TeamName, State, Note string
	Version                                                                                      int64
	Pages                                                                                        []Page
}
type ChapterChoice struct {
	ID, Label, Name string
	Order           int
}
type Upload struct {
	OriginalName                                         string
	RevisionID, PageID, Key, SourceKey, Checksum, Format string
	Position, Width, Height                              int
	Bytes                                                int64
}
type Reservation struct {
	PageID, SourceKey string
	Stored            bool
}
type Job struct {
	ID, ClaimID, PageID, RevisionID string
	Attempts                        int
	Page                            Page
}
type Variant struct {
	ID, PageID, Tier, ObjectKey, Mime, Checksum string
	Width, Height, Top                          int
	Bytes                                       int64
	Offset                                      float64
}
type ManifestPage struct {
	ID                      string
	Position, Width, Height int
	Variants                map[string][]Variant
}
type Manifest struct {
	HumanReviewer                                                                                                                        string
	HumanReviewedAt                                                                                                                      *time.Time
	PreparationMethod                                                                                                                    string
	HumanReviewed                                                                                                                        bool
	CoverID                                                                                                                              string
	Source, Credits                                                                                                                      string
	ReleaseID, RevisionID, TitleID, TitleSlug, TitleName, ChapterLabel, ProjectID, TeamID, TeamName, Kind, Direction, PreviousID, NextID string
	ChapterOrder                                                                                                                         int
	Pages                                                                                                                                []ManifestPage
	PublishedAt                                                                                                                          time.Time
}
type Asset struct {
	Variant
	Public     bool
	ModifiedAt time.Time
}
type Generated struct {
	Variant
	Content io.ReadSeeker
}
type Batch struct {
	Items []Generated
	Close func()
}
type CoverAsset struct {
	ID, ObjectKey, Mime, Checksum string
	Width, Height                 int
	Bytes                         int64
}
type ProcessedCover struct {
	CoverAsset
	Content io.Reader
	Close   func()
}
type CoverProcessor interface {
	ProcessCover(context.Context, io.Reader) (ProcessedCover, error)
}

type Repository interface {
	CreateDraft(context.Context, string, string, string, string) (StudioRevision, error)
	Revision(context.Context, string, string) (StudioRevision, error)
	ProjectChapters(context.Context, string, string, ...int) ([]ChapterChoice, error)
	ProjectRevisions(context.Context, string, string, ...int) ([]StudioRevision, error)
	Reserve(context.Context, string, Upload, string) (Reservation, error)
	Uploaded(context.Context, string, string, string) error
	UploadFailed(context.Context, string, string) error
	Reorder(context.Context, string, string, int64, []string) error
	DeletePage(context.Context, string, string, string, int64) error
	SaveAttribution(context.Context, string, string, int64, domain.Attribution) error
	Submit(context.Context, string, string, int64) error
	Review(context.Context, string, string, string, string) error
	Withdraw(context.Context, string, string, string) error
	Restore(context.Context, string, string, string) error
	SelectCover(context.Context, string, string, string) error
	CanUploadCover(context.Context, string, string) error
	SaveCover(context.Context, string, string, CoverAsset) error
	RetryPage(context.Context, string, string, string) error
	ModerationQueue(context.Context, string, ...int) ([]StudioRevision, error)
	LeaseJob(context.Context, string) (Job, bool, error)
	CompleteJob(context.Context, Job, []Variant) error
	FailJob(context.Context, Job, string) error
	Manifest(context.Context, string, string, string) (Manifest, error)
	Asset(context.Context, string, string) (Asset, error)
	Original(context.Context, string, string, string) (Variant, error)
}
type Storage interface {
	PutSource(context.Context, string, io.Reader, int64, string) error
	OpenSource(context.Context, string) (io.ReadCloser, error)
	PutDerivative(context.Context, string, io.Reader, int64, string) error
	OpenDerivative(context.Context, string, int64, int64) (io.ReadCloser, error)
}
type Processor interface {
	Process(context.Context, io.Reader, Page) (Batch, error)
}
