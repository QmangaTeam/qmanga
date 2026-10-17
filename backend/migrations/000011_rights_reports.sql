ALTER TABLE reports ADD COLUMN contact_email VARCHAR(254) NOT NULL DEFAULT '';
ALTER TABLE reports ADD COLUMN receipt_hash VARCHAR(64) NOT NULL DEFAULT '';
CREATE INDEX report_contact_limit ON reports(contact_email,created_at) WHERE contact_email<>'';
CREATE INDEX report_owner_cursor ON reports(user_id,created_at DESC,id DESC);
