---
id: TASK-25
title: Настроить Google OAuth проекта владельца
status: In Progress
assignee:
  - '@codex-identity'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 21:59'
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
- [ ] #1 Для сайта нужен OAuth-клиент в Google Cloud проекта владельца, с разрешёнными доменами и redirect URI для тестового и основного адресов. Личный `deka3ax@gmail.com` — учётная запись владельца для настройки консоли, а не адрес, который надо вшивать в продукт или выдавать другим пользователям. Секреты хранятся только на сервере вне Git.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Prepare exact server-only configuration and owner-console setup steps. 2. Inspect accessible configuration without exposing secrets. 3. Record external owner-project/consent/domain acceptance requirements; never fabricate credentials.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Prepared server-only GOOGLE_CLIENT_ID/GOOGLE_CLIENT_SECRET pair validation, .env.example and development/production Compose forwarding, exact callback derivation APP_ORIGIN + /api/v1/auth/google/callback, existing RUNBOOK owner-console steps. No accessible owner Cloud project/client/secret or confirmed test/production domain setup was provided; no credentials fabricated or personal email embedded. Next: owner-authorized Cloud project Web OAuth client creation, exact domains/redirects, verified consent/contact/privacy details, test-user consent checks, secret injection outside Git. This criterion remains unfulfilled and task stays In Progress; code readiness is not owner-console completion.
<!-- SECTION:NOTES:END -->
