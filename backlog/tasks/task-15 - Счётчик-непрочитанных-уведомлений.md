---
id: TASK-15
title: Счётчик непрочитанных уведомлений
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 22:31'
labels:
  - stage-4
dependencies: []
references:
  - UPDATE.md
ordinal: 15
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-4. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Счётчик непрочитанных уведомлений возле колокольчика: число или `99+`, доступное текстовое имя, обновление после прочтения и при возвращении на вкладку. Не показывать счётчик гостю. Список уведомлений — отдельные лёгкие карточки с явным непрочитанным состоянием.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Implement approved task criteria in existing architecture; add behavioral regressions, verify Docker integration and keep Review until CI/release. Legal/contact approval remains explicit.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Authenticated unread-count API with account binding, exact PostgreSQL count, capped 99+ display, accessible real count in bell label for header/mobile. Refresh on read via local invalidation, focus and visibility; guest count hidden. Generation/account checks reject stale in-flight results. Existing indexed user/is_read query; new real PostgreSQL test passes105 unread vs read/foreign account/idempotent marking. Browser and final integrated build pending.

Browser confirmed105 actual unread ->99+ visible, exact accessible105, marking read updates104, independent database change plusfocus refreshes4; guest hidesbadge andAPI401, foreign read cannotmarkowner. Full PostgreSQL suite103passes0actualskips. Remote acceptance/release pending.

Owner review fixed cross-account notification pagination/read state7548685 and shell refresh f34b2c7. Bound feed generations and original account checks prevent old rows/results after account switch.9da4468 completes an authorized mark-read after its link navigates away; f770ce9 delays identity response to exercise that exact unmount race. No production rate limit weakened.

Ревью владельца завершено для23 карточек на базе3302515: проверены реализация, права доступа, границы API/хранилищ и интеграция. Итоговые локальные проверки:106 Go-тестов/подтестов,114 прикладных и29 SSR-тестов,40/40 браузерных сценариев;0 ошибок и0 пропусков, Svelte0 ошибок/предупреждений. Логи tmp/review-complete-go.jsonl, tmp/review-complete-web-build.log, tmp/review-final-browser-acceptance.log. Точные auth/publishing/reading CI smoke пройдены на настоящих PostgreSQL/MinIO/Mailpit. Это локальная приёмка; Review сохранён до удалённого CI, staging и наблюдаемого выпуска того же SHA. Для TASK17 дополнительно нужны утверждение правовых текстов владельцем и действующий ящик поддержки; для TASK26 остаётся предрелизная проверка consent/contact.

Final read-navigation regression uses deterministic application unit coverage for view closure plus actual persisted PUT in browser. Browser waits for refreshed account heading before clicking after a cross-tab transition2d01db3; synthetic network hold removed0c3fb34. Final40/40 common acceptance passed.
<!-- SECTION:NOTES:END -->
