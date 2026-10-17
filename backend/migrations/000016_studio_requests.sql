ALTER TABLE teams ADD COLUMN review_status VARCHAR(16) NOT NULL DEFAULT 'pending' CHECK (review_status IN ('pending','approved','rejected'));
ALTER TABLE teams ADD COLUMN review_note TEXT NOT NULL DEFAULT '';
UPDATE teams SET review_status='approved' WHERE is_approved;

CREATE TABLE project_requests (
 id VARCHAR(36) PRIMARY KEY,
 team_id VARCHAR(36) NOT NULL REFERENCES teams(id),
 requested_by VARCHAR(36) NOT NULL REFERENCES users(id),
 title_id VARCHAR(36) REFERENCES titles(id),
 target_key TEXT NOT NULL,
 name_kk TEXT NOT NULL DEFAULT '',
 name_original TEXT NOT NULL DEFAULT '',
 title_slug VARCHAR(128) NOT NULL DEFAULT '',
 kind VARCHAR(16) NOT NULL DEFAULT 'manga',
 description_kk TEXT NOT NULL DEFAULT '',
 source_url TEXT NOT NULL DEFAULT '',
 note TEXT NOT NULL,
 status VARCHAR(16) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected','cancelled')),
 resolution TEXT NOT NULL DEFAULT '',
 reviewed_by VARCHAR(36) REFERENCES users(id),
 project_id VARCHAR(36) REFERENCES projects(id),
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX one_pending_project_request ON project_requests(team_id,target_key) WHERE status='pending';
CREATE INDEX project_request_queue ON project_requests(status,created_at,id);
CREATE INDEX project_request_team ON project_requests(team_id,created_at DESC,id DESC);
