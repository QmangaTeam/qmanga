CREATE TABLE categories (
 id text PRIMARY KEY,
 name_kk text NOT NULL CHECK (length(name_kk) BETWEEN 1 AND 80),
 slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND length(slug)<=64)
);
CREATE TABLE title_categories (
 title_id text NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
 category_id text NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
 PRIMARY KEY(title_id,category_id)
);
CREATE INDEX title_categories_category ON title_categories(category_id,title_id);
CREATE TABLE editorial_collections (
 id text PRIMARY KEY,
 name_kk text NOT NULL CHECK (length(name_kk) BETWEEN 1 AND 80),
 slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND length(slug)<=64),
 description text NOT NULL DEFAULT '' CHECK (length(description)<=4000),
 position integer NOT NULL DEFAULT 0 CHECK(position>=0),
 is_visible boolean NOT NULL DEFAULT false
);
CREATE TABLE collection_titles (
 collection_id text NOT NULL REFERENCES editorial_collections(id) ON DELETE CASCADE,
 title_id text NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
 position integer NOT NULL CHECK(position>=0),
 PRIMARY KEY(collection_id,title_id),
 UNIQUE(collection_id,position)
);
CREATE INDEX collection_titles_title ON collection_titles(title_id,collection_id);
