package postgres

import (
	"context"
	"net/url"
	"os"
	"qmanga/internal/catalog/application"
	"qmanga/internal/catalog/domain"
	platform "qmanga/internal/platform/postgres"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSeededTagSearchAndCombinedFilters(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	base, err := platform.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	schema := "tags_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = base.Pool().Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer base.Pool().Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	parsed, _ := url.Parse(databaseURL)
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := platform.New(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Pool().Exec(ctx, "INSERT INTO users(id,email,password_hash,role,is_verified) VALUES('admin','admin@tags.test','test','admin',true)"); err != nil {
		t.Fatal(err)
	}
	repo := New(db.Pool())
	service := application.New(repo, uuid.NewString)

	taxonomy, err := service.Taxonomy(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(taxonomy.Categories) != 6 {
		t.Fatalf("seeded categories missing: %d", len(taxonomy.Categories))
	}
	if len(taxonomy.Genres) != 713 {
		t.Fatalf("seeded tags missing: %d", len(taxonomy.Genres))
	}
	group := ""
	grouped := 0
	for _, item := range taxonomy.Genres {
		if item.CategorySlug == "" {
			t.Fatalf("tag without group: %+v", item)
		}
		if item.CategorySlug != group {
			group = item.CategorySlug
			grouped++
		}
		if item.NameRU == "" {
			t.Fatalf("tag without russian name: %+v", item)
		}
		if item.Name == "" {
			t.Fatalf("tag without kazakh name: %+v", item)
		}
	}
	if grouped != 6 {
		t.Fatalf("tags are not grouped by the six categories: %d", grouped)
	}
	horror := ""
	for _, item := range taxonomy.Genres {
		if item.Slug == "horror" {
			horror = item.Slug
			if item.Name != "Қорқыныш" || item.NameRU != "Ужасы" {
				t.Fatalf("seeded names lost: %+v", item)
			}
		}
	}
	if horror == "" {
		t.Fatal("seeded tag horror missing")
	}

	input := domain.Metadata{Slug: "tagged", Name: "Tagged", Kind: "manga", OriginalStatus: "ongoing", TranslationStatus: "ongoing", Direction: "ltr", Listed: true}
	horrorID := ""
	for _, item := range taxonomy.Genres {
		if item.Slug == "horror" {
			horrorID = item.ID
		}
	}
	input.GenreIDs = []string{horrorID}
	title, err := service.CreateTitle(ctx, "admin", input)
	if err != nil {
		t.Fatal(err)
	}
	other, err := service.CreateTitle(ctx, "admin", domain.Metadata{Slug: "plain", Name: "Plain", Kind: "manga", OriginalStatus: "ongoing", TranslationStatus: "ongoing", Direction: "ltr", Listed: true})
	if err != nil {
		t.Fatal(err)
	}

	page, err := service.List(ctx, application.Filter{Genres: []string{"horror"}})
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != title.ID {
		t.Fatalf("single tag filter failed: %+v %v", page, err)
	}
	page, err = service.List(ctx, application.Filter{Genres: []string{"horror", "magic"}})
	if err != nil || len(page.Items) != 0 {
		t.Fatalf("tags are not combined with AND: %+v %v", page, err)
	}
	page, err = service.List(ctx, application.Filter{Query: "қорқыныш"})
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != title.ID {
		t.Fatalf("search by kazakh tag name failed: %+v %v", page, err)
	}
	page, err = service.List(ctx, application.Filter{Query: "ужасы"})
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != title.ID {
		t.Fatalf("search by russian tag name failed: %+v %v", page, err)
	}
	page, err = service.List(ctx, application.Filter{Query: "tagged"})
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != title.ID {
		t.Fatalf("search by title name failed: %+v %v", page, err)
	}
	_ = other
	if _, err = service.List(ctx, application.Filter{Genres: []string{"not a slug"}}); err == nil {
		t.Fatal("invalid tag slug accepted")
	}
	many := make([]string, 0, 13)
	for index := 0; index < 13; index++ {
		many = append(many, "tag-"+strings.Repeat("x", index+1))
	}
	if _, err = service.List(ctx, application.Filter{Genres: many}); err == nil {
		t.Fatal("too many tags accepted")
	}
}
