# PLAN: Qmanga — Платформа манги и манхвы (MVP)

**Версия:** 0.1  
**Последнее обновление:** 2026-04-13  
**Статус:** В разработке (Planning phase)

---

## 1. РЕШЕНИЯ, ПРИНЯТЫЕ НА СОЗВОНЕ

### 1.1 Аутентификация

- **JWT + httpOnly Cookie**: 30 дней TTL (достаточно для MVP)
- **Google OAuth2**: Обязательно (Юзеры ленивые, не хотят заполнять формы)
- **Регистрация**: Автоматическая (при первом входе через Google создаем запись в БД)
- **Username**: Берем Name из Google. Если уже занят → добавляем 4 рандомные цифры (john_1234)
- **Profile Edit**: Язык выбора по UI (по умолчанию KZ), юзер может вручную сменить ник один раз в профиле
- ~~GitHub~~ ~~2FA~~ ~~Refresh tokens~~
- **Хранилище токена**: Redis (черный список при logout)

### 1.2 Ридер (Читалка)

- **Dark Mode**: На всем сайте, обязательно
- ~~TTS/OCR~~ ~~Blue Light Filter~~: Позже (слишком дорого для MVP)
- **Ленивая загрузка**: 5-7 изображений в DOM (IntersectionObserver)
- **LocalStorage**: Настройки ридера (качество, размер страницы, история)

### 1.3 Функциональность

- **Комментарии**: Только на Title и Chapter (не на каждой странице). Ручная модерация (админ удаляет спам)
- **Уведомления**: v2 (после запуска ридера). Для MVP: простой индикатор "New" в закладках
- **Рейтинг**: MVP (Live SQL AVG()). Redis кэш позже когда онлайн > 1000 юзеров
- **Фильтр по переводчикам**: MVP (Брендинг команд)
- **Локализация**: Fallback KZ → EN → RU. Жанры локализируются
- **Тестовые данные**: Пара глав для обкатки ридера

### 1.4 Деплой и DevOps

- **Хостинг**: GCore (Алматы) или PS.kz (пинг = конверсия)
- **CDN**: Go Proxy + Cloudflare (прямой доступ к MinIO — дыра в безопасности)
- **MinIO**: Приватный, доступ только через Go Image Proxy
- **CI/CD**: GitHub Actions
- **Контейнеризация**: Docker Compose (Go + PostgreSQL + Redis + MinIO)

### 1.5 Код и API

- **Strict Typing**: Go структуры ↔ Svelte интерфейсы (contract)
- **Image Proxy**: bimg/libvips для ресайза (Go service с Redis + S3 кэшированием)
- **Local First**: LocalStorage для UI настроек с самого старта
- **CORS**: Gin middleware для фронта (localhost:5173 в dev, domain.kz в prod)
  ```go
  // Dev
  AllowOrigins: []string{"http://localhost:5173"}
  // Prod
  AllowOrigins: []string{"https://qmanga.kz"}
  ```

---

## 2. СПРИНТЫ (Пересмотренные)

### Sprint 1: Инфра + БД + Auth

- [ ] Docker Compose: PostgreSQL 17, Redis, MinIO
- [ ] PostgreSQL DDL миграции (users, titles, chapters, pages, bookmarks, reading_history)
- [ ] JWT middleware (Gin)
- [ ] Google OAuth2 интеграция
- [ ] Простой endpoint для проверки (`GET /api/v1/health`)

### Sprint 2: Core API + Image Pipeline

- [ ] REST endpoints для каталога (`GET /api/v1/titles`, фильтры, поиск FTS)
- [ ] Endpoint для детали тайтла (`GET /api/v1/titles/:id`)
- [ ] Endpoint для получения глав (`GET /api/v1/chapters/:id`)
- [ ] **Image Pipeline**: `POST /api/admin/chapters/upload` (ZIP → фс конверсия через bimg)
- [ ] Image Proxy: `GET /api/v1/image/:key?q=low|medium` с кэшированием (Redis → S3 → Generate → Cache)
- [ ] Admin endpoints (CRUD тайтлов)

### Sprint 3: Frontend MVP

- [ ] SvelteKit Project setup (SPA mode)
- [ ] Layout: Header, Sidebar, Dark Mode toggle
- [ ] Каталог (Title List) + фильтры + поиск
- [ ] Страница тайтла (Metadata + список глав)
- [ ] API Client layer (контрактны запросы)
- [ ] LocalStorage Store для настроек

### Sprint 4: Ридер (Критичный)

- [ ] Читалка (longstrip, IntersectionObserver для ленивой загрузки)
- [ ] FAB (плавающая кнопка) + меню навигации
- [ ] Синхронизация прогресса (`POST /api/v1/user/history`)
- [ ] Локальные настройки (качество, размер, сохранение позиции)

### Sprint 5: User Features + Polish

- [ ] История чтения (`GET /api/v1/user/history`)
- [ ] Закладки (`POST /api/v1/user/bookmarks`)
- [ ] Рейтинг система (DB + UI для голосования)
- [ ] Фильтр по переводчикам (UI + API)
- [ ] Тестирование на 3G (Lighthouse)

---

## 3. КОНТРАКТНЫЕ ТИПЫ (Go ↔ Svelte)

> Эти типы должны совпадать точно. Генерируем один раз, используем везде.

### User (Auth)

```go
type User struct {
  ID        string    `json:"id"`           // UUID
  Username  string    `json:"username"`
  Email     string    `json:"email"`
  Role      string    `json:"role"`         // 'user', 'translator', 'admin'
  CreatedAt time.Time `json:"created_at"`
}
```

### Title (Манхва)

```go
type Title struct {
  ID          int       `json:"id"`
  TitleKz     string    `json:"title_kz"`
  TitleRu     string    `json:"title_ru"`
  TitleEn     string    `json:"title_en"`
  Description string    `json:"description"`
  CoverURL    string    `json:"cover_url"`
  Type        string    `json:"type"`        // 'manhwa', 'manga'
  Status      string    `json:"status"`      // 'ongoing', 'completed', 'dropped'
  ViewCount   int       `json:"view_count"`
  Rating      float64   `json:"rating"`
  CreatedAt   time.Time `json:"created_at"`
}
```

### Chapter

```go
type Chapter struct {
  ID            int       `json:"id"`
  TitleID       int       `json:"title_id"`
  VolumeNumber  int       `json:"volume_number"`
  ChapterNumber float64   `json:"chapter_number"` // 1, 1.5, 2
  ChapterName   string    `json:"chapter_name"`
  PageCount     int       `json:"page_count"`
  CreatedAt     time.Time `json:"created_at"`
}
```

### Page (Изображение)

```go
type Page struct {
  ID         int    `json:"id"`
  ChapterID  int    `json:"chapter_id"`
  OrderIndex int    `json:"order_index"`
  ImageKey   string `json:"image_key"`        // S3 path
  ImageURL   string `json:"image_url"`        // Полный URL для фронта
  Width      int    `json:"width"`           // Оригинальное разрешение
  Height     int    `json:"height"`
}
```

### ReadingHistory

```go
type ReadingHistory struct {
  UserID        string    `json:"user_id"`
  TitleID       int       `json:"title_id"`
  LastChapterID int       `json:"last_chapter_id"`
  LastPage      int       `json:"last_page"`
  UpdatedAt     time.Time `json:"updated_at"`
}
```

---

## 4. API ENDPOINTS (MVP)

### Public

- `GET /api/v1/titles` — каталог (фильтры: status, type, tag, genre; поиск; пагинация)
- `GET /api/v1/titles/:id` — детали тайтла + список глав (с локализацией: KZ→EN→RU fallback)
- `GET /api/v1/chapters/:id` — данные главы + массив Pages
- `GET /api/v1/image/:key?q=low|medium` — изображение (AVIF, on-demand resizing)
- `GET /api/v1/languages` — список доступных языков

### Auth Required (JWT)

- `POST /api/v1/user/history` — обновить прогресс
- `GET /api/v1/user/history` — получить историю (синхронизация с БД при входе)
- `POST /api/v1/user/bookmarks` — добавить/изменить закладку
- `PATCH /api/v1/user/profile` — изменить username (один раз)
- `PATCH /api/v1/user/language` — выбрать язык (сохраняется в profile)

### Admin Only

- `POST /api/admin/titles` — создать тайтл
- `PATCH /api/admin/titles/:id` — обновить тайтл
- `POST /api/admin/chapters/upload` — загрузить ZIP с главой (max 200MB)
- `GET /api/admin/moderation` — список на модерации (is_published=false)
- `PATCH /api/admin/titles/:id/publish` — опубликовать тайтл
- `PATCH /api/admin/chapters/:id/publish` — опубликовать главу
- `PATCH /api/admin/chapters/:id/reject` — отклонить главу (soft delete с причиной)

---

## 5. КЭШИРОВАНИЕ (Redis Strategy)

| Ключ                               | TTL | Инвалидация                    |
| ---------------------------------- | --- | ------------------------------ |
| `titles:all:{page}:{filters_hash}` | 1h  | При создании/обновлении тайтла |
| `title:{id}`                       | 2h  | При обновлении тайтла          |
| `chapters:{title_id}`              | 1h  | При добавлении главы           |
| `image:size:{key}:{quality}`       | 30d | Вручную при нужде              |
| `session:{token}`                  | 30d | При logout (в черный список)   |

---

## 6. ЗАГРУЗКА ИЗОБРАЖЕНИЙ (ZIP Pipeline + On-Demand Resizing)

```
1. Админ: POST /api/admin/chapters/upload
   Payload MultiPart:
   - title_id (int)
   - volume_number (int)
   - chapter_number (float)
   - archive (file.zip)
   Валидация: Content-Type (application/zip), Размер (max 200MB)

2. Go Backend:
   a. Распаковать archive.zip
   b. Отсортировать файлы по имени (01.jpg, 02.jpg, ...)
   c. Для каждого файла:
      - Прочитать в RAM
      - Если width > 2000px: Auto-resize до 2000px (Source)
      - Загрузить Source (максимальное качество) в MinIO:
        titles/{title_id}/{chapter_id}/{page_index}_source.avif
      - Сохранить metadata в PostgreSQL (table pages): width, height
      - Cache Invalidation: инвалидировать chapters:{title_id}
   d. Вернуть: { "chapter_id": 123, "pages": 25, "status": "success" }

3. Image Proxy (On-Demand Resizing):
   GET /api/v1/image/:key?q=low|medium|high
   - Юзер запрашивает low → Go проверяет S3
   - Если файл {key}_low.avif не существует:
     * Загруж Source из S3
     * Ресайзит до нужного размера (low: ~480px, medium: ~800px)
     * Сохраняет результат в S3: {key}_low.avif
   - Отдает юзеру
   - Redis cache: image:size:{key}:{quality} → 30 дней

4. S3 Folder Structure:
   titles/{title_id}/{chapter_id}/{page_index}_source.avif  (исходник)
   titles/{title_id}/{chapter_id}/{page_index}_low.avif     (генерируется)
   titles/{title_id}/{chapter_id}/{page_index}_medium.avif  (генерируется)
```

---

## 7. ФРОНТЕНД СТРУКТУРА (SvelteKit)

```
src/
├── routes/
│   ├── +page.svelte           # Home
│   ├── catalog/
│   │   └── +page.svelte       # Каталог
│   ├── title/
│   │   └── [id]/
│   │       └── +page.svelte   # Карточка тайтла
│   ├── read/
│   │   ├── [titleId]/
│   │   │   └── [chapterId]/
│   │   │       └── +page.svelte  # Читалка (KRITICAL)
│   │   └── +layout.svelte
│   ├── admin/
│   │   ├── +layout.svelte
│   │   ├── titles/+page.svelte
│   │   └── upload/+page.svelte
│   └── login/+page.svelte
├── lib/
│   ├── api/           # API Client
│   │   ├── client.ts
│   │   └── types.ts   # Те же типы что и Go
│   ├── stores/        # Svelte Stores
│   │   ├── auth.ts
│   │   ├── reader.ts  # Local settings
│   │   └── history.ts
│   ├── components/    # Reusable UI
│   │   ├── Header.svelte
│   │   ├── Sidebar.svelte
│   │   ├── TitleCard.svelte
│   │   └── Reader/
│   │       ├── LongStrip.svelte
│   │       ├── LazyImage.svelte
│   │       ├── FAB.svelte
│   │       └── SettingsMenu.svelte
│   └── utils/
│       ├── constants.ts
│       └── helpers.ts
└── styles/
    └── app.css        # Tailwind + Dark Mode
```

---

## 8. DATABASE (Обновленная DDL)

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  username VARCHAR(50) UNIQUE NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255),  -- NULL если user через Google
  google_id VARCHAR(255) UNIQUE,
  role VARCHAR(20) DEFAULT 'user',  -- 'user', 'translator', 'admin'
  language VARCHAR(10) DEFAULT 'kz', -- 'kz', 'en', 'ru'
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Genres (теги и категории с локализацией)
CREATE TABLE genres (
  id SERIAL PRIMARY KEY,
  slug VARCHAR(50) UNIQUE NOT NULL,
  name_kz VARCHAR(100) NOT NULL,
  name_en VARCHAR(100),
  name_ru VARCHAR(100),
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Titles (Манхва)
CREATE TABLE titles (
  id SERIAL PRIMARY KEY,
  title_kz VARCHAR(255) NOT NULL,
  title_ru VARCHAR(255),
  title_en VARCHAR(255),
  description_kz TEXT,
  description_ru TEXT,
  description_en TEXT,
  cover_url VARCHAR(255),
  type VARCHAR(20) DEFAULT 'manhwa',  -- 'manhwa', 'manga'
  status VARCHAR(20) DEFAULT 'ongoing',  -- 'ongoing', 'completed', 'dropped'
  is_published BOOLEAN DEFAULT false,    -- Модерация
  deleted_at TIMESTAMPTZ,                -- Soft delete
  view_count INT DEFAULT 0,
  rating FLOAT DEFAULT 0,
  translator_id UUID REFERENCES users(id),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Many-to-Many: Titles ↔ Genres
CREATE TABLE title_genres (
  title_id INT REFERENCES titles(id) ON DELETE CASCADE,
  genre_id INT REFERENCES genres(id) ON DELETE CASCADE,
  PRIMARY KEY (title_id, genre_id)
);

-- Chapters
CREATE TABLE chapters (
  id SERIAL PRIMARY KEY,
  title_id INT REFERENCES titles(id) ON DELETE CASCADE,
  volume_number INT DEFAULT 1,
  chapter_number DECIMAL(6,2) NOT NULL,
  chapter_name VARCHAR(255),
  is_published BOOLEAN DEFAULT false,  -- Модерация
  deleted_at TIMESTAMPTZ,                -- Soft delete
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(title_id, volume_number, chapter_number)
);

-- Pages
CREATE TABLE pages (
  id SERIAL PRIMARY KEY,
  chapter_id INT REFERENCES chapters(id) ON DELETE CASCADE,
  order_index INT NOT NULL,
  image_key VARCHAR(255) NOT NULL,  -- S3 path: titles/{title_id}/{chapter_id}/{page_index}_source.avif
  width INT,                        -- Оригинальная ширина
  height INT,                       -- Оригинальная высота
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(chapter_id, order_index)
);

-- Reading History
CREATE TABLE reading_history (
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  title_id INT REFERENCES titles(id) ON DELETE CASCADE,
  last_chapter_id INT REFERENCES chapters(id),
  last_page INT DEFAULT 1,
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (user_id, title_id)
);

-- Bookmarks
CREATE TABLE bookmarks (
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  title_id INT REFERENCES titles(id) ON DELETE CASCADE,
  status VARCHAR(20) NOT NULL,  -- 'reading', 'completed', 'dropped', 'planned'
  created_at TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (user_id, title_id)
);

-- Ratings (Live AVG для MVP)
CREATE TABLE ratings (
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  title_id INT REFERENCES titles(id) ON DELETE CASCADE,
  score INT CHECK (score >= 1 AND score <= 10),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (user_id, title_id)
);

-- Comments (v2, структура на будущее)
CREATE TABLE comments (
  id SERIAL PRIMARY KEY,
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  title_id INT REFERENCES titles(id) ON DELETE CASCADE,
  chapter_id INT REFERENCES chapters(id) ON DELETE SET NULL,
  content TEXT NOT NULL,
  is_deleted BOOLEAN DEFAULT false,  -- Soft delete
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_titles_fts ON titles USING GIN
  (to_tsvector('simple', title_kz || ' ' || COALESCE(title_ru, '') || ' ' || COALESCE(title_en, '')));
CREATE INDEX idx_titles_is_published ON titles(is_published) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_is_published ON chapters(is_published) WHERE deleted_at IS NULL;
```

---

## 9. ЛОКАЛЬНЫЕ НАСТРОЙКИ (LocalStorage Schema)

```js
// reader_settings (в LocalStorage)
{
  quality: 'medium',           // 'low', 'medium'
  pageWidth: '100%',           // '100%' или 'fixed' (с отступами)
  theme: 'dark',               // 'dark', 'light'
  language: 'kz',              // 'kz', 'en', 'ru' (по умолчанию KZ)
}

// reading_progress (в LocalStorage, перезаписуется тех из БД при логине)
{
  title_1: {
    chapter_id: 5,
    page: 3,
    last_read: 1681234567,     // timestamp
  },
  title_2: { ... }
}
```

---

## 10. УТВЕРЖДЕННЫЕ РЕШЕНИЯ (ROUND 2)

### 10.1 Теги и Категории

- **Фиксированный список**: Enum в коде или отдельная таблица `genres`
- **Модели доступа**: Переводчики выбирают из существующего, Админы добавляют новые
- **Почему**: Свободный ввод → помойка (экшн/экшен/action)

### 10.2 Модерация

- **Поле в БД**: `is_published` (bool) или `status` (enum: draft, pending, published) в `titles` и `chapters`
- **Flow**: Translator заливает → pending → Модератор чекает → published
- **Видимость**: Юзеры видят только `published` контент

### 10.3 Размер архивов

- **Max Upload**: 200 MB на главу
- **Обоснование**: Средняя глава WebP 15–40 MB, 200 MB покрывает HQ архивы

### 10.4 Image Proxy (Go)

- **Требование**: Все картинки через `GET /api/v1/image/:key?q=...`
- **Преимущества**:
  - На лету инжект вотермарков (позже)
  - Скрывает структуру S3
  - Cache-Control заголовки → CDN кэширование
- **TTL**: Картинки кэшируются на 30 дней

### 10.5 Аналитика

- **Internal Views**: Таблица `views` (title_id, user_id, timestamp) или Redis инкрементация + сброс в Postgres раз в час
- **External**: Umami или Plausible (self-hosted в Docker)
- **Зачем**: Рейтинг "Популярное за неделю"

### 10.6 Миграции БД

- **Tool**: `go-migrate` (отлично с pgx)
- **Структура**: `migrations/` папка с `.up.sql` и `.down.sql` файлами

---

---

## 12. УТВЕРЖДЕННЫЕ РЕШЕНИЯ (ROUND 3 — Image, Localization, CDN)

### 12.1 Image Pipeline (AVIF + On-Demand)

- **Формат**: AVIF (лучше чем WebP по сжатию)
- **Хранение**: Source (полное качество) сохраняется в S3 один раз
- **Генерация**: Low (480w), Medium (800w) генерируются на лету при первом запросе
- **S3 структура**: `titles/{title_id}/{chapter_id}/{page_index}_{quality}.avif`
- **Кэширование**: На 30 дней в Redis/CDN (Cache-Control заголовки)
- **Auto-resize**: При загрузке если width > 2000px → resize до 2000px перед сохранением Source

### 12.2 Локализация Контента

- **Fallback цепь**: KZ → EN → RU
- **Что локализуется**: Название, Описание, Жанры
- **Что НЕ менется**: Автор, Художник (остаются на языке оригинала/транслит)
- **Выбор языка**: По умолчанию KZ (qmanga.kz). Юзер выбирает в settings, сохраняется в БД (users.language)

### 12.3 Комментарии и Уведомления

- **Комментарии**: Только на Title и Chapter (не на каждой странице)
- **Модерация**: Ручная (админ нажимает Delete коммент)
- **Уведомления**: v2 (после запуска). Для MVP: индикатор "New" на карточках в закладках

### 12.4 Загрузка и Валидация

- **Метод**: multipart/form-data
- **Валидация**: Content-Type (image/\* или application/zip), Размер (max 200MB на ZIP)
- **Auto-optimize**: При загрузке Source — если width > 2000px, ресайзим до 2000px сразу

### 12.5 CDN и Безопасность

- **Прокси**: Go Image Proxy обязателен
- **CDN**: Cloudflare кэширует ответы Go (Cache-Control: public, max-age=2592000)
- **MinIO**: Приватный, доступ ТОЛЬКО через Go. Юзеры НЕ могут качать напрямую из S3
- **Контроль**: Go может логировать кто и что качает, инжектить вотермарки позже

### 12.6 Рейтинг система

- **Calculуция**: Live SQL AVG() для MVP
- **Обновление**: Юзер может менять свою оценку (UPDATE запрос)
- **Кэш**: Когда онлайн > 1000 юзеров → Redis AVG кэш с инвалидацией

### 12.7 Пагинация в Каталоге

- **Тип**: Infinite Scroll (стандарт для мобилок)
- **Лимит**: Подгружай по **20-30 тайтлов за раз** (оптимум: не перегружать DOM, не делать слишком длинные запросы)

### 12.8 Перезалив Главы (Update Strategy)

- **Логика**: Вариант A (Обновление существующей)
- **Flow**:
  1. Глава published → Модератор может только удалить
  2. Переводчик хочет fix → Заливает новый ZIP → статус переходит в `pending_update`
  3. Юзеры видят **старую версию** (pages хранят старые данные)
  4. Модератор одобрил → В S3 заменяются файлы, БД обновляется, статус → published
- **DB**: Добавить поле `updated_at` в chapters для отслеживания обновлений

### 12.9 Soft Delete (Удаление Контента)

- **API**: Удаленные тайтлы/главы НЕ отдаются через API (WHERE deleted_at IS NULL)
- **Закладки**: Если юзер откроет закладку на удаленный тайтл → "Тайтл удален или скрыт"
- **Мод-панель**: Админ может видеть удаленные (с фильтром) для восстановления если нужно

### 12.10 Image Quality (Manual Selection)

- **Управление**: Вручную в Settings Ридера (не автоматика)
- **Причина**: navigator.connection ошибается, юзер лучше знает свой лимит
- **Default**: Medium (480w) как стандарт

### 12.11 Сортировка Каталога (MVP Priority)

1. **По обновлениям** (дата последней главы) — главное для возвращаемости ⭐
2. **Популярные** (view_count) — для новых юзеров
3. **По рейтингу** (rating) — вспомогательная сортировка

### 12.12 Max Pages per Chapter

- **Лимит**: 150-200 страниц максимум
- **Валидация**: При загрузке ZIP проверяем количество файлов
- **Причина**: Защита от гигантских архивов, стабильность фронта

### 12.13 Длинные Стрипы (Long Strips)

- **Решение**: Не резать на части
- **Реализация**: Frontend: `object-fit: contain` + `width: 100%`
- **Причина**: AVIF хорошо справляется, нарезка усложнит синхронизацию pages

---

## 13. УТВЕРЖДЕННЫЕ РЕШЕНИЯ (ROUND 4 — Dev Setup, Docker, Code Style)

### 13.1 Код-стайл и линтеры

- **Go**:
  - `golangci-lint` (в CI)
  - `gofmt` (автоматический, перед коммитом)
- **Svelte**:
  - `eslint` + `prettier`
- **Коммиты**: `commitlint` (conventional commits: `feat:`, `fix:`, `chore:`)
- **.editorconfig**: Не нужен (каждый IDE/проект разбираются сами)

### 13.2 Docker Compose Порты

| Сервис         | Порт | Expose | Описание                     |
| -------------- | ---- | ------ | ---------------------------- |
| Backend (Go)   | 9009 | Да     | API (nginx proxy к нему)     |
| Frontend (dev) | 5173 | Да     | SvelteKit dev server         |
| PostgreSQL     | 5432 | Нет    | Internal только              |
| Redis          | 6379 | Нет    | Internal только              |
| MinIO API      | 9000 | Нет    | Internal (через Go proxy)    |
| MinIO Console  | 9001 | Да     | Admin panel (localhost:9001) |

### 13.3 MinIO Setup

- **Bucket**: Одна — `qmanga`
- **Структура внутри**:
  ```
  qmanga/
    titles/{title_id}/
      {chapter_id}/
        {page_index}_source.avif
        {page_index}_low.avif
        {page_index}_medium.avif
  ```
- **Access**: Приватный (no public URLs)

### 13.4 Git Workflow

- **Main Branch**: `main` (современный стандарт)
- **Branching**: Conventional Branching
  - `feat/...` — новая фича
  - `fix/...` — баг фикс
  - `chore/...` — рефакторинг, зависимости
- **Pull Requests**: **Обязательны** (5 человек, нужна code review)
- **Approver**: Один универсал/lead проверяет и мержит (держит структуру в чистоте)

### 13.5 Database Migrations

- **Инструмент**: `go-migrate` (простой, надежный, CLI)
- **Папка**: `/migrations` в корне репо
- **Формат**:
  ```
  000001_init_schema.up.sql   (CREATE TABLE...)
  000001_init_schema.down.sql (DROP TABLE...)
  000002_add_soft_delete.up.sql
  000002_add_soft_delete.down.sql
  ```
- **В коде**: `migrate.New("file://migrations", "postgres://...")` в init()

### 13.6 Environment Variables

- **Config файл**: `.env` (в `.gitignore`, никогда не коммитим)
- **Example**: `.env.example` в репо с "рыбой"
  ```
  DATABASE_URL=postgresql://user:password@localhost:5432/qmanga
  REDIS_URL=redis://localhost:6379
  MINIO_ENDPOINT=localhost:9000
  MINIO_ACCESS_KEY=minioadmin
  MINIO_SECRET_KEY=minioadmin
  JWT_SECRET=your-secret-key-here
  GOOGLE_CLIENT_ID=...
  GOOGLE_CLIENT_SECRET=...
  ```
- **Загрузка в Go**: `joho/godotenv` для локалки, `os.Getenv()` в prod (Docker ENV)

### 13.7 API Documentation (Swagger)

- **Библиотека**: `swaggo/swag` для Go/Gin
- **Генерация**: `swag init` (создает `/docs` папку с OpenAPI spec)
- **Доступ**: `GET /swagger/index.html` или `GET /docs/swagger.json`
- **Интеграция в Gin**:

  ```go
  import (
    _ "github.com/swaggo/gin-swagger/example/basic/docs"
    gs "github.com/swaggo/gin-swagger"
  )

  r.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))
  ```

- **Комментарии в коде**: Используй `// @Router /api/v1/titles [get]` над handlers
- **Почему**: Твои фронтенд-разработчики получат живой контракт вместо угадывания

### 13.8 Тестирование и Контроль Качества

- **API Testing** (на старте): Postman или Insomnia для быстрой проверки
- **Unit Tests**: Критичные части (Image Proxy, Auth) покрывай `go test`
- **CI**: GitHub Actions должен проверять `golangci-lint` и запускать тесты перед мержем
- **Pre-commit Hooks**: Локально `golangci-lint` перед коммитом

### 13.9 Seed Data (Тестовые данные)

- **Способ**: Отдельный скрипт `scripts/seed.go`
- **Запуск**: Вручную или через Makefile: `make seed`
- **Не в docker-compose**: Чтобы не засирать прод-БД случайно позже
- **Что заливаем**: Тестовые юзеры, жанры, структура. Реальные тайтлы позже (когда будут в наличии)

---

## 14. РЕПОЗИТОРИЙ СТРУКТУРА

```
qmanga/
├── .github/
│   └── workflows/
│       └── ci.yml              # GitHub Actions (lint, test, build)
├── migrations/
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   ├── 000002_add_genres.up.sql
│   └── 000002_add_genres.down.sql
├── scripts/
│   └── seed.go                 # Seed data для локальной разработки
├── backend/
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   ├── pkg/
│   │   ├── models/             # Go structs
│   │   ├── handlers/           # Route handlers
│   │   ├── middleware/         # Auth, CORS, logging
│   │   ├── service/            # Business logic
│   │   └── repository/         # Database queries
│   ├── docs/                   # Swagger/OpenAPI (auto-generated)
│   └── Dockerfile
├── frontend/
│   ├── src/
│   ├── package.json
│   └── Dockerfile
├── .env.example                # Template для .env
├── .gitignore
├── docker-compose.yml          # PostgreSQL, Redis, MinIO, Go backend
├── TASK.md                     # Этот файл
└── README.md
```

---

## 15. TECHNICAL DEBT (DO NOT START NOW)

- WebSockets для real-time обновлений (новые главы)
- Comments v2 + рекурсивные ответы
- Система рекомендаций
- Blue Light Filter
- PWA оптимизация
- OCR + TTS
- Advanced Analytics Dashboard

---

## 16. READY FOR SPRINT 1? ✅

**FINAL CHECKLIST:**

- [ ] Git репозиторий инициализирован (main branch, PR workflow готов)
- [ ] Docker Compose настроен (PostgreSQL 17, Redis, MinIO)
- [ ] PostgreSQL миграции готовы (go-migrate setup)
- [ ] JWT middleware (Go/Gin)
- [ ] Google OAuth2 интеграция
- [ ] Swagger документация (swaggo/swag)
- [ ] CORS middleware для фронта
- [ ] Health check endpoint (`GET /api/v1/health`)
- [ ] Seed скрипт (пустой, готов к заполнению)

**GO? 🚀**

Если да → Начинаем Sprint 1 с детальным планом!
