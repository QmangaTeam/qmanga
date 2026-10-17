package application

import (
	"context"
	"qmanga/internal/teams/domain"
	"time"
)

type Invitation struct {
	ID, TeamID, Email, Hash string
	Role                    domain.Role
	ExpiresAt               time.Time
}
type Member struct {
	UserID, Email string
	Role          domain.Role
}
type TeamMembership struct {
	Team domain.Team
	Role domain.Role
}
type Repository interface {
	Create(context.Context, string, domain.Team) error
	Access(context.Context, string, string) (domain.Access, error)
	ProjectAccess(context.Context, string, string) (domain.Access, error)
	ListForUser(context.Context, string) ([]TeamMembership, error)
	Pending(context.Context, string, ...int) ([]domain.Team, error)
	Approve(context.Context, string, string) error
	SaveInvitation(context.Context, string, Invitation) error
	AcceptInvitation(context.Context, string, string, time.Time) (string, error)
	Members(context.Context, string, string) ([]Member, error)
	RemoveMember(context.Context, string, string, string) error
}
type Tokens interface {
	NewID() string
	RandomToken() (string, error)
	Digest(string) string
}
type Mailer interface {
	Send(context.Context, string, string, string) error
}
type Service struct {
	repository Repository
	tokens     Tokens
	mailer     Mailer
	origin     string
}

func New(repository Repository, tokens Tokens, mailer Mailer, origin string) *Service {
	return &Service{repository: repository, tokens: tokens, mailer: mailer, origin: origin}
}
func (service *Service) Create(ctx context.Context, actorID, slug, name, description string) (domain.Team, error) {
	team, failure := domain.NewTeam(service.tokens.NewID(), slug, name, description)
	if failure != nil {
		return team, failure
	}
	return team, service.repository.Create(ctx, actorID, team)
}
func (service *Service) CheckProject(ctx context.Context, actorID, projectID string, action domain.Action) error {
	access, failure := service.repository.ProjectAccess(ctx, actorID, projectID)
	if failure != nil {
		return failure
	}
	if !access.Allows(action) {
		return domain.ErrForbidden
	}
	return nil
}
func (service *Service) Check(ctx context.Context, actorID, teamID string, action domain.Action) error {
	access, failure := service.repository.Access(ctx, actorID, teamID)
	if failure != nil {
		return failure
	}
	if !access.Allows(action) {
		return domain.ErrForbidden
	}
	return nil
}
func (service *Service) List(ctx context.Context, actorID string) ([]TeamMembership, error) {
	return service.repository.ListForUser(ctx, actorID)
}
func (service *Service) Pending(ctx context.Context, actorID string, offset ...int) ([]domain.Team, error) {
	return service.repository.Pending(ctx, actorID, offset...)
}
func (service *Service) Approve(ctx context.Context, actorID, teamID string) error {
	return service.repository.Approve(ctx, actorID, teamID)
}
func (service *Service) Invite(ctx context.Context, actorID, teamID, email string, role domain.Role) error {
	if !domain.Invitable(role) {
		return domain.ErrInvalid
	}
	normalized, failure := domain.InvitationEmail(email)
	if failure != nil {
		return failure
	}
	if failure = service.Check(ctx, actorID, teamID, domain.Manage); failure != nil {
		return failure
	}
	token, failure := service.tokens.RandomToken()
	if failure != nil {
		return failure
	}
	invitation := Invitation{ID: service.tokens.NewID(), TeamID: teamID, Email: normalized, Role: role, Hash: service.tokens.Digest(token), ExpiresAt: time.Now().Add(7 * 24 * time.Hour)}
	if failure = service.repository.SaveInvitation(ctx, actorID, invitation); failure != nil {
		return failure
	}
	return service.mailer.Send(ctx, normalized, "Qmanga — командаға шақыру", "Командаға қосылу үшін сілтемені ашыңыз:\n"+service.origin+"/teams/invitation?token="+token+"\n\nШақыру жеті күн жарамды. Поштаңыз расталған болуы керек.")
}
func (service *Service) Accept(ctx context.Context, actorID, token string) (string, error) {
	if len(token) != 43 {
		return "", domain.ErrExpired
	}
	return service.repository.AcceptInvitation(ctx, actorID, service.tokens.Digest(token), time.Now())
}
func (service *Service) Members(ctx context.Context, actorID, teamID string) ([]Member, error) {
	return service.repository.Members(ctx, actorID, teamID)
}
func (service *Service) Remove(ctx context.Context, actorID, teamID, userID string) error {
	return service.repository.RemoveMember(ctx, actorID, teamID, userID)
}
