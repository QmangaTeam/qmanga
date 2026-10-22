CREATE TABLE api_tokens (
 id varchar(36) PRIMARY KEY,
 user_id varchar(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 name varchar(100) NOT NULL,
 prefix varchar(16) NOT NULL,
 token_hash varchar(64) NOT NULL UNIQUE,
 created_at timestamptz NOT NULL,
 expires_at timestamptz NOT NULL,
 last_used_at timestamptz,
 revoked_at timestamptz,
 direct_publish boolean NOT NULL DEFAULT false,
 CHECK (expires_at > created_at)
);
CREATE INDEX api_tokens_owner ON api_tokens(user_id,created_at DESC);
