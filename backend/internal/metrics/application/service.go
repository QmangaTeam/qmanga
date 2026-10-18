package application

import (
	"context"
	"errors"
	"time"
)

var ErrForbidden = errors.New("metrics forbidden")
var ErrInvalid = errors.New("invalid metrics period")

type Day struct {
	Date                                                                            string
	Verified, Publications, Readers, Completions, Ratings, Libraries, Subscriptions int64
}
type Report struct {
	Days          []Day
	StartedAt     time.Time
	PeriodReaders int64
}
type Repository interface {
	Report(context.Context, string, bool, int) (Report, error)
}
type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository} }
func (service *Service) Report(ctx context.Context, actor string, admin bool, days int) (Report, error) {
	if days != 7 && days != 30 && days != 90 {
		return Report{}, ErrInvalid
	}
	return service.repository.Report(ctx, actor, admin, days)
}
