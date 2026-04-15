-- Fix existing tables: add missing columns and indexes if necessary
-- Add missing columns if older initial migration skipped them
ALTER TABLE titles ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE titles ADD COLUMN IF NOT EXISTS is_published BOOLEAN DEFAULT false;
ALTER TABLE chapters ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE chapters ADD COLUMN IF NOT EXISTS is_published BOOLEAN DEFAULT false;

-- Add additional title columns that may be missing from older initial migration
ALTER TABLE titles ADD COLUMN IF NOT EXISTS description_kz TEXT;
ALTER TABLE titles ADD COLUMN IF NOT EXISTS description_ru TEXT;
ALTER TABLE titles ADD COLUMN IF NOT EXISTS description_en TEXT;
ALTER TABLE titles ADD COLUMN IF NOT EXISTS view_count INT DEFAULT 0;
ALTER TABLE titles ADD COLUMN IF NOT EXISTS rating FLOAT DEFAULT 0;
ALTER TABLE titles ADD COLUMN IF NOT EXISTS translator_id UUID REFERENCES users(id);
ALTER TABLE titles ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

-- Ensure indexes exist (guarded by column existence where relevant)
DO $$
BEGIN
	IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='titles' AND column_name='is_published') THEN
		EXECUTE 'CREATE INDEX IF NOT EXISTS idx_titles_is_published ON titles(is_published) WHERE deleted_at IS NULL';
	END IF;
END
$$;

DO $$
BEGIN
	IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='chapters' AND column_name='is_published') THEN
		EXECUTE 'CREATE INDEX IF NOT EXISTS idx_chapters_is_published ON chapters(is_published) WHERE deleted_at IS NULL';
	END IF;
END
$$;
