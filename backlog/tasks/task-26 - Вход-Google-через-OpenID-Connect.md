---
id: TASK-26
title: Вход Google через OpenID Connect
status: Review
assignee:
  - '@codex-identity'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 23:04'
labels:
  - stage-7
dependencies:
  - TASK-25
  - TASK-18
references:
  - UPDATE.md
ordinal: 26
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-7. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Использовать OpenID Connect для **входа**, минимальные `openid email profile`; не запрашивать Gmail/Drive или другие данные «на всякий случай». Сервер проверяет `state`, ID token, issuer, audience, expiry, `email_verified` и устойчивый `sub`; создаёт обычную Qmanga-сессию. Связывание с существующим email требует безопасного подтверждения, без захвата аккаунта одним совпадением email.
- [ ] #2 На регистрации через Google запросить никнейм, если он ещё не задан. Отмена входа возвращает пользователя без потери текущей страницы. Сохраняется обычный вход по почте. Перед выпуском проверить consent screen и ограничения Google на тестовых пользователях.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Implement server OIDC authorization-code flow with state/nonce/PKCE and verified token claims. 2. Persist one-use flows and stable issuer/sub identities; refuse email-only linking. 3. Add disabled-until-configured UI, nickname registration and cancellation return path; test security decisions.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented server-only Google OIDC via coreos/go-oidc and oauth2: RS256 signature/issuer/audience/expiry, azp, nonce, verified email/stable sub; state cookie binding, ten-minute one-use PostgreSQL flow, S256 PKCE, ordinary Qmanga sessions. Existing email match explicitly refuses linking (email/password remains available), no access/refresh tokens persisted. New account requires nickname; cancellation restores safe original path. Google status/button disabled with absent server config; partial config rejected. Verified signed-token rejection tests, application replay/claims/redirect tests, HTTP disabled/cookie checks, Go identity/config/transport/bootstrap pass; frontend check zero errors (one concurrent Footer fixture warning). PostgreSQL integration and browser tests added for parent current-stack run. External owner Cloud client/consent/test-user and deployment acceptance remain pending TASK-25; not released.

Implementation SHA:d170196. Server bootstrap/Compose wiring belongs to parent integration commit. Real Google owner-project consent/client verification remains external; locally test signed mock-provider behavior and disabled UI only. Existing email accounts are never auto-linked; self-service linking is deliberately unavailable.

Actual Google end-to-end verification now succeeded in owner browser after explicit consent: initial login requested missingnickname, owner selected maqsatto, second flow verified IDtoken and created standard session returning /settings with verifiedemail. Console/secret configured TASK25, rootlocalenabled, isolatedsandbox remains disabled for regression. Publicrelease still requires consentbranding/approvedlegalpages/supportmailbox/deployment.

Review fix7548685 routes successful Google sessions through a one-use 60-second session-bound HttpOnly proof before awaited browser cache cleanup and cross-tab account notification. Cancel/error/unmarked/replayed/wrong-session completion does not clear caches. Actual Google repeat sign-in was rechecked on rebuilt local API and returned /settings as maqsatto without a completion loop. Backend token/transport, server proof and browser isolation regressions cover the boundary. AC2 retains the external pre-release consent/contact acceptance gate.

Ревью владельца завершено для23 карточек на базе3302515: проверены реализация, права доступа, границы API/хранилищ и интеграция. Итоговые локальные проверки:106 Go-тестов/подтестов,114 прикладных и29 SSR-тестов,40/40 браузерных сценариев;0 ошибок и0 пропусков, Svelte0 ошибок/предупреждений. Логи tmp/review-complete-go.jsonl, tmp/review-complete-web-build.log, tmp/review-final-browser-acceptance.log. Точные auth/publishing/reading CI smoke пройдены на настоящих PostgreSQL/MinIO/Mailpit. Это локальная приёмка; Review сохранён до удалённого CI, staging и наблюдаемого выпуска того же SHA. Для TASK17 дополнительно нужны утверждение правовых текстов владельцем и действующий ящик поддержки; для TASK26 остаётся предрелизная проверка consent/contact.

Remote10d39b8 CI passed and exactSHA serves staging. Deployed /api/v1/auth/google/status currently reports enabled:false because serverOAuth pair is not yet installed. Localowner Google acceptance remains valid; remoteconfiguration/consent/contact and release acceptance still pending. Keep Review.
<!-- SECTION:NOTES:END -->
