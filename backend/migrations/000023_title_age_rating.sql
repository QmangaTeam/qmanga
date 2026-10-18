-- Additive: old applications can omit this column; 0 means unclassified, never a presumed 12+.
ALTER TABLE titles ADD COLUMN age_rating SMALLINT NOT NULL DEFAULT 0
    CHECK (age_rating IN (0, 12, 16, 18));
CREATE INDEX titles_age_rating_updated ON titles(age_rating, updated_at DESC, id DESC);
