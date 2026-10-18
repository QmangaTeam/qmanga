---
id: TASK-26
title: Вход Google через OpenID Connect
status: To Do
assignee: []
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 21:20'
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
- [ ] #1 Использовать OpenID Connect для **входа**, минимальные `openid email profile`; не запрашивать Gmail/Drive или другие данные «на всякий случай». Сервер проверяет `state`, ID token, issuer, audience, expiry, `email_verified` и устойчивый `sub`; создаёт обычную Qmanga-сессию. Связывание с существующим email требует безопасного подтверждения, без захвата аккаунта одним совпадением email.
- [ ] #2 На регистрации через Google запросить никнейм, если он ещё не задан. Отмена входа возвращает пользователя без потери текущей страницы. Сохраняется обычный вход по почте. Перед выпуском проверить consent screen и ограничения Google на тестовых пользователях.
<!-- AC:END -->
