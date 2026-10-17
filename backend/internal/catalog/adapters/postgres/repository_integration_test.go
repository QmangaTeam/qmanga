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
	"qmanga/internal/catalog/application"
	"qmanga/internal/catalog/domain"
	platform "qmanga/internal/platform/postgres"
)

func TestCatalogueVisibilityAliasesAndChapterOrder(testingTarget *testing.T) {
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
	execute("INSERT INTO users(id,email,password_hash,role,is_verified) VALUES('admin','admin@example.test','test','admin',true),('reader','reader@example.test','test','reader',true),('foreign','foreign@example.test','test','reader',true)")
	execute("INSERT INTO teams(id,slug,name,is_approved) VALUES('team','dala','Дала',true),('pending','pending','Pending',false)")
	execute("INSERT INTO team_memberships(id,team_id,user_id,role) VALUES('membership','team','reader','editor')")
	repository := New(database.Pool())
	service := application.New(repository, uuid.NewString)
	genre, failure := service.CreateGenre(ctx, "admin", "Шытырман", "adventure")
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	input := domain.Metadata{Slug: "dala-uni", Name: "Дала үні", Kind: "manga", OriginalStatus: "ongoing", TranslationStatus: "ongoing", Direction: "rtl", Aliases: []string{"ӨРКЕН ӘЛЕМ", "100% нақты"}}
	input.GenreIDs = []string{genre.ID}
	title, failure := service.CreateTitle(ctx, "admin", input)
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	if _, failure = service.CreateTitle(ctx, "reader", input); !errors.Is(failure, domain.ErrForbidden) {
		testingTarget.Fatal("reader created title")
	}
	if _, failure = service.CreateProject(ctx, "admin", title.ID, "pending"); !errors.Is(failure, domain.ErrForbidden) {
		testingTarget.Fatalf("unapproved team project accepted: %v", failure)
	}
	projectID, failure := service.CreateProject(ctx, "admin", title.ID, "team")
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	input.Description = "Жаңартылған сипаттама"
	if _, failure = service.UpdateTitle(ctx, "reader", title.ID, input); !errors.Is(failure, domain.ErrForbidden) {
		testingTarget.Fatal("reader edited metadata")
	}
	if _, failure = service.UpdateTitle(ctx, "admin", title.ID, input); failure != nil {
		testingTarget.Fatal(failure)
	}
	privateDetail, failure := service.AdminDetail(ctx, "admin", title.ID)
	if failure != nil || privateDetail.Title.Description != input.Description {
		testingTarget.Fatal("metadata editing failed", failure)
	}
	chapters := make([]string, 0)
	for position, label := range []string{"1", "1.1", "1.10", "2", "Арнайы"} {
		chapterID, failure := service.CreateChapter(ctx, "admin", title.ID, label, "1", "", position+1)
		if failure != nil {
			testingTarget.Fatal(failure)
		}
		chapters = append(chapters, chapterID)
	}
	page, failure := service.List(ctx, application.Filter{})
	if failure != nil || len(page.Items) != 0 {
		testingTarget.Fatalf("draft title leaked: %+v %v", page, failure)
	}
	if _, failure = service.Detail(ctx, title.Slug); !errors.Is(failure, domain.ErrNotFound) {
		testingTarget.Fatal("draft detail exposed")
	}
	for position, chapterID := range chapters {
		releaseID, revisionID := uuid.NewString(), uuid.NewString()
		execute("INSERT INTO releases(id,chapter_id,project_id,team_id,status,current_revision_id) VALUES($1,$2,$3,'team','published',$4)", releaseID, chapterID, projectID, revisionID)
		state := "published"
		if position == 4 {
			state = "submitted"
		}
		execute("INSERT INTO revisions(id,release_id,status) VALUES($1,$2,$3)", revisionID, releaseID, state)
	}
	page, failure = service.List(ctx, application.Filter{Query: "өркен әлем"})
	if failure != nil || len(page.Items) != 1 || page.Items[0].ChapterCount != 4 {
		testingTarget.Fatalf("alias/publication filtering failed: %+v %v", page, failure)
	}
	for _, filter := range []application.Filter{{Genre: "missing"}, {Kind: "manhwa"}, {OriginalStatus: "completed"}, {TranslationStatus: "stopped"}} {
		filtered, failure := service.List(ctx, filter)
		if failure != nil || len(filtered.Items) != 0 {
			testingTarget.Fatalf("filter ignored: %+v %v", filter, failure)
		}
	}
	filtered, failure := service.List(ctx, application.Filter{Genre: "adventure", OriginalStatus: "ongoing", TranslationStatus: "ongoing", Kind: "manga"})
	if failure != nil || len(filtered.Items) != 1 {
		testingTarget.Fatal("combined filters failed", failure)
	}
	detail, failure := service.Detail(ctx, title.Slug)
	if failure != nil || len(detail.Projects) != 1 || detail.Title.Direction != "rtl" || len(detail.Genres) != 1 {
		testingTarget.Fatalf("public detail projects: %+v %v", detail, failure)
	}
	team, failure := service.Team(ctx, "dala")
	if failure != nil || len(team.Projects) != 1 {
		testingTarget.Fatalf("public team projects: %+v %v", team, failure)
	}
	if _, failure = service.Team(ctx, "pending"); !errors.Is(failure, domain.ErrNotFound) {
		testingTarget.Fatal("unapproved team exposed")
	}
	page, failure = service.List(ctx, application.Filter{Query: "%"})
	if failure != nil || len(page.Items) != 1 {
		testingTarget.Fatal("literal percent search failed", failure)
	}
	page, failure = service.List(ctx, application.Filter{Query: "_"})
	if failure != nil || len(page.Items) != 0 {
		testingTarget.Fatal("wildcard search was not escaped", failure)
	}
	chapterPage, failure := service.Chapters(ctx, title.Slug, 0, 2)
	if failure != nil || len(chapterPage.Items) != 2 || chapterPage.NextOrder == nil || chapterPage.Items[1].Label != "1.1" {
		testingTarget.Fatalf("chapter cursor: %+v %v", chapterPage, failure)
	}
	chapterPage, failure = service.Chapters(ctx, title.Slug, *chapterPage.NextOrder, 2)
	if failure != nil || len(chapterPage.Items) != 2 || chapterPage.Items[0].Label != "1.10" || chapterPage.NextOrder != nil {
		testingTarget.Fatalf("fractional labels reordered or private revision exposed: %+v %v", chapterPage, failure)
	}
	for _, actorID := range []string{"reader", "foreign"} {
		projects, failure := service.StudioProjects(ctx, actorID)
		expected := 1
		if actorID == "foreign" {
			expected = 0
		}
		if failure != nil || len(projects) != expected {
			testingTarget.Fatal("studio project isolation failed", failure)
		}
	}
	input.Slug = "ai-zholy"
	input.Name = "Ай жолы"
	second, failure := service.CreateTitle(ctx, "admin", input)
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	secondProject, failure := service.CreateProject(ctx, "admin", second.ID, "team")
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	secondChapter, failure := service.CreateChapter(ctx, "admin", second.ID, "1", "", "", 1)
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	execute("INSERT INTO releases(id,chapter_id,project_id,team_id,status,current_revision_id) VALUES('second-release',$1,$2,'team','published','second-revision')", secondChapter, secondProject)
	execute("INSERT INTO revisions(id,release_id,status) VALUES('second-revision','second-release','published')")
	execute("UPDATE titles SET updated_at='2026-10-17T00:00:00Z'")
	for _, sort := range []string{"updated", "name"} {
		first, failure := service.List(ctx, application.Filter{Limit: 1, Sort: sort})
		if failure != nil || len(first.Items) != 1 || first.Next == nil {
			testingTarget.Fatalf("first keyset page: %+v %v", first, failure)
		}
		last, failure := service.List(ctx, application.Filter{Limit: 1, Sort: sort, After: first.Next})
		if failure != nil || len(last.Items) != 1 || last.Next != nil || last.Items[0].ID == first.Items[0].ID {
			testingTarget.Fatalf("duplicate/skipped keyset tie: %+v %v", last, failure)
		}
	}
	execute("UPDATE users SET role='reader' WHERE id='admin'")
	if failure = repository.CreateTitle(ctx, "admin", title); !errors.Is(failure, domain.ErrForbidden) {
		testingTarget.Fatal("adapter allowed revoked administrator")
	}
}
