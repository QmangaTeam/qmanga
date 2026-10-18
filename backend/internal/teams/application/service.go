package application

import (
	"context"
	"qmanga/internal/teams/domain"
	"strings"
	"time"
	"unicode/utf8"
)

type Invitation struct {
	ID, TeamID, Email, Hash string
	Role                    domain.Role
	ExpiresAt               time.Time
}
type InvitationSummary struct {
	ID, Email, Status string
	Role              domain.Role
	ExpiresAt         time.Time
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
	ListForUser(context.Context, string, ...int) ([]TeamMembership, error)
	Pending(context.Context, string, ...int) ([]domain.Team, error)
	Approve(context.Context, string, string) error
	Reject(context.Context, string, string, string) error
	Update(context.Context, string, domain.Team) error
	SaveInvitation(context.Context, string, Invitation) error
	Invitations(context.Context, string, string, ...int) ([]InvitationSummary, error)
	RevokeInvitation(context.Context, string, string, string) error
	AcceptInvitation(context.Context, string, string, time.Time) (string, error)
	Members(context.Context, string, string, ...int) ([]Member, error)
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
func (service *Service) Create(ctx context.Context, actorID, slug, name, description string, telegram ...string) (domain.Team, error) {
	team, failure := domain.NewTeam(service.tokens.NewID(), slug, name, description, telegram...)
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
func (service *Service) List(ctx context.Context, actorID string, offset ...int) ([]TeamMembership, error) {
	return service.repository.ListForUser(ctx, actorID, offset...)
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
func (service *Service) Members(ctx context.Context, actorID, teamID string, offset ...int) ([]Member, error) {
	return service.repository.Members(ctx, actorID, teamID, offset...)
}
func (service *Service) Remove(ctx context.Context, actorID, teamID, userID string) error {
	return service.repository.RemoveMember(ctx, actorID, teamID, userID)
}

func (service *Service) Reject(ctx context.Context, actor, id, reason string) error {
	reason = strings.TrimSpace(reason)
	if utf8.RuneCountInString(reason) < 8 || utf8.RuneCountInString(reason) > 2000 {
		return domain.ErrInvalid
	}
	return service.repository.Reject(ctx, actor, id, reason)
}
func (service *Service) Update(ctx context.Context, actor, id, slug, name, description string, telegram ...string) error {
	team, failure := domain.NewTeam(id, slug, name, description, telegram...)
	if failure != nil {
		return failure
	}
	return service.repository.Update(ctx, actor, team)
}
func (service *Service) Invitations(ctx context.Context, actor, team string, offset ...int) ([]InvitationSummary, error) {
	return service.repository.Invitations(ctx, actor, team, offset...)
}
func (service *Service) RevokeInvitation(ctx context.Context, actor, team, id string) error {
	return service.repository.RevokeInvitation(ctx, actor, team, id)
}

// CheckAvatar permits verified team owners and editors to change team imagery.
func (service *Service) CheckAvatar(ctx context.Context, actorID, teamID string) error {
	access, err := service.repository.Access(ctx, actorID, teamID)
	if err != nil {
		return err
	}
	if !access.Verified || (access.Role != domain.Owner && access.Role != domain.Editor) {
		return domain.ErrForbidden
	}
	return nil
}
