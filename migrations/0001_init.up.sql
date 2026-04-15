-- Initial schema: users and titles (minimal)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  username VARCHAR(50) UNIQUE NOT NULL,
  email VARCHAR(255) UNIQUE,
  role VARCHAR(20) DEFAULT 'user',
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS titles (
  id SERIAL PRIMARY KEY,
  title_kz VARCHAR(255) NOT NULL,
  title_ru VARCHAR(255),
  title_en VARCHAR(255),
  description TEXT,
  cover_url VARCHAR(255),
  type VARCHAR(20) DEFAULT 'manhwa',
  status VARCHAR(20) DEFAULT 'ongoing',
  created_at TIMESTAMPTZ DEFAULT NOW()
);
