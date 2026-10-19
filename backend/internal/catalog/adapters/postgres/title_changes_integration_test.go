package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"net/url"
	"os"
	"qmanga/internal/catalog/application"
	"qmanga/internal/catalog/domain"
	platform "qmanga/internal/platform/postgres"
	media "qmanga/internal/publishing/adapters/postgres"
	mediaDomain "qmanga/internal/publishing/domain"
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
	execute("INSERT INTO genres(id,name_kk,slug) VALUES('genre','Genre','genre')")
	execute("INSERT INTO categories(id,name_kk,slug) VALUES('initial-category','Initial category','initial-category')")
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
	data.GenreIDs = []string{"genre"}
	data.CategoryIDs = []string{"initial-category"}
	input := domain.TitleChangeInput{BaseVersion: current.Snapshot.Version(), Data: data, Reason: "Official source", Links: []string{"https://example.org/source"}, CoverAction: "keep", BannerAction: "keep"}
	input.CoverAction = "replace"
	input.BannerAction = "replace"
	assets := []application.TitleChangeAsset{}
	for position, purpose := range []string{"cover", "banner", "evidence"} {
		assets = append(assets, application.TitleChangeAsset{ID: purpose + "-asset", Purpose: purpose, ObjectKey: "title-changes/request/" + purpose + ".webp", Mime: "image/webp", Checksum: strings.Repeat("a", 64), Width: 100, Height: 100, Bytes: 100, Position: position + 1})
	}
	request, err := repo.CreateChange(ctx, "editor", "project", "request", input, assets)
	if err != nil {
		testingTarget.Fatal(err)
	}
	mediaRepo := media.New(database.Pool())
	if _, err = mediaRepo.Asset(ctx, "", "cover-asset"); !errors.Is(err, mediaDomain.ErrNotFound) {
		testingTarget.Fatal("pending cover leaked", err)
	}
	if _, err = mediaRepo.Asset(ctx, "foreign", "evidence-asset"); !errors.Is(err, mediaDomain.ErrNotFound) {
		testingTarget.Fatal("evidence leaked", err)
	}
	if asset, err := mediaRepo.Asset(ctx, "editor", "evidence-asset"); err != nil || asset.Public {
		testingTarget.Fatal("editor evidence", err)
	}
	if request.Status != "pending" {
		testingTarget.Fatal(request.Status)
	}
	var legacy int
	database.Pool().QueryRow(ctx, "SELECT count(*) FROM title_cover_assets WHERE id='cover-asset'").Scan(&legacy)
	if legacy != 0 {
		testingTarget.Fatal("pending artwork entered legacy public cover table")
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
	if err = database.Pool().QueryRow(ctx, "SELECT count(*) FROM title_cover_assets WHERE id='cover-asset' AND title_id='title' AND object_key='title-changes/request/cover.webp'").Scan(&legacy); err != nil || legacy != 1 {
		testingTarget.Fatal("approved cover unavailable to previous API during rollback", legacy, err)
	}
	var genres, categories int
	if err = database.Pool().QueryRow(ctx, "SELECT (SELECT count(*) FROM title_genres WHERE title_id='title' AND genre_id='genre'),(SELECT count(*) FROM title_categories WHERE title_id='title' AND category_id='initial-category')").Scan(&genres, &categories); err != nil || genres != 1 || categories != 1 {
		testingTarget.Fatal("approved taxonomy not applied", genres, categories, err)
	}
	if asset, err := mediaRepo.Asset(ctx, "", "cover-asset"); err != nil || !asset.Public {
		testingTarget.Fatal("approved cover not public", err)
	}
	if asset, err := mediaRepo.Asset(ctx, "", "banner-asset"); err != nil || !asset.Public {
		testingTarget.Fatal("approved banner not public", err)
	}
	if _, err = mediaRepo.Asset(ctx, "", "evidence-asset"); !errors.Is(err, mediaDomain.ErrNotFound) {
		testingTarget.Fatal("approved evidence leaked", err)
	}
	input.CoverAction = "keep"
	input.BannerAction = "keep"
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
