---
id: TASK-18
title: Никнейм при регистрации
status: Review
assignee:
  - '@codex-identity'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 22:31'
labels:
  - stage-5
dependencies: []
references:
  - UPDATE.md
ordinal: 18
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-5. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Регистрация спрашивает никнейм сразу; проверяет длину, пробелы и недопустимые символы на клиенте и сервере. Существующие аккаунты с пустым `display_name` продолжают работать и могут заполнить его в настройках. Не показывать email как публичное имя.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Require a validated nickname at registration while preserving legacy accounts. 2. Add domain, transport and client regression checks. 3. Record verified results and release acceptance remaining.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented required nickname across registration UI/application/Go handler; Unicode letters/numbers and single internal spaces plus _.-; legacy empty profiles remain usable. Existing settings can fill nickname; CreateAccount now persists display_name. Verified bun domain test 1 pass/11 assertions; svelte-check 0 errors/warnings; go identity domain/application/postgres passes (DB integration skipped without TEST_DATABASE_URL). Full transport test pending concurrent catalog changes and parent real stack/browser acceptance. Release not verified.

Implementation SHAs:42e1d18 (nickname) and76654ff (strict login fixture correction). Parent real PostgreSQL focused flow subsequently passed; current release/browser acceptance remains tracked by parent.

Actual disposable-stack browser suite now verifies missingnickname and email-shaped nickname both422/error.code validation, valid Kazakh nickname persisted through login, and nickname input minlength2. Targetedbrowser1test passed; finaldeployedSHA remains pending.

Owner-requested code review confirmed client/server nickname validation, legacy-profile compatibility and public-name fallback without exposing email; exact CI registration smoke now includes display_name (d044eea). Local acceptance from real PostgreSQL/MinIO/Mailpit and browser passes; remote CI/staging/release remains required.

Ревью владельца завершено для23 карточек на базе3302515: проверены реализация, права доступа, границы API/хранилищ и интеграция. Итоговые локальные проверки:106 Go-тестов/подтестов,114 прикладных и29 SSR-тестов,40/40 браузерных сценариев;0 ошибок и0 пропусков, Svelte0 ошибок/предупреждений. Логи tmp/review-complete-go.jsonl, tmp/review-complete-web-build.log, tmp/review-final-browser-acceptance.log. Точные auth/publishing/reading CI smoke пройдены на настоящих PostgreSQL/MinIO/Mailpit. Это локальная приёмка; Review сохранён до удалённого CI, staging и наблюдаемого выпуска того же SHA. Для TASK17 дополнительно нужны утверждение правовых текстов владельцем и действующий ящик поддержки; для TASK26 остаётся предрелизная проверка consent/contact.
<!-- SECTION:NOTES:END -->
