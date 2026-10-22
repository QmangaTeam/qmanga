package postgres

import (
	"context"
	"github.com/google/uuid"
	"net/url"
	"os"
	platform "qmanga/internal/platform/postgres"
	"qmanga/internal/publishing/domain"
	"strings"
	"testing"
	"time"
)

func TestRevisionCountsUseSnapshotAfterLockWait(testingTarget *testing.T) {
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

	execute("INSERT INTO releases(id,project_id,chapter_id,team_id)VALUES('release','project','chapter','team')")
	execute("INSERT INTO revisions(id,release_id,revision_num,status,version)VALUES('revision','release',1,'processing',2)")
	execute("INSERT INTO pages(id,revision_id,position,width,height,byte_size,source_key,checksum,image_format,upload_key,status)VALUES('page1','revision',1,100,200,20,'source1','checksum1','png','key1','queued')")
	writer, err := database.Pool().Begin(ctx)
	if err != nil {
		testingTarget.Fatal(err)
	}
	defer writer.Rollback(ctx)
	if _, err = writer.Exec(ctx, "SELECT id FROM revisions WHERE id='revision' FOR UPDATE"); err != nil {
		testingTarget.Fatal(err)
	}
	reader, err := database.Pool().Begin(ctx)
	if err != nil {
		testingTarget.Fatal(err)
	}
	defer reader.Rollback(ctx)
	type outcome struct {
		revision domain.Revision
		err      error
	}
	result := make(chan outcome, 1)
	go func() { revision, err := lockRevision(ctx, reader, "revision"); result <- outcome{revision, err} }()
	// Wait for an actual PostgreSQL lock wait, not an assumed scheduler delay.
	for {
		var blocked bool
		if err = database.Pool().QueryRow(ctx, "SELECT cardinality(pg_blocking_pids($1))>0", reader.Conn().PgConn().PID()).Scan(&blocked); err != nil {
			testingTarget.Fatal(err)
		}
		if blocked {
			break
		}
		select {
		case <-ctx.Done():
			testingTarget.Fatal(ctx.Err())
		case <-time.After(time.Millisecond):
		}
	}
	if _, err = writer.Exec(ctx, "INSERT INTO pages(id,revision_id,position,width,height,byte_size,source_key,checksum,image_format,upload_key,status)VALUES('page2','revision',2,100,200,20,'source2','checksum2','png','key2','uploading')"); err != nil {
		testingTarget.Fatal(err)
	}
	if _, err = writer.Exec(ctx, "UPDATE revisions SET version=version+1 WHERE id='revision'"); err != nil {
		testingTarget.Fatal(err)
	}
	if err = writer.Commit(ctx); err != nil {
		testingTarget.Fatal(err)
	}
	actual := <-result
	if actual.err != nil {
		testingTarget.Fatal(actual.err)
	}
	if actual.revision.Version() != 3 || actual.revision.PageCount() != 2 {
		testingTarget.Fatalf("stale aggregate after lock wait: version=%d pages=%d; want version=3 pages=2", actual.revision.Version(), actual.revision.PageCount())
	}
	if err = actual.revision.PageReady(); err != nil {
		testingTarget.Fatal(err)
	}
	if actual.revision.State() != domain.Processing {
		testingTarget.Fatal("uploading page incorrectly marked revision ready")
	}
}
