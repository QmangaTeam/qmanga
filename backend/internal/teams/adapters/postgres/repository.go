package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"qmanga/internal/teams/application"
	"qmanga/internal/teams/domain"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

type queryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func translate(failure error) error {
	if errors.Is(failure, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var postgresError *pgconn.PgError
	if errors.As(failure, &postgresError) && postgresError.Code == "23505" {
		return domain.ErrConflict
	}
	return failure
}
func accessFor(ctx context.Context, database queryer, actorID, teamID string) (domain.Access, error) {
	var access domain.Access
	failure := database.QueryRow(ctx, `SELECT account.is_verified,account.role,coalesce(team.is_approved,false),coalesce(member.role,'')
	FROM users account LEFT JOIN teams team ON team.id=$2 LEFT JOIN team_memberships member ON member.team_id=team.id AND member.user_id=account.id WHERE account.id=$1`, actorID, teamID).Scan(&access.Verified, &access.GlobalRole, &access.Approved, &access.Role)
	if errors.Is(failure, pgx.ErrNoRows) {
		return access, domain.ErrForbidden
	}
	return access, failure
}
func (repository *Repository) Access(ctx context.Context, actorID, teamID string) (domain.Access, error) {
	return accessFor(ctx, repository.pool, actorID, teamID)
}
func (repository *Repository) ProjectAccess(ctx context.Context, actorID, projectID string) (domain.Access, error) {
	var teamID string
	if failure := repository.pool.QueryRow(ctx, "SELECT team_id FROM projects WHERE id=$1", projectID).Scan(&teamID); failure != nil {
		if errors.Is(failure, pgx.ErrNoRows) {
			return domain.Access{}, domain.ErrForbidden
		}
		return domain.Access{}, failure
	}
	return repository.Access(ctx, actorID, teamID)
}
func (repository *Repository) Create(ctx context.Context, actorID string, team domain.Team) error {
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return failure
	}
	defer transaction.Rollback(ctx)
	var verified bool
	if failure = transaction.QueryRow(ctx, "SELECT is_verified FROM users WHERE id=$1 FOR SHARE", actorID).Scan(&verified); failure != nil {
		return domain.ErrForbidden
	}
	if !verified {
		return domain.ErrForbidden
	}
	_, failure = transaction.Exec(ctx, "INSERT INTO teams(id,slug,name,description) VALUES($1,$2,$3,$4)", team.ID, team.Slug, team.Name, team.Description)
	if failure != nil {
		return translate(failure)
	}
	_, failure = transaction.Exec(ctx, "INSERT INTO team_memberships(id,team_id,user_id,role) VALUES($1,$2,$3,'owner')", uuid.NewString(), team.ID, actorID)
	if failure != nil {
		return failure
	}
	return transaction.Commit(ctx)
}
func (repository *Repository) withAccess(ctx context.Context, actorID, teamID string, action domain.Action, mutation func(pgx.Tx) error) error {
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return failure
	}
	defer transaction.Rollback(ctx)
	var lockedID string
	if failure = transaction.QueryRow(ctx, "SELECT id FROM teams WHERE id=$1 FOR UPDATE", teamID).Scan(&lockedID); failure != nil {
		return translate(failure)
	}
	if failure = transaction.QueryRow(ctx, "SELECT id FROM users WHERE id=$1 FOR SHARE", actorID).Scan(&lockedID); failure != nil {
		return domain.ErrForbidden
	}
	access, failure := accessFor(ctx, transaction, actorID, teamID)
	if failure != nil {
		return failure
	}
	if !access.Allows(action) {
		return domain.ErrForbidden
	}
	if failure = mutation(transaction); failure != nil {
		return translate(failure)
	}
	return transaction.Commit(ctx)
}
func (repository *Repository) ListForUser(ctx context.Context, actorID string) ([]application.TeamMembership, error) {
	rows, failure := repository.pool.Query(ctx, `SELECT team.id,team.slug,team.name,team.description,team.is_approved,member.role FROM teams team JOIN team_memberships member ON member.team_id=team.id WHERE member.user_id=$1 ORDER BY team.name LIMIT 100`, actorID)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := make([]application.TeamMembership, 0)
	for rows.Next() {
		var membership application.TeamMembership
		team := &membership.Team
		if failure = rows.Scan(&team.ID, &team.Slug, &team.Name, &team.Description, &team.Approved, &membership.Role); failure != nil {
			return nil, failure
		}
		items = append(items, membership)
	}
	return items, rows.Err()
}
func (repository *Repository) Pending(ctx context.Context, actorID string) ([]domain.Team, error) {
	access, failure := repository.Access(ctx, actorID, "")
	if failure != nil {
		return nil, failure
	}
	if !access.Allows(domain.Approve) {
		return nil, domain.ErrForbidden
	}
	rows, failure := repository.pool.Query(ctx, "SELECT id,slug,name,description,is_approved FROM teams WHERE NOT is_approved ORDER BY created_at LIMIT 100")
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := make([]domain.Team, 0)
	for rows.Next() {
		var team domain.Team
		if failure = rows.Scan(&team.ID, &team.Slug, &team.Name, &team.Description, &team.Approved); failure != nil {
			return nil, failure
		}
		items = append(items, team)
	}
	return items, rows.Err()
}
func (repository *Repository) Approve(ctx context.Context, actorID, teamID string) error {
	return repository.withAccess(ctx, actorID, teamID, domain.Approve, func(transaction pgx.Tx) error {
		_, failure := transaction.Exec(ctx, "UPDATE teams SET is_approved=true WHERE id=$1", teamID)
		if failure != nil {
			return failure
		}
		_, failure = transaction.Exec(ctx, "INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id) VALUES($1,$2,'approve','team',$3)", uuid.NewString(), actorID, teamID)
		return failure
	})
}
func (repository *Repository) SaveInvitation(ctx context.Context, actorID string, invitation application.Invitation) error {
	return repository.withAccess(ctx, actorID, invitation.TeamID, domain.Manage, func(transaction pgx.Tx) error {
		if !domain.Invitable(invitation.Role) {
			return domain.ErrInvalid
		}
		_, failure := transaction.Exec(ctx, `INSERT INTO team_invitations(id,team_id,email,role,token,expires_at) VALUES($1,$2,$3,$4,$5,$6)
		ON CONFLICT(team_id,email) DO UPDATE SET role=excluded.role,token=excluded.token,expires_at=excluded.expires_at`, invitation.ID, invitation.TeamID, invitation.Email, invitation.Role, invitation.Hash, invitation.ExpiresAt)
		return failure
	})
}
func (repository *Repository) AcceptInvitation(ctx context.Context, actorID, hash string, now time.Time) (string, error) {
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return "", failure
	}
	defer transaction.Rollback(ctx)
	var teamID string
	if failure = transaction.QueryRow(ctx, "SELECT team_id FROM team_invitations WHERE token=$1", hash).Scan(&teamID); failure != nil {
		if errors.Is(failure, pgx.ErrNoRows) {
			return "", domain.ErrExpired
		}
		return "", failure
	}
	var lockedID string
	if failure = transaction.QueryRow(ctx, "SELECT id FROM teams WHERE id=$1 FOR UPDATE", teamID).Scan(&lockedID); failure != nil {
		return "", translate(failure)
	}
	var email string
	var verified bool
	if failure = transaction.QueryRow(ctx, "SELECT email,is_verified FROM users WHERE id=$1 FOR SHARE", actorID).Scan(&email, &verified); failure != nil || !verified {
		return "", domain.ErrForbidden
	}
	var inviteEmail string
	var role domain.Role
	failure = transaction.QueryRow(ctx, "DELETE FROM team_invitations WHERE token=$1 AND expires_at>$2 RETURNING email,role", hash, now).Scan(&inviteEmail, &role)
	if errors.Is(failure, pgx.ErrNoRows) {
		return "", domain.ErrExpired
	}
	if failure != nil {
		return "", failure
	}
	if email != inviteEmail || !domain.Invitable(role) {
		return "", domain.ErrForbidden
	}
	var existingOwner bool
	if failure = transaction.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM team_memberships WHERE team_id=$1 AND user_id=$2 AND role='owner')", teamID, actorID).Scan(&existingOwner); failure != nil {
		return "", failure
	}
	if existingOwner {
		return "", domain.ErrConflict
	}
	_, failure = transaction.Exec(ctx, `INSERT INTO team_memberships(id,team_id,user_id,role) VALUES($1,$2,$3,$4)
	ON CONFLICT(team_id,user_id) DO UPDATE SET role=excluded.role`, uuid.NewString(), teamID, actorID, role)
	if failure != nil {
		return "", failure
	}
	return teamID, transaction.Commit(ctx)
}
func (repository *Repository) Members(ctx context.Context, actorID, teamID string) ([]application.Member, error) {
	access, failure := repository.Access(ctx, actorID, teamID)
	if failure != nil {
		return nil, failure
	}
	if !access.Allows(domain.Manage) {
		return nil, domain.ErrForbidden
	}
	rows, failure := repository.pool.Query(ctx, "SELECT member.user_id,account.email,member.role FROM team_memberships member JOIN users account ON account.id=member.user_id WHERE member.team_id=$1 ORDER BY member.created_at LIMIT 100", teamID)
	if failure != nil {
		return nil, failure
	}
	defer rows.Close()
	items := make([]application.Member, 0)
	for rows.Next() {
		var member application.Member
		if failure = rows.Scan(&member.UserID, &member.Email, &member.Role); failure != nil {
			return nil, failure
		}
		items = append(items, member)
	}
	return items, rows.Err()
}
func (repository *Repository) RemoveMember(ctx context.Context, actorID, teamID, userID string) error {
	return repository.withAccess(ctx, actorID, teamID, domain.Manage, func(transaction pgx.Tx) error {
		var role domain.Role
		if failure := transaction.QueryRow(ctx, "SELECT role FROM team_memberships WHERE team_id=$1 AND user_id=$2", teamID, userID).Scan(&role); failure != nil {
			return translate(failure)
		}
		if role == domain.Owner {
			return domain.ErrConflict
		}
		_, failure := transaction.Exec(ctx, "DELETE FROM team_memberships WHERE team_id=$1 AND user_id=$2", teamID, userID)
		return failure
	})
}
