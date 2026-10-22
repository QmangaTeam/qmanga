package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"net/url"
	"os"
	"qmanga/internal/identity/application"
	"qmanga/internal/identity/domain"
	databaseAdapter "qmanga/internal/platform/postgres"
	"strings"
	"testing"
	"time"
)

func TestAccountProfilesPostgres(t *testing.T) {
	connection := os.Getenv("TEST_DATABASE_URL")
	if connection == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base, err := databaseAdapter.New(ctx, connection)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	schema := "profiles_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = base.Pool().Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer base.Pool().Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	parsed, _ := url.Parse(connection)
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := databaseAdapter.New(ctx, parsed.String())
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
	exec(`INSERT INTO users(id,email,password_hash,is_verified) VALUES('owner','owner@test.local','hash',true),('viewer','viewer@test.local','hash',true),('unverified','unverified@test.local','hash',false)`)
	r := New(db.Pool())
	p, err := r.Preferences(ctx, "owner")
	if err != nil || p.Version != 0 || p.PublicProfile || !p.AllowFollow || !p.NotifyReleases {
		t.Fatalf("defaults %+v %v", p, err)
	}
	if _, err = r.PublicProfile(ctx, "", "owner"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("private profile exposed", err)
	}
	own, err := r.PublicProfile(ctx, "owner", "owner")
	if err != nil || !own.IsOwner {
		t.Fatal("owner cannot read private", err)
	}
	p.Bio = "Қазақша"
	p.PublicProfile = true
	p.ShowLibrary = true
	p.ShowActivity = true
	p, err = r.SavePreferences(ctx, "owner", p)
	if err != nil || p.Version != 1 {
		t.Fatal("first save", p, err)
	}
	if _, err = r.SavePreferences(ctx, "owner", application.DefaultPreferences()); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("stale create accepted", err)
	}
	stale := p
	p.Bio = "Өзгерді"
	p, err = r.SavePreferences(ctx, "owner", p)
	if err != nil || p.Version != 2 {
		t.Fatal("second save", p, err)
	}
	if _, err = r.SavePreferences(ctx, "owner", stale); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("stale update accepted", err)
	}
	if err = r.FollowProfile(ctx, "unverified", "owner", true); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("unverified followed", err)
	}
	for i := 0; i < 2; i++ {
		if err = r.FollowProfile(ctx, "viewer", "owner", true); err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err = db.Pool().QueryRow(ctx, `SELECT count(*) FROM outbox_events WHERE event_type='profile_followed' AND payload->>'user_id'='owner' AND payload->>'actor_id'='viewer'`).Scan(&count); err != nil || count != 1 {
		t.Fatal("follow dedupe", count, err)
	}
	pub, err := r.PublicProfile(ctx, "viewer", "owner")
	if err != nil || !pub.IsFollowing || pub.FollowersCount != 1 {
		t.Fatal("follow read model", pub, err)
	}
	exec(`INSERT INTO titles(id,slug,name_kk,is_listed) VALUES('visible','visible','Көрінеді',true),('hidden','hidden','Жасырын',false); INSERT INTO library_entries(id,user_id,title_id) VALUES('v','owner','visible'),('h','owner','hidden')`)
	pub, err = r.PublicProfile(ctx, "", "owner")
	if err != nil || len(pub.Library) != 1 || pub.Library[0].TitleID != "visible" {
		t.Fatal("library disclosure", pub, err)
	}
	exec(`INSERT INTO teams(id,slug,name,is_approved) VALUES('team','team','Team',true); INSERT INTO projects(id,title_id,team_id) VALUES('project','hidden','team'); INSERT INTO chapters(id,title_id,label,order_num) VALUES('chapter','hidden','1',1); INSERT INTO releases(id,chapter_id,project_id,team_id,status,current_revision_id) VALUES('release','chapter','project','team','published','revision'); INSERT INTO revisions(id,release_id,status) VALUES('revision','release','published'); INSERT INTO reading_activity(day,user_id,release_id) VALUES(current_date,'owner','release')`)
	pub, err = r.PublicProfile(ctx, "", "owner")
	if err != nil || len(pub.Activity) != 1 || len(pub.Library) != 2 {
		t.Fatal("published activity missing", pub, err)
	}
	exec(`UPDATE releases SET status='withdrawn' WHERE id='release'`)
	pub, err = r.PublicProfile(ctx, "", "owner")
	if err != nil || len(pub.Activity) != 0 || len(pub.Library) != 1 {
		t.Fatal("withdrawn disclosure", pub, err)
	}
	p.ShowLibrary = false
	p.ShowActivity = false
	p, err = r.SavePreferences(ctx, "owner", p)
	if err != nil {
		t.Fatal(err)
	}
	pub, err = r.PublicProfile(ctx, "", "owner")
	if err != nil || len(pub.Library) != 0 || len(pub.Activity) != 0 {
		t.Fatal("section privacy ignored", pub, err)
	}
	own, err = r.PublicProfile(ctx, "owner", "owner")
	if err != nil || len(own.Library) != 1 {
		t.Fatal("owner cannot see own library", own, err)
	}
	p.PublicProfile = false
	p, err = r.SavePreferences(ctx, "owner", p)
	if err != nil {
		t.Fatal(err)
	}
	if err = r.FollowProfile(ctx, "unverified", "owner", true); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("private follow", err)
	}
	if err = r.FollowProfile(ctx, "viewer", "owner", false); err != nil {
		t.Fatal("private unfollow blocked", err)
	}
	if _, err = r.PublicProfile(ctx, "viewer", "owner"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("private profile exposed to former follower", err)
	}
	// Concurrent forms based on one version must not overwrite one another.
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() { _, e := r.SavePreferences(ctx, "owner", p); results <- e }()
	}
	successes, conflicts := 0, 0
	for i := 0; i < 2; i++ {
		e := <-results
		if e == nil {
			successes++
		} else if errors.Is(e, domain.ErrConflict) {
			conflicts++
		} else {
			t.Fatal(e)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent saves: successes=%d conflicts=%d", successes, conflicts)
	}
	exec(`INSERT INTO titles(id,slug,name_kk,is_listed) SELECT 'cap-'||i,'cap-'||i,'Тайтл '||i,true FROM generate_series(1,42) i; INSERT INTO library_entries(id,user_id,title_id) SELECT 'cap-library-'||i,'owner','cap-'||i FROM generate_series(1,42) i`)
	own, err = r.PublicProfile(ctx, "owner", "owner")
	if err != nil || len(own.Library) != 40 {
		t.Fatal("library cap", len(own.Library), err)
	}
	exec(`UPDATE users SET is_blocked=true WHERE id='owner'`)
	if _, err = r.PublicProfile(ctx, "owner", "owner"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("blocked profile exposed", err)
	}
}
