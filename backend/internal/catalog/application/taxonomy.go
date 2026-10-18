package application

import (
	"context"
	"qmanga/internal/catalog/domain"
	"strings"
)

type Taxonomy struct {
	Genres, Categories []Genre
	Collections        []domain.Collection
}
type TaxonomyRepository interface {
	Taxonomy(context.Context, bool) (Taxonomy, error)
	SaveTaxon(context.Context, string, string, Genre) error
	DeleteTaxon(context.Context, string, string, string) error
	SaveCollection(context.Context, string, domain.Collection) error
	SetCategories(context.Context, string, string, []string) error
}

func (service *Service) taxonomyRepository() (TaxonomyRepository, error) {
	repository, ok := service.repository.(TaxonomyRepository)
	if !ok {
		return nil, domain.ErrInvalid
	}
	return repository, nil
}
func (service *Service) Taxonomy(ctx context.Context, actor string) (Taxonomy, error) {
	if actor != "" {
		if err := service.admin(ctx, actor); err != nil {
			return Taxonomy{}, err
		}
	}
	repository, err := service.taxonomyRepository()
	if err != nil {
		return Taxonomy{}, err
	}
	return repository.Taxonomy(ctx, actor != "")
}
func (service *Service) SaveTaxon(ctx context.Context, actor, kind, id, name, slug string) (Genre, error) {
	if err := service.admin(ctx, actor); err != nil {
		return Genre{}, err
	}
	if (kind != "genres" && kind != "categories") || !domain.ValidGenre(name, slug) {
		return Genre{}, domain.ErrInvalid
	}
	if id == "" {
		id = service.newID()
	}
	item := Genre{ID: id, Name: strings.TrimSpace(name), Slug: slug}
	repository, err := service.taxonomyRepository()
	if err != nil {
		return item, err
	}
	return item, repository.SaveTaxon(ctx, actor, kind, item)
}
func (service *Service) DeleteTaxon(ctx context.Context, actor, kind, id string) error {
	if err := service.admin(ctx, actor); err != nil {
		return err
	}
	if kind != "genres" && kind != "categories" && kind != "collections" {
		return domain.ErrInvalid
	}
	repository, err := service.taxonomyRepository()
	if err != nil {
		return err
	}
	return repository.DeleteTaxon(ctx, actor, kind, id)
}
func (service *Service) SaveCollection(ctx context.Context, actor string, item domain.Collection) (domain.Collection, error) {
	if err := service.admin(ctx, actor); err != nil {
		return item, err
	}
	if item.ID == "" {
		item.ID = service.newID()
	}
	item.Name = strings.TrimSpace(item.Name)
	if !item.Valid() {
		return item, domain.ErrInvalid
	}
	repository, err := service.taxonomyRepository()
	if err != nil {
		return item, err
	}
	return item, repository.SaveCollection(ctx, actor, item)
}
func (service *Service) SetCategories(ctx context.Context, actor, title string, ids []string) error {
	if err := service.admin(ctx, actor); err != nil {
		return err
	}
	if !domain.ValidReferences(ids, 24) {
		return domain.ErrInvalid
	}
	repository, err := service.taxonomyRepository()
	if err != nil {
		return err
	}
	return repository.SetCategories(ctx, actor, title, ids)
}
