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
	"qmanga/internal/community/domain"
	platform "qmanga/internal/platform/postgres"
)

func TestNestedRepliesKeepImmediateAddresseeAndFlatThread(t *testing.T) {
	connection := os.Getenv("TEST_DATABASE_URL")
	if connection == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base, err := platform.New(ctx, connection)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	schema := "nested_comments_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = base.Pool().Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer base.Pool().Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	parsed, _ := url.Parse(connection)
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
	_, err = database.Pool().Exec(ctx, `INSERT INTO users(id,email,password_hash,is_verified) VALUES('one','one@test.invalid','hash',true),('two','two@test.invalid','hash',true);
INSERT INTO titles(id,slug,name_kk,is_listed) VALUES('title','title','Title',true),('other','other','Other',true)`)
	if err != nil {
		t.Fatal(err)
	}
	repository := New(database.Pool())
	create := func(user, title, parent, body string) (string, error) {
		id := uuid.NewString()
		_, err := repository.Comment(ctx, user, domain.Comment{ID: id, TitleID: title, ParentID: parent, Body: body})
		return id, err
	}
	root, err := create("one", "title", "", "Root")
	if err != nil {
		t.Fatal(err)
	}
	reply, err := create("two", "title", root, "Reply")
	if err != nil {
		t.Fatal(err)
	}
	nested, err := create("one", "title", reply, "Nested")
	if err != nil {
		t.Fatal(err)
	}
	thread, err := repository.Discussion(ctx, "title", "", "")
	if err != nil || len(thread.Comments) != 3 {
		t.Fatalf("thread: %#v %v", thread, err)
	}
	var rootCount int
	for _, item := range thread.Comments {
		if item.ID == root {
			rootCount = item.Replies
		}
		if item.ID == nested && (item.ParentID != reply || item.RootID != root) {
			t.Fatalf("nested address lost: %#v", item)
		}
	}
	if rootCount != 2 {
		t.Fatalf("flat thread count: %d", rootCount)
	}
	focused, err := repository.Discussion(ctx, "title", "", "", nested)
	if err != nil || len(focused.Comments) != 3 {
		t.Fatalf("focused nested context: %#v %v", focused, err)
	}
	wrong, err := repository.Discussion(ctx, "other", "", "", nested)
	if err != nil || len(wrong.Comments) != 0 {
		t.Fatalf("cross-title focus leaked: %#v %v", wrong, err)
	}
	page, err := repository.Replies(ctx, root, "")
	if err != nil || len(page.Comments) != 2 {
		t.Fatalf("reply page: %#v %v", page, err)
	}
	// A focused reply beyond the first contiguous batch must not advance paging.
	if _, err = database.Pool().Exec(ctx, `INSERT INTO comments(id,user_id,title_id,body,created_at) VALUES('paged-root','one','title','Paging root',now()-interval '1 hour');
 INSERT INTO comments(id,user_id,title_id,parent_id,root_id,body,created_at) SELECT 'paged-reply-'||lpad(n::text,2,'0'),'one','title','paged-root','paged-root','Reply '||n,now()-interval '30 minutes'+n*interval '1 second' FROM generate_series(1,30)n`); err != nil {
		t.Fatal(err)
	}
	focusedPage, err := repository.Discussion(ctx, "title", "", "", "paged-reply-25")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	cursor := ""
	for _, item := range focusedPage.Comments {
		if item.ParentID != "" {
			seen[item.ID] = true
		}
		if item.ID == "paged-root" {
			cursor = item.ReplyCursor
		}
	}
	if cursor != "paged-reply-10" {
		t.Fatalf("focused reply advanced contiguous cursor: %q", cursor)
	}
	for step := 0; step < 3; step++ {
		page, err := repository.Replies(ctx, "paged-root", cursor)
		if err != nil {
			t.Fatal(err)
		}
		for _, item := range page.Comments {
			seen[item.ID] = true
		}
		if page.Next == "" {
			break
		}
		cursor = page.Next
	}
	if len(seen) != 30 {
		t.Fatalf("focused thread lost replies: %d", len(seen))
	}
	if _, err = create("two", "other", nested, "Wrong title"); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("cross-title reply: %v", err)
	}
	if _, err = database.Pool().Exec(ctx, "UPDATE comments SET is_hidden=true WHERE id=$1", reply); err != nil {
		t.Fatal(err)
	}
	if _, err = create("one", "title", reply, "Hidden target"); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("hidden parent reply: %v", err)
	}
}
