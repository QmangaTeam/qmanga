---
id: TASK-11
title: 'Модель жанров, категорий и подборок'
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 22:31'
labels:
  - stage-3
dependencies: []
references:
  - UPDATE.md
ordinal: 11
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-3. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Развести **жанры** (уже существуют), **метки/категории содержания** и **редакционные подборки**. Подборка имеет заголовок, описание, порядок, видимость и выбранные тайтлы; жанр — таксономия для фильтра; возраст — отдельное поле, не жанр. Не смешивать эти сущности в одну таблицу.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add separate category and editorial collection entities, persistence and invariants. 2. Verify migration and repository behavior.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented separate categories and ordered, visible editorial collections in migration 000024; genres and age remain independent. Added domain and isolated PostgreSQL regression. Go catalog/HTTP unit tests pass; PostgreSQL run pending harness URL.

Implementation commits: 433aa31, e85e4e5, b9058a8. Domain/application authorization and validation tests pass; isolated PostgreSQL test is ready for coordinated acceptance. Genres, categories, collections and age rating remain separate. Next: run real PostgreSQL migration/persistence regression, then leave Review pending observed release.

Actual shared acceptance now passed: full Go/PostgreSQL suite103 test/subtests with0actualskips; taxonomy browser CRUD/authorization/combined filters passed in integrated Docker run. Implementation433aa31/e85e4e5/b9058a8; public team avatar e3365ea. Remote CI/staging/release remains pending.

Ревью владельца завершено для23 карточек на базе3302515: проверены реализация, права доступа, границы API/хранилищ и интеграция. Итоговые локальные проверки:106 Go-тестов/подтестов,114 прикладных и29 SSR-тестов,40/40 браузерных сценариев;0 ошибок и0 пропусков, Svelte0 ошибок/предупреждений. Логи tmp/review-complete-go.jsonl, tmp/review-complete-web-build.log, tmp/review-final-browser-acceptance.log. Точные auth/publishing/reading CI smoke пройдены на настоящих PostgreSQL/MinIO/Mailpit. Это локальная приёмка; Review сохранён до удалённого CI, staging и наблюдаемого выпуска того же SHA. Для TASK17 дополнительно нужны утверждение правовых текстов владельцем и действующий ящик поддержки; для TASK26 остаётся предрелизная проверка consent/contact.
<!-- SECTION:NOTES:END -->
