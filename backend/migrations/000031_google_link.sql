ALTER TABLE google_auth_flows
 ADD COLUMN link_user_id varchar(36) REFERENCES users(id) ON DELETE CASCADE,
 ADD COLUMN link_session_hash varchar(64),
 ADD CONSTRAINT google_link_binding CHECK ((link_user_id IS NULL) = (link_session_hash IS NULL));
