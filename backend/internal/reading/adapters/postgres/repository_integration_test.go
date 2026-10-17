package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/url"
	"os"
	communityPostgres "qmanga/internal/community/adapters/postgres"
	communityApp "qmanga/internal/community/application"
	communityDomain "qmanga/internal/community/domain"
	notificationPostgres "qmanga/internal/notifications/adapters/postgres"
	platform "qmanga/internal/platform/postgres"
	"qmanga/internal/reading/application"
	"qmanga/internal/reading/domain"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type queryCounter struct{ count atomic.Int64 }

func (counter *queryCounter) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	counter.count.Add(1)
	return ctx
}
func (counter *queryCounter) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func TestReadingIsolationReplayAndConflict(testingTarget *testing.T) {
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
	schema := "reading_" + strings.ReplaceAll(uuid.NewString(), "-", "")
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
	execute := func(query string) {
		testingTarget.Helper()
		if _, failure := database.Pool().Exec(ctx, query); failure != nil {
			testingTarget.Fatal(failure)
		}
	}
	execute(`INSERT INTO users(id,email,password_hash)VALUES('first','first@example.test','test'),('second','second@example.test','test');INSERT INTO titles(id,slug,name_kk)VALUES('title','title','Атау');INSERT INTO teams(id,slug,name,is_approved)VALUES('team','team','Топ',true);INSERT INTO projects(id,title_id,team_id)VALUES('project','title','team');INSERT INTO chapters(id,title_id,label,order_num)VALUES('chapter','title','1',10);INSERT INTO releases(id,chapter_id,project_id,team_id,status,current_revision_id)VALUES('release','chapter','project','team','published','revision');INSERT INTO revisions(id,release_id,status)VALUES('revision','release','published');INSERT INTO pages(id,revision_id,position)VALUES('page','revision',1)`)
	repository := New(database.Pool())
	selection := application.Selection{TitleID: "title", Status: "reading", Favorite: true, PreferredTeamID: "team", Notify: true}
	if failure = repository.Save(ctx, "first", selection); failure != nil {
		testingTarget.Fatal(failure)
	}
	own, failure := repository.List(ctx, "first", "")
	if failure != nil || len(own) != 1 || !own[0].Favorite || own[0].Unseen != 1 {
		testingTarget.Fatal(own, failure)
	}
	foreign, failure := repository.List(ctx, "second", "")
	stale := selection
	selection.Version = own[0].Version
	selection.Status = "paused"
	if failure = repository.Save(ctx, "first", selection); failure != nil {
		testingTarget.Fatal(failure)
	}
	if failure = repository.Save(ctx, "first", stale); !errors.Is(failure, domain.ErrConflict) {
		testingTarget.Fatal("stale library change accepted", failure)
	}
	selection.Version++
	selection.Status = "reading"
	if failure = repository.Save(ctx, "first", selection); failure != nil {
		testingTarget.Fatal(failure)
	}
	if failure != nil || len(foreign) != 0 {
		testingTarget.Fatal("library crossed accounts", failure)
	}
	event := domain.Event{ID: "event1", TitleID: "title", DeviceID: "phone", Sequence: 1, ReleaseID: "release", RevisionID: "revision", PageID: "page", Fraction: 1, Completed: true}
	result, failure := repository.Advance(ctx, "first", event)
	if failure != nil || result.Progress.Version != 1 || result.Progress.FurthestOrder != 10 {
		testingTarget.Fatal(result, failure)
	}
	replay, failure := repository.Advance(ctx, "first", event)
	if failure != nil || replay.Outcome != "duplicate" || replay.Progress.Version != 1 {
		testingTarget.Fatal(replay, failure)
	}
	event.ID = "event2"
	event.DeviceID = "tablet"
	result, failure = repository.Advance(ctx, "first", event)
	if failure != nil || result.Outcome != "conflict" || result.Progress.Version != 1 {
		testingTarget.Fatal(result, failure)
	}
	event.BaseVersion = 1
	result, failure = repository.Advance(ctx, "first", event)
	if failure != nil || result.Outcome != "accepted" || result.Progress.Version != 2 {
		testingTarget.Fatal(result, failure)
	}
	lostResponse := domain.Event{ID: "event1", TitleID: "title", DeviceID: "phone", Sequence: 1, ReleaseID: "release", RevisionID: "revision", PageID: "page", Fraction: 1, Completed: true}
	retried, failure := repository.Advance(ctx, "first", lostResponse)
	if failure != nil || retried.Outcome != "conflict" || retried.Progress.Version != 2 {
		testingTarget.Fatal("a lost acknowledgment silently rebased across another device", retried, failure)
	}
	event.ID = "event3"
	event.BaseVersion = 2
	result, failure = repository.Advance(ctx, "first", event)
	if failure != nil || result.Outcome != "stale" {
		testingTarget.Fatal(result, failure)
	}
	event.ID = "event4"
	event.Sequence = 2
	event.PageID = "foreign"
	if _, failure = repository.Advance(ctx, "first", event); !errors.Is(failure, domain.ErrNotFound) {
		testingTarget.Fatal("foreign page accepted", failure)
	}
	other, failure := repository.Progress(ctx, "second", "title")
	if failure != nil || other.Version != 0 {
		testingTarget.Fatal("progress crossed accounts", failure)
	}
	execute(`UPDATE users SET is_verified=true;`)
	community := communityApp.New(communityPostgres.New(database.Pool()), uuid.NewString)
	if failure = community.Rating(ctx, "first", "title", 8); failure != nil {
		testingTarget.Fatal(failure)
	}
	if failure = community.Rating(ctx, "first", "title", 10); failure != nil {
		testingTarget.Fatal(failure)
	}
	if failure = community.Rating(ctx, "second", "title", 6); failure != nil {
		testingTarget.Fatal(failure)
	}
	commentID, failure := community.Comment(ctx, "first", communityDomain.Comment{TitleID: "title", ReleaseID: "release", Body: "Қазақша пікір", Spoiler: true})
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	replyID, failure := community.Comment(ctx, "second", communityDomain.Comment{TitleID: "title", ReleaseID: "release", ParentID: commentID, Body: "Жауап мәтіні"})
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	if _, failure = community.Comment(ctx, "first", communityDomain.Comment{TitleID: "title", ReleaseID: "release", ParentID: replyID, Body: "Терең жауап"}); !errors.Is(failure, communityDomain.ErrInvalid) {
		testingTarget.Fatal("nested reply accepted", failure)
	}
	if failure = community.Edit(ctx, "second", commentID, "Өзгенің пікірін ауыстыру", false); !errors.Is(failure, communityDomain.ErrForbidden) {
		testingTarget.Fatal("foreign edit", failure)
	}
	discussion, failure := community.Discussion(ctx, "title", "release", "")
	if failure != nil || len(discussion.Comments) != 2 || discussion.Score != 8 || discussion.Ratings != 2 {
		testingTarget.Fatal(discussion, failure)
	}
	for index := 0; index < 15; index++ {
		if _, failure = database.Pool().Exec(ctx, `INSERT INTO comments(id,user_id,title_id,release_id,parent_id,body)VALUES($1,'second','title','release',$2,'Қосымша жауап')`, uuid.NewString(), commentID); failure != nil {
			testingTarget.Fatal(failure)
		}
	}
	discussion, failure = community.Discussion(ctx, "title", "release", "")
	if failure != nil || len(discussion.Comments) != 11 {
		testingTarget.Fatal("unbounded replies", len(discussion.Comments), failure)
	}
	moreReplies, failure := community.Replies(ctx, commentID, discussion.Comments[len(discussion.Comments)-1].ID)
	if failure != nil || len(moreReplies.Comments) != 6 {
		testingTarget.Fatal("reply pagination lost data", moreReplies, failure)
	}
	reportID, failure := community.Report(ctx, "second", communityApp.Report{TitleID: "title", ReleaseID: "release", TargetType: "page", TargetID: "page", PageID: "page", Category: "image", Reason: "Бет анық жүктелмейді"})
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	if _, failure = community.Reports(ctx, "second", ""); !errors.Is(failure, communityDomain.ErrForbidden) {
		testingTarget.Fatal("reader accessed reports", failure)
	}
	execute(`UPDATE users SET role='moderator' WHERE id='first'`)
	reports, failure := community.Reports(ctx, "first", "")
	if failure != nil || len(reports) != 1 {
		testingTarget.Fatal(reports, failure)
	}
	if failure = community.Visibility(ctx, "first", commentID, true, "Спойлер ескертуі дұрыс емес"); failure != nil {
		testingTarget.Fatal(failure)
	}
	discussion, failure = community.Discussion(ctx, "title", "release", "")
	if failure != nil || discussion.Comments[0].Body != "" || !discussion.Comments[0].Hidden {
		testingTarget.Fatal("hidden text leaked", failure)
	}
	if failure = community.Resolve(ctx, "first", reportID, "resolved", "Бет тексеріліп түзетілді"); failure != nil {
		testingTarget.Fatal(failure)
	}
	notifications := notificationPostgres.New(database.Pool())
	if worked, failure := notifications.Deliver(ctx); failure != nil || !worked {
		testingTarget.Fatal("reply delivery", failure)
	}
	items, failure := notifications.List(ctx, "first", "")
	if failure != nil || len(items) != 1 || items[0].Kind != "comment_reply" {
		testingTarget.Fatal(items, failure)
	}
	execute(`INSERT INTO outbox_events(id,event_type,payload)VALUES('event-first','release_published','{"release_id":"release","revision_id":"revision"}'),('event-second','release_published','{"release_id":"release","revision_id":"revision"}')`)
	for attempt := 0; attempt < 2; attempt++ {
		if worked, failure := notifications.Deliver(ctx); failure != nil || !worked {
			testingTarget.Fatal("release delivery", failure)
		}
	}
	items, failure = notifications.List(ctx, "first", "")
	if failure != nil || len(items) != 2 {
		testingTarget.Fatal("duplicate notification", items, failure)
	}
	if failure = notifications.Read(ctx, "second", items[0].ID); failure != nil {
		testingTarget.Fatal(failure)
	}
	items, failure = notifications.List(ctx, "first", "")
	if failure != nil || items[0].Read {
		testingTarget.Fatal("foreign notification marked", failure)
	}
	execute(`INSERT INTO library_entries(id,user_id,title_id)VALUES('library-second','second','title'); INSERT INTO teams(id,slug,name,is_approved)VALUES('alternate','alternate','Екінші топ',true); INSERT INTO projects(id,title_id,team_id)VALUES('alternate-project','title','alternate'); INSERT INTO releases(id,chapter_id,project_id,team_id,status,current_revision_id)VALUES('alternate-release','chapter','alternate-project','alternate','published','alternate-revision'); INSERT INTO revisions(id,release_id,status)VALUES('alternate-revision','alternate-release','published'); INSERT INTO outbox_events(id,event_type,payload)VALUES('new-original','release_published','{"release_id":"release","revision_id":"revision"}'),('new-alternate','release_published','{"release_id":"alternate-release","revision_id":"alternate-revision"}')`)
	for attempt := 0; attempt < 2; attempt++ {
		if _, failure = notifications.Deliver(ctx); failure != nil {
			testingTarget.Fatal(failure)
		}
	}
	ungrouped, failure := notifications.List(ctx, "second", "")
	if failure != nil || len(ungrouped) != 1 {
		testingTarget.Fatal("logical chapter duplicated across teams", len(ungrouped), failure)
	}
	execute("INSERT INTO titles(id,slug,name_kk,is_listed)VALUES('announcement','announcement','Анонс',true),('private-title','private-title','Private',false)")
	if _, failure = community.Discussion(ctx, "announcement", "", ""); failure != nil {
		testingTarget.Fatal("announcement discussion inaccessible", failure)
	}
	if _, failure = community.RightsReport(ctx, communityApp.RightsRequest{TitleSlug: "announcement", Email: "rights@example.test", Reason: "Functional test announcement report"}); failure != nil {
		testingTarget.Fatal("announcement rights contact inaccessible", failure)
	}
	if _, failure = community.RightsReport(ctx, communityApp.RightsRequest{TitleSlug: "private-title", Email: "rights@example.test", Reason: "Private title should remain private"}); !errors.Is(failure, communityDomain.ErrNotFound) {
		testingTarget.Fatal("draft exposed through rights form", failure)
	}
	communityService := community
	receipt, failure := communityService.RightsReport(ctx, communityApp.RightsRequest{TitleSlug: "title", Email: "rights@example.test", Reason: "Мен құқық иесімін, материалды тексеріңіз."})
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	if _, failure = communityService.RightsStatus(ctx, receipt.ID, "wrong-token"); !errors.Is(failure, communityDomain.ErrNotFound) {
		testingTarget.Fatal("anonymous receipt leaked", failure)
	}
	status, failure := communityService.RightsStatus(ctx, receipt.ID, receipt.Token)
	if failure != nil || status.Status != "pending" {
		testingTarget.Fatal("receipt status missing", status, failure)
	}
	if failure = community.Resolve(ctx, "first", receipt.ID, "resolved", "Материал тексерілді."); failure != nil {
		testingTarget.Fatal(failure)
	}
	status, failure = communityService.RightsStatus(ctx, receipt.ID, receipt.Token)
	if failure != nil || status.Resolution != "Материал тексерілді." {
		testingTarget.Fatal("resolution not visible to reporter", status, failure)
	}
	personal, failure := communityService.MyReports(ctx, "second", "")
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	for _, report := range personal {
		if report.ID == receipt.ID {
			testingTarget.Fatal("anonymous report leaked into account history")
		}
	}
	testingTarget.Run("library page does not issue a progress query per title", func(testingTarget *testing.T) {
		execute(`INSERT INTO users(id,email,password_hash) VALUES('paging','paging@example.test','test');
		 INSERT INTO titles(id,slug,name_kk) SELECT 'lib-'||lpad(number::text,3,'0'),'lib-'||number,'Кітап '||number FROM generate_series(1,110) number;
		 INSERT INTO library_entries(id,user_id,title_id,status) SELECT 'entry-'||id,'paging',id,'reading' FROM titles WHERE id LIKE 'lib-%'`)
		counter := &queryCounter{}
		config := database.Pool().Config().Copy()
		config.ConnConfig.Tracer = counter
		pool, failure := pgxpool.NewWithConfig(ctx, config)
		if failure != nil {
			testingTarget.Fatal(failure)
		}
		defer pool.Close()
		items, failure := New(pool).List(ctx, "paging", "")
		if failure != nil || len(items) != 100 {
			testingTarget.Fatalf("library page: %d %v", len(items), failure)
		}
		if count := counter.count.Load(); count != 1 {
			testingTarget.Fatalf("100 rows required %d queries, want one joined query", count)
		}
		execute(`UPDATE library_entries SET status='planned',is_favorite=true WHERE user_id='paging' AND title_id='lib-001'; UPDATE titles SET name_kk='Қашықтағы алмас' WHERE id='lib-001'; UPDATE titles SET name_kk='100% ready' WHERE id='lib-002'`)
		service := application.New(New(pool))
		filtered, failure := service.Library(ctx, "paging", application.LibraryFilter{Query: "ҚАШЫҚТАҒЫ", Status: "planned", Favorite: true})
		if failure != nil || filtered.Total != 1 || len(filtered.Items) != 1 || filtered.Items[0].TitleID != "lib-001" || filtered.Next != "" {
			testingTarget.Fatalf("complete filtered library: %+v %v", filtered, failure)
		}
		firstPage, failure := service.Library(ctx, "paging", application.LibraryFilter{})
		if failure != nil || firstPage.Total != 110 || len(firstPage.Items) != 100 || firstPage.Next == "" {
			testingTarget.Fatalf("first page: %+v %v", firstPage, failure)
		}
		lastPage, failure := service.Library(ctx, "paging", application.LibraryFilter{After: firstPage.Next})
		if failure != nil || lastPage.Total != 110 || len(lastPage.Items) != 10 || lastPage.Next != "" {
			testingTarget.Fatalf("last page/count: %+v %v", lastPage, failure)
		}
		literal, failure := service.Library(ctx, "paging", application.LibraryFilter{Query: "%"})
		if failure != nil || literal.Total != 1 {
			testingTarget.Fatalf("literal search: %+v %v", literal, failure)
		}
		missing, failure := service.Library(ctx, "first", application.LibraryFilter{Query: "Қашықтағы"})
		if failure != nil || missing.Total != 0 || len(missing.Items) != 0 {
			testingTarget.Fatal("filtered library leaked across accounts", failure)
		}
		if _, failure := service.Library(ctx, "paging", application.LibraryFilter{Status: "invalid"}); !errors.Is(failure, domain.ErrInvalid) {
			testingTarget.Fatal("invalid status accepted", failure)
		}
		execute(`INSERT INTO titles(id,slug,name_kk,cover_asset_id) SELECT 'recent-'||number,'recent-'||number,'Жақында '||number,'cover-'||number FROM generate_series(1,3) number;
		 INSERT INTO projects(id,title_id,team_id) SELECT 'recent-project-'||number,'recent-'||number,'team' FROM generate_series(1,3) number;
		 INSERT INTO chapters(id,title_id,label,order_num) SELECT 'recent-chapter-'||number,'recent-'||number,'1',1 FROM generate_series(1,3) number;
		 INSERT INTO releases(id,chapter_id,project_id,team_id,status,current_revision_id) SELECT 'recent-release-'||number,'recent-chapter-'||number,'recent-project-'||number,'team','published','recent-revision-'||number FROM generate_series(1,3) number;
		 INSERT INTO revisions(id,release_id,status) SELECT 'recent-revision-'||number,'recent-release-'||number,'published' FROM generate_series(1,3) number;
		 INSERT INTO pages(id,revision_id,position) SELECT 'recent-page-'||number,'recent-revision-'||number,1 FROM generate_series(1,3) number;
		 INSERT INTO reading_progress(id,user_id,title_id,release_id,revision_id,page_id,page_position,fraction,version,updated_at) SELECT 'recent-progress-'||number,'paging','recent-'||number,'recent-release-'||number,'recent-revision-'||number,'recent-page-'||number,1,0.5,1,'2026-10-18 00:00:00+00'::timestamptz + number*interval '1 minute' FROM generate_series(1,3) number`)
		recent, failure := service.Recent(ctx, "paging", 2)
		if failure != nil || len(recent) != 2 || recent[0].TitleID != "recent-3" || recent[1].TitleID != "recent-2" || recent[0].CoverID != "cover-3" || recent[0].Availability != "available" || recent[0].PagePosition != 1 || recent[0].Progress.Fraction != 0.5 {
			testingTarget.Fatalf("activity projection independent of library: %+v %v", recent, failure)
		}
		execute(`UPDATE releases SET status='withdrawn' WHERE id='recent-release-3'; INSERT INTO revisions(id,release_id,status)VALUES('updated-revision','recent-release-2','published'); UPDATE releases SET current_revision_id='updated-revision' WHERE id='recent-release-2'`)
		recent, failure = service.Recent(ctx, "paging", 4)
		if failure != nil || recent[0].Availability != "withdrawn" || recent[1].Availability != "updated" {
			testingTarget.Fatalf("availability: %+v %v", recent, failure)
		}
		references, failure := service.References(ctx, []string{"recent-release-1", "recent-release-3"})
		if failure != nil || len(references) != 1 || references[0].TitleID != "recent-1" {
			testingTarget.Fatal("public metadata exposed withdrawn reference", references, failure)
		}
	})
}
