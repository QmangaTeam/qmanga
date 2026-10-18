---
id: TASK-14
title: Меню аккаунта по аватару
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
ordinal: 14
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-4. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Нажатие на аватар в шапке открывает меню на месте: профиль/настройки, библиотека, кабинет переводчика при наличии доступа, «Управление» для админа, выход. На телефоне — компактное модальное меню. Самостоятельная страница `/menu` остаётся адресуемой для нижней навигации, пока пользовательский поток не проверен.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Implement approved task criteria in existing architecture; add behavioral regressions, verify Docker integration and keep Review until CI/release. Legal/contact approval remains explicit.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Avatar opens account dialog in place on desktop/mobile. Profile/settings, library, translator cabinet only for approved membership, Management for verifiedadmin and separate moderation for verifiedmoderator, actual logout with visible failures. /menu retained for mobile navigation. Initials fallback, user avatar refresh after confirmed own upload, close44px and Escape/focus return. Browser and SSR checks pass both locales/themes and320/360/desktop. Remote acceptance/release pending.

Ревью владельца завершено для23 карточек на базе3302515: проверены реализация, права доступа, границы API/хранилищ и интеграция. Итоговые локальные проверки:106 Go-тестов/подтестов,114 прикладных и29 SSR-тестов,40/40 браузерных сценариев;0 ошибок и0 пропусков, Svelte0 ошибок/предупреждений. Логи tmp/review-complete-go.jsonl, tmp/review-complete-web-build.log, tmp/review-final-browser-acceptance.log. Точные auth/publishing/reading CI smoke пройдены на настоящих PostgreSQL/MinIO/Mailpit. Это локальная приёмка; Review сохранён до удалённого CI, staging и наблюдаемого выпуска того же SHA. Для TASK17 дополнительно нужны утверждение правовых текстов владельцем и действующий ящик поддержки; для TASK26 остаётся предрелизная проверка consent/contact.
<!-- SECTION:NOTES:END -->
