-- Migration: 000034_catalog_filter_indexes.sql
-- Description: Correct expression index for the publication formats filter.
--
-- jsonb_path_ops supports @>, @? and @@ but not the ?| operator, and an index on the
-- whole publication document cannot serve an operator applied to publication->'formats'.
-- The format filter therefore needs a jsonb_ops expression index. The previous
-- whole-document index was not used by any catalogue query.

DROP INDEX IF EXISTS titles_publication_idx;

CREATE INDEX titles_formats_idx ON titles USING gin ((publication->'formats') jsonb_ops);
