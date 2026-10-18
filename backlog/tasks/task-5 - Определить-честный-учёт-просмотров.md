---
id: TASK-5
title: Определить честный учёт просмотров
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 21:58'
labels:
  - stage-2
dependencies: []
references:
  - UPDATE.md
ordinal: 5
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-2. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Если вводится учёт просмотров, сначала определить событие и защиту от повторов/ботов, затем добавить честную цифру. До этого вместо просмотров не рисовать фиктивные значения.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Define truthful event/privacy contracts in ARCHITECTURE; persist transactional daily events and distinct reading completions; expose bounded 7/30/90-day scoped dashboards with accessible tables; verify authorization, deduplication, migration and UI.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented contract in docs/ARCHITECTURE.md. No public views counter: media/image hits, preloads and anonymous requests excluded. Dashboard completion is explicitly self-reported accepted last-page progress, deduplicated account/translation/UTC-day, not proof of human reading. 90-day activity retention and no IP/email/fingerprint collection. Code d809ebd, 8ead0ae, ee4ed3e; actual PostgreSQL/reading/HTTP tests passed 2026-10-19. Release/acceptance of same SHA remains pending.
<!-- SECTION:NOTES:END -->
