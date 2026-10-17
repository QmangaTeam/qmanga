ALTER TABLE reading_events ADD COLUMN request_hash CHAR(64),ADD COLUMN accepted_version BIGINT NOT NULL DEFAULT 0;
