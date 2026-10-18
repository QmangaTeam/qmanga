---
id: TASK-25
title: Настроить Google OAuth проекта владельца
status: Review
assignee:
  - '@codex-identity'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 22:31'
labels:
  - stage-7
dependencies: []
references:
  - UPDATE.md
ordinal: 25
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-7. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Для сайта нужен OAuth-клиент в Google Cloud проекта владельца, с разрешёнными доменами и redirect URI для тестового и основного адресов. Личный `deka3ax@gmail.com` — учётная запись владельца для настройки консоли, а не адрес, который надо вшивать в продукт или выдавать другим пользователям. Секреты хранятся только на сервере вне Git.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Prepare exact server-only configuration and owner-console setup steps. 2. Inspect accessible configuration without exposing secrets. 3. Record external owner-project/consent/domain acceptance requirements; never fabricate credentials.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Owner signed into Google Cloud and explicitly confirmed saving callbacks and creating additional secret. Reused existing Qmanga Web client in owner project ace-vial-473816-m8; verified authorizeddomain maqsatto.dev and saved callbacks localhost8080, qmanga-dev.maqsatto.dev and qmanga.maqsatto.dev under /api/v1/auth/google/callback. Newsecret stored only root ignored.env mode0600, API/worker restarted; real localGoogle login successfully created maqsatto and opened settings with verifiedemail. No secret inGit/tasknotes. Console remains Testing; public branding/approved legal links/working support mailbox and deployedserver secrets remain release gates. Existing supportemail was pre-existing console setting; future supportmailbox notyetactive. No publicproductionrelease claimed.

Owner-requested review reconfirmed actual repeat Google login on rebuilt local API7548685; same approved identity scopes, standard Qmanga session, no additional credentials or grants created during review. Remote deployment and production consent/contact acceptance remain pending.

Ревью владельца завершено для23 карточек на базе3302515: проверены реализация, права доступа, границы API/хранилищ и интеграция. Итоговые локальные проверки:106 Go-тестов/подтестов,114 прикладных и29 SSR-тестов,40/40 браузерных сценариев;0 ошибок и0 пропусков, Svelte0 ошибок/предупреждений. Логи tmp/review-complete-go.jsonl, tmp/review-complete-web-build.log, tmp/review-final-browser-acceptance.log. Точные auth/publishing/reading CI smoke пройдены на настоящих PostgreSQL/MinIO/Mailpit. Это локальная приёмка; Review сохранён до удалённого CI, staging и наблюдаемого выпуска того же SHA. Для TASK17 дополнительно нужны утверждение правовых текстов владельцем и действующий ящик поддержки; для TASK26 остаётся предрелизная проверка consent/contact.
<!-- SECTION:NOTES:END -->
