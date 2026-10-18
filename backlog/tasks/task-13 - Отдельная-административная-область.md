---
id: TASK-13
title: Отдельная административная область
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
ordinal: 13
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-4. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 В меню оставить одну ясную кнопку «Управление» только для администратора. Она открывает отдельную полноэкранную административную область с левой боковой навигацией на компьютере и выдвижной навигацией на телефоне. Разделы: обзор, тайтлы/главы, заявки/команды, жанры/категории/подборки, модерация, пользователи, обращения, аудит. Действующие права модератора и администратора не смешивать.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Implement approved task criteria in existing architecture; add behavioral regressions, verify Docker integration and keep Review until CI/release. Legal/contact approval remains explicit.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Separate full-height administrative shell with sticky left desktop navigation and native mobile drawer. /admin overview, /admin/catalog titles/chapters, /admin/teams, /admin/requests, /admin/taxonomy, /admin/moderation, /admin/accounts, /admin/reports, /admin/audit. Server layout verifies authenticated+verified admin; underlying APIs retain authorization. Moderators retain /moderation without admin access. Menu offers one Management link. Browser checks pass kk/ru x light/dark,320/360/desktop; guestredirect and reader403 tested. Native close/Escape/focus, reducedmotion and no overflow verified. Remote CI/release pending.

Ревью владельца завершено для23 карточек на базе3302515: проверены реализация, права доступа, границы API/хранилищ и интеграция. Итоговые локальные проверки:106 Go-тестов/подтестов,114 прикладных и29 SSR-тестов,40/40 браузерных сценариев;0 ошибок и0 пропусков, Svelte0 ошибок/предупреждений. Логи tmp/review-complete-go.jsonl, tmp/review-complete-web-build.log, tmp/review-final-browser-acceptance.log. Точные auth/publishing/reading CI smoke пройдены на настоящих PostgreSQL/MinIO/Mailpit. Это локальная приёмка; Review сохранён до удалённого CI, staging и наблюдаемого выпуска того же SHA. Для TASK17 дополнительно нужны утверждение правовых текстов владельцем и действующий ящик поддержки; для TASK26 остаётся предрелизная проверка consent/contact.
<!-- SECTION:NOTES:END -->
