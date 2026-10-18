---
id: TASK-21
title: Предпочтительная команда из читаемой главы
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 21:58'
labels:
  - stage-5
dependencies: []
references:
  - UPDATE.md
ordinal: 21
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-5. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Если проектов несколько, определять предпочтительную команду из читаемой главы; иначе оставлять выбор пустым. Команду можно менять позже.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Infer a new library preference from accepted reading when projects compete; retain explicit later edits; verify integration.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Initial library save infers preferred team from actually accepted reading only with >1 projects. Single-project frontend auto-selection removed; no reading leaves blank. Explicit subsequent changes/clearing remain authoritative. d809ebd + 8ead0ae; actual PostgreSQL metrics regression verifies inferred team and explicit clearing; reading adapters/application/domain and HTTP tests pass. Local only; release/acceptance pending.
<!-- SECTION:NOTES:END -->
