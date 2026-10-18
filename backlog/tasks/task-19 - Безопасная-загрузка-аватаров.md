---
id: TASK-19
title: Безопасная загрузка аватаров
status: Review
assignee:
  - '@codex-identity'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 22:01'
labels:
  - stage-5
dependencies: []
references:
  - UPDATE.md
ordinal: 19
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-5. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Пользователь и команда загружают аватар через Go-медиа шлюз: JPEG/PNG/WebP, лимит размера, кадрирование/уменьшение и удаление метаданных. Свой аватар меняет владелец аккаунта, командный — её владелец/редактор по правам. Пустое состояние использует инициалы.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add bounded JPEG/PNG/WebP avatar processing and private storage metadata. 2. Authorize user ownership and team owner/editor roles through existing access port. 3. Expose same-origin Go media and upload routes, settings/team controls, tests.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented bounded 5 MiB JPEG/PNG/WebP upload, 8M pixel cap, center crop to 256px JPEG, metadata stripping by re-encode, four concurrent upload/media slots and deadlines. Own-user permission and verified team owner/editor permission checked before decoding. Private MinIO derivative storage, PostgreSQL current-avatar lookup, approved-team/nonblocked-user media visibility; old object keys never public. Same-origin media supports GET/HEAD/ETag/ranges with authorization before304. Profile and team editor controls with initials fallback. Go identity/teams/HTTP tests pass (DB-gated cases require parent real-stack run); frontend check pending composition wiring. No release acceptance yet.

Follow-up: bootstrap/browser composition wired; svelte-check zero errors. Added application ownership-before-processing tests, image decoder tests, PostgreSQL visibility test and real browser register/avatar/media suite. Parent will run current real stack before acceptance; no deployed release confirmed.

Security follow-up: uploads require original rendered actor via X-Qmanga-Account; fresh CSRF after cookie switch cannot overwrite a different account. Added typed actor header regression. Objects get durable pending-GC reservation before upload; replacement transaction queues only old key, failed persistence queues unreferenced new key, cleanup retries boundedly and excludes active keys; added migration000028 and cleanup tests. Absent avatar returns local404/HEAD no body. Confirmed Go identity/HTTP/S3 tests and svelte-check0errors0warnings. Header refresh hint published after successful own-avatar upload.

Implementation SHAs:22c8b47 (avatar pipeline/UI) andef70106 (accountbinding, durable garbage cleanup, missing-avatar404). Remaining review: parent current-build real browser suite and deployed release; public team-avatar presentation delegated to catalog owner if included in this acceptance.
<!-- SECTION:NOTES:END -->
