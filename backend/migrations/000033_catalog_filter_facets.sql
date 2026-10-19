-- Migration: 000033_catalog_filter_facets.sql
-- Description: Group order and indexes for the catalogue filter panel.
--
-- The filter panel follows the reference order of the Senkuro taxonomy groups instead of
-- alphabetical order. Publication facets (year, source, country, formats) live in the
-- titles.publication JSONB document; expression indexes keep the range and equality filters
-- cheap as the catalogue grows.

ALTER TABLE categories ADD COLUMN position integer NOT NULL DEFAULT 0 CHECK (position >= 0);

UPDATE categories SET position = ordering.position
FROM (VALUES
    ('main-cast', 1),
    ('traits', 2),
    ('setting', 3),
    ('themes', 4),
    ('elements', 5),
    ('demographics', 6)
) AS ordering(slug, position)
WHERE categories.slug = ordering.slug;

CREATE INDEX categories_position_idx ON categories(position, slug);
CREATE INDEX titles_publication_idx ON titles USING gin (publication jsonb_path_ops);
CREATE INDEX titles_source_idx ON titles ((publication->>'source'));
CREATE INDEX titles_country_idx ON titles ((publication->>'country'));
