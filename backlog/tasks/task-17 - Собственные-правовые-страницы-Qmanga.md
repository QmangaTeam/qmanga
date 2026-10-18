---
id: TASK-17
title: Собственные правовые страницы Qmanga
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
ordinal: 17
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-4. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Добавить **собственные** страницы Qmanga: пользовательское соглашение, правила авторских прав/DMCA, публичная оферта только если реально есть платные услуги, информация об авторском праве и контакт для претензий. Ссылки на документы Senkuro из задания — пример состава футера, не тексты или адреса для Qmanga. До публикации юридические тексты проверяются владельцем/юристом, а контакт должен реально принимать письма. Подпись: © 2026 Qmanga, текущая версия сборки и короткое описание сервиса без несуществующих форматов.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Implement approved task criteria in existing architecture; add behavioral regressions, verify Docker integration and keep Review until CI/release. Legal/contact approval remains explicit.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Original bilingual Qmanga terms and copyright policy drafts implemented at /legal/terms and /legal/copyright, noindex and explicit draft status pending owner/legal review. No paid-service offer because none exists. Existing anonymous /rights workflow is actual current complaint channel. Owner supplied qmanga-support@maqsatto.dev as a future mailbox, explicitly not active yet; rendered as planned text, never a working mailto link. Footer carries build version (exact RELEASE_SHA in deployment) and copyright2026. External release gate: owner/legal approval and create/test mailbox before publishing final policies.

Added own bilingual privacy draft /legal/privacy for Google consent readiness; actual browser verifies all3legal pages, explicitdraft/noindex and futuremailbox is not advertised as mailto. Acceptance criterion intentionally remains unchecked: owner/legal approval, operator/logretention details and actual mailbox creation/receipt test are outstanding. Implemented reviewable drafts; no public legal release claimed.

Ревью владельца завершено для23 карточек на базе3302515: проверены реализация, права доступа, границы API/хранилищ и интеграция. Итоговые локальные проверки:106 Go-тестов/подтестов,114 прикладных и29 SSR-тестов,40/40 браузерных сценариев;0 ошибок и0 пропусков, Svelte0 ошибок/предупреждений. Логи tmp/review-complete-go.jsonl, tmp/review-complete-web-build.log, tmp/review-final-browser-acceptance.log. Точные auth/publishing/reading CI smoke пройдены на настоящих PostgreSQL/MinIO/Mailpit. Это локальная приёмка; Review сохранён до удалённого CI, staging и наблюдаемого выпуска того же SHA. Для TASK17 дополнительно нужны утверждение правовых текстов владельцем и действующий ящик поддержки; для TASK26 остаётся предрелизная проверка consent/contact.
<!-- SECTION:NOTES:END -->
