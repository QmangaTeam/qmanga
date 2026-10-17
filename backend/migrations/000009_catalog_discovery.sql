ALTER TABLE titles ADD COLUMN authors TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE titles ADD COLUMN is_featured BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX title_created_cursor ON titles(created_at DESC,id DESC);
CREATE INDEX release_update_cursor ON releases(updated_at DESC,id DESC) WHERE status='published';
