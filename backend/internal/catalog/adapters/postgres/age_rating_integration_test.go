package postgres

import (
	"context"
	"errors"
	"fmt"
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

func TestAgeRatingPersistenceFilteringAndLegacyWriters(testingTarget *testing.T) {
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

	execute("INSERT INTO users(id,email,password_hash,role,is_verified) VALUES('admin','age-admin@example.test','test','admin',true),('reader','age-reader@example.test','test','reader',true)")
	repo := New(database.Pool())
	service := application.New(repo, uuid.NewString)
	input := domain.Metadata{Slug: "age-unknown", Name: "Age unknown", Kind: "manga", OriginalStatus: "ongoing", TranslationStatus: "ongoing", Direction: "ltr", Listed: true}
	ids := map[int]string{}
	for _, age := range []int{0, 12, 16, 18} {
		input.Slug = fmt.Sprintf("age-%d", age)
		input.Name = fmt.Sprintf("Age %d", age)
		input.AgeRating = &age
		title, err := service.CreateTitle(ctx, "admin", input)
		if err != nil {
			testingTarget.Fatal(err)
		}
		ids[age] = title.ID
		detail, err := repo.Detail(ctx, input.Slug)
		if err != nil || detail.Title.AgeRating == nil || *detail.Title.AgeRating != age {
			testingTarget.Fatal("classification not persisted", err)
		}
	}
	for _, age := range []int{0, 12, 16, 18} {
		page, err := service.List(ctx, application.Filter{AgeRatings: []int{age}, Limit: 1, Sort: "name"})
		if err != nil || len(page.Items) != 1 || page.Items[0].ID != ids[age] || page.Items[0].AgeRating != age {
			testingTarget.Fatal("incorrect exact age filter", age, err, page)
		}
	}
	input.Slug = "age-18"
	input.Name = "Updated by old client"
	input.AgeRating = nil
	if _, err := service.UpdateTitle(ctx, "admin", ids[18], input); err != nil {
		testingTarget.Fatal(err)
	}
	detail, err := repo.AdminDetail(ctx, "admin", ids[18])
	if err != nil || detail.Title.AgeRating == nil || *detail.Title.AgeRating != 18 {
		testingTarget.Fatal("legacy PUT erased rating", err)
	}
	if _, err := database.Pool().Exec(ctx, "UPDATE titles SET age_rating=14 WHERE id=$1", ids[18]); err == nil {
		testingTarget.Fatal("database accepted invalid rating")
	}
	zero := 0
	input.AgeRating = &zero
	if _, err := service.UpdateTitle(ctx, "admin", ids[18], input); err != nil {
		testingTarget.Fatal(err)
	}
	detail, err = repo.AdminDetail(ctx, "admin", ids[18])
	if err != nil || *detail.Title.AgeRating != 0 {
		testingTarget.Fatal("explicit unknown did not clear classification", err)
	}
	if _, err := service.UpdateTitle(ctx, "reader", ids[18], input); !errors.Is(err, domain.ErrForbidden) {
		testingTarget.Fatal("reader changed classification", err)
	}
	invalid := 14
	if _, err := service.List(ctx, application.Filter{AgeRatings: []int{invalid}}); !errors.Is(err, domain.ErrInvalid) {
		testingTarget.Fatal("invalid filter accepted", err)
	}
	execute("INSERT INTO titles(id,slug,name_kk,is_listed) VALUES('legacy','legacy','Legacy title',true)")
	var legacy int
	if err := database.Pool().QueryRow(ctx, "SELECT age_rating FROM titles WHERE id='legacy'").Scan(&legacy); err != nil || legacy != 0 {
		testingTarget.Fatal("old insert incompatible", err)
	}
}
