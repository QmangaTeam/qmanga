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
	"strings"
	"testing"
	"time"
)

func TestTaxonomyPersistenceAuthorizationAndFilters(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base, err := platform.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	schema := "taxonomy_" + strings.ReplaceAll(uuid.NewString(), "-", "")
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
	if _, err = db.Pool().Exec(ctx, "INSERT INTO users(id,email,password_hash,role,is_verified) VALUES('admin','admin@taxonomy.test','test','admin',true),('reader','reader@taxonomy.test','test','reader',true)"); err != nil {
		t.Fatal(err)
	}
	repo := New(db.Pool())
	service := application.New(repo, uuid.NewString)
	category, err := service.SaveTaxon(ctx, "admin", "categories", "", "Қиял әлемдері", "fantasy-worlds")
	if err != nil {
		t.Fatal(err)
	}
	genre, err := service.SaveTaxon(ctx, "admin", "genres", "", "Adventure", "adventure-tales")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.SaveTaxon(ctx, "reader", "categories", "", "No", "no"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("reader wrote category", err)
	}
	adult := 18
	input := domain.Metadata{Slug: "first", Name: "First", Kind: "manga", OriginalStatus: "ongoing", TranslationStatus: "ongoing", Direction: "ltr", Listed: true, AgeRating: &adult, GenreIDs: []string{genre.ID}}
	first, err := service.CreateTitle(ctx, "admin", input)
	if err != nil {
		t.Fatal(err)
	}
	input.Slug = "second"
	input.Name = "Second"
	second, err := service.CreateTitle(ctx, "admin", input)
	if err != nil {
		t.Fatal(err)
	}
	input.Slug = "hidden"
	input.Name = "Hidden"
	input.Listed = false
	hidden, err := service.CreateTitle(ctx, "admin", input)
	if err != nil {
		t.Fatal(err)
	}
	for _, title := range []string{first.ID, second.ID} {
		if err = service.SetCategories(ctx, "admin", title, []string{category.ID}); err != nil {
			t.Fatal(err)
		}
	}
	collection, err := service.SaveCollection(ctx, "admin", domain.Collection{Name: "Editor choice", Slug: "editor-choice", Description: "Independent editorial selection", Position: 3, Visible: true, TitleIDs: []string{second.ID, hidden.ID, first.ID}})
	if err != nil {
		t.Fatal(err)
	}
	taxonomy, err := service.Taxonomy(ctx, "")
	if err != nil || len(taxonomy.Collections) != 1 || len(taxonomy.Collections[0].TitleIDs) != 2 || taxonomy.Collections[0].TitleIDs[0] != second.ID {
		t.Fatal("public visibility or order lost", err, taxonomy)
	}
	adminTaxonomy, err := service.Taxonomy(ctx, "admin")
	if err != nil || len(adminTaxonomy.Collections[0].TitleIDs) != 3 {
		t.Fatal("admin lost hidden member", err)
	}
	page, err := service.List(ctx, application.Filter{Category: category.Slug, Collection: collection.Slug, Genres: []string{genre.Slug}, AgeRatings: []int{adult}, Sort: "name", Limit: 1})
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != first.ID || page.Next == nil {
		t.Fatal("combined filters/pagination failed", err, page)
	}
	page, err = service.List(ctx, application.Filter{Category: category.Slug, Collection: collection.Slug, Genres: []string{genre.Slug}, AgeRatings: []int{adult}, Sort: "name", Limit: 1, After: page.Next})
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != second.ID || page.Next != nil {
		t.Fatal("next filter page failed", err, page)
	}
	detail, err := repo.Detail(ctx, "first")
	if err != nil || len(detail.Categories) != 1 || len(detail.Collections) != 1 {
		t.Fatal("title labels missing", err, detail)
	}
	if err = service.SetCategories(ctx, "admin", first.ID, []string{"missing"}); err == nil {
		t.Fatal("unknown category accepted")
	}
	detail, err = repo.Detail(ctx, "first")
	if err != nil || len(detail.Categories) != 1 {
		t.Fatal("failed assignment not atomic", err)
	}
	category.Name = "Revised category"
	if _, err = service.SaveTaxon(ctx, "admin", "categories", category.ID, category.Name, category.Slug); err != nil {
		t.Fatal(err)
	}
	collection.Visible = false
	collection.TitleIDs = []string{first.ID}
	if _, err = service.SaveCollection(ctx, "admin", collection); err != nil {
		t.Fatal(err)
	}
	page, err = service.List(ctx, application.Filter{Collection: collection.Slug})
	if err != nil || len(page.Items) != 0 {
		t.Fatal("hidden collection publicly filterable", err)
	}
	detail, err = repo.Detail(ctx, "first")
	if err != nil || len(detail.Collections) != 0 {
		t.Fatal("hidden collection labels leaked", err)
	}
	for _, kind := range []string{"categories", "genres", "collections"} {
		id := category.ID
		if kind == "genres" {
			id = genre.ID
		}
		if kind == "collections" {
			id = collection.ID
		}
		if err = service.DeleteTaxon(ctx, "admin", kind, id); err != nil {
			t.Fatal(err)
		}
	}
	detail, err = repo.Detail(ctx, "first")
	if err != nil || len(detail.Categories) != 0 || len(detail.Genres) != 0 {
		t.Fatal("taxonomy delete did not unlink", err)
	}
}
