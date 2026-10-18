---
id: TASK-18
title: Никнейм при регистрации
status: To Do
assignee: []
created_date: '2026-10-18 21:20'
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
