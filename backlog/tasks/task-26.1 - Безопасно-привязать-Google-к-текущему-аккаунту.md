---
id: TASK-26.1
title: Безопасно привязать Google к текущему аккаунту
status: Review
assignee:
  - '@codex-google'
created_date: '2026-10-19 01:31'
updated_date: '2026-10-19 01:38'
labels: []
dependencies: []
parent_task_id: TASK-26
ordinal: 34026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Формулировка владельца — дословно: «Да, добавить привязку к текущему аккаунту».

Контекст подтверждения: на staging реальный Google-вход встретил существующий аккаунт с той же почтой; автоматическое связывание по одному email запрещено. Владелец подтвердил отдельную безопасную привязку в настройках после обычного входа. Дополнительное описание: связать проверенный Google issuer/sub с текущим аккаунтом, сохранив его ID, роли, проекты, библиотеку и ручные настройки профиля; дальнейший вход Google возвращает именно этот аккаунт.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 В настройках авторизованного аккаунта есть статус Google и действие привязки; после успешной привязки обычный Google-вход возвращает тот же аккаунт.
- [x] #2 Старт привязки требует действующую проверенную сессию, expected-account и CSRF; callback привязан к тому же аккаунту и сессии и отвергает выход, смену/отзыв сессии, replay и неверные Google claims.
- [x] #3 Совпадение email само по себе не связывает аккаунты; Google identity другого аккаунта не переносится. ID, роли, проекты, библиотека и пользовательские правки профиля сохраняются.
- [x] #4 Реальные PostgreSQL и HTTP/application/browser регрессии пройдены, затем дополнение проверено независимым субагентом перед выпуском.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Добавить отдельный flow привязки к account/session, переиспользуя state/nonce/PKCE и проверенный Google token. Транзакционно проверять текущую сессию/права и уникальность issuer/sub. UI настройки + account-bound gateway, сохранить обычный вход/ручной профиль. Реальный Google callback на staging, повторный login и независимый review; затем CI и prod того же SHA.

Reuse registered callback and OIDC checks; migration 000031 adds actor/session binding to existing one-use flows. Add verified session/CSRF/expected-account protected link start and private status; callback requires exact saved actor and session and repository transaction rechecks live session under locks before immutable identity insert. Add account-bound settings gateway/control and regression tests for replay, session switches/revocation, foreign identities and unchanged profile. Parent wires composition, runs shared browser, independent review and release.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented explicit linking only from authenticated verified settings account: private account-bound status and CSRF/Origin/expected-account start, existing OIDC state/nonce/PKCE callback with exact actor/session binding. Link marker is dispatch-only; saved login/link flow kinds reject downgrade and forged markers. One-use flow consumed before failure; live account/session checked again under transaction row locks and DB clock at insertion. No email-only linking or issuer/sub transfer; link inserts only external identity, preserving profile/avatar/role/ID/session/import/library/projects. Tests: initial missing-API application and gateway regressions failed before implementation; application security regressions then green. Actual PostgreSQL identity+HTTP full packages passed; latest Google-specific repository and HTTP regressions passed after DB-clock hardening. Parent full Go ./... actual PostgreSQL/codecs passed; 126/126 frontend src/lib tests passed; integrated parent Svelte check 0 errors/warnings; diff and browser syntax check clean. Independent final_owner_review reports no material findings, including final expiry SQL. Parent composition/settings-route wiring included in feature. New browser google-link.test.mjs prepared for parent shared fixture run; browser AC1/4 and actual staging/production link acceptance remain pending. Keep Review until acceptance and observed release of same SHA.

Implementation SHA: a9617ee. Parent owns shared browser/staging/release acceptance and remaining AC1/4 evidence; no deployment or push performed by identity implementer.

Родитель: полный итоговый browser прогон 50/50, без ошибок/пропусков (tmp/google-link-final-browser.log); UI linking account/CSRF, обе локали, темы,320/360/1280 прошли. Full Go ./... и повтор locked-session PostgreSQL/HTTP успешны; итоговая api/web сборка здорова. Независимый final_owner_review проверил весь addition и финальную SQL-проверку срока сессии: замечаний нет. Настоящая привязка Google существующего владельца и production release остаются отдельной приёмкой после CI/staging.
<!-- SECTION:NOTES:END -->
