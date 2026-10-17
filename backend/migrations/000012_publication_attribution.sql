ALTER TABLE revisions ADD COLUMN source_note TEXT NOT NULL DEFAULT '';
ALTER TABLE revisions ADD COLUMN credits TEXT NOT NULL DEFAULT '';
ALTER TABLE revisions ADD COLUMN rights_confirmed BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE revisions ADD COLUMN rights_confirmed_by TEXT REFERENCES users(id);
