CREATE TABLE account_preferences (
 user_id varchar(36) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
 version bigint NOT NULL DEFAULT 1 CHECK(version>0),
 bio varchar(500) NOT NULL DEFAULT '', public_profile boolean NOT NULL DEFAULT false,
 show_library boolean NOT NULL DEFAULT false, show_activity boolean NOT NULL DEFAULT false,
 allow_follow boolean NOT NULL DEFAULT true, notify_releases boolean NOT NULL DEFAULT true,
 notify_replies boolean NOT NULL DEFAULT true, notify_followers boolean NOT NULL DEFAULT false,
 email_digest boolean NOT NULL DEFAULT false,
 digest_after timestamptz NOT NULL DEFAULT now()+interval '7 days'
);
CREATE TABLE account_follows (
 follower_id varchar(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 followee_id varchar(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(follower_id,followee_id), CHECK(follower_id<>followee_id)
);
CREATE INDEX account_follows_followee ON account_follows(followee_id,follower_id);
