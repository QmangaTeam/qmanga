package postgres

import (
	"context"
	"github.com/google/uuid"
	"net/url"
	"os"
	"qmanga/internal/notifications/application"
	platform "qmanga/internal/platform/postgres"
	"strings"
	"testing"
	"time"
)

func TestUnreadCountIsolationAndRead(t *testing.T) {
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
	schema := "unread_" + strings.ReplaceAll(uuid.NewString(), "-", "")
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
	_, err = database.Pool().Exec(ctx, `INSERT INTO users(id,email,password_hash)VALUES('one','one@test.local','x'),('two','two@test.local','x'); INSERT INTO notifications(id,user_id,type,title_kk,body_kk) SELECT 'n'||n,'one','comment_reply','title','body' FROM generate_series(1,105)n; INSERT INTO notifications(id,user_id,type,title_kk,body_kk,is_read) VALUES('other','two','comment_reply','title','body',false),('read','one','comment_reply','title','body',true);`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = database.Pool().Exec(ctx, `UPDATE notifications SET dedupe_key='reply:legacy-comment' WHERE id='n1'; UPDATE notifications SET data=jsonb_build_object('comment_id','new-comment') WHERE id='n2'`); err != nil {
		t.Fatal(err)
	}
	repository := New(database.Pool())
	allUnread := []application.Notification{}
	after := ""
	for {
		items, failure := repository.List(ctx, "one", after, true)
		if failure != nil {
			t.Fatal(failure)
		}
		for _, item := range items {
			if item.Read {
				t.Fatal("read item in unread query")
			}
		}
		allUnread = append(allUnread, items...)
		if len(items) < 50 {
			break
		}
		after = items[len(items)-1].ID
	}
	comments := map[string]string{}
	for _, item := range allUnread {
		comments[item.ID] = item.CommentID
	}
	if comments["n1"] != "legacy-comment" || comments["n2"] != "new-comment" {
		t.Fatalf("comment destinations: %v", comments)
	}
	if len(allUnread) != 105 {
		t.Fatalf("unread server pagination=%d", len(allUnread))
	}

	count, err := repository.Unread(ctx, "one")
	if err != nil || count != 105 {
		t.Fatalf("real count=%d err=%v", count, err)
	}
	if err = repository.Read(ctx, "two", "n1"); err != nil {
		t.Fatal(err)
	}
	count, err = repository.Unread(ctx, "one")
	if err != nil || count != 105 {
		t.Fatal("foreign read changed owner count")
	}
	if err = repository.Read(ctx, "one", "n1"); err != nil {
		t.Fatal(err)
	}
	if err = repository.Read(ctx, "one", "n1"); err != nil {
		t.Fatal(err)
	}
	count, err = repository.Unread(ctx, "one")
	if err != nil || count != 104 {
		t.Fatalf("idempotent read count=%d %v", count, err)
	}
	count, err = repository.Unread(ctx, "two")
	if err != nil || count != 1 {
		t.Fatalf("other account count=%d %v", count, err)
	}
}
