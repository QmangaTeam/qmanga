ALTER TABLE library_entries ADD COLUMN selection_version BIGINT NOT NULL DEFAULT 1;
ALTER TABLE library_entries ADD COLUMN is_removed BOOLEAN NOT NULL DEFAULT false;
