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
	"qmanga/internal/publishing/domain"
)

func TestDiscardDraftKeepsPublishedRevisionAndCancelsProcessing(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base, err := platform.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	schema := "discard_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = base.Pool().Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer base.Pool().Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	parsed, _ := url.Parse(databaseURL)
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	database, err := platform.New(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err = database.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	run := func(sql string, args ...any) {
		t.Helper()
		if _, failure := database.Pool().Exec(ctx, sql, args...); failure != nil {
			t.Fatal(failure)
		}
	}
	run(`INSERT INTO users(id,email,password_hash,is_verified) VALUES('owner','owner@example.test','hash',true),('other','other@example.test','hash',true);
INSERT INTO teams(id,slug,name,is_approved) VALUES('team','team','Team',true);
INSERT INTO team_memberships(id,team_id,user_id,role) VALUES('membership','team','owner','owner');
INSERT INTO titles(id,slug,name_kk) VALUES('title','title','Title');
INSERT INTO projects(id,title_id,team_id) VALUES('project','title','team');
INSERT INTO chapters(id,title_id,label,order_num) VALUES('chapter','title','1',1)`)
	repository := New(database.Pool())
	empty, err := repository.CreateDraft(ctx, "owner", "project", "chapter", uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	if err = repository.DiscardDraft(ctx, "other", empty.ID, empty.Version); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("foreign discard: %v", err)
	}
	if err = repository.DiscardDraft(ctx, "owner", empty.ID, empty.Version+1); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("stale discard: %v", err)
	}
	if err = repository.DiscardDraft(ctx, "owner", empty.ID, empty.Version); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = database.Pool().QueryRow(ctx, "SELECT count(*) FROM releases").Scan(&count); err != nil || count != 0 {
		t.Fatalf("empty release remained: %d %v", count, err)
	}
	if _, err = repository.Revision(ctx, "owner", empty.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("discarded draft still readable: %v", err)
	}

	run(`INSERT INTO releases(id,chapter_id,project_id,team_id,status,current_revision_id) VALUES('published','chapter','project','team','published','published-revision');
INSERT INTO revisions(id,release_id,revision_num,status) VALUES('published-revision','published',1,'published');
INSERT INTO pages(id,revision_id,position,width,height,byte_size,source_key,checksum,status) VALUES('published-page','published-revision',1,100,200,100,'source',repeat('a',64),'ready');
INSERT INTO page_variants(id,page_id,tier,object_key,width,height,byte_size,checksum,mime) VALUES('published-variant','published-page','eco','object',100,200,10,repeat('b',64),'image/webp')`)
	correction, err := repository.CreateDraft(ctx, "owner", "project", "chapter", uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	if correction.ID == "published-revision" || len(correction.Pages) != 1 {
		t.Fatal("correction was not cloned")
	}
	run("INSERT INTO jobs(id,type,payload,status) VALUES('work','process_page',jsonb_build_object('page_id',$1::text),'leased')", correction.Pages[0].ID)
	if err = repository.DiscardDraft(ctx, "owner", correction.ID, correction.Version); err != nil {
		t.Fatal(err)
	}
	if err = database.Pool().QueryRow(ctx, "SELECT count(*) FROM revisions WHERE id='published-revision'").Scan(&count); err != nil || count != 1 {
		t.Fatalf("published revision lost: %d %v", count, err)
	}
	if err = database.Pool().QueryRow(ctx, "SELECT count(*) FROM pages WHERE revision_id=$1", correction.ID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("correction pages remained: %d %v", count, err)
	}
	var status string
	if err = database.Pool().QueryRow(ctx, "SELECT status FROM jobs WHERE id='work'").Scan(&status); err != nil || status != "cancelled" {
		t.Fatalf("worker was not cancelled: %s %v", status, err)
	}
}
