ALTER TABLE comments ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE reports ADD COLUMN title_id VARCHAR(36) REFERENCES titles(id),ADD COLUMN release_id VARCHAR(36) REFERENCES releases(id),ADD COLUMN page_id VARCHAR(36),ADD COLUMN resolved_at TIMESTAMPTZ;
CREATE INDEX comments_thread_page ON comments(title_id,release_id,created_at DESC,id DESC);
CREATE INDEX comments_parent ON comments(parent_id,created_at,id);
CREATE INDEX comments_user_limit ON comments(user_id,created_at);
CREATE INDEX reports_user_limit ON reports(user_id,created_at);
ALTER TABLE notifications ADD COLUMN event_id VARCHAR(36),ADD COLUMN dedupe_key VARCHAR(128);
CREATE UNIQUE INDEX notification_delivery_once ON notifications(user_id,dedupe_key);
CREATE INDEX notification_cursor ON notifications(user_id,created_at DESC,id DESC);
