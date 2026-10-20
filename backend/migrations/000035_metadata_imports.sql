-- Source identity is independent of canonical taxonomy IDs. Unresolved tags never create genres.
CREATE TABLE metadata_tag_mappings (
 source text NOT NULL CHECK(source IN('shikimori','anilist')),
 external_id text NOT NULL CHECK(length(external_id) BETWEEN 1 AND 160),
 name text NOT NULL CHECK(length(name) <= 160),
 status text NOT NULL CHECK(status IN('unmatched','ambiguous','mapped')),
 matches text[] NOT NULL DEFAULT '{}',
 genre_id text REFERENCES genres(id) ON DELETE SET NULL,
 category_id text REFERENCES categories(id) ON DELETE SET NULL,
 reviewed_by text REFERENCES users(id) ON DELETE SET NULL,
 updated_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(source,external_id), CHECK(genre_id IS NULL OR category_id IS NULL)
);
CREATE TABLE metadata_import_receipts (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 actor_id text REFERENCES users(id) ON DELETE SET NULL,
 project_id text REFERENCES projects(id) ON DELETE SET NULL,
 source text NOT NULL, external_id text NOT NULL, source_url text NOT NULL,
 snapshot jsonb NOT NULL, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX metadata_import_receipts_actor ON metadata_import_receipts(actor_id,created_at DESC);
