package postgres

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"qmanga/internal/catalog/application"
	"qmanga/internal/catalog/domain"
)

func (repository *Repository) Taxonomy(ctx context.Context, admin bool) (application.Taxonomy, error) {
	result := application.Taxonomy{Genres: []application.Genre{}, Categories: []application.Genre{}, Collections: []domain.Collection{}}
	var err error
	result.Genres, err = repository.Genres(ctx)
	if err != nil {
		return result, err
	}
	rows, err := repository.pool.Query(ctx, "SELECT id,name_kk,name_ru,slug FROM categories ORDER BY position,name_kk,id")
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var item application.Genre
		if err = rows.Scan(&item.ID, &item.Name, &item.NameRU, &item.Slug); err != nil {
			rows.Close()
			return result, err
		}
		result.Categories = append(result.Categories, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	condition := "is_visible"
	if admin {
		condition = "true"
	}
	rows, err = repository.pool.Query(ctx, `SELECT collection.id,collection.name_kk,collection.slug,collection.description,collection.position,collection.is_visible,ARRAY(SELECT mapping.title_id FROM collection_titles mapping JOIN titles title ON title.id=mapping.title_id WHERE mapping.collection_id=collection.id AND ($1 OR `+titleVisible+`) ORDER BY mapping.position) FROM editorial_collections collection WHERE `+condition+` ORDER BY collection.position,collection.id`, admin)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.Collection
		if err = rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.Position, &item.Visible, &item.TitleIDs); err != nil {
			return result, err
		}
		result.Collections = append(result.Collections, item)
	}
	return result, rows.Err()
}
func (repository *Repository) SaveTaxon(ctx context.Context, actor, kind string, item application.Genre) error {
	table := "genres"
	if kind == "categories" {
		table = "categories"
	}
	return repository.withAdmin(ctx, actor, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "INSERT INTO "+table+"(id,name_kk,slug) VALUES($1,$2,$3) ON CONFLICT(id) DO UPDATE SET name_kk=excluded.name_kk,slug=excluded.slug", item.ID, item.Name, item.Slug)
		if err != nil {
			return err
		}
		return audit(ctx, tx, actor, kind, item.ID)
	})
}
func (repository *Repository) DeleteTaxon(ctx context.Context, actor, kind, id string) error {
	table := "genres"
	if kind == "categories" {
		table = "categories"
	}
	if kind == "collections" {
		table = "editorial_collections"
	}
	return repository.withAdmin(ctx, actor, func(tx pgx.Tx) error {
		result, err := tx.Exec(ctx, "DELETE FROM "+table+" WHERE id=$1", id)
		if err != nil {
			return err
		}
		if result.RowsAffected() != 1 {
			return domain.ErrNotFound
		}
		_, err = tx.Exec(ctx, "INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id) VALUES($1,$2,'delete',$3,$4)", uuid.NewString(), actor, kind, id)
		return err
	})
}
func (repository *Repository) SaveCollection(ctx context.Context, actor string, item domain.Collection) error {
	return repository.withAdmin(ctx, actor, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO editorial_collections(id,name_kk,slug,description,position,is_visible) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(id) DO UPDATE SET name_kk=excluded.name_kk,slug=excluded.slug,description=excluded.description,position=excluded.position,is_visible=excluded.is_visible`, item.ID, item.Name, item.Slug, item.Description, item.Position, item.Visible)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "DELETE FROM collection_titles WHERE collection_id=$1", item.ID); err != nil {
			return err
		}
		for position, id := range item.TitleIDs {
			if _, err = tx.Exec(ctx, "INSERT INTO collection_titles(collection_id,title_id,position) VALUES($1,$2,$3)", item.ID, id, position); err != nil {
				return err
			}
		}
		return audit(ctx, tx, actor, "collection", item.ID)
	})
}
func (repository *Repository) SetCategories(ctx context.Context, actor, title string, ids []string) error {
	return repository.withAdmin(ctx, actor, func(tx pgx.Tx) error {
		var id string
		if err := tx.QueryRow(ctx, "SELECT id FROM titles WHERE id=$1 FOR UPDATE", title).Scan(&id); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "DELETE FROM title_categories WHERE title_id=$1", title); err != nil {
			return err
		}
		for _, category := range ids {
			if _, err := tx.Exec(ctx, "INSERT INTO title_categories(title_id,category_id) VALUES($1,$2)", title, category); err != nil {
				return err
			}
		}
		return audit(ctx, tx, actor, "title_categories", title)
	})
}
