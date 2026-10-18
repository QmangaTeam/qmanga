---
id: TASK-24
title: Рабочие панели администратора и переводчика
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 22:31'
labels:
  - stage-6
dependencies:
  - TASK-23
references:
  - UPDATE.md
ordinal: 24
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-6. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Администратор видит обзор сайта с периодом 7/30/90 дней, числами и простыми доступными графиками. Переводчик видит только статистику своих проектов/команд: чтение его опубликованных переводов, подписки/библиотеки и публикации. Нет доступа к чужим данным и email читателей. Пустой период честно обозначен.
- [x] #2 Серверные агрегаты имеют предсказуемую нагрузку на Oracle Free Tier: индексы, ограниченные запросы и кэш/фоновые сводки при необходимости. График читается как таблица и не требует hover.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Define truthful event/privacy contracts in ARCHITECTURE; persist transactional daily events and distinct reading completions; expose bounded 7/30/90-day scoped dashboards with accessible tables; verify authorization, deduplication, migration and UI.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented real /admin overview and /studio/stats with 7/30/90-day UTC periods, event totals, distinct period readers, completion bars and visible accessible data table. Private no-store API authorizes current approved memberships/admin server-side, returns no emails/reader identifiers. Queries have 5s deadline, indexed <=90-day ranges, preaggregated counters. No team-follow feature invented; team libraries/notifications attributed only to explicit preferred team. Empty state/collection start explicit. d809ebd, 8ead0ae, ee4ed3e. Actual PG+HTTP target suite, 42 reading and 2 DTO frontend tests, frontend boundaries and web:check (0 errors/warnings) pass. Added full browser metrics regression; pending integrated browser run and release/acceptance.

Fresh integrated browser metrics tests both PASS in tmp/all16-final-browser.log: real admin 7/30/90-day API/table data, guest/admin isolation, own translator scope, no PII, 320/360/1280px both themes. Visually inspected 320 light and 360 dark screenshots. Found minimum-height ticks for zero days and corrected to zero in a555f2f (web rebuild pending). Stronger actual PostgreSQL foreign-team exclusion test PASS0.454s. Final same-SHA release and owner acceptance remain pending; Review retained.

Ревью владельца завершено для23 карточек на базе3302515: проверены реализация, права доступа, границы API/хранилищ и интеграция. Итоговые локальные проверки:106 Go-тестов/подтестов,114 прикладных и29 SSR-тестов,40/40 браузерных сценариев;0 ошибок и0 пропусков, Svelte0 ошибок/предупреждений. Логи tmp/review-complete-go.jsonl, tmp/review-complete-web-build.log, tmp/review-final-browser-acceptance.log. Точные auth/publishing/reading CI smoke пройдены на настоящих PostgreSQL/MinIO/Mailpit. Это локальная приёмка; Review сохранён до удалённого CI, staging и наблюдаемого выпуска того же SHA. Для TASK17 дополнительно нужны утверждение правовых текстов владельцем и действующий ящик поддержки; для TASK26 остаётся предрелизная проверка consent/contact.
<!-- SECTION:NOTES:END -->
