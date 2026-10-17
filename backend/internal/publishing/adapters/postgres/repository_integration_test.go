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
	input := application.Upload{RevisionID: draft.ID, PageID: uuid.NewString(), Key: "file-1", SourceKey: uuid.NewString(), Position: 1, Bytes: 20, Width: 100, Height: 200, Format: "png", Checksum: strings.Repeat("a", 64)}
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
	if failure = repository.Submit(ctx, "owner", draft.ID, draft.Version); failure != nil {
		testingTarget.Fatal(failure)
	}
	if _, failure = repository.Reserve(ctx, "owner", input, "new-claim"); failure == nil {
		testingTarget.Fatal("submitted pages mutable")
	}
	if failure = repository.Review(ctx, "owner", draft.ID, "publish", ""); !errors.Is(failure, domain.ErrForbidden) {
		testingTarget.Fatal("owner published own revision")
	}
	if failure = repository.Review(ctx, "admin", draft.ID, "publish", ""); failure != nil {
		testingTarget.Fatal(failure)
	}
	manifest, failure := repository.Manifest(ctx, "", draft.ReleaseID, "")
	if failure != nil || manifest.RevisionID != draft.ID || len(manifest.Pages) != 1 {
		testingTarget.Fatal("published manifest missing", failure)
	}
	replacement, failure := repository.CreateDraft(ctx, "owner", "project", "chapter", uuid.NewString())
	if failure != nil {
		testingTarget.Fatal(failure)
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
}
