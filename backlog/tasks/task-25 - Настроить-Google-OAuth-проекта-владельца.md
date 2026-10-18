---
id: TASK-25
title: Настроить Google OAuth проекта владельца
status: To Do
assignee: []
created_date: '2026-10-18 21:20'
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
