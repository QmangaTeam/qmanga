ALTER TABLE sessions ADD COLUMN csrf_hash VARCHAR(64) NOT NULL DEFAULT '';
CREATE INDEX sessions_user_idx ON sessions(user_id);
CREATE INDEX sessions_expiry_idx ON sessions(expires_at);
CREATE UNIQUE INDEX tokens_user_kind_idx ON tokens(user_id,token_type);
CREATE UNIQUE INDEX users_normalized_email_idx ON users(lower(email));
