CREATE TABLE avatars (
 owner_kind text NOT NULL CHECK(owner_kind IN ('user','team')),
 owner_id varchar(36) NOT NULL,
 object_key text NOT NULL UNIQUE,
 bytes bigint NOT NULL CHECK(bytes > 0 AND bytes <= 5242880),
 updated_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(owner_kind,owner_id)
);
