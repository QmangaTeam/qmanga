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

func TestHomeRankingsUseActualReadersRatingsAndPublishedTitles(testingTarget *testing.T) {
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
	schema := "rankings_" + strings.ReplaceAll(uuid.NewString(), "-", "")
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
	execute("INSERT INTO users(id,email,password_hash,is_blocked) VALUES('one','one@test.invalid','test',false),('two','two@test.invalid','test',false),('blocked-one','blocked-one@test.invalid','test',true),('blocked-two','blocked-two@test.invalid','test',true)")
	execute("INSERT INTO teams(id,slug,name,is_approved) VALUES('team','team','Team',true)")
	for _, title := range []struct{ id, kind string }{{"most-read", "manhwa"}, {"best-rated", "manhwa"}, {"manhua", "manhua"}} {
		execute("INSERT INTO titles(id,slug,name_kk,type) VALUES($1,$1,$1,$2)", title.id, title.kind)
		execute("INSERT INTO projects(id,title_id,team_id) VALUES($1,$2,'team')", title.id+"-project", title.id)
		execute("INSERT INTO chapters(id,title_id,label,order_num) VALUES($1,$2,'1',1)", title.id+"-chapter", title.id)
		execute("INSERT INTO releases(id,chapter_id,project_id,team_id,status,current_revision_id) VALUES($1,$2,$3,'team','published',$4)", title.id+"-release", title.id+"-chapter", title.id+"-project", title.id+"-revision")
		execute("INSERT INTO revisions(id,release_id,status) VALUES($1,$2,'published')", title.id+"-revision", title.id+"-release")
	}
	execute("INSERT INTO titles(id,slug,name_kk,type,is_listed) VALUES('announcement','announcement','Announcement','manhwa',true)")
	execute("INSERT INTO ratings(id,user_id,title_id,score) VALUES('rating-one','one','most-read',8),('rating-two','one','best-rated',9),('blocked-rating-one','blocked-one','most-read',10),('blocked-rating-two','blocked-two','most-read',10),('announcement-rating','one','announcement',10)")
	for _, item := range []struct{ user, title string }{{"one", "most-read"}, {"two", "most-read"}, {"one", "best-rated"}, {"blocked-one", "best-rated"}, {"blocked-two", "best-rated"}} {
		execute("INSERT INTO reading_progress(id,user_id,title_id,release_id) VALUES($1,$2,$3,$4)", uuid.NewString(), item.user, item.title, item.title+"-release")
	}
	service := application.New(New(database.Pool()), uuid.NewString)
	mostRead, failure := service.Rankings(ctx, "", 16)
	if failure != nil || len(mostRead) != 3 || mostRead[0].ID != "most-read" || mostRead[1].ID != "best-rated" {
		testingTarget.Fatal("ranking did not count only actual unblocked readers", mostRead, failure)
	}
	top, failure := service.Rankings(ctx, "manhwa", 16)
	if failure != nil || len(top) != 2 || top[0].ID != "best-rated" || top[0].Score != 9 || top[1].Score != 8 {
		testingTarget.Fatal("type ranking did not use actual unblocked ratings/publications", top, failure)
	}
	otherKind, failure := service.Rankings(ctx, "manhua", 16)
	if failure != nil || len(otherKind) != 1 || otherKind[0].ID != "manhua" {
		testingTarget.Fatal("ranking mixed title kinds", otherKind, failure)
	}
	limited, failure := service.Rankings(ctx, "", 1)
	if failure != nil || len(limited) != 1 {
		testingTarget.Fatal("ranking ignored limit", limited, failure)
	}
	if _, failure = service.Rankings(ctx, "invalid", 16); !errors.Is(failure, domain.ErrInvalid) {
		testingTarget.Fatal("invalid ranking kind accepted", failure)
	}
}
