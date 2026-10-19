-- A receipt proves that this browser created this account; ordinary login grants no import.
CREATE TABLE registration_reading_imports (
 user_id varchar(36) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
 receipt_hash varchar(64) NOT NULL,
 expires_at timestamptz NOT NULL,
 completed_at timestamptz
);
CREATE INDEX registration_reading_import_expiry ON registration_reading_imports(expires_at);
