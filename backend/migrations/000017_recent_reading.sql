CREATE INDEX reading_recent_activity ON reading_progress(user_id,updated_at DESC,title_id DESC);
