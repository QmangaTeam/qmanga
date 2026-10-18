package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/url"
	"os"
	"qmanga/internal/metrics/application"
	platform "qmanga/internal/platform/postgres"
	reading "qmanga/internal/reading/adapters/postgres"
	readingApp "qmanga/internal/reading/application"
	readingDomain "qmanga/internal/reading/domain"
	"strings"
	"testing"
	"time"
)

func TestMetricsScopeEventsAndReadingSuggestion(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	base, err := platform.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	schema := "metrics_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = base.Pool().Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer base.Pool().Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	parsed, _ := url.Parse(databaseURL)
	parameters := parsed.Query()
	parameters.Set("search_path", schema)
	parsed.RawQuery = parameters.Encode()
	db, err := platform.New(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	exec := func(sql string) {
		t.Helper()
		if _, err := db.Pool().Exec(ctx, sql); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO users(id,email,password_hash,role,is_verified) VALUES('admin','admin@test.example','x','admin',true),('reader','reader@test.example','x','reader',true),('translator','translator@test.example','x','reader',true),('other','other@test.example','x','reader',true); INSERT INTO teams(id,name,slug,is_approved) VALUES('own','Own','own',true),('foreign','Foreign','foreign',true);INSERT INTO team_memberships(id,user_id,team_id,role) VALUES('membership','translator','own','member');INSERT INTO titles(id,name_kk,slug) VALUES('title','Title','title');INSERT INTO projects(id,title_id,team_id) VALUES('project','title','own'),('foreign-project','title','foreign');`)
	for i := 1; i <= 3; i++ {
		exec(fmt.Sprintf(`INSERT INTO chapters(id,title_id,label,order_num) VALUES('c%d','title','%d',%d);INSERT INTO releases(id,chapter_id,project_id,team_id,status,current_revision_id) VALUES('r%d','c%d','project','own','published','v%d');INSERT INTO revisions(id,release_id,status) VALUES('v%d','r%d','published');INSERT INTO pages(id,revision_id,position) VALUES('p%d','v%d',1);`, i, i, i, i, i, i, i, i, i, i))
	}
	repo := reading.New(db.Pool())
	for i := 1; i <= 3; i++ {
		event := readingDomain.Event{ID: fmt.Sprint(i), DeviceID: "device", Sequence: int64(i), BaseVersion: int64(i - 1), TitleID: "title", ReleaseID: fmt.Sprint("r", i), RevisionID: fmt.Sprint("v", i), PageID: fmt.Sprint("p", i), Fraction: 1, Completed: true}
		result, err := repo.Advance(ctx, "reader", event)
		if err != nil || result.Outcome != "accepted" {
			t.Fatal(result, err)
		}
		if _, err = repo.Advance(ctx, "reader", event); err != nil {
			t.Fatal(err)
		}
		eligible, err := repo.Suggestion(ctx, "reader", "title")
		if err != nil || eligible != (i == 3) {
			t.Fatalf("completion %d: eligible %v, %v", i, eligible, err)
		}
	}
	// Another revision/translation of the same chapter does not add a logical completion.
	exec(`INSERT INTO reading_completions SELECT 'reader','title','c3' ON CONFLICT DO NOTHING`)
	if err = repo.Save(ctx, "reader", readingApp.Selection{TitleID: "title", Status: "reading", Notify: true}); err != nil {
		t.Fatal(err)
	}
	entry, err := repo.Selection(ctx, "reader", "title")
	if err != nil || entry.PreferredTeamID != "own" {
		t.Fatal(entry, err)
	}
	eligible, err := repo.Suggestion(ctx, "reader", "title")
	if err != nil || eligible {
		t.Fatal(eligible, err)
	}
	explicit := readingApp.Selection{TitleID: "title", Status: "reading", Notify: true, Version: entry.Version}
	if err = repo.Save(ctx, "reader", explicit); err != nil {
		t.Fatal(err)
	}
	entry, err = repo.Selection(ctx, "reader", "title")
	if err != nil || entry.PreferredTeamID != "" {
		t.Fatal("explicit clear ignored", entry, err)
	}
	exec(`INSERT INTO ratings(id,user_id,title_id,score) VALUES('rating','reader','title',8);UPDATE ratings SET score=9 WHERE id='rating';UPDATE releases SET status='draft' WHERE id='r1';UPDATE releases SET status='published' WHERE id='r1';INSERT INTO reading_activity VALUES((now() AT TIME ZONE 'UTC')::date-91,'other','r1',true);UPDATE metrics_collection SET pruned_on=(now() AT TIME ZONE 'UTC')::date-1;INSERT INTO reading_activity VALUES((now() AT TIME ZONE 'UTC')::date,'other','r1',false);`)
	metrics := New(db.Pool())
	report, err := metrics.Report(ctx, "admin", true, 7)
	if err != nil || len(report.Days) != 7 {
		t.Fatal(report, err)
	}
	today := report.Days[6]
	if today.Verified != 4 || today.Publications != 3 || today.Completions != 3 || today.Readers != 2 || today.Libraries != 1 || today.Subscriptions != 1 || today.Ratings != 1 || report.PeriodReaders != 2 {
		t.Fatalf("unexpected aggregates: %+v %+v", today, report)
	}
	own, err := metrics.Report(ctx, "translator", false, 30)
	if err != nil || own.Days[29].Completions != 3 || own.Days[29].Verified != 0 || own.Days[29].Libraries != 1 {
		t.Fatal(own, err)
	}
	if _, err = metrics.Report(ctx, "reader", true, 7); !errors.Is(err, application.ErrForbidden) {
		t.Fatal("reader saw admin report", err)
	}
	exec(`DELETE FROM team_memberships WHERE user_id='translator'`)
	if _, err = metrics.Report(ctx, "translator", false, 7); !errors.Is(err, application.ErrForbidden) {
		t.Fatal("former member retained access", err)
	}
	exec(`INSERT INTO titles(id,name_kk,slug) VALUES('single','Single','single');INSERT INTO projects(id,title_id,team_id) VALUES('single-project','single','own');`)
	for _, title := range []string{"single", "title"} {
		if err = repo.Save(ctx, "other", readingApp.Selection{TitleID: title, Status: "reading"}); err != nil {
			t.Fatal(err)
		}
		entry, err = repo.Selection(ctx, "other", title)
		if err != nil || entry.PreferredTeamID != "" {
			t.Fatal("inferred preference without actual competing-project reading", entry, err)
		}
	}
	if err = repo.Remove(ctx, "reader", "title"); err != nil {
		t.Fatal(err)
	}
	entry, err = repo.Selection(ctx, "reader", "title")
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.Save(ctx, "reader", readingApp.Selection{TitleID: "title", Status: "reading", Version: entry.Version}); err != nil {
		t.Fatal(err)
	}
	entry, err = repo.Selection(ctx, "reader", "title")
	if err != nil || entry.PreferredTeamID != "own" {
		t.Fatal("restored library entry lost reading-derived preference", entry, err)
	}
	var old int
	if err = db.Pool().QueryRow(ctx, `SELECT count(*) FROM reading_activity WHERE day<(now() AT TIME ZONE 'UTC')::date-89`).Scan(&old); err != nil || old != 0 {
		t.Fatal("retention", old, err)
	}
}
