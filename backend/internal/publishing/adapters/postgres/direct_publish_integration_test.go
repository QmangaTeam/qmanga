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

func TestDirectPublishAuthorizationAtomicityAndRetry(t *testing.T) {
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
	schema := "direct_" + strings.ReplaceAll(uuid.NewString(), "-", "")
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
	run(`INSERT INTO users(id,email,password_hash,is_verified,role)VALUES('owner','owner@example.test','hash',true,'admin');
INSERT INTO teams(id,slug,name,is_approved)VALUES('team','team','Team',true);
INSERT INTO team_memberships(id,team_id,user_id,role)VALUES('member','team','owner','owner');
INSERT INTO titles(id,slug,name_kk)VALUES('title','title','Title');
INSERT INTO projects(id,title_id,team_id)VALUES('project','title','team');
INSERT INTO chapters(id,title_id,label,order_num)VALUES('chapter','title','1',1);
INSERT INTO releases(id,chapter_id,project_id,team_id)VALUES('release','chapter','project','team');
INSERT INTO revisions(id,release_id,revision_num,status,version,source_note,credits,rights_confirmed,rights_confirmed_by)VALUES('revision','release',1,'ready',5,'https://source.test','Owner',true,'owner');
INSERT INTO pages(id,revision_id,position,width,height,byte_size,source_key,checksum,status)VALUES('page','revision',1,100,200,100,'source',repeat('a',64),'ready');
INSERT INTO page_variants(id,page_id,tier,object_key,width,height,byte_size,checksum,mime)VALUES('variant','page','eco','object',100,200,10,repeat('b',64),'image/webp');
INSERT INTO api_tokens(id,user_id,name,prefix,token_hash,created_at,expires_at,direct_publish)VALUES('token','owner','Importer','qmt_test',repeat('a',64),now()-interval '2 days',now()+interval '1 day',true)`)
	repository := New(database.Pool())
	cases := []struct {
		name, change, restore string
		version               int64
		want                  error
	}{
		{"revoked", "UPDATE api_tokens SET revoked_at=now()", "UPDATE api_tokens SET revoked_at=NULL", 5, domain.ErrForbidden},
		{"expired", "UPDATE api_tokens SET expires_at=now()-interval '1 second'", "UPDATE api_tokens SET expires_at=now()+interval '1 day'", 5, domain.ErrForbidden},
		{"permission", "UPDATE api_tokens SET direct_publish=false", "UPDATE api_tokens SET direct_publish=true", 5, domain.ErrForbidden},
		{"role", "UPDATE users SET role='user'", "UPDATE users SET role='admin'", 5, domain.ErrForbidden},
		{"blocked", "UPDATE users SET is_blocked=true", "UPDATE users SET is_blocked=false", 5, domain.ErrForbidden},
		{"unverified", "UPDATE users SET is_verified=false", "UPDATE users SET is_verified=true", 5, domain.ErrForbidden},
		{"foreign token owner", "INSERT INTO users(id,email,password_hash,is_verified,role)VALUES('other','other@example.test','hash',true,'admin'); UPDATE api_tokens SET user_id='other'", "UPDATE api_tokens SET user_id='owner'", 5, domain.ErrForbidden},
		{"membership absent", "DELETE FROM team_memberships", "INSERT INTO team_memberships(id,team_id,user_id,role)VALUES('member','team','owner','owner')", 5, domain.ErrForbidden},
		{"team not approved", "UPDATE teams SET is_approved=false", "UPDATE teams SET is_approved=true", 5, domain.ErrForbidden},
		{"project inactive", "UPDATE projects SET status='paused'", "UPDATE projects SET status='active'", 5, domain.ErrForbidden},
		{"stale", "", "", 4, domain.ErrConflict},
		{"attribution", "UPDATE revisions SET rights_confirmed=false", "UPDATE revisions SET rights_confirmed=true", 5, domain.ErrTransition},
		{"failed page", "UPDATE revisions SET status='processing'; UPDATE pages SET status='failed'", "UPDATE pages SET status='ready'; UPDATE revisions SET status='ready'", 5, domain.ErrTransition},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			if item.change != "" {
				run(item.change)
			}
			err := repository.DirectPublish(ctx, "owner", "token", "revision", item.version)
			if !errors.Is(err, item.want) {
				t.Fatalf("got %v want %v", err, item.want)
			}
			if item.restore != "" {
				run(item.restore)
			}
		})
	}
	if err := repository.DirectPublish(ctx, "owner", "missing", "revision", 5); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("missing token: %v", err)
	}
	if err := repository.DirectPublish(ctx, "owner", "token", "revision", 5); err != nil {
		t.Fatal(err)
	}
	if err := repository.DirectPublish(ctx, "owner", "token", "revision", 5); err != nil {
		t.Fatal(err)
	}
	var state, token string
	var version int64
	var count int
	if err := database.Pool().QueryRow(ctx, "SELECT status,version FROM revisions WHERE id='revision'").Scan(&state, &version); err != nil || state != "published" || version != 7 {
		t.Fatalf("state=%s version=%d err=%v", state, version, err)
	}
	if err := database.Pool().QueryRow(ctx, "SELECT count(*) FROM outbox_events WHERE event_type IN('release_published','publication_decision')").Scan(&count); err != nil || count != 2 {
		t.Fatalf("events=%d err=%v", count, err)
	}
	if err := database.Pool().QueryRow(ctx, "SELECT metadata->>'api_token_id' FROM audit_logs WHERE action='direct_publish' AND user_id='owner'").Scan(&token); err != nil || token != "token" {
		t.Fatalf("audit token=%s err=%v", token, err)
	}
	run("UPDATE api_tokens SET revoked_at=now()")
	if err := repository.DirectPublish(ctx, "owner", "token", "revision", 5); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("revoked retry: %v", err)
	}
}
