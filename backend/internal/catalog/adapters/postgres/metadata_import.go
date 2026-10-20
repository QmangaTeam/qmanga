package postgres

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"qmanga/internal/catalog/application"
	"qmanga/internal/catalog/domain"
	"strings"
	"unicode"
)

func (r *Repository) AuthorizeImport(ctx context.Context, actor, project string) error {
	if project != "" {
		_, err := r.ChangeContext(ctx, actor, project)
		return err
	}
	allowed, err := r.IsAdmin(ctx, actor)
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrForbidden
	}
	return nil
}
func tagKey(value string) string {
	return strings.Join(strings.FieldsFunc(strings.ToLower(value), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) }), " ")
}
func (r *Repository) MatchImportTags(ctx context.Context, source string, tags []application.ExternalTag) (genres, categories []string, candidates []application.ImportCandidate, err error) {
	genres = []string{}
	categories = []string{}
	candidates = []application.ImportCandidate{}
	taxonomy, err := r.Taxonomy(ctx, false)
	if err != nil {
		return
	}
	type target struct{ id, kind string }
	aliases := map[string][]target{}
	for kind, items := range map[string][]application.Genre{"genres": taxonomy.Genres, "categories": taxonomy.Categories} {
		for _, item := range items {
			seen := map[string]bool{}
			for _, name := range []string{item.Name, item.NameRU, item.Slug} {
				key := tagKey(name)
				if key != "" && !seen[key] {
					aliases[key] = append(aliases[key], target{item.ID, kind})
					seen[key] = true
				}
			}
		}
	}
	seen := map[string]bool{}
	for _, tag := range tags {
		if tag.ID == "" || len(tag.ID) > 160 || len([]rune(tag.Name)) > 160 {
			continue
		}
		var genre, category string
		failure := r.pool.QueryRow(ctx, `SELECT coalesce(genre_id,''),coalesce(category_id,'') FROM metadata_tag_mappings WHERE source=$1 AND external_id=$2 AND status='mapped'`, source, tag.ID).Scan(&genre, &category)
		if failure != nil && failure != pgx.ErrNoRows {
			err = failure
			return
		}
		matches := map[target]bool{}
		if genre != "" {
			matches[target{genre, "genres"}] = true
		} else if category != "" {
			matches[target{category, "categories"}] = true
		} else {
			for _, name := range append([]string{tag.Name}, tag.Aliases...) {
				for _, match := range aliases[tagKey(name)] {
					matches[match] = true
				}
			}
		}
		candidate := application.ImportCandidate{Source: source, ID: tag.ID, Name: tag.Name, Status: "unmatched", Matches: []string{}}
		for match := range matches {
			candidate.Matches = append(candidate.Matches, match.id)
		}
		if len(matches) == 1 {
			candidate.Status = "mapped"
			for match := range matches {
				key := match.kind + match.id
				if !seen[key] {
					if match.kind == "genres" {
						genres = append(genres, match.id)
					} else {
						categories = append(categories, match.id)
					}
					seen[key] = true
				}
				if match.kind == "genres" {
					genre = match.id
				} else {
					category = match.id
				}
			}
		} else if len(matches) > 1 {
			candidate.Status = "ambiguous"
		}
		_, err = r.pool.Exec(ctx, `INSERT INTO metadata_tag_mappings(source,external_id,name,status,matches,genre_id,category_id) VALUES($1,$2,$3,$4,$5,nullif($6,''),nullif($7,'')) ON CONFLICT(source,external_id) DO UPDATE SET name=excluded.name,status=CASE WHEN metadata_tag_mappings.reviewed_by IS NOT NULL THEN metadata_tag_mappings.status ELSE excluded.status END,matches=excluded.matches,genre_id=CASE WHEN metadata_tag_mappings.reviewed_by IS NOT NULL THEN metadata_tag_mappings.genre_id ELSE excluded.genre_id END,category_id=CASE WHEN metadata_tag_mappings.reviewed_by IS NOT NULL THEN metadata_tag_mappings.category_id ELSE excluded.category_id END,updated_at=now()`, source, tag.ID, tag.Name, candidate.Status, candidate.Matches, genre, category)
		if err != nil {
			return
		}
		if candidate.Status != "mapped" {
			candidates = append(candidates, candidate)
		}
	}
	return
}
func (r *Repository) RecordImport(ctx context.Context, actor, project string, item application.ImportTitle) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `INSERT INTO metadata_import_receipts(actor_id,project_id,source,external_id,source_url,snapshot) VALUES($1,nullif($2,''),$3,$4,$5,$6)`, actor, project, item.Source, item.ID, item.URL, data)
	return err
}
func (r *Repository) ImportCandidates(ctx context.Context, actor string) ([]application.ImportCandidate, error) {
	if err := r.AuthorizeImport(ctx, actor, ""); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `SELECT source,external_id,name,status,matches FROM metadata_tag_mappings WHERE status <> 'mapped' OR (genre_id IS NULL AND category_id IS NULL) ORDER BY updated_at DESC,source,external_id LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []application.ImportCandidate{}
	for rows.Next() {
		var item application.ImportCandidate
		if err = rows.Scan(&item.Source, &item.ID, &item.Name, &item.Status, &item.Matches); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (r *Repository) ResolveImportCandidate(ctx context.Context, actor, source, id, kind, target string) error {
	return r.withAdmin(ctx, actor, func(tx pgx.Tx) error {
		table := "genres"
		column := "genre_id"
		other := "category_id"
		if kind == "categories" {
			table = "categories"
			column = "category_id"
			other = "genre_id"
		}
		var exists bool
		if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM "+table+" WHERE id=$1)", target).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return domain.ErrInvalid
		}
		result, err := tx.Exec(ctx, "UPDATE metadata_tag_mappings SET status='mapped',"+column+"=$3,"+other+"=NULL,reviewed_by=$4,updated_at=now() WHERE source=$1 AND external_id=$2", source, id, target, actor)
		if err != nil {
			return err
		}
		if result.RowsAffected() != 1 {
			return domain.ErrNotFound
		}
		metadata, _ := json.Marshal(map[string]string{"source": source, "external_id": id, "kind": kind, "target": target})
		_, err = tx.Exec(ctx, `INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id,metadata) VALUES($1,$2,'resolve','metadata_tag_mapping',$1,$3)`, uuid.NewString(), actor, metadata)
		return err
	})
}
