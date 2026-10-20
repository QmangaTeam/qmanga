---
id: TASK-101
title: Поиск и импорт метаданных тайтла из Shikimori
status: Review
assignee:
  - '@codex'
created_date: '2026-10-20 16:49'
updated_date: '2026-10-20 17:44'
labels:
  - web
  - publishing
  - imports
dependencies: []
priority: high
type: feature
ordinal: 45003
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Владелец: «пишешь в поиск выбираешь тайтл и он автоматом загружает метаданные имеющиеся». Первый источник — Shikimori. Выбор результата автоматически заполняет форму и теги через общую политику сопоставления; ручное переписывание каждого поля не требуется. Перед реализацией сверить действующий документированный API/лимиты источника.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 В форме создания/изменения тайтла есть поиск по Shikimori с различимыми результатами (названия/тип/год и обложка при наличии), loading/empty/error/retry; пользователь явно выбирает нужный тайтл.
- [x] #2 После выбора автоматически заполняются доступные поддерживаемые названия, оригинал/алиасы, описание, авторы, тип/статусы, даты/публикация, изображения и теги; отсутствующие поля не выдумываются и не очищают уже введённые данные.
- [x] #3 Источник, внешний ID и происхождение данных видимы; смена выбранного результата не теряет несохранённое без согласия, заполнение не публикует тайтл само по себе, существующие права/модерация/версии сохранены.
- [x] #4 Импорт находится за application port/source adapter, имеет кеш/таймауты/лимиты и безопасную обработку URL/изображений; API fixtures и browser сценарий поиск→выбор→проверка→сохранение покрыты.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Confirm official source contracts; add bounded server adapters and common import application port. 2. Add explicit search/select/autofill in existing admin and moderated forms, retaining absent/manual values and source provenance. 3. Verify source fixtures, mapping, browser flow and existing checks; independent review and release remain with coordinator.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented common server/application import port and Shikimori adapter, private search/select/cover routes, bounded cache/timeouts/source pacing and expected-account authorization. Admin and studio forms autofill only available supported fields; absent values preserve authored data, replacement requires consent, canonical form context is locked during async selection. Source identity and selection snapshot persist in metadata_import_receipts; source link enters supported publication/evidence fields. Studio cover is reprocessed to WebP and added to the existing moderated multipart proposal. Admin metadata form previews source artwork and explicitly explains that applying artwork requires the existing project metadata-change proposal; it never bypasses artwork moderation. Independent review fixes cover generation invalidation, accurate cover-failure copy, form locks and bounded image concurrency. Local checks: catalog/HTTP/bootstrap Go tests pass; full PostgreSQL suite was run successfully by coordinator before final endpoint follow-up; Svelte check 0 errors/0 warnings; frontend boundaries pass; merge tests pass. Browser search/select/confirmation/admin real metadata save and studio multipart retry passed against the local production bundle; expanded final browser checks pending coordinator rebuild. No production release is claimed.

Final implementation SHAs: 3f7ae72, f96809e, ec8e34f, d3fb15f, 6be46e4. Independent metadata review by cbz_import reported no remaining blocking findings after the context/busy/keyboard and candidate-editor fixes. Real rebuilt local API -> live Shikimori search (10 results) -> selection external 13 -> source authors/tags mapping -> metadata receipt -> reprocessed WebP cover all returned HTTP 200; 1 author, 4 canonical genres, 0 unresolved tags, cover 27066 bytes, provenance host shikimori.io. Live AniList same pipeline also passed. Final browser cases cover admin explicit confirmation and real metadata save, studio cover preview/provenance retention and moderated multipart retry, cover-fetch failure truthfulness, pending import field/context locks, and one-editor candidate keyboard mapping. Source fixture tests, latest targeted Go checks, Svelte 0/0 and frontend boundaries pass. Local implementation is ready for integration review; authorized push, full Verify gate for migration 000035, observed staging/production release and production SHA confirmation remain coordinator-owned.

Migration 000035 is additive: new metadata_tag_mappings and metadata_import_receipts tables/index only; existing title metadata, taxonomy IDs, publication data, revisions and offline storage schemas are not rewritten. Application rollback can leave these new tables in place; no production volume/data deletion or down migration is required. The migration requires the full Verify gate before deploy-ready under the existing release contract.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Search and explicit source selection fill supported metadata and canonical tags, preserve absent/manual values, retain source provenance and prepare studio artwork through existing moderation. Shikimori .io and AniList live APIs and the local production bundle were verified. Admin artwork application remains intentionally routed through project metadata proposals. Awaiting observed release; local verification is not production acceptance.
<!-- SECTION:FINAL_SUMMARY:END -->
