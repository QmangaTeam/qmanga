---
id: TASK-64
title: Теги Senkuro в поиске каталога
status: Done
assignee: []
created_date: '2026-10-19 19:29'
updated_date: '2026-10-19 22:27'
labels:
  - catalog
  - taxonomy
  - frontend
dependencies: []
references:
  - 'https://senkuro.me/browse/manga'
  - docs/ui/DESIGN-REVIEW.md
  - backlog/tasks/task-11 - Модель жанров
  - категорий и подборок.md
modified_files:
  - apps/web/src/lib/features/catalog/application/catalog.ts
  - >-
    apps/web/src/lib/features/catalog/infrastructure/http-catalog-gateway.test.ts
  - apps/web/src/lib/features/catalog/infrastructure/http-catalog-gateway.ts
  - apps/web/src/lib/features/catalog/presentation/AdminCatalogue.svelte
  - apps/web/src/lib/features/catalog/presentation/Catalogue.svelte
  - apps/web/src/lib/features/catalog/presentation/CatalogueFilters.svelte
  - apps/web/src/lib/features/catalog/presentation/TagSearch.svelte
  - apps/web/src/lib/features/catalog/presentation/Title.svelte
  - apps/web/src/routes/catalog/+page.server.ts
  - apps/web/tests/browser/tag-search.test.mjs
  - apps/web/tests/browser/taxonomy.test.mjs
  - apps/web/tests/ui/TagSearch.fixture.svelte
  - apps/web/tests/ui/tag-search.test.ts
  - backend/internal/catalog/adapters/postgres/repository.go
  - backend/internal/catalog/adapters/postgres/repository_integration_test.go
  - backend/internal/catalog/adapters/postgres/senkuro_tags_integration_test.go
  - backend/internal/catalog/adapters/postgres/taxonomy.go
  - backend/internal/catalog/adapters/postgres/taxonomy_integration_test.go
  - backend/internal/catalog/application/service.go
  - backend/internal/catalog/application/service_test.go
  - backend/internal/catalog/domain/title.go
  - backend/internal/transport/http/catalog.go
  - backend/internal/transport/http/catalog_taxonomy.go
  - backend/internal/transport/http/catalog_test.go
  - backend/migrations/000032_senkuro_taxonomy.sql
type: feature
ordinal: 39026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Запрос владельца (20.10.2026): перенести с Senkuro (/browse/manga) все теги и связанные справочники в поиск Qmanga и добавить их на production. Источник — их GraphQL api.senkuro.me/graphql, запрос allLabels: 1125 меток в 7 группах, из них 719 без рейтинга EXPLICIT. Решения владельца: брать 719 не-18+ меток; названия на двух языках интерфейса Qmanga (казахский и русский); отдельный поиск по тегам в фильтре каталога по образцу Senkuro; доставка в prod миграцией-сидом в репозитории.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Таксономия Senkuro перенесена: 6 групп становятся категориями Qmanga, 713 меток — тегами с казахским и русским названием, привязкой к группе и иерархией внутри группы; данные лежат в миграции-сиде, повторный прогон идемпотентен.
- [x] #2 В фильтре каталога работает отдельный поиск по тегам: подсказки сгруппированы по категориям, теги выбираются множественно, выбранное видно чипами, панель одинаково работает в desktop-колонке и в мобильном диалоге на 320 и 360 px в обеих темах и локалях.
- [x] #3 Фильтрация каталога по нескольким тегам работает через API (логика AND), а текстовый поиск каталога находит тайтлы по названию тега.
- [x] #4 Название тега в интерфейсе соответствует локали (казахский/русский), админская таксономия не ломается, старый одиночный параметр genre продолжает работать.
- [x] #5 Проверено: полный Go-набор с PostgreSQL, web-тесты и svelte-check, браузерная приёмка каталога; релиз того же SHA на staging и production с внешней проверкой.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Снять таксономию Senkuro (allLabels) и разложить на группы/метки с иерархией. 2. Перевести 713 названий на казахский. 3. Миграция 000032: name_ru для genres/categories, category_id и parent_slug для genres, сид данных. 4. Backend: Genre с двумя именами и группой, мульти-фильтр по тегам (AND), текстовый поиск по названию тега. 5. Frontend: компонент TagSearch (подсказки по группам, чипы, мультивыбор), замена селекта жанра, локализованные названия. 6. Тесты Go/web, браузерная приёмка, коммит, push, CI, release.sh staging и production, внешняя проверка.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Реализовано в коммитах 1aca5f5 (backend+миграция), 0f74f3c (frontend), e005b50 (UX-правки и правка браузерного теста), 276d0c5 (новая браузерная регрессия).

Данные: backend/migrations/000032_senkuro_taxonomy.sql — 6 категорий и 713 тегов Senkuro (719 не-18+ минус сами рейтинговые/EXPLICIT-метки), у каждого тега казахское name и русское name_ru, category_id, parent_slug, position. Сид идемпотентен (ON CONFLICT (slug) DO NOTHING), слаг нормализован к ^[a-z0-9]+(-[a-z0-9]+)*$, 4 аномалии источника исправлены. Миграции встроены в бинарь и применяются при старте API.

Backend: Genre получил NameRU/CategorySlug/ParentSlug/Position; Filter.Genres []string; List фильтрует по нескольким тегам (AND через EXISTS), текстовый поиск ищет и по name/name_ru тега; /api/v1/taxonomy и /api/v1/titles отдают name_ru/category_slug/parent_slug/position. Повторяющиеся genre сворачиваются, невалидный слаг или больше 12 тегов — 400/422.

Frontend: TagSearch.svelte (подсказки по группам, чипы, мультивыбор, имена по локали) заменил селект жанра в CatalogueFilters.svelte (desktop-колонка и мобильный диалог), чипы активных фильтров в Catalogue.svelte, локализованные теги в Title.svelte, в AdminCatalogue.svelte вместо 713 чекбоксов — тот же поиск.

Проверки: sudo docker compose --profile tools run --rm go-tests go test ./... -count=1 — зелёный с PostgreSQL (включая senkuro_tags_integration_test); bun test src/lib/features src/lib/shared src/lib/server — 127 pass; bun run test:ui — 38 pass; bun run check — 0 ошибок; check:architecture, check:ui-contrast, build — зелёные. Браузерная приёмка в управляемом Chrome на локальном стеке: каталог 320/360/desktop, kk/ru, тёмная и светлая темы; подсказки сгруппированы, чипы добавляются/снимаются, форма отправляет genre=<slug>, каталог сужается до помеченного тайтла, поиск по казахскому и русскому названию тега находит тайтл, админка сохраняет теги в title_genres. Браузерные регрессии локально: taxonomy, catalogue-filters, search, admin, tag-search — зелёные (остальные файлы требуют полного CI-фикстура, без него падают на 401 ещё до кода).
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-10-19 22:27

---

20.10.2026, AC #5 закрыт выпуском.

Production и staging: bb8aec38d6c019e08d940b6209e794795ca956e3 (тот же SHA, что в TASK-65). CI Verify Qmanga 39333412292 успешен, все три ARM64 images опубликованы; staging poller поднял SHA, production переведён release.sh production. Внешне: /api/v1/genres = 713, /api/v1/taxonomy = 6 групп в порядке Senkuro, панель каталога проверена Playwright (kk/ru, обе темы, desktop/mobile).

---
<!-- COMMENTS:END -->
