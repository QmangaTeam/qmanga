---
id: TASK-23
title: Определить события и агрегаты метрик
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 22:31'
labels:
  - stage-6
dependencies: []
references:
  - UPDATE.md
ordinal: 23
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-6. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Перед графиками определить реальные события и хранение агрегатов: новые подтверждённые пользователи по дням; опубликованные главы; активные читатели с прогрессом; оценки и подписки на тайтлы/команды. Для «просмотров» — отдельная спецификация дедупликации, приватности и срока хранения; нельзя считать каждый запрос фрагмента изображения просмотром.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Define truthful event/privacy contracts in ARCHITECTURE; persist transactional daily events and distinct reading completions; expose bounded 7/30/90-day scoped dashboards with accessible tables; verify authorization, deduplication, migration and UI.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Transactional daily counters: confirmation, first translation publication, first rating, new/restored library entries and newly enabled title notifications. Accepted progress ledger supplies distinct daily/period readers and daily deduplicated completions. Historical dates not fabricated; collection start visible. Existing product has no separate team-follow feature: subscriptions mean title notification opt-ins; team scope only explicit preferred team, approved with owner scope. Contract docs/ARCHITECTURE.md; migration27; commits d809ebd, ee4ed3e. Actual PostgreSQL event/dedup/retention/scoping tests pass; no external release yet.

Ревью владельца завершено для23 карточек на базе3302515: проверены реализация, права доступа, границы API/хранилищ и интеграция. Итоговые локальные проверки:106 Go-тестов/подтестов,114 прикладных и29 SSR-тестов,40/40 браузерных сценариев;0 ошибок и0 пропусков, Svelte0 ошибок/предупреждений. Логи tmp/review-complete-go.jsonl, tmp/review-complete-web-build.log, tmp/review-final-browser-acceptance.log. Точные auth/publishing/reading CI smoke пройдены на настоящих PostgreSQL/MinIO/Mailpit. Это локальная приёмка; Review сохранён до удалённого CI, staging и наблюдаемого выпуска того же SHA. Для TASK17 дополнительно нужны утверждение правовых текстов владельцем и действующий ящик поддержки; для TASK26 остаётся предрелизная проверка consent/contact.
<!-- SECTION:NOTES:END -->
