CREATE TABLE external_identities (
 issuer text NOT NULL,
 subject text NOT NULL CHECK(length(subject)>0 AND length(subject)<=255),
 user_id varchar(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(issuer,subject),
 UNIQUE(issuer,user_id)
);
CREATE TABLE google_auth_flows (
 state_hash varchar(64) PRIMARY KEY,
 nonce varchar(64) NOT NULL,
 verifier varchar(128) NOT NULL,
 return_path text NOT NULL,
 nickname varchar(160) NOT NULL DEFAULT '',
 expires_at timestamptz NOT NULL
);
CREATE INDEX google_auth_flows_expiry ON google_auth_flows(expires_at);
