-- Notifications owns retry/delivery state; account preference changes remain in Identity.
CREATE TABLE notification_digest_state (
 user_id varchar(36) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
 last_sent_at timestamptz,
 available_at timestamptz NOT NULL DEFAULT now()
);
