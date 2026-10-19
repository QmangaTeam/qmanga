---
id: TASK-26
title: Вход Google через OpenID Connect
status: Review
assignee:
  - '@codex-identity'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-19 01:07'
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
- [ ] #2 При первом входе через Google почта, ник и аватар берутся из профиля Google; отдельный ручной ввод ника не требуется. Отмена сохраняет исходную страницу, обычный вход по почте остаётся доступен.
- [ ] #3 Перед выпуском проверены Google consent screen и ограничения доступа; реальный вход на production создаёт обычную сессию с данными Google.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Implement server OIDC authorization-code flow with state/nonce/PKCE and verified token claims. 2. Persist one-use flows and stable issuer/sub identities; refuse email-only linking. 3. Add disabled-until-configured UI, nickname registration and cancellation return path; test security decisions.

Owner profile defaults: add verified Google name/picture claims, derive safe nickname without manual prompt, import bounded Google-host avatar through existing sanitizer only for newly created accounts; preserve existing profiles and stable identity security. Verify unit/provider/real PostgreSQL regressions; parent owns production wiring and release.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented server-only Google OIDC via coreos/go-oidc and oauth2: RS256 signature/issuer/audience/expiry, azp, nonce, verified email/stable sub; state cookie binding, ten-minute one-use PostgreSQL flow, S256 PKCE, ordinary Qmanga sessions. Existing email match explicitly refuses linking (email/password remains available), no access/refresh tokens persisted. New account requires nickname; cancellation restores safe original path. Google status/button disabled with absent server config; partial config rejected. Verified signed-token rejection tests, application replay/claims/redirect tests, HTTP disabled/cookie checks, Go identity/config/transport/bootstrap pass; frontend check zero errors (one concurrent Footer fixture warning). PostgreSQL integration and browser tests added for parent current-stack run. External owner Cloud client/consent/test-user and deployment acceptance remain pending TASK-25; not released.

Implementation SHA:d170196. Server bootstrap/Compose wiring belongs to parent integration commit. Real Google owner-project consent/client verification remains external; locally test signed mock-provider behavior and disabled UI only. Existing email accounts are never auto-linked; self-service linking is deliberately unavailable.

Actual Google end-to-end verification now succeeded in owner browser after explicit consent: initial login requested missingnickname, owner selected maqsatto, second flow verified IDtoken and created standard session returning /settings with verifiedemail. Console/secret configured TASK25, rootlocalenabled, isolatedsandbox remains disabled for regression. Publicrelease still requires consentbranding/approvedlegalpages/supportmailbox/deployment.

Review fix7548685 routes successful Google sessions through a one-use 60-second session-bound HttpOnly proof before awaited browser cache cleanup and cross-tab account notification. Cancel/error/unmarked/replayed/wrong-session completion does not clear caches. Actual Google repeat sign-in was rechecked on rebuilt local API and returned /settings as maqsatto without a completion loop. Backend token/transport, server proof and browser isolation regressions cover the boundary. AC2 retains the external pre-release consent/contact acceptance gate.

Ревью владельца завершено для23 карточек на базе3302515: проверены реализация, права доступа, границы API/хранилищ и интеграция. Итоговые локальные проверки:106 Go-тестов/подтестов,114 прикладных и29 SSR-тестов,40/40 браузерных сценариев;0 ошибок и0 пропусков, Svelte0 ошибок/предупреждений. Логи tmp/review-complete-go.jsonl, tmp/review-complete-web-build.log, tmp/review-final-browser-acceptance.log. Точные auth/publishing/reading CI smoke пройдены на настоящих PostgreSQL/MinIO/Mailpit. Это локальная приёмка; Review сохранён до удалённого CI, staging и наблюдаемого выпуска того же SHA. Для TASK17 дополнительно нужны утверждение правовых текстов владельцем и действующий ящик поддержки; для TASK26 остаётся предрелизная проверка consent/contact.

Remote10d39b8 CI passed and exactSHA serves staging. Deployed /api/v1/auth/google/status currently reports enabled:false because serverOAuth pair is not yet installed. Localowner Google acceptance remains valid; remoteconfiguration/consent/contact and release acceptance still pending. Keep Review.

Уточнение владельца: «аву сразу с гугла ник с гугла и почту тоже». Заменяет прежнее требование обязательного ручного ввода ника при Google-регистрации. Дополнительное описание: использовать данные Google как начальные данные нового профиля; повторный вход не должен затирать сделанные пользователем изменения профиля.

Google profile defaults implemented in 78323df: verified signed name/picture claims initialize new users; safe nickname derived from Google name (email local-part/fixed fallback when unsuitable), verified email retained. No manual Google nickname required; repeat issuer/sub login never changes edited name/email/avatar. New AvatarImporter permits only HTTPS lh3–lh6.googleusercontent.com, denies redirects and non-public resolved addresses, pins checked DNS addresses, bounds response to 5 MiB/8 seconds, then calls existing sanitized origin-avatar Upload pipeline. Picture unavailable/invalid leaves normal fallback without blocking login; no hotlink or new scope. No migration required. Verified: initial nickname regression failed before implementation; SSRF CGN/0-range regression failed then passed. Full go test ./... passed on actual PostgreSQL; after final IP hardening identity/... and transport/http/... passed again. bun run web:check: 0 errors/warnings; git diff --check clean. Parent owns bootstrap wiring (construct avatarService before NewGoogle and pass googleoidc.NewAvatarImporter(avatarService)), new google-profile browser regression run, independent review, production OAuth configuration and release acceptance. No production action, push, service restart or shared browser fixture performed by identity agent. AC2 browser integration remains pending parent verification; AC3 production acceptance remains pending.
<!-- SECTION:NOTES:END -->
