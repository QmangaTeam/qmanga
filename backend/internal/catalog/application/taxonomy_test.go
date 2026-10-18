package application

import (
	"context"
	"errors"
	"qmanga/internal/catalog/domain"
	"testing"
)

type taxonomyRepository struct {
	Repository
	allowed        bool
	writes         int
	lastCollection domain.Collection
}

func (repo *taxonomyRepository) IsAdmin(context.Context, string) (bool, error) {
	return repo.allowed, nil
}
func (repo *taxonomyRepository) Taxonomy(context.Context, bool) (Taxonomy, error) {
	return Taxonomy{}, nil
}
func (repo *taxonomyRepository) SaveTaxon(context.Context, string, string, Genre) error {
	repo.writes++
	return nil
}
func (repo *taxonomyRepository) DeleteTaxon(context.Context, string, string, string) error {
	repo.writes++
	return nil
}
func (repo *taxonomyRepository) SaveCollection(_ context.Context, _ string, item domain.Collection) error {
	repo.writes++
	repo.lastCollection = item
	return nil
}
func (repo *taxonomyRepository) SetCategories(context.Context, string, string, []string) error {
	repo.writes++
	return nil
}
func TestTaxonomyRejectsUnauthorizedAndInvalidMutations(t *testing.T) {
	ctx := context.Background()
	repo := &taxonomyRepository{}
	service := New(repo, func() string { return "new-id" })
	if _, err := service.SaveTaxon(ctx, "reader", "categories", "", "Name", "name"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal(err)
	}
	if _, err := service.SaveCollection(ctx, "reader", domain.Collection{Name: "Name", Slug: "name"}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal(err)
	}
	if err := service.SetCategories(ctx, "reader", "title", []string{}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal(err)
	}
	if err := service.DeleteTaxon(ctx, "reader", "genres", "genre"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal(err)
	}
	if repo.writes != 0 {
		t.Fatal("unauthorized writes reached persistence")
	}
	repo.allowed = true
	if _, err := service.SaveTaxon(ctx, "admin", "age", "", "Name", "name"); !errors.Is(err, domain.ErrInvalid) {
		t.Fatal("age mixed with taxonomy", err)
	}
	if _, err := service.SaveCollection(ctx, "admin", domain.Collection{Name: "Name", Slug: "name", TitleIDs: []string{"duplicate", "duplicate"}}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatal(err)
	}
	if err := service.SetCategories(ctx, "admin", "title", []string{"same", "same"}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatal(err)
	}
	if repo.writes != 0 {
		t.Fatal("invalid writes reached persistence")
	}
	item, err := service.SaveCollection(ctx, "admin", domain.Collection{Name: "  Әлем ", Slug: "world", Position: 2, TitleIDs: []string{"second", "first"}})
	if err != nil || item.ID != "new-id" || item.Name != "Әлем" || repo.lastCollection.TitleIDs[0] != "second" {
		t.Fatal("collection normalization changed selection order", err, item)
	}
}
