ALTER TABLE titles ADD COLUMN banner_asset_id text NOT NULL DEFAULT '';
CREATE TABLE title_change_requests (
 id text PRIMARY KEY,
 title_id varchar(36) NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
 project_id varchar(36) REFERENCES projects(id) ON DELETE SET NULL,
 team_id varchar(36) REFERENCES teams(id) ON DELETE SET NULL,
 requested_by varchar(36) REFERENCES users(id) ON DELETE SET NULL,
 reviewed_by varchar(36) REFERENCES users(id) ON DELETE SET NULL,
 before_data jsonb NOT NULL CHECK(jsonb_typeof(before_data)='object'),
 proposed_data jsonb NOT NULL CHECK(jsonb_typeof(proposed_data)='object'),
 base_version text NOT NULL CHECK(base_version ~ '^[a-f0-9]{64}$'),
 reason text NOT NULL CHECK(length(reason) BETWEEN 3 AND 2000),
 links jsonb NOT NULL DEFAULT '[]' CHECK(jsonb_typeof(links)='array' AND jsonb_array_length(links)<=10),
 cover_action text NOT NULL CHECK(cover_action IN('keep','replace','remove')),
 banner_action text NOT NULL CHECK(banner_action IN('keep','replace','remove')),
 cover_page_id text NOT NULL DEFAULT '',
 cover_asset_id text NOT NULL DEFAULT '',
 banner_asset_id text NOT NULL DEFAULT '',
 status text NOT NULL DEFAULT 'pending' CHECK(status IN('pending','approved','rejected')),
 resolution text NOT NULL DEFAULT '' CHECK(length(resolution)<=2000),
 created_at timestamptz NOT NULL DEFAULT now(),
 reviewed_at timestamptz,
 CHECK((status='pending')=(reviewed_at IS NULL)),
 CHECK(status<>'rejected' OR length(resolution)>=3),
 CHECK((cover_action='replace')=(cover_asset_id<>'')),
 CHECK((banner_action='replace')=(banner_asset_id<>''))
);
CREATE UNIQUE INDEX title_change_pending_project ON title_change_requests(project_id) WHERE status='pending';
CREATE INDEX title_change_queue ON title_change_requests(status,created_at DESC,id DESC);
CREATE INDEX title_change_project_history ON title_change_requests(project_id,created_at DESC,id DESC);
CREATE INDEX title_change_team_rate ON title_change_requests(team_id,created_at);
CREATE TABLE title_change_assets (
 id text PRIMARY KEY,
 request_id text NOT NULL REFERENCES title_change_requests(id) ON DELETE CASCADE,
 purpose text NOT NULL CHECK(purpose IN('cover','banner','evidence')),
 object_key text NOT NULL UNIQUE,
 mime text NOT NULL CHECK(mime='image/webp'),
 checksum text NOT NULL CHECK(checksum ~ '^[a-f0-9]{64}$'),
 width integer NOT NULL CHECK(width>0),
 height integer NOT NULL CHECK(height>0),
 byte_size bigint NOT NULL CHECK(byte_size>0),
 position integer NOT NULL CHECK(position>0),
 UNIQUE(request_id,position),
 CHECK((purpose='cover' AND width<=720 AND height<=1200 AND byte_size<=26214400)
 OR (purpose='banner' AND width<=1600 AND height<=800 AND byte_size<=26214400)
 OR (purpose='evidence' AND width<=2000 AND height<=3000 AND byte_size<=5242880))
);
CREATE UNIQUE INDEX title_change_single_artwork ON title_change_assets(request_id,purpose) WHERE purpose IN('cover','banner');
