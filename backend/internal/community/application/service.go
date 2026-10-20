package application

import (
	"context"
	"qmanga/internal/community/domain"
	"time"
)

type Comment struct {
	DisplayName string
	domain.Comment
	CreatedAt time.Time
	Replies   int
}
type Discussion struct {
	Comments []Comment
	Next     string
	Score    float64
	Ratings  int
}
type ReportDecision struct {
	Status    string    `json:"status"`
	Action    string    `json:"action"`
	ActorName string    `json:"actor_name"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}
type Report struct {
	TitleName, TitleSlug, TeamName, TeamSlug, ChapterLabel, ReleaseStatus, RevisionID                  string
	CommentBody, CommentAuthor                                                                         string
	CommentHidden, CommentDeleted                                                                      bool
	PagePosition                                                                                       int
	History                                                                                            []ReportDecision
	ContactEmail                                                                                       string
	ID, UserID, TitleID, ReleaseID, PageID, TargetType, TargetID, Category, Reason, Status, Resolution string
	CreatedAt                                                                                          time.Time
}
type Repository interface {
	RightsReport(context.Context, ReportReceipt, RightsRequest) error
	RightsStatus(context.Context, string, string) (Report, error)
	MyReports(context.Context, string, string) ([]Report, error)
	Replies(context.Context, string, string) (Discussion, error)
	Discussion(context.Context, string, string, string, ...string) (Discussion, error)
	Comment(context.Context, string, domain.Comment) (string, error)
	Edit(context.Context, string, string, string, bool) error
	Delete(context.Context, string, string) error
	Rating(context.Context, string, string, int) error
	MyRating(context.Context, string, string) (int, error)
	Report(context.Context, string, Report) (string, error)
	Reports(context.Context, string, string, ...string) ([]Report, error)
	Resolve(context.Context, string, string, string, string) error
	Visibility(context.Context, string, string, bool, string) error
}
type Service struct {
	repository Repository
	newID      func() string
}

func (service *Service) Replies(ctx context.Context, parent, after string) (Discussion, error) {
	return service.repository.Replies(ctx, parent, after)
}

func New(repository Repository, newID func() string) *Service { return &Service{repository, newID} }
func (service *Service) Discussion(ctx context.Context, title, release, after string, focus ...string) (Discussion, error) {
	return service.repository.Discussion(ctx, title, release, after, focus...)
}
func (service *Service) Comment(ctx context.Context, actor string, input domain.Comment) (string, error) {
	if !domain.ValidBody(input.Body) || input.TitleID == "" {
		return "", domain.ErrInvalid
	}
	input.ID = service.newID()
	return service.repository.Comment(ctx, actor, input)
}
func (service *Service) Edit(ctx context.Context, actor, id, body string, spoiler bool) error {
	if !domain.ValidBody(body) {
		return domain.ErrInvalid
	}
	return service.repository.Edit(ctx, actor, id, body, spoiler)
}
func (service *Service) Delete(ctx context.Context, actor, id string) error {
	return service.repository.Delete(ctx, actor, id)
}
func (service *Service) Rating(ctx context.Context, actor, title string, score int) error {
	if score < 0 || score > 10 {
		return domain.ErrInvalid
	}
	return service.repository.Rating(ctx, actor, title, score)
}
func (service *Service) MyRating(ctx context.Context, actor, title string) (int, error) {
	return service.repository.MyRating(ctx, actor, title)
}
func (service *Service) Report(ctx context.Context, actor string, input Report) (string, error) {
	if !domain.ValidCategory(input.Category) || !domain.ValidReason(input.Reason) {
		return "", domain.ErrInvalid
	}
	input.ID = service.newID()
	return service.repository.Report(ctx, actor, input)
}
func (service *Service) Reports(ctx context.Context, actor, after string, filters ...string) ([]Report, error) {
	return service.repository.Reports(ctx, actor, after, filters...)
}
func (service *Service) Resolve(ctx context.Context, actor, id, status, reason string) error {
	if !domain.ValidReason(reason) || (status != "resolved" && status != "dismissed") {
		return domain.ErrInvalid
	}
	return service.repository.Resolve(ctx, actor, id, status, reason)
}
func (service *Service) Visibility(ctx context.Context, actor, id string, hidden bool, reason string) error {
	if !domain.ValidReason(reason) {
		return domain.ErrInvalid
	}
	return service.repository.Visibility(ctx, actor, id, hidden, reason)
}
