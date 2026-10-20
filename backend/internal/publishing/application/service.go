package application

import (
	"context"
	"io"
	"strings"
	"time"

	"qmanga/internal/publishing/domain"
)

type Service struct {
	repository     Repository
	storage        Storage
	coverProcessor CoverProcessor
	newID          func() string
}

func New(repository Repository, storage Storage, coverProcessor CoverProcessor, newID func() string) *Service {
	return &Service{repository: repository, storage: storage, coverProcessor: coverProcessor, newID: newID}
}
func (service *Service) CreateDraft(ctx context.Context, actor, project, chapter string) (StudioRevision, error) {
	return service.repository.CreateDraft(ctx, actor, project, chapter, service.newID())
}
func (service *Service) DiscardDraft(ctx context.Context, actor, id string, version int64) error {
	if version < 1 {
		return domain.ErrConflict
	}
	return service.repository.DiscardDraft(ctx, actor, id, version)
}
func (service *Service) Upload(ctx context.Context, actor string, input Upload, content io.Reader) (Reservation, error) {
	if input.Key == "" || len(input.Key) > 128 || strings.ContainsAny(input.Key, "\r\n") || input.Position < 1 || input.Position > 300 || len(input.Checksum) != 64 || len(input.OriginalName) > 255 || strings.ContainsAny(input.OriginalName, "\r\n\x00") {
		return Reservation{}, domain.ErrInvalidImage
	}
	input.PageID = service.newID()
	input.SourceKey = service.newID()
	claim := service.newID()
	reservation, failure := service.repository.Reserve(ctx, actor, input, claim)
	if failure != nil || reservation.Stored {
		return reservation, failure
	}
	if failure = service.storage.PutSource(ctx, reservation.SourceKey, content, input.Bytes, "application/octet-stream"); failure != nil {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = service.repository.UploadFailed(cleanup, reservation.PageID, claim)
		return reservation, failure
	}
	return reservation, service.repository.Uploaded(ctx, actor, reservation.PageID, claim)
}

func (service *Service) Revision(ctx context.Context, actor, id string) (StudioRevision, error) {
	return service.repository.Revision(ctx, actor, id)
}
func (service *Service) ProjectChapters(ctx context.Context, actor, id string, offset ...int) ([]ChapterChoice, error) {
	return service.repository.ProjectChapters(ctx, actor, id, offset...)
}
func (service *Service) ProjectRevisions(ctx context.Context, actor, id string, offset ...int) ([]StudioRevision, error) {
	return service.repository.ProjectRevisions(ctx, actor, id, offset...)
}
func (service *Service) Reorder(ctx context.Context, actor, id string, version int64, ids []string) error {
	return service.repository.Reorder(ctx, actor, id, version, ids)
}
func (service *Service) DeletePage(ctx context.Context, actor, id, page string, version int64) error {
	return service.repository.DeletePage(ctx, actor, id, page, version)
}
func (service *Service) SaveAttribution(ctx context.Context, actor, id string, version int64, input domain.Attribution) error {
	if !input.Valid() {
		return domain.ErrTransition
	}
	return service.repository.SaveAttribution(ctx, actor, id, version, input)
}
func (service *Service) Submit(ctx context.Context, actor, id string, version int64) error {
	return service.repository.Submit(ctx, actor, id, version)
}
func (service *Service) Review(ctx context.Context, actor, id, decision, reason string) error {
	if len(reason) > 4000 || (decision == "reject" && strings.TrimSpace(reason) == "") {
		return domain.ErrTransition
	}
	return service.repository.Review(ctx, actor, id, decision, reason)
}
func (service *Service) Withdraw(ctx context.Context, actor, id, reason string) error {
	if len(reason) > 4000 || strings.TrimSpace(reason) == "" {
		return domain.ErrTransition
	}
	return service.repository.Withdraw(ctx, actor, id, reason)
}
func (service *Service) Restore(ctx context.Context, actor, id, reason string) error {
	if len(reason) > 4000 || strings.TrimSpace(reason) == "" {
		return domain.ErrTransition
	}
	return service.repository.Restore(ctx, actor, id, reason)
}
func (service *Service) SelectCover(ctx context.Context, actor, project, page string) error {
	return service.repository.SelectCover(ctx, actor, project, page)
}
func (service *Service) UploadCover(ctx context.Context, actor, project string, source io.Reader) (string, error) {
	if failure := service.repository.CanUploadCover(ctx, actor, project); failure != nil {
		return "", failure
	}
	cover, failure := service.coverProcessor.ProcessCover(ctx, source)
	if failure != nil {
		return "", failure
	}
	defer cover.Close()
	cover.ID = service.newID()
	cover.ObjectKey = cover.ID
	if failure = service.storage.PutDerivative(ctx, cover.ObjectKey, cover.Content, cover.Bytes, cover.Mime); failure != nil {
		return "", failure
	}
	if failure = service.repository.SaveCover(ctx, actor, project, cover.CoverAsset); failure != nil {
		return "", failure
	}
	return cover.ID, nil
}
func (service *Service) RetryPage(ctx context.Context, actor, revision, page string) error {
	return service.repository.RetryPage(ctx, actor, revision, page)
}
func (service *Service) ModerationQueue(ctx context.Context, actor string, offset ...int) ([]StudioRevision, error) {
	return service.repository.ModerationQueue(ctx, actor, offset...)
}
func (service *Service) Manifest(ctx context.Context, actor, release, preview string) (Manifest, error) {
	return service.repository.Manifest(ctx, actor, release, preview)
}
func (service *Service) Asset(ctx context.Context, actor, id string) (Asset, error) {
	return service.repository.Asset(ctx, actor, id)
}
func (service *Service) Original(ctx context.Context, actor, release, page string) (Variant, error) {
	return service.repository.Original(ctx, actor, release, page)
}
func (service *Service) Stream(ctx context.Context, asset Asset, offset, length int64) (io.ReadCloser, error) {
	return service.storage.OpenDerivative(ctx, asset.ObjectKey, offset, length)
}

type Worker struct {
	repository Repository
	storage    Storage
	processor  Processor
	newID      func() string
}

func NewWorker(repository Repository, storage Storage, processor Processor, newID func() string) *Worker {
	return &Worker{repository, storage, processor, newID}
}
func (worker *Worker) Once(ctx context.Context) (bool, error) {
	job, exists, failure := worker.repository.LeaseJob(ctx, worker.newID())
	if failure != nil || !exists {
		return false, failure
	}
	operation, cancel := context.WithTimeout(ctx, 270*time.Second)
	defer cancel()
	failure = worker.process(operation, job)
	if failure != nil {
		cleanup, finish := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer finish()
		_ = worker.repository.FailJob(cleanup, job, "processing_failed")
	}
	return true, failure
}
func (worker *Worker) process(ctx context.Context, job Job) error {
	source, failure := worker.storage.OpenSource(ctx, job.Page.SourceKey)
	if failure != nil {
		return failure
	}
	defer source.Close()
	batch, failure := worker.processor.Process(ctx, source, job.Page)
	if failure != nil {
		return failure
	}
	defer batch.Close()
	variants := make([]Variant, 0, len(batch.Items))
	for _, generated := range batch.Items {
		generated.ID = worker.newID()
		generated.ObjectKey = generated.ID
		generated.PageID = job.PageID
		if failure = worker.storage.PutDerivative(ctx, generated.ObjectKey, generated.Content, generated.Bytes, generated.Mime); failure != nil {
			return failure
		}
		variants = append(variants, generated.Variant)
	}
	return worker.repository.CompleteJob(ctx, job, variants)
}

func (service *Service) FilteredModerationQueue(ctx context.Context, actor, team, kind string, offset ...int) ([]StudioRevision, error) {
	return service.repository.FilteredModerationQueue(ctx, actor, team, kind, offset...)
}
