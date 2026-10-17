package application

import (
	"context"
	"qmanga/internal/catalog/domain"
	"strings"
	"unicode/utf8"
)

type StudioProject struct {
	Project  Project
	Role     string
	Approved bool
}

func (service *Service) RequestProject(ctx context.Context, actor string, input domain.ProjectRequestInput) (domain.ProjectRequest, error) {
	request, failure := domain.NewProjectRequest(service.newID(), input)
	if failure != nil {
		return request, failure
	}
	return request, service.repository.CreateProjectRequest(ctx, actor, request)
}
func (service *Service) ProjectRequests(ctx context.Context, actor string, admin bool, after string) ([]domain.ProjectRequest, error) {
	if len(after) > 36 {
		return nil, domain.ErrInvalid
	}
	return service.repository.ProjectRequests(ctx, actor, admin, after)
}
func (service *Service) ReviewProjectRequest(ctx context.Context, actor, id, decision, reason, titleID string) (domain.ProjectRequest, error) {
	if !domain.ValidRequestDecision(decision, reason) || len(titleID) > 36 {
		return domain.ProjectRequest{}, domain.ErrInvalid
	}
	return service.repository.ReviewProjectRequest(ctx, actor, id, decision, strings.TrimSpace(reason), titleID)
}
func (service *Service) CancelProjectRequest(ctx context.Context, actor, id string) error {
	return service.repository.CancelProjectRequest(ctx, actor, id)
}
func (service *Service) StudioProject(ctx context.Context, actor, id string) (StudioProject, error) {
	return service.repository.StudioProject(ctx, actor, id)
}
func (service *Service) AdminProject(ctx context.Context, actor, id string) (Project, error) {
	return service.repository.AdminProject(ctx, actor, id)
}
func (service *Service) CreateStudioChapter(ctx context.Context, actor, project, label, volume, name string, order int) (string, error) {
	label = strings.TrimSpace(label)
	volume = strings.TrimSpace(volume)
	name = strings.TrimSpace(name)
	validationOrder := order
	if validationOrder == 0 {
		validationOrder = 1
	}
	if domain.ValidateChapter(label, validationOrder) != nil || utf8.RuneCountInString(volume) > 32 || utf8.RuneCountInString(name) > 200 {
		return "", domain.ErrInvalid
	}
	return service.repository.CreateStudioChapter(ctx, actor, project, label, volume, name, order)
}
