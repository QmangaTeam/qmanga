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
	genre, failure := service.CreateGenre(ctx, "admin", "Шытырман", "adventure-tales")
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	input := domain.Metadata{Slug: "dala-uni", Name: "Дала үні", Kind: "manga", OriginalStatus: "ongoing", TranslationStatus: "ongoing", Direction: "rtl", Aliases: []string{"ӨРКЕН ӘЛЕМ", "100% нақты"}}
	input.GenreIDs = []string{genre.ID}
	input.Publication = domain.Publication{Year: 2023, Publishers: []string{"Publisher"}, Country: "Корея", Source: "Новелла", Formats: []string{"Webtoon"}, Links: []domain.ExternalLink{{Label: "Official", URL: "https://example.com/title"}}}
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
	testingTarget.Run("administrative project is read only and does not grant studio membership", func(testingTarget *testing.T) {
		execute("INSERT INTO users(id,email,password_hash,role,is_verified,is_blocked) VALUES('owner','owner@example.test','test','reader',true,false),('member','member@example.test','test','reader',true,false),('moderator','moderator@example.test','test','moderator',true,false),('blocked-admin','blocked-admin@example.test','test','admin',true,true)")
		execute("INSERT INTO team_memberships(id,team_id,user_id,role) VALUES('owner-membership','team','owner','owner'),('member-membership','team','member','member')")
		defer execute("DELETE FROM users WHERE id IN ('owner','member','moderator','blocked-admin')")
		defer execute("DELETE FROM team_memberships WHERE id IN ('owner-membership','member-membership')")
		item, failure := service.AdminProject(ctx, "admin", projectID)
		if failure != nil || item.ID != projectID || item.TitleName != "Дала үні" || item.TeamName != "Дала" || item.Status != "active" {
			testingTarget.Fatalf("admin project context: %+v %v", item, failure)
		}
		for _, actor := range []string{"reader", "owner", "member", "moderator", "blocked-admin", "foreign"} {
			if _, failure := service.AdminProject(ctx, actor, projectID); !errors.Is(failure, domain.ErrForbidden) {
				testingTarget.Fatalf("%s administrative access: %v", actor, failure)
			}
		}
		for _, actor := range []string{"admin", "moderator", "blocked-admin", "foreign"} {
			if _, failure := service.StudioProject(ctx, actor, projectID); !errors.Is(failure, domain.ErrForbidden) {
				testingTarget.Fatalf("%s acquired studio membership: %v", actor, failure)
			}
		}
		for actor, role := range map[string]string{"owner": "owner", "reader": "editor", "member": "member"} {
			item, failure := service.StudioProject(ctx, actor, projectID)
			if failure != nil || item.Role != role {
				testingTarget.Fatalf("%s studio role changed: %+v %v", actor, item, failure)
			}
		}
		if _, failure := service.CreateStudioChapter(ctx, "member", projectID, "member-forbidden", "", "", 0); !errors.Is(failure, domain.ErrForbidden) {
			testingTarget.Fatalf("member gained editor permission: %v", failure)
		}
		if _, failure := service.AdminProject(ctx, "admin", "missing"); !errors.Is(failure, domain.ErrNotFound) {
			testingTarget.Fatalf("missing project: %v", failure)
		}
	})
	input.Description = "Жаңартылған сипаттама"
	input.Authors = []string{"Қаламгер"}
	input.Featured = true
	if _, failure = service.UpdateTitle(ctx, "reader", title.ID, input); !errors.Is(failure, domain.ErrForbidden) {
		testingTarget.Fatal("reader edited metadata")
	}
	if _, failure = service.UpdateTitle(ctx, "admin", title.ID, input); failure != nil {
		testingTarget.Fatal(failure)
	}
	privateDetail, failure := service.AdminDetail(ctx, "admin", title.ID)
	if failure == nil && (privateDetail.Title.Publication.Year != 2023 || len(privateDetail.Title.Publication.Links) != 1 || privateDetail.ChapterCount != 0) {
		testingTarget.Fatal("publication metadata round trip", privateDetail)
	}
	if failure != nil || privateDetail.Title.Description != input.Description || len(privateDetail.Title.Authors) != 1 || !privateDetail.Title.Featured {
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
	execute("UPDATE titles SET cover_asset_id='cover-test' WHERE id=$1", title.ID)
	publicDetail, detailFailure := service.Detail(ctx, title.Slug)
	if detailFailure != nil || publicDetail.ChapterCount != 4 || publicDetail.Title.Publication.Country != "Корея" {
		testingTarget.Fatal("public facts must count visible chapters and preserve metadata", publicDetail, detailFailure)
	}
	updates, failure := service.Updates(ctx, "", "", 2)
	if failure != nil || len(updates.Items) != 2 || updates.Next == "" {
		testingTarget.Fatal("release feed missing", updates, failure)
	}
	if updates.Items[0].CoverID != "cover-test" {
		testingTarget.Fatal("update thumbnail did not follow the title cover", updates.Items[0])
	}
	remaining, failure := service.Updates(ctx, updates.Next, "", 2)
	if failure != nil || len(remaining.Items) != 2 || remaining.Next != "" || remaining.Items[0].ReleaseID == updates.Items[0].ReleaseID {
		testingTarget.Fatal("feed cursor/private revision", remaining, failure)
	}
	picks, failure := service.List(ctx, application.Filter{Featured: true, Sort: "new"})
	if failure != nil || len(picks.Items) != 1 {
		testingTarget.Fatal("editorial selection/new sort", picks, failure)
	}
	page, failure = service.List(ctx, application.Filter{Query: "өркен әлем"})
	if failure != nil || len(page.Items) != 1 || page.Items[0].ChapterCount != 4 {
		testingTarget.Fatalf("alias/publication filtering failed: %+v %v", page, failure)
	}
	for _, filter := range []application.Filter{{Genres: []string{"missing"}}, {Kinds: []string{"manhwa"}}, {OriginalStatuses: []string{"completed"}}, {TranslationStatuses: []string{"stopped"}}} {
		filtered, failure := service.List(ctx, filter)
		if failure != nil || len(filtered.Items) != 0 {
			testingTarget.Fatalf("filter ignored: %+v %v", filter, failure)
		}
	}
	filtered, failure := service.List(ctx, application.Filter{Genres: []string{"adventure-tales"}, OriginalStatuses: []string{"ongoing"}, TranslationStatuses: []string{"ongoing"}, Kinds: []string{"manga"}})
	if failure != nil || len(filtered.Items) != 1 {
		testingTarget.Fatal("combined filters failed", failure)
	}
	for _, facet := range []struct {
		name   string
		filter application.Filter
		count  int
	}{
		{"chapters exact", application.Filter{Query: "өркен әлем", ChaptersMin: 4, ChaptersMax: 4}, 1},
		{"chapters above", application.Filter{Query: "өркен әлем", ChaptersMin: 5}, 0},
		{"chapters below", application.Filter{Query: "өркен әлем", ChaptersMax: 3}, 0},
		{"year exact", application.Filter{Query: "өркен әлем", YearMin: 2023, YearMax: 2023}, 1},
		{"year mismatch", application.Filter{Query: "өркен әлем", YearMax: 2022}, 0},
		{"format", application.Filter{Query: "өркен әлем", Formats: []string{"Webtoon"}}, 1},
		{"source", application.Filter{Query: "өркен әлем", Sources: []string{"Новелла"}}, 1},
		{"country", application.Filter{Query: "өркен әлем", Countries: []string{"Корея"}}, 1},
	} {
		page, failure := service.List(ctx, facet.filter)
		if failure != nil || len(page.Items) != facet.count {
			testingTarget.Fatalf("facet %s failed: %+v %v", facet.name, page, failure)
		}
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
	searchPage, failure := service.Chapters(ctx, title.Slug, 0, 1, application.ChapterFilter{Query: "1.1", Sort: "desc"})
	if failure != nil || len(searchPage.Items) != 1 || searchPage.Items[0].Label != "1.10" || searchPage.NextOrder == nil {
		testingTarget.Fatalf("descending chapter search failed: %+v %v", searchPage, failure)
	}
	searchPage, failure = service.Chapters(ctx, title.Slug, *searchPage.NextOrder, 1, application.ChapterFilter{Query: "1.1", Sort: "desc"})
	if failure != nil || len(searchPage.Items) != 1 || searchPage.Items[0].Label != "1.1" || searchPage.NextOrder != nil {
		testingTarget.Fatalf("descending chapter cursor failed: %+v %v", searchPage, failure)
	}
	for _, query := range []string{"%", "_", "Арнайы"} {
		searchPage, failure = service.Chapters(ctx, title.Slug, 0, 50, application.ChapterFilter{Query: query})
		if failure != nil || len(searchPage.Items) != 0 {
			testingTarget.Fatalf("wildcard or private chapter exposed: %q %+v %v", query, searchPage, failure)
		}
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
	execute("UPDATE users SET role='admin' WHERE id='admin'")
	announcement := input
	announcement.Slug = "announced-title"
	announcement.Name = "Announcement"
	announcement.Aliases = []string{}
	announcement.Listed = true
	if _, failure = service.CreateTitle(ctx, "admin", announcement); failure != nil {
		testingTarget.Fatal(failure)
	}
	listed, failure := service.List(ctx, application.Filter{Query: "announcement"})
	if failure != nil || len(listed.Items) != 1 {
		testingTarget.Fatal("explicit announcement missing", failure)
	}
	publishedOnly, failure := service.List(ctx, application.Filter{Query: "announcement", HasChapters: true})
	if failure != nil || len(publishedOnly.Items) != 0 {
		testingTarget.Fatal("unpublished title in published-only filter", failure)
	}
	execute("INSERT INTO titles(id,slug,name_kk,type) SELECT 'extra-'||number,'extra-'||number,'Extra '||number,'manga' FROM generate_series(1,105)number")
	searched, searchErr := repository.SearchAdminTitles(ctx, "admin", "Extra 105")
	if searchErr != nil || len(searched) != 1 || searched[0].ID != "extra-105" {
		testingTarget.Fatal("full admin search", searched, searchErr)
	}
	firstPage, failure := repository.AdminTitles(ctx, "admin")
	if failure != nil || len(firstPage) != 100 {
		testingTarget.Fatal("first admin page", len(firstPage), failure)
	}
	nextPage, failure := repository.AdminTitles(ctx, "admin", 100)
	if failure != nil || len(nextPage) == 0 {
		testingTarget.Fatal("older titles inaccessible", failure)
	}
	for _, first := range firstPage {
		for _, next := range nextPage {
			if first.ID == next.ID {
				testingTarget.Fatal("pagination overlaps")
			}
		}
	}

	testingTarget.Run("translator requests approval and creates chapters without catalogue administration", func(testingTarget *testing.T) {
		request, failure := service.RequestProject(ctx, "reader", domain.ProjectRequestInput{TeamID: "team", Name: "Жаңа аударма", Kind: "manhwa", SourceURL: "https://example.test/story", Note: "Команда жаңа аударма дайындайды"})
		if failure != nil {
			testingTarget.Fatal("translator cannot request project", failure)
		}
		if _, failure = service.RequestProject(ctx, "foreign", domain.ProjectRequestInput{TeamID: "team", TitleID: title.ID, Note: "Бөтен командаға өтінім"}); !errors.Is(failure, domain.ErrForbidden) {
			testingTarget.Fatal("foreign request accepted", failure)
		}
		own, failure := service.ProjectRequests(ctx, "reader", false, "")
		if failure != nil || len(own) != 1 {
			testingTarget.Fatal("request history missing", failure)
		}
		hidden, failure := service.ProjectRequests(ctx, "foreign", false, "")
		if failure != nil || len(hidden) != 0 {
			testingTarget.Fatal("request history leaks", failure)
		}
		if _, failure = service.ProjectRequests(ctx, "reader", true, ""); !errors.Is(failure, domain.ErrForbidden) {
			testingTarget.Fatal("reader accessed admin queue", failure)
		}
		if _, failure = service.ReviewProjectRequest(ctx, "reader", request.ID, "approve", "", ""); !errors.Is(failure, domain.ErrForbidden) {
			testingTarget.Fatal("self approval allowed", failure)
		}
		filtered, filterFailure := service.ProjectRequests(ctx, "admin", true, "", application.ProjectRequestFilter{Team: "team", Kind: "new"})
		if filterFailure != nil || len(filtered) != 1 || filtered[0].ID != request.ID {
			testingTarget.Fatal("server request filter lost matching row", filterFailure)
		}
		filtered, filterFailure = service.ProjectRequests(ctx, "admin", true, "", application.ProjectRequestFilter{Kind: "existing"})
		if filterFailure != nil || len(filtered) != 0 {
			testingTarget.Fatal("request type filter is not server-wide", filterFailure)
		}
		filtered, filterFailure = service.ProjectRequests(ctx, "admin", true, "", application.ProjectRequestFilter{Team: "not-matching"})
		if filterFailure != nil || len(filtered) != 0 {
			testingTarget.Fatal("request team filter ignored", filterFailure)
		}
		accepted, failure := service.ReviewProjectRequest(ctx, "admin", request.ID, "approve", "", "")
		if failure != nil || accepted.ProjectID == "" || accepted.Status != "approved" {
			testingTarget.Fatal("approval did not create project", accepted, failure)
		}
		repeated, failure := service.ReviewProjectRequest(ctx, "admin", request.ID, "approve", "", "")
		if failure != nil || repeated.ProjectID != accepted.ProjectID {
			testingTarget.Fatal("approval replay duplicated project", failure)
		}
		chapter, failure := service.CreateStudioChapter(ctx, "reader", accepted.ProjectID, "1", "1", "Басталуы", 0)
		if failure != nil {
			testingTarget.Fatal("editor cannot create chapter", failure)
		}
		same, failure := service.CreateStudioChapter(ctx, "reader", accepted.ProjectID, "1", "1", "Басталуы", 0)
		if failure != nil || same != chapter {
			testingTarget.Fatal("chapter retry duplicated logical chapter", failure)
		}
		if _, failure = service.CreateStudioChapter(ctx, "reader", accepted.ProjectID, "1", "1", "Басталуы", 8); !errors.Is(failure, domain.ErrConflict) {
			testingTarget.Fatal("existing chapter silently discarded explicit order", failure)
		}
		if _, failure = service.CreateStudioChapter(ctx, "reader", accepted.ProjectID, "2", "", "", 0); failure != nil {
			testingTarget.Fatal(failure)
		}
		if _, failure = service.CreateStudioChapter(ctx, "reader", accepted.ProjectID, "2", "", "New name", 0); !errors.Is(failure, domain.ErrConflict) {
			testingTarget.Fatal("existing chapter silently discarded new name", failure)
		}
		if _, failure = service.CreateStudioChapter(ctx, "foreign", accepted.ProjectID, "2", "", "", 0); !errors.Is(failure, domain.ErrForbidden) {
			testingTarget.Fatal("foreign chapter created", failure)
		}
		execute("UPDATE team_memberships SET role='member' WHERE user_id='reader'")
		if _, failure = service.CreateStudioChapter(ctx, "reader", accepted.ProjectID, "2", "", "", 0); !errors.Is(failure, domain.ErrForbidden) {
			testingTarget.Fatal("regular member created shared chapter", failure)
		}
		execute("UPDATE team_memberships SET role='editor' WHERE user_id='reader'")
		proposed, failure := service.RequestProject(ctx, "reader", domain.ProjectRequestInput{TeamID: "team", TitleID: listed.Items[0].ID, Note: "Жарияланған анонсты аударамыз"})
		if failure != nil {
			testingTarget.Fatal(failure)
		}
		if _, failure = service.RequestProject(ctx, "reader", proposed.ProjectRequestInput); !errors.Is(failure, domain.ErrConflict) {
			testingTarget.Fatal("duplicate pending request", failure)
		}
		rejected, failure := service.ReviewProjectRequest(ctx, "admin", proposed.ID, "reject", "Дереккөзді нақтылау қажет", "")
		if failure != nil || rejected.Resolution == "" {
			testingTarget.Fatal("rejection reason missing", failure)
		}
		retry, failure := service.RequestProject(ctx, "reader", proposed.ProjectRequestInput)
		if failure != nil {
			testingTarget.Fatal("cannot request after rejection", failure)
		}
		if failure = service.CancelProjectRequest(ctx, "foreign", retry.ID); !errors.Is(failure, domain.ErrForbidden) {
			testingTarget.Fatal("foreign cancellation allowed", failure)
		}
		if failure = service.CancelProjectRequest(ctx, "reader", retry.ID); failure != nil {
			testingTarget.Fatal("cannot cancel own request", failure)
		}
	})
	testingTarget.Run("project filtering precedes paging and counts the complete membership scope", func(testingTarget *testing.T) {
		execute(`INSERT INTO teams(id,slug,name,is_approved) VALUES('page-team','page-team','Басқа команда',true);
		 INSERT INTO team_memberships(id,team_id,user_id,role) VALUES('page-member','page-team','reader','editor');
		 INSERT INTO titles(id,slug,name_kk) SELECT 'project-title-'||number,'project-title-'||number,'Paging '||lpad(number::text,3,'0') FROM generate_series(1,110) number;
		 INSERT INTO projects(id,title_id,team_id) SELECT 'page-project-'||number,'project-title-'||number,CASE WHEN number=110 THEN 'page-team' ELSE 'team' END FROM generate_series(1,110) number`)
		page, failure := service.StudioProjectPage(ctx, "reader", application.ProjectFilter{Query: "Paging"})
		if failure != nil || page.Total != 110 || len(page.Items) != 100 || page.Next == "" {
			testingTarget.Fatalf("complete projects: %+v %v", page, failure)
		}
		filtered, failure := service.StudioProjectPage(ctx, "reader", application.ProjectFilter{Query: "Paging", TeamID: "page-team"})
		if failure != nil || filtered.Total != 1 || len(filtered.Items) != 1 || filtered.Items[0].ID != "page-project-110" || filtered.Next != "" {
			testingTarget.Fatalf("filtered project beyond page: %+v %v", filtered, failure)
		}
		hidden, failure := service.StudioProjectPage(ctx, "foreign", application.ProjectFilter{TeamID: "page-team"})
		if failure != nil || hidden.Total != 0 || len(hidden.Items) != 0 {
			testingTarget.Fatal("projects leaked through team filter", failure)
		}
	})
	testingTarget.Run("project cards expose own chapter count and public title rating", func(testingTarget *testing.T) {
		execute(`INSERT INTO chapters(id,title_id,label,order_num) VALUES('card-chapter','project-title-110','1',1);
        INSERT INTO releases(id,chapter_id,project_id,team_id,status) VALUES('card-release','card-chapter','page-project-110','page-team','draft');
        INSERT INTO ratings(id,user_id,title_id,score) VALUES('card-rating','reader','project-title-110',8)`)
		page, failure := service.StudioProjectPage(ctx, "reader", application.ProjectFilter{TeamID: "page-team"})
		if failure != nil || len(page.Items) != 1 {
			testingTarget.Fatalf("card page: %+v %v", page, failure)
		}
		item := page.Items[0]
		if item.ChapterCount != 1 || item.RatingCount != 1 || item.Score != 8 {
			testingTarget.Fatalf("card statistics: %+v", item)
		}
	})
	testingTarget.Run("project next work comes from scoped revisions and preserves unknown as absent", func(testingTarget *testing.T) {
		execute(`INSERT INTO revisions(id,release_id,status,revision_num) VALUES('card-ready','card-release','ready',1),('card-changes','card-release','changes_requested',2)`)
		page, failure := service.StudioProjectPage(ctx, "reader", application.ProjectFilter{TeamID: "page-team"})
		if failure != nil || len(page.Items) != 1 || page.Items[0].NextRevisionID != "card-changes" || page.Items[0].NextRevisionState != "changes_requested" {
			testingTarget.Fatalf("next work: %+v %v", page, failure)
		}
		hidden, failure := service.StudioProjectPage(ctx, "foreign", application.ProjectFilter{TeamID: "page-team"})
		if failure != nil || len(hidden.Items) != 0 {
			testingTarget.Fatalf("foreign work leaked: %+v %v", hidden, failure)
		}
		execute(`INSERT INTO revisions(id,release_id,status,revision_num) VALUES('card-published','card-release','published',3)`)
		page, failure = service.StudioProjectPage(ctx, "reader", application.ProjectFilter{TeamID: "page-team"})
		if failure != nil || len(page.Items) != 1 || page.Items[0].NextRevisionID != "" || page.Items[0].NextRevisionState != "" {
			testingTarget.Fatalf("no pending work: %+v %v", page, failure)
		}
	})

	testingTarget.Run("title comment count excludes hidden deleted and chapter messages", func(testingTarget *testing.T) {
		for _, query := range []string{`INSERT INTO comments(id,user_id,title_id,body) VALUES('count-root','reader',$1::varchar,'Visible root')`, `INSERT INTO comments(id,user_id,title_id,parent_id,root_id,body) VALUES('count-reply','reader',$1::varchar,'count-root','count-root','Visible reply')`, `INSERT INTO comments(id,user_id,title_id,body,is_hidden) VALUES('count-hidden','reader',$1::varchar,'Hidden',true)`, `INSERT INTO comments(id,user_id,title_id,body,deleted_at) VALUES('count-deleted','reader',$1::varchar,'Deleted',now())`, `INSERT INTO comments(id,user_id,title_id,release_id,body) SELECT 'count-chapter','reader',$1::varchar,release.id,'Chapter message' FROM releases release JOIN chapters chapter ON chapter.id=release.chapter_id WHERE chapter.title_id=$1::varchar LIMIT 1`} {
			execute(query, title.ID)
		}
		detail, failure := service.Detail(ctx, title.Slug)
		if failure != nil || detail.CommentCount != 2 {
			testingTarget.Fatalf("visible title comments: %+v %v", detail, failure)
		}
	})
	testingTarget.Run("305 published chapters are searched before pagination and ordered without duplicates", func(testingTarget *testing.T) {
		execute(`INSERT INTO chapters(id,title_id,label,title_kk,order_num) SELECT 'long-chapter-'||number,$1,'Long '||number,'Ұзақ '||number,1000+number FROM generate_series(1,305) number`, title.ID)
		execute(`INSERT INTO releases(id,chapter_id,project_id,team_id,status,current_revision_id) SELECT 'long-release-'||number,'long-chapter-'||number,$1,'team','published','long-revision-'||number FROM generate_series(1,305) number`, projectID)
		execute(`INSERT INTO revisions(id,release_id,status) SELECT 'long-revision-'||number,'long-release-'||number,'published' FROM generate_series(1,305) number`)
		found, failure := service.Chapters(ctx, title.Slug, 0, 50, application.ChapterFilter{Query: "Long 305", Sort: "desc"})
		if failure != nil || len(found.Items) != 1 || found.Items[0].Order != 1305 {
			testingTarget.Fatalf("far chapter: %+v %v", found, failure)
		}
		for _, sort := range []string{"asc", "desc"} {
			seen := map[string]bool{}
			after := 0
			previous := 0
			for {
				page, failure := service.Chapters(ctx, title.Slug, after, 100, application.ChapterFilter{Query: "Long", Sort: sort})
				if failure != nil {
					testingTarget.Fatal(failure)
				}
				for _, chapter := range page.Items {
					if seen[chapter.ID] || previous != 0 && (sort == "asc" && chapter.Order <= previous || sort == "desc" && chapter.Order >= previous) {
						testingTarget.Fatal("duplicate or wrong chapter order")
					}
					seen[chapter.ID] = true
					previous = chapter.Order
				}
				if page.NextOrder == nil {
					break
				}
				after = *page.NextOrder
			}
			if len(seen) != 305 {
				testingTarget.Fatalf("%s lost chapters: %d", sort, len(seen))
			}
		}
	})
}
