package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"qmanga/internal/publishing/application"
	"qmanga/internal/publishing/domain"
	"strings"
)

// A correction owns new page and variant records but shares immutable object bytes.
// Prefer the last rejected attempt so corrections do not discard submitted work.
func cloneRevision(ctx context.Context, tx pgx.Tx, id, release string) error {
	var source string
	err := tx.QueryRow(ctx, `SELECT id FROM revisions WHERE release_id=$1 AND id<>$2 AND status IN ('published','changes_requested') ORDER BY revision_num DESC LIMIT 1`, release, id).Scan(&source)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE revisions destination SET source_note=source.source_note,credits=source.credits,rights_confirmed=source.rights_confirmed,rights_confirmed_by=source.rights_confirmed_by,preparation_method=source.preparation_method FROM revisions source WHERE destination.id=$1 AND source.id=$2`, id, source); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `SELECT id FROM pages WHERE revision_id=$1 AND status='ready' ORDER BY position`, source)
	if err != nil {
		return err
	}
	var originals []string
	for rows.Next() {
		var old string
		if err = rows.Scan(&old); err != nil {
			rows.Close()
			return err
		}
		originals = append(originals, old)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, old := range originals {
		page := uuid.NewString()
		if _, err = tx.Exec(ctx, `INSERT INTO pages(id,revision_id,position,width,height,byte_size,source_key,checksum,status,image_format,upload_key,original_name) SELECT $1,$2,position,width,height,byte_size,source_key,checksum,status,image_format,upload_key,original_name FROM pages WHERE id=$3`, page, id, old); err != nil {
			return err
		}
		// Generate UUIDs in Go; PostgreSQL installations need no UUID extension.
		variants, err := tx.Query(ctx, `SELECT id FROM page_variants WHERE page_id=$1`, old)
		if err != nil {
			return err
		}
		var ids []string
		for variants.Next() {
			var v string
			if err = variants.Scan(&v); err != nil {
				variants.Close()
				return err
			}
			ids = append(ids, v)
		}
		err = variants.Err()
		variants.Close()
		if err != nil {
			return err
		}
		for _, v := range ids {
			if _, err = tx.Exec(ctx, `INSERT INTO page_variants(id,page_id,tier,object_key,width,height,byte_size,offset_y,checksum,mime) SELECT $1,$2,tier,object_key,width,height,byte_size,offset_y,checksum,mime FROM page_variants WHERE id=$3`, uuid.NewString(), page, v); err != nil {
				return err
			}
		}
	}
	if len(originals) > 0 {
		_, err = tx.Exec(ctx, `UPDATE revisions SET status='ready' WHERE id=$1`, id)
	}
	return err
}

func (repository *Repository) SelectCover(ctx context.Context, actor, project, page string) error {
	return repository.withProject(ctx, actor, project, "submit", func(tx pgx.Tx) error {
		var title, previous string
		if err := tx.QueryRow(ctx, `SELECT title.id,coalesce(title.cover_asset_id,'') FROM titles title JOIN projects project ON project.title_id=title.id WHERE project.id=$1 FOR UPDATE OF title`, project).Scan(&title, &previous); err != nil {
			return err
		}
		if page == "" {
			if _, err := tx.Exec(ctx, `UPDATE titles SET cover_asset_id='',cover_managed=true,updated_at=now() WHERE id=$1`, title); err != nil {
				return err
			}
			metadata, _ := json.Marshal(map[string]any{"before": map[string]string{"cover_asset_id": previous}, "after": map[string]string{"cover_asset_id": ""}})
			_, err := tx.Exec(ctx, `INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id,metadata) VALUES($1,$2,'remove_cover','title',$3,$4)`, uuid.NewString(), actor, title, metadata)
			return err
		}
		id := uuid.NewString()
		result, err := tx.Exec(ctx, `INSERT INTO title_cover_assets(id,title_id,object_key,mime,checksum,width,height,byte_size,selected_by) SELECT $1,$2,variant.object_key,variant.mime,variant.checksum,variant.width,variant.height,variant.byte_size,$3 FROM page_variants variant JOIN pages page ON page.id=variant.page_id JOIN revisions revision ON revision.id=page.revision_id JOIN releases publication ON publication.id=revision.release_id WHERE page.id=$4 AND publication.project_id=$5 AND page.status='ready' AND variant.tier='eco' ORDER BY variant.offset_y LIMIT 1`, id, title, actor, page, project)
		if err != nil {
			return err
		}
		if result.RowsAffected() != 1 {
			return domain.ErrNotFound
		}
		if _, err = tx.Exec(ctx, `UPDATE titles SET cover_asset_id=$2,cover_managed=true,updated_at=now() WHERE id=$1`, title, id); err != nil {
			return err
		}
		metadata, _ := json.Marshal(map[string]any{"before": map[string]string{"cover_asset_id": previous}, "after": map[string]string{"cover_asset_id": id}})
		_, err = tx.Exec(ctx, `INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id,metadata) VALUES($1,$2,'select_cover','title',$3,$4)`, uuid.NewString(), actor, title, metadata)
		return err
	})
}

func (repository *Repository) CanUploadCover(ctx context.Context, actor, project string) error {
	return repository.withProject(ctx, actor, project, "submit", func(pgx.Tx) error { return nil })
}

func (repository *Repository) SaveCover(ctx context.Context, actor, project string, cover application.CoverAsset) error {
	if cover.ID == "" || cover.ObjectKey == "" || cover.Mime != "image/webp" || len(cover.Checksum) != 64 || cover.Width <= 0 || cover.Height <= 0 || cover.Bytes <= 0 {
		return domain.ErrInvalidImage
	}
	return repository.withProject(ctx, actor, project, "submit", func(tx pgx.Tx) error {
		var title, previous string
		if err := tx.QueryRow(ctx, `SELECT title.id,coalesce(title.cover_asset_id,'') FROM titles title JOIN projects project ON project.title_id=title.id WHERE project.id=$1 FOR UPDATE OF title`, project).Scan(&title, &previous); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO title_cover_assets(id,title_id,object_key,mime,checksum,width,height,byte_size,selected_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, cover.ID, title, cover.ObjectKey, cover.Mime, cover.Checksum, cover.Width, cover.Height, cover.Bytes, actor); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE titles SET cover_asset_id=$2,cover_managed=true,updated_at=now() WHERE id=$1`, title, cover.ID); err != nil {
			return err
		}
		metadata, _ := json.Marshal(map[string]any{"before": map[string]string{"cover_asset_id": previous}, "after": map[string]string{"cover_asset_id": cover.ID}})
		_, err := tx.Exec(ctx, `INSERT INTO audit_logs(id,user_id,action,entity_type,entity_id,metadata) VALUES($1,$2,'upload_cover','title',$3,$4)`, uuid.NewString(), actor, title, metadata)
		return err
	})
}

func (repository *Repository) RetryPage(ctx context.Context, actor, id, page string) error {
	return repository.withRevision(ctx, actor, id, "upload", func(tx pgx.Tx, revision domain.Revision) error {
		if !editable(revision) {
			return domain.ErrImmutable
		}
		// Upload failures require reselecting bytes. Processing failures retain the source.
		result, err := tx.Exec(ctx, `UPDATE pages SET status='queued',failure_code='' WHERE id=$1 AND revision_id=$2 AND status='failed' AND failure_code='processing_failed'`, page, id)
		if err != nil {
			return err
		}
		if result.RowsAffected() != 1 {
			return domain.ErrTransition
		}
		if _, err = tx.Exec(ctx, `UPDATE jobs SET status='pending',attempts=0,available_at=now(),leased_until=NULL,last_error=NULL WHERE dedupe_key=$1 AND status='failed'`, "page:"+page); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE revisions SET status='processing',version=version+1,updated_at=now() WHERE id=$1`, id)
		return err
	})
}

func (repository *Repository) Restore(ctx context.Context, actor, release, reason string) error {
	if strings.TrimSpace(reason) == "" || len(reason) > 4000 {
		return domain.ErrTransition
	}
	var id string
	if err := repository.pool.QueryRow(ctx, `SELECT current_revision_id FROM releases WHERE id=$1 AND status='withdrawn'`, release).Scan(&id); err != nil {
		return translate(err)
	}
	return repository.withRevision(ctx, actor, id, "review", func(tx pgx.Tx, revision domain.Revision) error {
		if revision.State() != domain.Published {
			return domain.ErrTransition
		}
		result, err := tx.Exec(ctx, `UPDATE releases SET status='published',updated_at=now() WHERE id=$1 AND current_revision_id=$2 AND status='withdrawn'`, release, id)
		if err != nil {
			return err
		}
		if result.RowsAffected() != 1 {
			return domain.ErrConflict
		}
		return audit(ctx, tx, actor, "restore", release, reason)
	})
}

// Automatic first cover is a separate safe image too. Later chapters cannot replace it.
func ensureDefaultCover(ctx context.Context, tx pgx.Tx, actor, chapter, revision string) error {
	var title, existing string
	var managed bool
	if err := tx.QueryRow(ctx, `SELECT title.id,coalesce(title.cover_asset_id,''),title.cover_managed FROM titles title JOIN chapters chapter ON chapter.title_id=title.id WHERE chapter.id=$1 FOR UPDATE OF title`, chapter).Scan(&title, &existing, &managed); err != nil {
		return err
	}
	if existing != "" || managed {
		_, err := tx.Exec(ctx, `UPDATE titles SET updated_at=now() WHERE id=$1`, title)
		return err
	}
	id := uuid.NewString()
	result, err := tx.Exec(ctx, `INSERT INTO title_cover_assets(id,title_id,object_key,mime,checksum,width,height,byte_size,selected_by) SELECT $1,$2,variant.object_key,variant.mime,variant.checksum,variant.width,variant.height,variant.byte_size,$3 FROM page_variants variant JOIN pages page ON page.id=variant.page_id WHERE page.revision_id=$4 AND page.status='ready' AND variant.tier='eco' ORDER BY page.position,variant.offset_y LIMIT 1`, id, title, actor, revision)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return domain.ErrTransition
	}
	_, err = tx.Exec(ctx, `UPDATE titles SET cover_asset_id=$2,updated_at=now() WHERE id=$1`, title, id)
	return err
}
