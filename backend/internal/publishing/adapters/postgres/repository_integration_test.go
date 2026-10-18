package postgres

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	notificationPostgres "qmanga/internal/notifications/adapters/postgres"
	platform "qmanga/internal/platform/postgres"
	"qmanga/internal/publishing/application"
	"qmanga/internal/publishing/domain"
)

func TestPublicationIsolationAndAtomicReplacement(testingTarget *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		testingTarget.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base, failure := platform.New(ctx, databaseURL)
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	defer base.Close()
	schema := "publishing_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, failure = base.Pool().Exec(ctx, "CREATE SCHEMA "+schema); failure != nil {
		testingTarget.Fatal(failure)
	}
	defer base.Pool().Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	parsed, _ := url.Parse(databaseURL)
	parameters := parsed.Query()
	parameters.Set("search_path", schema)
	parsed.RawQuery = parameters.Encode()
	database, failure := platform.New(ctx, parsed.String())
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	defer database.Close()
	if failure = database.Migrate(ctx); failure != nil {
		testingTarget.Fatal(failure)
	}
	execute := func(query string, args ...any) {
		testingTarget.Helper()
		if _, failure := database.Pool().Exec(ctx, query, args...); failure != nil {
			testingTarget.Fatal(failure)
		}
	}
	execute("INSERT INTO users(id,email,password_hash,is_verified,role)VALUES('owner','owner@example.test','test',true,'reader'),('other','other@example.test','test',true,'reader'),('admin','admin@example.test','test',true,'admin')")
	execute("INSERT INTO teams(id,slug,name,is_approved)VALUES('team','demo','Дала',true)")
	execute("INSERT INTO team_memberships(id,team_id,user_id,role)VALUES('membership','team','owner','owner')")
	execute("INSERT INTO titles(id,slug,name_kk)VALUES('title','demo','Дала')")
	execute("INSERT INTO projects(id,title_id,team_id)VALUES('project','title','team')")
	execute("INSERT INTO chapters(id,title_id,label,order_num)VALUES('chapter','title','1',1)")
	repository := New(database.Pool())
	draft, failure := repository.CreateDraft(ctx, "owner", "project", "chapter", uuid.NewString())
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	if _, failure = repository.Revision(ctx, "other", draft.ID); !errors.Is(failure, domain.ErrForbidden) {
		testingTarget.Fatal("foreign draft exposed", failure)
	}
	input := application.Upload{OriginalName: "page-01.png", RevisionID: draft.ID, PageID: uuid.NewString(), Key: "file-1", SourceKey: uuid.NewString(), Position: 1, Bytes: 20, Width: 100, Height: 200, Format: "png", Checksum: strings.Repeat("a", 64)}
	reservation, failure := repository.Reserve(ctx, "owner", input, "claim")
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	if failure = repository.Uploaded(ctx, "owner", reservation.PageID, "claim"); failure != nil {
		testingTarget.Fatal(failure)
	}
	if duplicate, failure := repository.Reserve(ctx, "owner", input, "duplicate"); failure != nil || !duplicate.Stored || duplicate.PageID != reservation.PageID {
		testingTarget.Fatal("upload retry duplicated page", failure)
	}
	job, exists, failure := repository.LeaseJob(ctx, "worker-1")
	if failure != nil || !exists {
		testingTarget.Fatal("job missing", failure)
	}
	if _, exists, failure = repository.LeaseJob(ctx, "worker-2"); failure != nil || exists {
		testingTarget.Fatal("job leased twice", failure)
	}
	if failure = repository.Submit(ctx, "owner", draft.ID, 3); failure == nil {
		testingTarget.Fatal("processing revision submitted")
	}
	variants := []application.Variant{}
	for _, tier := range []string{"eco", "standard", "high", "original"} {
		variants = append(variants, application.Variant{ID: uuid.NewString(), PageID: reservation.PageID, Tier: tier, ObjectKey: uuid.NewString(), Width: 100, Height: 200, Bytes: 10, Mime: "image/webp", Checksum: strings.Repeat("b", 64)})
	}
	if failure = repository.CompleteJob(ctx, job, variants[:3]); failure == nil {
		testingTarget.Fatal("partial variants marked ready")
	}
	execute("UPDATE jobs SET leased_until=now()-interval '1 second' WHERE id=$1", job.ID)
	reclaimed, exists, failure := repository.LeaseJob(ctx, "worker-recovered")
	if failure != nil || !exists || reclaimed.ID != job.ID {
		testingTarget.Fatal("expired job not recovered", failure)
	}
	if failure = repository.CompleteJob(ctx, job, variants); !errors.Is(failure, domain.ErrConflict) {
		testingTarget.Fatal("stale worker committed", failure)
	}
	job = reclaimed
	if failure = repository.CompleteJob(ctx, job, variants); failure != nil {
		testingTarget.Fatal(failure)
	}
	draft, failure = repository.Revision(ctx, "owner", draft.ID)
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	if failure = repository.Submit(ctx, "owner", draft.ID, draft.Version); failure == nil {
		testingTarget.Fatal("publication without rights confirmation accepted")
	}
	if failure = repository.SaveAttribution(ctx, "owner", draft.ID, draft.Version, domain.Attribution{Source: "Permission from rights holder", Credits: "Translation team", Confirmed: true, PreparationMethod: "machine_assisted", HumanReviewed: true}); failure != nil {
		testingTarget.Fatal(failure)
	}
	draft, failure = repository.Revision(ctx, "owner", draft.ID)
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	if failure = repository.Submit(ctx, "owner", draft.ID, draft.Version); failure != nil {
		testingTarget.Fatal(failure)
	}
	if _, failure = repository.Reserve(ctx, "owner", input, "new-claim"); failure == nil {
		testingTarget.Fatal("submitted pages mutable")
	}
	if failure = repository.Review(ctx, "owner", draft.ID, "publish", ""); !errors.Is(failure, domain.ErrForbidden) {
		testingTarget.Fatal("owner published own revision")
	}
	execute("UPDATE revisions SET rights_confirmed=false WHERE id=$1", draft.ID)
	if failure = repository.Review(ctx, "admin", draft.ID, "publish", ""); !errors.Is(failure, domain.ErrTransition) {
		testingTarget.Fatal("legacy unconfirmed submission published", failure)
	}
	execute("UPDATE revisions SET rights_confirmed=true WHERE id=$1", draft.ID)
	if failure = repository.Review(ctx, "admin", draft.ID, "publish", ""); failure != nil {
		testingTarget.Fatal(failure)
	}
	notifications := notificationPostgres.New(database.Pool())
	for {
		worked, deliveryFailure := notifications.Deliver(ctx)
		if deliveryFailure != nil {
			testingTarget.Fatal(deliveryFailure)
		}
		if !worked {
			break
		}
	}
	decisions, failure := notifications.List(ctx, "owner", "")
	if failure != nil || len(decisions) != 1 || decisions[0].Kind != "publication_published" {
		testingTarget.Fatal("publication decision missing", decisions, failure)
	}
	foreignDecisions, failure := notifications.List(ctx, "other", "")
	if failure != nil || len(foreignDecisions) != 0 {
		testingTarget.Fatal("decision crossed team", foreignDecisions, failure)
	}
	manifest, failure := repository.Manifest(ctx, "", draft.ReleaseID, "")
	if failure != nil || manifest.RevisionID != draft.ID || len(manifest.Pages) != 1 || manifest.PreparationMethod != "machine_assisted" || !manifest.HumanReviewed || manifest.HumanReviewedAt == nil {
		testingTarget.Fatal("published manifest missing", failure)
	}
	replacement, failure := repository.CreateDraft(ctx, "owner", "project", "chapter", uuid.NewString())
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	if len(replacement.Pages) != 1 || replacement.Pages[0].ID == draft.Pages[0].ID || replacement.Pages[0].OriginalName != "page-01.png" || replacement.Pages[0].ThumbnailID == "" || replacement.Credits != draft.Credits || replacement.State != "ready" || replacement.HumanReviewed || replacement.HumanReviewedAt != nil {
		testingTarget.Fatal("correction did not clone pages and credits", replacement)
	}
	if _, err := repository.Asset(ctx, "", replacement.Pages[0].ThumbnailID); err == nil {
		testingTarget.Fatal("draft derivative exposed")
	}
	if _, err := repository.Asset(ctx, "", variants[3].ID); err == nil {
		testingTarget.Fatal("private original exposed")
	}
	if err := repository.SelectCover(ctx, "other", "project", replacement.Pages[0].ID); !errors.Is(err, domain.ErrForbidden) {
		testingTarget.Fatal("foreign cover accepted", err)
	}
	if err := repository.SelectCover(ctx, "owner", "project", replacement.Pages[0].ID); err != nil {
		testingTarget.Fatal(err)
	}
	var coverID string
	if err := database.Pool().QueryRow(ctx, "SELECT cover_asset_id FROM titles WHERE id='title'").Scan(&coverID); err != nil {
		testingTarget.Fatal(err)
	}
	if _, err := repository.Asset(ctx, "", coverID); err != nil {
		testingTarget.Fatal("managed cover unavailable", err)
	}
	if failure = repository.Review(ctx, "admin", replacement.ID, "publish", ""); failure == nil {
		testingTarget.Fatal("incomplete replacement published")
	}
	manifest, failure = repository.Manifest(ctx, "", draft.ReleaseID, "")
	if failure != nil || manifest.RevisionID != draft.ID {
		testingTarget.Fatal("failed replacement broke active publication")
	}
	if failure = repository.Withdraw(ctx, "admin", draft.ReleaseID, "Тексеру"); failure != nil {
		testingTarget.Fatal(failure)
	}
	if _, failure = repository.Asset(ctx, "", variants[0].ID); failure == nil {
		testingTarget.Fatal("withdrawn media still public")
	}
	if _, failure = repository.Manifest(ctx, "", draft.ReleaseID, ""); !errors.Is(failure, domain.ErrWithdrawn) {
		testingTarget.Fatal("withdrawn manifest exposed", failure)
	}

	if _, err := repository.Asset(ctx, "", coverID); err != nil {
		testingTarget.Fatal("withdrawal removed independently selected cover", err)
	}
	if err := repository.Restore(ctx, "owner", draft.ReleaseID, "Rights checked"); !errors.Is(err, domain.ErrForbidden) {
		testingTarget.Fatal("owner restored release", err)
	}
	if err := repository.Restore(ctx, "admin", draft.ReleaseID, ""); !errors.Is(err, domain.ErrTransition) {
		testingTarget.Fatal("reasonless restore accepted", err)
	}
	if err := repository.Restore(ctx, "admin", draft.ReleaseID, "Rights checked"); err != nil {
		testingTarget.Fatal(err)
	}
	if restored, err := repository.Manifest(ctx, "", draft.ReleaseID, ""); err != nil || restored.RevisionID != draft.ID {
		testingTarget.Fatal("restore changed revision", err)
	}
	if err := repository.Submit(ctx, "owner", replacement.ID, replacement.Version); err != nil {
		testingTarget.Fatal(err)
	}
	if err := repository.Review(ctx, "admin", replacement.ID, "publish", ""); err != nil {
		testingTarget.Fatal(err)
	}
	if active, err := repository.Manifest(ctx, "", draft.ReleaseID, ""); err != nil || active.RevisionID != replacement.ID || active.CoverID != coverID {
		testingTarget.Fatal("replacement changed managed cover or failed to switch active revision", err)
	}
	if _, err := repository.Asset(ctx, "", variants[0].ID); err == nil {
		testingTarget.Fatal("superseded publication page remained public")
	}
	replacement, failure = repository.CreateDraft(ctx, "owner", "project", "chapter", uuid.NewString())
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	if err := repository.Reorder(ctx, "owner", replacement.ID, replacement.Version, []string{}); err != nil {
		testingTarget.Fatal(err)
	}
	if _, err := repository.Asset(ctx, "", coverID); err != nil {
		testingTarget.Fatal("draft deletion removed independent cover", err)
	}
	testingTarget.Run("staged page deletion is atomic and versioned", func(testingTarget *testing.T) {
		for position := 1; position <= 2; position++ {
			pageInput := input
			pageInput.RevisionID = replacement.ID
			pageInput.PageID = uuid.NewString()
			pageInput.Key = pageInput.PageID
			pageInput.SourceKey = pageInput.PageID
			pageInput.Position = position
			reserved, failure := repository.Reserve(ctx, "owner", pageInput, "order-claim")
			if failure != nil {
				testingTarget.Fatal(failure)
			}
			if failure = repository.Uploaded(ctx, "owner", reserved.PageID, "order-claim"); failure != nil {
				testingTarget.Fatal(failure)
			}
		}
		var failedPage string
		if err := database.Pool().QueryRow(ctx, "SELECT id FROM pages WHERE revision_id=$1 AND position=2", replacement.ID).Scan(&failedPage); err != nil {
			testingTarget.Fatal(err)
		}
		execute("UPDATE pages SET status='failed',failure_code='processing_failed' WHERE id=$1", failedPage)
		execute("UPDATE jobs SET status='failed',attempts=max_attempts WHERE dedupe_key=$1", "page:"+failedPage)
		if err := repository.RetryPage(ctx, "other", replacement.ID, failedPage); !errors.Is(err, domain.ErrForbidden) {
			testingTarget.Fatal("foreign retry accepted", err)
		}
		if err := repository.RetryPage(ctx, "owner", replacement.ID, failedPage); err != nil {
			testingTarget.Fatal(err)
		}
		var retried string
		if err := database.Pool().QueryRow(ctx, "SELECT status FROM jobs WHERE dedupe_key=$1", "page:"+failedPage).Scan(&retried); err != nil || retried != "pending" {
			testingTarget.Fatal("retry did not schedule processing", retried, err)
		}
		before, failure := repository.Revision(ctx, "owner", replacement.ID)
		if failure != nil || len(before.Pages) != 2 {
			testingTarget.Fatal("draft pages missing", failure)
		}
		if failure = repository.Reorder(ctx, "owner", before.ID, before.Version, []string{"foreign-page"}); !errors.Is(failure, domain.ErrForbidden) {
			testingTarget.Fatal("foreign page accepted", failure)
		}
		unchanged, failure := repository.Revision(ctx, "owner", before.ID)
		if failure != nil || len(unchanged.Pages) != 2 || unchanged.Version != before.Version {
			testingTarget.Fatal("invalid order deleted pages")
		}
		kept := before.Pages[1].ID
		if failure = repository.Reorder(ctx, "owner", before.ID, before.Version, []string{kept}); failure != nil {
			testingTarget.Fatal(failure)
		}
		after, failure := repository.Revision(ctx, "owner", before.ID)
		if failure != nil || len(after.Pages) != 1 || after.Pages[0].ID != kept || after.Pages[0].Position != 1 || after.Version != before.Version+1 {
			testingTarget.Fatal("staged deletion not applied", failure)
		}
		if failure = repository.Reorder(ctx, "owner", before.ID, before.Version, []string{}); !errors.Is(failure, domain.ErrConflict) {
			testingTarget.Fatal("stale delete accepted", failure)
		}
		var status string
		if failure = database.Pool().QueryRow(ctx, "SELECT status FROM jobs WHERE payload->>'page_id'=$1", before.Pages[0].ID).Scan(&status); failure != nil || status != "cancelled" {
			testingTarget.Fatal("deleted page job remains active", status, failure)
		}
	})
}
