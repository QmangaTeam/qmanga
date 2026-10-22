package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"net/url"
	"os"
	"qmanga/internal/notifications/application"
	platform "qmanga/internal/platform/postgres"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestModerationTelegramQueue(t *testing.T) {
	address := os.Getenv("TEST_DATABASE_URL")
	if address == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base, err := platform.New(ctx, address)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	schema := "telegram_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = base.Pool().Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer base.Pool().Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	parsed, _ := url.Parse(address)
	params := parsed.Query()
	params.Set("search_path", schema)
	parsed.RawQuery = params.Encode()
	database, err := platform.New(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err = database.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	exec := func(query string) {
		t.Helper()
		if _, err := database.Pool().Exec(ctx, query); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO users(id,email,password_hash) VALUES('owner','owner@test.local','x');
 INSERT INTO teams(id,slug,name,is_approved,review_status) VALUES('approved','approved','Approved',true,'approved'),('pending','pending','Pending',false,'pending');
 INSERT INTO titles(id,slug,name_kk) VALUES('title','title','Қазақша тайтл');
 INSERT INTO projects(id,title_id,team_id) VALUES('project','title','approved');
 INSERT INTO chapters(id,title_id,label,order_num) VALUES('chapter','title','1',1);
 INSERT INTO releases(id,chapter_id,project_id,team_id) VALUES('release','chapter','project','approved');
 INSERT INTO revisions(id,release_id,status) VALUES('revision','release','ready');
 UPDATE revisions SET status='submitted' WHERE id='revision';
 UPDATE revisions SET status='submitted' WHERE id='revision';
 INSERT INTO project_requests(id,team_id,requested_by,target_key,note,name_kk) VALUES('request','approved','owner','new-title','Request details','New title');
 INSERT INTO title_change_requests(id,title_id,team_id,requested_by,before_data,proposed_data,base_version,reason,cover_action,banner_action) VALUES('change','title','approved','owner','{}','{}',repeat('a',64),'Change title details','keep','keep');
 INSERT INTO reports(id,user_id,title_id,category,reason,target_type,target_id) VALUES('report','owner','title','other','Report detail','title','title');`)
	count := func() int {
		t.Helper()
		var n int
		if err := database.Pool().QueryRow(ctx, "SELECT count(*) FROM moderation_telegram_queue").Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if n := count(); n != 5 {
		t.Fatalf("expected five moderation sources, got %d", n)
	}
	tx, err := database.Pool().Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO teams(id,slug,name) VALUES('rolled-back','rolled-back','Rollback')`); err != nil {
		t.Fatal(err)
	}
	tx.Rollback(ctx)
	if count() != 5 {
		t.Fatal("rollback leaked a notification")
	}
	exec(`UPDATE revisions SET status='changes_requested' WHERE id='revision';UPDATE revisions SET status='submitted' WHERE id='revision';UPDATE teams SET name='Pending edited' WHERE id='pending';`)
	if count() != 7 {
		t.Fatal("resubmission was lost")
	}
	r := New(database.Pool())
	sender := &queueTestSender{fail: true}
	if worked, err := r.DeliverModeration(ctx, sender); !worked || err == nil {
		t.Fatal("failure not returned")
	}
	var failed, delivered int
	if err := database.Pool().QueryRow(ctx, `SELECT count(*) FILTER(WHERE attempts=1 AND delivered_at IS NULL AND available_at>now()),count(*) FILTER(WHERE delivered_at IS NOT NULL) FROM moderation_telegram_queue`).Scan(&failed, &delivered); err != nil || failed != 1 || delivered != 0 {
		t.Fatal("failed message was not durably deferred", err)
	}
	exec(`UPDATE moderation_telegram_queue SET available_at=now()`)
	sender.fail = false
	for i := 0; i < 7; i++ {
		if worked, err := r.DeliverModeration(ctx, sender); !worked || err != nil {
			t.Fatal("delivery failed", err)
		}
	}
	if worked, err := r.DeliverModeration(ctx, sender); worked || err != nil {
		t.Fatal("delivered events replayed", err)
	}
	if sender.sent.Load() != 7 {
		t.Fatal("wrong delivery count")
	}
}

type queueTestSender struct {
	fail bool
	sent atomic.Int32
}

func (s *queueTestSender) Send(_ context.Context, m application.ModerationMessage) error {
	if s.fail {
		return errors.New("unavailable")
	}
	if m.Label == "" || m.Path == "" {
		return errors.New("missing submission context")
	}
	s.sent.Add(1)
	return nil
}
