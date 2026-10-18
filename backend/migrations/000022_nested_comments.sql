ALTER TABLE comments ADD COLUMN root_id VARCHAR(36) REFERENCES comments(id) ON DELETE CASCADE;
UPDATE comments SET root_id=parent_id WHERE parent_id IS NOT NULL;
CREATE INDEX comments_root_page ON comments(root_id,created_at,id) WHERE root_id IS NOT NULL;
