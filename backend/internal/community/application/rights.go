package application

import (
	"context"
	"net/mail"
	"qmanga/internal/community/domain"
	"regexp"
	"strings"
)

type RightsRequest struct{ TitleSlug, Email, Reason string }
type ReportReceipt struct{ ID, Token string }

var titleSlug = regexp.MustCompile("^[a-z0-9]+(-[a-z0-9]+)*$")

func (service *Service) RightsReport(ctx context.Context, input RightsRequest) (ReportReceipt, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.TitleSlug = strings.TrimSpace(input.TitleSlug)
	address, failure := mail.ParseAddress(input.Email)
	if failure != nil || address.Address != input.Email || len(input.Email) > 254 || !strings.Contains(input.Email, "@") || !domain.ValidReason(input.Reason) || len(input.TitleSlug) > 128 || !titleSlug.MatchString(input.TitleSlug) {
		return ReportReceipt{}, domain.ErrInvalid
	}
	receipt := ReportReceipt{ID: service.newID(), Token: service.newID() + service.newID()}
	if failure = service.repository.RightsReport(ctx, receipt, input); failure != nil {
		return ReportReceipt{}, failure
	}
	return receipt, nil
}
func (service *Service) RightsStatus(ctx context.Context, id, token string) (Report, error) {
	if len(id) > 36 || len(token) != 72 {
		return Report{}, domain.ErrNotFound
	}
	return service.repository.RightsStatus(ctx, id, token)
}
func (service *Service) MyReports(ctx context.Context, actor, after string) ([]Report, error) {
	if len(after) > 36 {
		return nil, domain.ErrInvalid
	}
	return service.repository.MyReports(ctx, actor, after)
}
