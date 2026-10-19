package application

import (
	"context"
	"encoding/hex"
	"io"
	"qmanga/internal/catalog/domain"
	"strings"
	"time"
	"unicode/utf8"
)

type TitleChangeAsset struct {
	ID, Purpose             string
	ObjectKey               string `json:"-"`
	Mime, Checksum          string
	Width, Height, Position int
	Bytes                   int64
}
type ChangeImageSource struct {
	Purpose string
	Reader  io.Reader
}
type ProcessedChangeImage struct {
	Asset   TitleChangeAsset
	Content io.Reader
	Close   func()
}
type TitleChangeContext struct {
	ProjectID, TitleID, TitleName, TeamName, Slug string
	Base                                          domain.Metadata `json:"-"`
	Snapshot                                      domain.ChangeSnapshot
}
type TitleChange struct {
	ID, ProjectID, TitleID, TitleName, TeamName, RequestedBy, RequestedName string
	Status, Reason, Resolution, ReviewedBy, ReviewedName, BaseVersion       string
	CoverAction, BannerAction, CoverAssetID, BannerAssetID, CoverPageID     string
	Links                                                                   []string
	Before                                                                  domain.ChangeSnapshot
	Proposed                                                                domain.ProposedTitleData
	Assets                                                                  []TitleChangeAsset
	CreatedAt                                                               time.Time
	ReviewedAt                                                              *time.Time
}
type TitleChangeDetail struct {
	Request   TitleChange
	Current   domain.ChangeSnapshot
	CanReview bool
}
type TitleChangeRepository interface {
	ChangeContext(context.Context, string, string) (TitleChangeContext, error)
	CreateChange(context.Context, string, string, string, domain.TitleChangeInput, []TitleChangeAsset) (TitleChange, error)
	ChangeExists(context.Context, string) (bool, error)
	ChangeCoverPage(context.Context, string, string, string) (TitleChangeAsset, error)
	ProjectChanges(context.Context, string, string, string) ([]TitleChange, error)
	ChangeQueue(context.Context, string, string) ([]TitleChange, error)
	ChangeDetail(context.Context, string, string) (TitleChangeDetail, error)
	ReviewChange(context.Context, string, string, string, string) error
}
type ChangeImageStorage interface {
	PutDerivative(context.Context, string, io.Reader, int64, string) error
	OpenDerivative(context.Context, string, int64, int64) (io.ReadCloser, error)
	DeleteDerivative(context.Context, string) error
}
type ChangeImageProcessor interface {
	ProcessChangeImage(context.Context, io.Reader, string) (ProcessedChangeImage, error)
}
type TitleChanges struct {
	repository TitleChangeRepository
	storage    ChangeImageStorage
	processor  ChangeImageProcessor
	newID      func() string
}

func NewTitleChanges(repository TitleChangeRepository, storage ChangeImageStorage, processor ChangeImageProcessor, newID func() string) *TitleChanges {
	return &TitleChanges{repository, storage, processor, newID}
}
func validChangeID(value string) bool {
	return value != "" && len(value) <= 36 && strings.TrimSpace(value) == value
}
func (service *TitleChanges) Context(ctx context.Context, actor, project string) (TitleChangeContext, error) {
	if !validChangeID(project) {
		return TitleChangeContext{}, domain.ErrInvalid
	}
	return service.repository.ChangeContext(ctx, actor, project)
}
func (service *TitleChanges) Submit(ctx context.Context, actor, project string, input domain.TitleChangeInput, sources []ChangeImageSource) (TitleChange, error) {
	current, err := service.Context(ctx, actor, project)
	if err != nil {
		return TitleChange{}, err
	}
	evidence, cover, banner := 0, 0, 0
	for _, source := range sources {
		if source.Reader == nil {
			return TitleChange{}, domain.ErrInvalid
		}
		switch source.Purpose {
		case "evidence":
			evidence++
		case "cover":
			cover++
		case "banner":
			banner++
		default:
			return TitleChange{}, domain.ErrInvalid
		}
	}
	if cover > 1 || banner > 1 {
		return TitleChange{}, domain.ErrInvalid
	}
	if err = input.Validate(evidence, cover == 1, banner == 1); err != nil {
		return TitleChange{}, err
	}
	if input.BaseVersion != current.Snapshot.Version() {
		return TitleChange{}, domain.ErrConflict
	}
	proposed, err := input.Data.Apply(current.TitleID, current.Base)
	if err != nil {
		return TitleChange{}, err
	}
	input.Data.Name = proposed.Name
	input.Data.Authors = proposed.Authors
	input.Data.Aliases = proposed.Aliases
	candidate := domain.ChangeSnapshot{Data: input.Data, CoverID: current.Snapshot.CoverID, BannerID: current.Snapshot.BannerID}
	if input.CoverAction == "keep" && input.BannerAction == "keep" && candidate.Version() == current.Snapshot.Version() {
		return TitleChange{}, domain.ErrInvalid
	}
	if input.CoverAction == "remove" && current.Snapshot.CoverID == "" || input.BannerAction == "remove" && current.Snapshot.BannerID == "" {
		return TitleChange{}, domain.ErrInvalid
	}
	if input.CoverPageID != "" {
		page, failure := service.repository.ChangeCoverPage(ctx, actor, project, input.CoverPageID)
		if failure != nil {
			return TitleChange{}, failure
		}
		reader, failure := service.storage.OpenDerivative(ctx, page.ObjectKey, 0, -1)
		if failure != nil {
			return TitleChange{}, failure
		}
		defer reader.Close()
		sources = append(sources, ChangeImageSource{Purpose: "cover", Reader: reader})
	}
	id := service.newID()
	assets := []TitleChangeAsset{}
	stored := []string{}
	committed, attempted := false, false
	defer func() {
		if committed || len(stored) == 0 {
			return
		}
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		// A lost commit acknowledgement must never cause deletion of a committed request's images.
		if attempted {
			exists, failure := service.repository.ChangeExists(cleanup, id)
			if failure != nil || exists {
				return
			}
		}
		for _, key := range stored {
			_ = service.storage.DeleteDerivative(cleanup, key)
		}
	}()
	for position, source := range sources {
		asset, failure := service.storeImage(ctx, id, position+1, source, func(key string) { stored = append(stored, key) })
		if failure != nil {
			return TitleChange{}, failure
		}
		assets = append(assets, asset)
	}
	attempted = true
	request, err := service.repository.CreateChange(ctx, actor, project, id, input, assets)
	if err != nil {
		return TitleChange{}, err
	}
	committed = true
	return request, nil
}
func (service *TitleChanges) storeImage(ctx context.Context, request string, position int, source ChangeImageSource, reserve func(string)) (TitleChangeAsset, error) {
	processed, err := service.processor.ProcessChangeImage(ctx, source.Reader, source.Purpose)
	if err != nil {
		return TitleChangeAsset{}, err
	}
	if processed.Close != nil {
		defer processed.Close()
	}
	asset := processed.Asset
	maxWidth, maxHeight, maxBytes := 720, 1200, int64(25<<20)
	switch source.Purpose {
	case "banner":
		maxWidth, maxHeight = 1600, 800
	case "evidence":
		maxWidth, maxHeight, maxBytes = 2000, 3000, 5<<20
	}
	hash, hashError := hex.DecodeString(asset.Checksum)
	if asset.Mime != "image/webp" || hashError != nil || len(hash) != 32 || asset.Width < 1 || asset.Width > maxWidth || asset.Height < 1 || asset.Height > maxHeight || asset.Bytes < 1 || asset.Bytes > maxBytes || processed.Content == nil {
		return TitleChangeAsset{}, domain.ErrInvalid
	}
	asset.ID = service.newID()
	asset.Purpose = source.Purpose
	asset.Position = position
	asset.ObjectKey = "title-changes/" + request + "/" + asset.ID + ".webp"
	reserve(asset.ObjectKey)
	if err = service.storage.PutDerivative(ctx, asset.ObjectKey, processed.Content, asset.Bytes, asset.Mime); err != nil {
		return TitleChangeAsset{}, err
	}
	return asset, nil
}
func (service *TitleChanges) ProjectList(ctx context.Context, actor, project, after string) ([]TitleChange, error) {
	if !validChangeID(project) || len(after) > 36 {
		return nil, domain.ErrInvalid
	}
	return service.repository.ProjectChanges(ctx, actor, project, after)
}
func (service *TitleChanges) Queue(ctx context.Context, actor, after string) ([]TitleChange, error) {
	if len(after) > 36 {
		return nil, domain.ErrInvalid
	}
	return service.repository.ChangeQueue(ctx, actor, after)
}
func (service *TitleChanges) Detail(ctx context.Context, actor, id string) (TitleChangeDetail, error) {
	if !validChangeID(id) {
		return TitleChangeDetail{}, domain.ErrInvalid
	}
	return service.repository.ChangeDetail(ctx, actor, id)
}
func (service *TitleChanges) Review(ctx context.Context, actor, id, decision, reason string) error {
	reason = strings.TrimSpace(reason)
	if !validChangeID(id) || (decision != "approve" && decision != "reject") || utf8.RuneCountInString(reason) > 2000 || decision == "reject" && utf8.RuneCountInString(reason) < 3 {
		return domain.ErrInvalid
	}
	return service.repository.ReviewChange(ctx, actor, id, decision, reason)
}
