CREATE TABLE avatar_garbage (
 object_key text PRIMARY KEY,
 ready_at timestamptz NOT NULL
);
CREATE INDEX avatar_garbage_ready ON avatar_garbage(ready_at);
