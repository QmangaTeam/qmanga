package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"qmanga/internal/publishing/application"
	"qmanga/internal/publishing/domain"
	teamDomain "qmanga/internal/teams/domain"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool} }

type queryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func translate(failure error) error {
	if errors.Is(failure, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var databaseError *pgconn.PgError
	if errors.As(failure, &databaseError) {
		switch databaseError.Code {
		case "23505":
			return domain.ErrConflict
		case "23503", "23514", "22001":
			return domain.ErrTransition
		}
	}
	return failure
}
func access(ctx context.Context, database queryer, actor, projectID, action string) error {
	var allowed teamDomain.Access
	var status string
	failure := database.QueryRow(ctx, `SELECT account.is_verified,account.role,team.is_approved,coalesce(member.role,''),project.status FROM projects project JOIN teams team ON team.id=project.team_id JOIN users account ON account.id=$1 AND NOT account.is_blocked LEFT JOIN team_memberships member ON member.team_id=team.id AND member.user_id=account.id WHERE project.id=$2`, actor, projectID).Scan(&allowed.Verified, &allowed.GlobalRole, &allowed.Approved, &allowed.Role, &status)
	if errors.Is(failure, pgx.ErrNoRows) {
		return domain.ErrForbidden
	}
	if failure != nil {
		return failure
	}
	if action == "review" {
		if allowed.Verified && (allowed.GlobalRole == "admin" || allowed.GlobalRole == "moderator") {
			return nil
		}
		return domain.ErrForbidden
	}
	if !allowed.Allows(teamDomain.Action(action)) || (action != "read_draft" && status != "active") {
		return domain.ErrForbidden
	}
	return nil
}
func (repository *Repository) withProject(ctx context.Context, actor, projectID, action string, mutation func(pgx.Tx) error) error {
	transaction, failure := repository.pool.Begin(ctx)
	if failure != nil {
		return failure
	}
	defer transaction.Rollback(ctx)
	var locked string
	if failure = transaction.QueryRow(ctx, "SELECT team.id FROM teams team JOIN projects project ON project.team_id=team.id WHERE project.id=$1 FOR UPDATE OF team", projectID).Scan(&locked); failure != nil {
		return translate(failure)
	}
	if failure = transaction.QueryRow(ctx, "SELECT id FROM users WHERE id=$1 AND NOT is_blocked FOR SHARE", actor).Scan(&locked); failure != nil {
		return domain.ErrForbidden
	}
	if failure = access(ctx, transaction, actor, projectID, action); failure != nil {
		return failure
	}
	if failure = mutation(transaction); failure != nil {
		return translate(failure)
	}
	return transaction.Commit(ctx)
}
func (repository *Repository) withRevision(ctx context.Context, actor, id, action string, mutation func(pgx.Tx, domain.Revision) error) error {
	var projectID string
	if failure := repository.pool.QueryRow(ctx, "SELECT publication.project_id FROM revisions revision JOIN releases publication ON publication.id=revision.release_id WHERE revision.id=$1", id).Scan(&projectID); failure != nil {
		return translate(failure)
	}
	return repository.withProject(ctx, actor, projectID, action, func(transaction pgx.Tx) error {
		revision, failure := lockRevision(ctx, transaction, id)
		if failure != nil {
			return failure
		}
		return mutation(transaction, revision)
	})
}
func lockRevision(ctx context.Context, transaction pgx.Tx, id string) (domain.Revision, error) {
	var releaseID string
	var state domain.State
	var version, bytes int64
	var count, ready int
	failure := transaction.QueryRow(ctx, `SELECT revision.release_id,revision.status,revision.version,(SELECT count(*) FROM pages WHERE revision_id=revision.id),(SELECT count(*) FROM pages WHERE revision_id=revision.id AND status='ready'),coalesce((SELECT sum(byte_size) FROM pages WHERE revision_id=revision.id),0) FROM revisions revision WHERE revision.id=$1 FOR UPDATE OF revision`, id).Scan(&releaseID, &state, &version, &count, &ready, &bytes)
	if failure != nil {
		return domain.Revision{}, translate(failure)
	}
	return domain.RestoreRevision(id, releaseID, state, version, count, ready, bytes)
}
func persist(ctx context.Context, transaction pgx.Tx, revision domain.Revision) error {
	_, failure := transaction.Exec(ctx, "UPDATE revisions SET status=$2,version=$3,updated_at=now() WHERE id=$1", revision.ID(), revision.State(), revision.Version())
	return failure
}
func editable(revision domain.Revision) bool {
	return revision.State() == domain.Draft || revision.State() == domain.Processing || revision.State() == domain.Ready
}
func audit(ctx context.Context, transaction pgx.Tx, actor, action, entityID, reason string) error {
	details := map[string]any{"reason": reason}
	if action == "withdraw" {
		details["before"] = map[string]string{"status": "published"}
		details["after"] = map[string]string{"status": "withdrawn"}
	}
	if action == "restore" {
		details["before"] = map[string]string{"status": "withdrawn"}
		details["after"] = map[string]string{"status": "published"}
	}
	metadata, _ := json.Marshal(details)
	_, failure := transaction.Exec(ctx, "INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id,metadata) VALUES($1,$2,$3,'release',$4,$5)", uuid.NewString(), actor, action, entityID, metadata)
	return failure
}
func (repository *Repository) CreateDraft(ctx context.Context, actor, projectID, chapterID, id string) (application.StudioRevision, error) {
	actualID := id
	failure := repository.withProject(ctx, actor, projectID, "upload", func(transaction pgx.Tx) error {
		var teamID string
		if failure := transaction.QueryRow(ctx, "SELECT project.team_id FROM projects project JOIN chapters chapter ON chapter.title_id=project.title_id WHERE project.id=$1 AND chapter.id=$2", projectID, chapterID).Scan(&teamID); failure != nil {
			return translate(failure)
		}
		var releaseID string
		if failure := transaction.QueryRow(ctx, `INSERT INTO releases(id,chapter_id,project_id,team_id)VALUES($1,$2,$3,$4) ON CONFLICT(chapter_id,team_id)DO UPDATE SET project_id=releases.project_id RETURNING id`, uuid.NewString(), chapterID, projectID, teamID).Scan(&releaseID); failure != nil {
			return failure
		}
		failure := transaction.QueryRow(ctx, "SELECT id FROM revisions WHERE release_id=$1 AND status IN ('draft','processing','ready')", releaseID).Scan(&actualID)
		if failure == nil {
			return nil
		}
		if !errors.Is(failure, pgx.ErrNoRows) {
			return failure
		}
		actualID = id
		_, failure = transaction.Exec(ctx, "INSERT INTO revisions(id,release_id,revision_num) SELECT $1::varchar,$2::varchar,coalesce(max(revision_num),0)+1 FROM revisions WHERE release_id=$2::varchar", id, releaseID)
		if failure != nil {
			return failure
		}
		return cloneRevision(ctx, transaction, id, releaseID)
	})
	if failure != nil {
		return application.StudioRevision{}, failure
	}
	return repository.Revision(ctx, actor, actualID)
}
func (repository *Repository) Reserve(ctx context.Context, actor string, input application.Upload, claim string) (application.Reservation, error) {
	var result application.Reservation
	failure := repository.withRevision(ctx, actor, input.RevisionID, "upload", func(transaction pgx.Tx, revision domain.Revision) error {
		if !editable(revision) {
			return domain.ErrImmutable
		}
		var checksum, status string
		var lease *time.Time
		failure := transaction.QueryRow(ctx, "SELECT id,source_key,checksum,status,leased_until FROM pages WHERE revision_id=$1 AND upload_key=$2", input.RevisionID, input.Key).Scan(&result.PageID, &result.SourceKey, &checksum, &status, &lease)
		if failure == nil {
			if checksum != input.Checksum {
				return domain.ErrConflict
			}
			if status == "queued" || status == "ready" {
				result.Stored = true
				return nil
			}
			if status == "uploading" && lease != nil && lease.After(time.Now()) {
				return domain.ErrConflict
			}
			_, failure = transaction.Exec(ctx, "UPDATE pages SET status='uploading',lease_token=$2,leased_until=now()+interval '15 minutes',failure_code='' WHERE id=$1", result.PageID, claim)
			return failure
		}
		if !errors.Is(failure, pgx.ErrNoRows) {
			return failure
		}
		if failure = revision.ReservePage(input.Bytes, input.Width, input.Height, input.Format); failure != nil {
			return failure
		}
		_, failure = transaction.Exec(ctx, `INSERT INTO pages(id,revision_id,position,width,height,byte_size,source_key,checksum,image_format,upload_key,lease_token,original_name,leased_until)VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,now()+interval '15 minutes')`, input.PageID, input.RevisionID, input.Position, input.Width, input.Height, input.Bytes, input.SourceKey, input.Checksum, input.Format, input.Key, claim, input.OriginalName)
		if failure != nil {
			return failure
		}
		result = application.Reservation{PageID: input.PageID, SourceKey: input.SourceKey}
		if _, failure = transaction.Exec(ctx, "UPDATE revisions SET human_reviewed=false,human_reviewed_by=NULL,human_reviewed_at=NULL WHERE id=$1", input.RevisionID); failure != nil {
			return failure
		}
		return persist(ctx, transaction, revision)
	})
	return result, failure
}
func (repository *Repository) Uploaded(ctx context.Context, actor, pageID, claim string) error {
	var revisionID string
	if failure := repository.pool.QueryRow(ctx, "SELECT revision_id FROM pages WHERE id=$1", pageID).Scan(&revisionID); failure != nil {
		return translate(failure)
	}
	return repository.withRevision(ctx, actor, revisionID, "upload", func(transaction pgx.Tx, revision domain.Revision) error {
		if !editable(revision) {
			return domain.ErrImmutable
		}
		result, failure := transaction.Exec(ctx, "UPDATE pages SET status='queued',leased_until=NULL WHERE id=$1 AND lease_token=$2 AND status='uploading'", pageID, claim)
		if failure != nil {
			return failure
		}
		if result.RowsAffected() != 1 {
			return domain.ErrConflict
		}
		payload, _ := json.Marshal(map[string]string{"page_id": pageID})
		_, failure = transaction.Exec(ctx, `INSERT INTO jobs(id,type,payload,dedupe_key)VALUES($1,'process_page',$2,$3)ON CONFLICT(dedupe_key)DO UPDATE SET status='pending',attempts=0,available_at=now(),last_error=NULL WHERE jobs.status='failed'`, uuid.NewString(), payload, "page:"+pageID)
		if failure != nil {
			return failure
		}
		_, failure = transaction.Exec(ctx, "UPDATE revisions SET version=version+1,updated_at=now() WHERE id=$1", revisionID)
		return failure
	})
}
func (repository *Repository) UploadFailed(ctx context.Context, pageID, claim string) error {
	_, failure := repository.pool.Exec(ctx, "UPDATE pages SET status='failed',failure_code='upload_failed',leased_until=NULL WHERE id=$1 AND lease_token=$2 AND status='uploading'", pageID, claim)
	return failure
}
func (repository *Repository) Reorder(ctx context.Context, actor, id string, version int64, pageIDs []string) error {
	return repository.withRevision(ctx, actor, id, "upload", func(transaction pgx.Tx, revision domain.Revision) error {
		if !editable(revision) {
			return domain.ErrImmutable
		}
		if version != revision.Version() {
			return domain.ErrConflict
		}
		if len(pageIDs) > revision.PageCount() {
			return domain.ErrTransition
		}

		if _, failure := transaction.Exec(ctx, "UPDATE jobs SET status='cancelled' WHERE type='process_page' AND payload->>'page_id' IN(SELECT id FROM pages WHERE revision_id=$1 AND NOT(id=ANY($2::varchar[])))", id, pageIDs); failure != nil {
			return failure
		}
		if _, failure := transaction.Exec(ctx, "DELETE FROM pages WHERE revision_id=$1 AND NOT(id=ANY($2::varchar[]))", id, pageIDs); failure != nil {
			return failure
		}
		seen := map[string]bool{}
		for position, pageID := range pageIDs {
			if seen[pageID] {
				return domain.ErrConflict
			}
			seen[pageID] = true
			result, failure := transaction.Exec(ctx, "UPDATE pages SET position=$3 WHERE id=$1 AND revision_id=$2", pageID, id, position+1)
			if failure != nil {
				return failure
			}
			if result.RowsAffected() != 1 {
				return domain.ErrForbidden
			}
		}
		_, failure := transaction.Exec(ctx, "UPDATE revisions SET human_reviewed=false,human_reviewed_by=NULL,human_reviewed_at=NULL,status=CASE WHEN NOT EXISTS(SELECT 1 FROM pages WHERE revision_id=$1) THEN 'draft' WHEN NOT EXISTS(SELECT 1 FROM pages WHERE revision_id=$1 AND status<>'ready') THEN 'ready' ELSE 'processing' END,version=version+1,updated_at=now() WHERE id=$1", id)
		return failure
	})
}
func (repository *Repository) DeletePage(ctx context.Context, actor, id, pageID string, version int64) error {
	return repository.withRevision(ctx, actor, id, "upload", func(transaction pgx.Tx, revision domain.Revision) error {
		if !editable(revision) {
			return domain.ErrImmutable
		}
		if revision.Version() != version {
			return domain.ErrConflict
		}
		result, failure := transaction.Exec(ctx, "DELETE FROM pages WHERE id=$1 AND revision_id=$2", pageID, id)
		if failure != nil {
			return failure
		}
		if result.RowsAffected() != 1 {
			return domain.ErrNotFound
		}
		if _, failure = transaction.Exec(ctx, "UPDATE jobs SET status='cancelled' WHERE payload->>'page_id'=$1 AND type='process_page'", pageID); failure != nil {
			return failure
		}
		if _, failure = transaction.Exec(ctx, "UPDATE pages SET position=ordered.position FROM(SELECT id,row_number()OVER(ORDER BY position)position FROM pages WHERE revision_id=$1)ordered WHERE pages.id=ordered.id", id); failure != nil {
			return failure
		}
		_, failure = transaction.Exec(ctx, `UPDATE revisions SET human_reviewed=false,human_reviewed_by=NULL,human_reviewed_at=NULL,status=CASE WHEN NOT EXISTS(SELECT 1 FROM pages WHERE revision_id=$1)THEN 'draft' WHEN NOT EXISTS(SELECT 1 FROM pages WHERE revision_id=$1 AND status<>'ready')THEN 'ready' ELSE 'processing' END,version=version+1,updated_at=now() WHERE id=$1`, id)
		return failure
	})
}
func (repository *Repository) Submit(ctx context.Context, actor, id string, version int64) error {
	return repository.withRevision(ctx, actor, id, "submit", func(transaction pgx.Tx, revision domain.Revision) error {
		if version != revision.Version() {
			return domain.ErrConflict
		}
		var confirmed bool
		if failure := transaction.QueryRow(ctx, "SELECT rights_confirmed AND length(trim(source_note))>=8 AND length(trim(credits))>=2 AND rights_confirmed_by IS NOT NULL FROM revisions WHERE id=$1", id).Scan(&confirmed); failure != nil {
			return failure
		}
		if !confirmed {
			return domain.ErrTransition
		}
		if failure := revision.Submit(); failure != nil {
			return failure
		}
		return persist(ctx, transaction, revision)
	})
}
func (repository *Repository) Review(ctx context.Context, actor, id, decision, reason string) error {
	return repository.withRevision(ctx, actor, id, "review", func(transaction pgx.Tx, revision domain.Revision) error {
		if decision == "reject" {
			if failure := revision.Reject(reason); failure != nil {
				return failure
			}
			if failure := persist(ctx, transaction, revision); failure != nil {
				return failure
			}
			if _, failure := transaction.Exec(ctx, "UPDATE revisions SET moderation_note=$2 WHERE id=$1", id, reason); failure != nil {
				return failure
			}
			if failure := publicationDecision(ctx, transaction, id, "rejected", reason); failure != nil {
				return failure
			}
			return audit(ctx, transaction, actor, "reject", revision.ReleaseID(), reason)
		}
		if decision != "publish" {
			return domain.ErrTransition
		}
		var confirmed bool
		if failure := transaction.QueryRow(ctx, "SELECT rights_confirmed FROM revisions WHERE id=$1", id).Scan(&confirmed); failure != nil {
			return failure
		}
		if !confirmed {
			return domain.ErrTransition
		}
		return publishRevision(ctx, transaction, actor, revision, reason)
	})
}
func (repository *Repository) Withdraw(ctx context.Context, actor, releaseID, reason string) error {
	var id string
	if failure := repository.pool.QueryRow(ctx, "SELECT current_revision_id FROM releases WHERE id=$1 AND current_revision_id IS NOT NULL", releaseID).Scan(&id); failure != nil {
		return translate(failure)
	}
	return repository.withRevision(ctx, actor, id, "review", func(transaction pgx.Tx, revision domain.Revision) error {
		release := domain.Release{ID: releaseID, ActiveRevisionID: id}
		if failure := release.Withdraw(reason); failure != nil {
			return failure
		}
		result, failure := transaction.Exec(ctx, "UPDATE releases SET status='withdrawn',updated_at=now() WHERE id=$1 AND current_revision_id=$2 AND status='published'", releaseID, id)
		if failure != nil {
			return failure
		}
		if result.RowsAffected() != 1 {
			return domain.ErrConflict
		}
		if _, failure := transaction.Exec(ctx, "UPDATE titles SET cover_asset_id='' WHERE cover_asset_id IN(SELECT variant.id FROM page_variants variant JOIN pages page ON page.id=variant.page_id WHERE page.revision_id=$1)", id); failure != nil {
			return failure
		}
		if failure := publicationDecision(ctx, transaction, id, "withdrawn", reason); failure != nil {
			return failure
		}
		return audit(ctx, transaction, actor, "withdraw", releaseID, reason)
	})
}
func publicationDecision(ctx context.Context, transaction pgx.Tx, revisionID, decision, reason string) error {
	payload, failure := json.Marshal(map[string]string{"revision_id": revisionID, "decision": decision, "reason": reason})
	if failure != nil {
		return failure
	}
	_, failure = transaction.Exec(ctx, "INSERT INTO outbox_events(id,event_type,payload)VALUES($1,'publication_decision',$2)", uuid.NewString(), payload)
	return failure
}

// publishRevision is shared by moderation and the explicitly privileged bearer path.
// The caller has locked and authorized the revision in the same transaction.
func publishRevision(ctx context.Context, transaction pgx.Tx, actor string, revision domain.Revision, reason string) error {
	id := revision.ID()
	var release domain.Release
	if failure := transaction.QueryRow(ctx, "SELECT id,project_id,chapter_id FROM releases WHERE id=$1 FOR UPDATE", revision.ReleaseID()).Scan(&release.ID, &release.ProjectID, &release.ChapterID); failure != nil {
		return failure
	}
	if failure := release.Publish(&revision); failure != nil {
		return failure
	}
	if failure := persist(ctx, transaction, revision); failure != nil {
		return failure
	}
	if _, failure := transaction.Exec(ctx, "UPDATE releases SET current_revision_id=$2,status='published',updated_at=now() WHERE id=$1", release.ID, id); failure != nil {
		return failure
	}
	if failure := ensureDefaultCover(ctx, transaction, actor, release.ChapterID, id); failure != nil {
		return failure
	}

	payload, _ := json.Marshal(map[string]string{"release_id": release.ID, "revision_id": id})
	if _, failure := transaction.Exec(ctx, "INSERT INTO outbox_events(id,event_type,payload)VALUES($1,'release_published',$2)", uuid.NewString(), payload); failure != nil {
		return failure
	}
	if failure := publicationDecision(ctx, transaction, id, "published", reason); failure != nil {
		return failure
	}
	return audit(ctx, transaction, actor, "publish", release.ID, reason)
}
