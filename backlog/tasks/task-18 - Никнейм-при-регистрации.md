---
id: TASK-18
title: Никнейм при регистрации
status: Review
assignee:
  - '@codex-identity'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 21:54'
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
- [ ] #1 Регистрация спрашивает никнейм сразу; проверяет длину, пробелы и недопустимые символы на клиенте и сервере. Существующие аккаунты с пустым `display_name` продолжают работать и могут заполнить его в настройках. Не показывать email как публичное имя.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Require a validated nickname at registration while preserving legacy accounts. 2. Add domain, transport and client regression checks. 3. Record verified results and release acceptance remaining.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented required nickname across registration UI/application/Go handler; Unicode letters/numbers and single internal spaces plus _.-; legacy empty profiles remain usable. Existing settings can fill nickname; CreateAccount now persists display_name. Verified bun domain test 1 pass/11 assertions; svelte-check 0 errors/warnings; go identity domain/application/postgres passes (DB integration skipped without TEST_DATABASE_URL). Full transport test pending concurrent catalog changes and parent real stack/browser acceptance. Release not verified.
<!-- SECTION:NOTES:END -->
