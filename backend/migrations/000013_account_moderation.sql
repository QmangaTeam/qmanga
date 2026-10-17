ALTER TABLE users ADD COLUMN is_blocked BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN management_version BIGINT NOT NULL DEFAULT 0;
CREATE INDEX audit_cursor ON audit_logs(created_at DESC,id DESC);
