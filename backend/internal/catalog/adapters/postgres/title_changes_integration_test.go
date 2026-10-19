package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"net/url"
	"os"
	"qmanga/internal/catalog/domain"
	platform "qmanga/internal/platform/postgres"
	"strings"
	"testing"
	"time"
)

func TestTitleChangeApprovalAndConflict(testingTarget *testing.T) {
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
	schema := "catalogue_" + strings.ReplaceAll(uuid.NewString(), "-", "")
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
	execute := func(query string, arguments ...any) {
		testingTarget.Helper()
		if _, failure := database.Pool().Exec(ctx, query, arguments...); failure != nil {
			testingTarget.Fatal(failure)
		}
	}

	execute("INSERT INTO users(id,email,password_hash,role,is_verified) VALUES('admin','admin@changes.test','test','admin',true),('editor','editor@changes.test','test','reader',true),('foreign','foreign@changes.test','test','reader',true),('mod','mod@changes.test','test','moderator',true)")
	execute("INSERT INTO teams(id,slug,name,is_approved) VALUES('team','team','Team',true)")
	execute("INSERT INTO team_memberships(id,team_id,user_id,role) VALUES('membership','team','editor','editor')")
	execute("INSERT INTO titles(id,slug,name_kk,type,direction) VALUES('title','title','Original','manga','ltr')")
	execute("INSERT INTO projects(id,title_id,team_id) VALUES('project','title','team')")
	repo := New(database.Pool())
	current, err := repo.ChangeContext(ctx, "editor", "project")
	if err != nil {
		testingTarget.Fatal(err)
	}
	if _, err = repo.ChangeContext(ctx, "foreign", "project"); !errors.Is(err, domain.ErrForbidden) {
		testingTarget.Fatal("foreign context", err)
	}
	data := current.Snapshot.Data
	data.Name = "Changed"
	input := domain.TitleChangeInput{BaseVersion: current.Snapshot.Version(), Data: data, Reason: "Official source", Links: []string{"https://example.org/source"}, CoverAction: "keep", BannerAction: "keep"}
	request, err := repo.CreateChange(ctx, "editor", "project", "request", input, nil)
	if err != nil {
		testingTarget.Fatal(err)
	}
	if request.Status != "pending" {
		testingTarget.Fatal(request.Status)
	}
	var name string
	database.Pool().QueryRow(ctx, "SELECT name_kk FROM titles WHERE id='title'").Scan(&name)
	if name != "Original" {
		testingTarget.Fatal("changed before moderation")
	}
	if err = repo.ReviewChange(ctx, "editor", "request", "approve", ""); !errors.Is(err, domain.ErrForbidden) {
		testingTarget.Fatal("editor reviewed", err)
	}
	if err = repo.ReviewChange(ctx, "mod", "request", "approve", ""); err != nil {
		testingTarget.Fatal(err)
	}
	database.Pool().QueryRow(ctx, "SELECT name_kk FROM titles WHERE id='title'").Scan(&name)
	if name != "Changed" {
		testingTarget.Fatal("not applied", name)
	}
	current, err = repo.ChangeContext(ctx, "editor", "project")
	if err != nil {
		testingTarget.Fatal(err)
	}
	input.BaseVersion = current.Snapshot.Version()
	input.Data = current.Snapshot.Data
	input.Data.Name = "Stale change"
	if _, err = repo.CreateChange(ctx, "editor", "project", "stale", input, nil); err != nil {
		testingTarget.Fatal(err)
	}
	execute("INSERT INTO categories(id,name_kk,slug) VALUES('category','Category','category')")
	if err = repo.SetCategories(ctx, "admin", "title", []string{"category"}); err != nil {
		testingTarget.Fatal(err)
	}
	if err = repo.ReviewChange(ctx, "mod", "stale", "approve", ""); !errors.Is(err, domain.ErrConflict) {
		testingTarget.Fatal("stale overwrite", err)
	}
	if err = repo.ReviewChange(ctx, "mod", "stale", "reject", "Data changed, submit again"); err != nil {
		testingTarget.Fatal(err)
	}
	var count int
	database.Pool().QueryRow(ctx, "SELECT count(*) FROM notifications WHERE user_id='editor'").Scan(&count)
	if count != 2 {
		testingTarget.Fatal("missing decisions", count)
	}
}
