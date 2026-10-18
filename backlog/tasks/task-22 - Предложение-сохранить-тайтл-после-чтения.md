---
id: TASK-22
title: Предложение сохранить тайтл после чтения
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
ordinal: 22
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-5. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Во время чтения после 3–4 **разных завершённых глав одного тайтла**, если он ещё не в библиотеке, ненавязчиво предложить добавить. Не показывать гостю как обманчивое сохранение, не повторять после отказа в той же сессии, не блокировать чтение.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Persist distinct chapter completions, query authenticated library eligibility, show a nonblocking reader prompt with per-session dismissal; verify policies and integration.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Distinct account/title/logical-chapter completion receipts; same chapter across translation/revision cannot increase count. At 3 completed chapters, server checks absent/not-removed library and reader shows nonblocking inline prompt only for authenticated accepted progress. Refusal persisted per account/title in sessionStorage; guest/account-change/late-response tests pass. d809ebd + 8ead0ae; 42 frontend reading tests and actual PostgreSQL completion-threshold/library tests pass. Local only; final integrated browser/release acceptance pending.
<!-- SECTION:NOTES:END -->
