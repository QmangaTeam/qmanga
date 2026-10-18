---
id: TASK-7
title: Возрастная маркировка тайтлов
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 22:31'
labels:
  - stage-2
dependencies: []
references:
  - UPDATE.md
ordinal: 7
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-2. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Возрастная маркировка на тайтле и карточках: минимум 12+/16+/18+, 18+ заметно красная; возраст хранится на сервере и фильтруется.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Регрессиями зафиксировать допустимые значения 0/12/16/18 и отклонение остальных. 2. Добавить совместимую миграцию age_rating=0 для прежних тайтлов, запись/чтение и точный серверный фильтр age_rating; старый PUT без поля сохраняет маркировку. 3. Прокинуть поле через DTO/адаптеры, админку, карточку/полки/каталог, URL-фильтр и красную метку 18+. 4. Реальные PostgreSQL/API/browser проверки, полный build и интеграционный набор; коммиты по завершённым частям, Review до CI/выпуска.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Backend slice verified: migration 000023 adds constrained smallint age_rating default 0 and composite age/update/id index. Domain validates 0/12/16/18, summary/detail DTO expose integer, public age_rating filter exact (including 0). Update omission uses nullable command field and SQL COALESCE to preserve classification; explicit 0 clears. Real isolated PostgreSQL tests verify persistence, filtering, DB rejection of 14, legacy insert/update, admin-only changes and clearing. HTTP regression before implementation accepted age_rating=14 with 200; now rejects invalid/malformed values. Full Go/PostgreSQL/codec suite passed 76 test/subtests across 23 packages, no actual skips: tmp/age-full-go.jsonl. Frontend adapter 3 tests and SSR 27 passed; app diagnostics 0 errors/warnings. UI/browser integration and Docker deployment still pending.

Frontend completed: age metadata travels through validated adapters; admin can set/clear classification; title, catalogue cards and home cover shelves show the shared AgeRating component. 18+ uses a red semantic surface with checked contrast in both themes. Unclassified content has no invented badge. Catalogue exact age filter, active chips, query preservation and cursor pagination are wired through URL/server. Targeted browser test covers storage via API/admin UI, 0/12/16/18, invalid values, old PUT preserving 18, non-admin 403, adult red badge, unclassified absence, filtering/pagination and kk/ru × light/dark × 320/360/1280. Final Docker frontend build: 91 app tests + 28 SSR tests, 0 failures; app/preview diagnostics 0 errors/warnings, architecture/contrast pass. Full browser suite: 22/22 pass, no skips; full Go/PostgreSQL/codec: 76 test/subtests, 23 packages, no actual skips. Backend commit 070824b. Logs tmp/age-full-go.jsonl, tmp/age-web-final-build.log, tmp/age-full-browser.log; screenshots tmp/age-final-screenshots/. Local Docker updated with additive migration; remote CI/deployment and physical devices not performed. Review until observed CI/staging acceptance/release.

Ревью владельца завершено для23 карточек на базе3302515: проверены реализация, права доступа, границы API/хранилищ и интеграция. Итоговые локальные проверки:106 Go-тестов/подтестов,114 прикладных и29 SSR-тестов,40/40 браузерных сценариев;0 ошибок и0 пропусков, Svelte0 ошибок/предупреждений. Логи tmp/review-complete-go.jsonl, tmp/review-complete-web-build.log, tmp/review-final-browser-acceptance.log. Точные auth/publishing/reading CI smoke пройдены на настоящих PostgreSQL/MinIO/Mailpit. Это локальная приёмка; Review сохранён до удалённого CI, staging и наблюдаемого выпуска того же SHA. Для TASK17 дополнительно нужны утверждение правовых текстов владельцем и действующий ящик поддержки; для TASK26 остаётся предрелизная проверка consent/contact.
<!-- SECTION:NOTES:END -->
