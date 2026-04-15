-- Full schema for Sprint 1 (idempotent)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  username VARCHAR(50) UNIQUE NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255),
  google_id VARCHAR(255) UNIQUE,
  role VARCHAR(20) DEFAULT 'user',
  language VARCHAR(10) DEFAULT 'kz',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Genres
CREATE TABLE IF NOT EXISTS genres (
  id SERIAL PRIMARY KEY,
  slug VARCHAR(50) UNIQUE NOT NULL,
  name_kz VARCHAR(100) NOT NULL,
  name_en VARCHAR(100),
  name_ru VARCHAR(100),
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Titles
CREATE TABLE IF NOT EXISTS titles (
  id SERIAL PRIMARY KEY,
  title_kz VARCHAR(255) NOT NULL,
  title_ru VARCHAR(255),
  title_en VARCHAR(255),
  description_kz TEXT,
  description_ru TEXT,
  description_en TEXT,
  cover_url VARCHAR(255),
  type VARCHAR(20) DEFAULT 'manhwa',
  status VARCHAR(20) DEFAULT 'ongoing',
  is_published BOOLEAN DEFAULT false,
  deleted_at TIMESTAMPTZ,
  view_count INT DEFAULT 0,
  rating FLOAT DEFAULT 0,
  translator_id UUID REFERENCES users(id),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- title_genres
CREATE TABLE IF NOT EXISTS title_genres (
  title_id INT REFERENCES titles(id) ON DELETE CASCADE,
  genre_id INT REFERENCES genres(id) ON DELETE CASCADE,
  PRIMARY KEY (title_id, genre_id)
);

-- Chapters
CREATE TABLE IF NOT EXISTS chapters (
  id SERIAL PRIMARY KEY,
  title_id INT REFERENCES titles(id) ON DELETE CASCADE,
  volume_number INT DEFAULT 1,
  chapter_number DECIMAL(6,2) NOT NULL,
  chapter_name VARCHAR(255),
  is_published BOOLEAN DEFAULT false,
  deleted_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(title_id, volume_number, chapter_number)
);

-- Pages
CREATE TABLE IF NOT EXISTS pages (
  id SERIAL PRIMARY KEY,
  chapter_id INT REFERENCES chapters(id) ON DELETE CASCADE,
  order_index INT NOT NULL,
  image_key VARCHAR(255) NOT NULL,
  width INT,
  height INT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(chapter_id, order_index)
);

-- Reading History
CREATE TABLE IF NOT EXISTS reading_history (
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  title_id INT REFERENCES titles(id) ON DELETE CASCADE,
  last_chapter_id INT REFERENCES chapters(id),
  last_page INT DEFAULT 1,
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (user_id, title_id)
);

-- Bookmarks
CREATE TABLE IF NOT EXISTS bookmarks (
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  title_id INT REFERENCES titles(id) ON DELETE CASCADE,
  status VARCHAR(20) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (user_id, title_id)
);

-- Ratings
CREATE TABLE IF NOT EXISTS ratings (
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  title_id INT REFERENCES titles(id) ON DELETE CASCADE,
  score INT CHECK (score >= 1 AND score <= 10),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (user_id, title_id)
);

-- Comments
CREATE TABLE IF NOT EXISTS comments (
  id SERIAL PRIMARY KEY,
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  title_id INT REFERENCES titles(id) ON DELETE CASCADE,
  chapter_id INT REFERENCES chapters(id) ON DELETE SET NULL,
  content TEXT NOT NULL,
  is_deleted BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_titles_fts ON titles USING GIN (to_tsvector('simple', title_kz || ' ' || COALESCE(title_ru, '') || ' ' || COALESCE(title_en, '')));
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='titles' AND column_name='deleted_at') THEN
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_titles_is_published ON titles(is_published) WHERE deleted_at IS NULL';
  END IF;
END
$$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='chapters' AND column_name='deleted_at') THEN
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_chapters_is_published ON chapters(is_published) WHERE deleted_at IS NULL';
  END IF;
END
$$;
