---
id: TASK-21
title: Предпочтительная команда из читаемой главы
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 22:31'
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

Additional actual PostgreSQL regressions cover no-reading/multiple-project blank selection, one-project blank selection, restoration inference and explicit later clear. Restore correction ab0ec11. Real browser reading-suggestion.test.mjs PASS (two tests, 29.4s) verifies preference from actually read alternate teamB, editable via real Title dialog, and prompt Add restores current teamA.

Ревью владельца завершено для23 карточек на базе3302515: проверены реализация, права доступа, границы API/хранилищ и интеграция. Итоговые локальные проверки:106 Go-тестов/подтестов,114 прикладных и29 SSR-тестов,40/40 браузерных сценариев;0 ошибок и0 пропусков, Svelte0 ошибок/предупреждений. Логи tmp/review-complete-go.jsonl, tmp/review-complete-web-build.log, tmp/review-final-browser-acceptance.log. Точные auth/publishing/reading CI smoke пройдены на настоящих PostgreSQL/MinIO/Mailpit. Это локальная приёмка; Review сохранён до удалённого CI, staging и наблюдаемого выпуска того же SHA. Для TASK17 дополнительно нужны утверждение правовых текстов владельцем и действующий ящик поддержки; для TASK26 остаётся предрелизная проверка consent/contact.
<!-- SECTION:NOTES:END -->
