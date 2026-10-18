---
id: TASK-24
title: Рабочие панели администратора и переводчика
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 21:58'
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
- [ ] #1 Администратор видит обзор сайта с периодом 7/30/90 дней, числами и простыми доступными графиками. Переводчик видит только статистику своих проектов/команд: чтение его опубликованных переводов, подписки/библиотеки и публикации. Нет доступа к чужим данным и email читателей. Пустой период честно обозначен.
- [ ] #2 Серверные агрегаты имеют предсказуемую нагрузку на Oracle Free Tier: индексы, ограниченные запросы и кэш/фоновые сводки при необходимости. График читается как таблица и не требует hover.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Define truthful event/privacy contracts in ARCHITECTURE; persist transactional daily events and distinct reading completions; expose bounded 7/30/90-day scoped dashboards with accessible tables; verify authorization, deduplication, migration and UI.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented real /admin overview and /studio/stats with 7/30/90-day UTC periods, event totals, distinct period readers, completion bars and visible accessible data table. Private no-store API authorizes current approved memberships/admin server-side, returns no emails/reader identifiers. Queries have 5s deadline, indexed <=90-day ranges, preaggregated counters. No team-follow feature invented; team libraries/notifications attributed only to explicit preferred team. Empty state/collection start explicit. d809ebd, 8ead0ae, ee4ed3e. Actual PG+HTTP target suite, 42 reading and 2 DTO frontend tests, frontend boundaries and web:check (0 errors/warnings) pass. Added full browser metrics regression; pending integrated browser run and release/acceptance.
<!-- SECTION:NOTES:END -->
