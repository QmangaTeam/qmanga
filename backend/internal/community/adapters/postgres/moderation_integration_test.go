package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"net/url"
	"os"
	"qmanga/internal/community/domain"
	databaseAdapter "qmanga/internal/platform/postgres"
	"strings"
	"testing"
	"time"
)

func TestModerationContextHistoryAndCursor(tester *testing.T) {
	connection := os.Getenv("TEST_DATABASE_URL")
	if connection == "" {
		tester.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base, failure := databaseAdapter.New(ctx, connection)
	if failure != nil {
		tester.Fatal(failure)
	}
	defer base.Close()
	schema := "community_moderation_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, failure = base.Pool().Exec(ctx, "CREATE SCHEMA "+schema); failure != nil {
		tester.Fatal(failure)
	}
	defer base.Pool().Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	parsed, _ := url.Parse(connection)
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	database, failure := databaseAdapter.New(ctx, parsed.String())
	if failure != nil {
		tester.Fatal(failure)
	}
	defer database.Close()
	if failure = database.Migrate(ctx); failure != nil {
		tester.Fatal(failure)
	}
	fixture := `INSERT INTO users(id,email,password_hash,role,is_verified,display_name) VALUES('admin','admin@test.local','hash','admin',true,'Admin'),('reader','reader@test.local','hash','reader',true,'Reader');
 INSERT INTO titles(id,slug,name_kk) VALUES('title','story','Story');
 INSERT INTO teams(id,slug,name) VALUES('team','team','Team');
 INSERT INTO projects(id,title_id,team_id) VALUES('project','title','team');
 INSERT INTO chapters(id,title_id,label,order_num) VALUES('chapter','title','12',1);
 INSERT INTO releases(id,chapter_id,project_id,team_id,status) VALUES('release','chapter','project','team','published');
 INSERT INTO revisions(id,release_id,status) VALUES('revision','release','published');
 INSERT INTO pages(id,revision_id,position) VALUES('page','revision',3);
 UPDATE releases SET current_revision_id='revision' WHERE id='release';
 INSERT INTO comments(id,user_id,title_id,release_id,body) VALUES('comment','reader','title','release','Reported text');
 INSERT INTO reports(id,user_id,title_id,release_id,page_id,category,reason,target_type,target_id) SELECT 'report-'||lpad(n::text,3,'0'),'reader','title','release','page','spam','Repeated spam','comment','comment' FROM generate_series(1,51) n;`
	if _, failure = database.Pool().Exec(ctx, fixture); failure != nil {
		tester.Fatal(failure)
	}
	repository := New(database.Pool())
	if _, failure = repository.Reports(ctx, "reader", ""); !errors.Is(failure, domain.ErrForbidden) {
		tester.Fatal("report context leaked", failure)
	}
	reports, failure := repository.Reports(ctx, "admin", "")
	if failure != nil || len(reports) != 50 {
		tester.Fatal("first report page", len(reports), failure)
	}
	item := reports[0]
	if item.TitleName != "Story" || item.TeamName != "Team" || item.ChapterLabel != "12" || item.PagePosition != 3 || item.CommentBody != "Reported text" || item.CommentAuthor != "Reader" {
		tester.Fatal("missing context", item)
	}
	cursor := reports[49].ID
	for _, report := range reports {
		if failure = repository.Resolve(ctx, "admin", report.ID, "resolved", "Checked and resolved"); failure != nil {
			tester.Fatal(failure)
		}
	}
	reports, failure = repository.Reports(ctx, "admin", cursor)
	if failure != nil || len(reports) != 1 || reports[0].ID != "report-051" {
		tester.Fatal("cursor lost after resolving batch", reports, failure)
	}
	if len(reports[0].History) != 10 {
		tester.Fatal("bounded decision history missing", reports[0].History)
	}
	for _, hidden := range []bool{true, false} {
		if failure = repository.Visibility(ctx, "admin", "comment", hidden, "Checked visibility"); failure != nil {
			tester.Fatal(failure)
		}
	}
	var hide, restore int
	if failure = database.Pool().QueryRow(ctx, `SELECT count(*) FILTER(WHERE action='hide_comment' AND metadata->'before'->>'hidden'='false' AND metadata->'after'->>'hidden'='true'),count(*) FILTER(WHERE action='restore_comment' AND metadata->'before'->>'hidden'='true' AND metadata->'after'->>'hidden'='false') FROM audit_logs`).Scan(&hide, &restore); failure != nil || hide != 1 || restore != 1 {
		tester.Fatal("ambiguous visibility history", hide, restore, failure)
	}
	if failure = repository.Delete(ctx, "reader", "comment"); failure != nil {
		tester.Fatal(failure)
	}
	if failure = repository.Visibility(ctx, "admin", "comment", false, "Restore deleted text"); !errors.Is(failure, domain.ErrNotFound) {
		tester.Fatal("restored deleted comment", failure)
	}
}
