---
id: TASK-27
title: Исправить подсветку поля в диалоге поиска
status: To Do
assignee: []
created_date: '2026-10-18 21:21'
labels:
  - ui
  - bug
dependencies: []
references:
  - apps/web/src/lib/shared/presentation/ui/AppHeader.svelte
ordinal: 1026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
В диалоге поиска пользователь заметил лишний оранжевый контур вокруг вложенного поля. Источник: замечание владельца от 19 октября 2026 года на /menu, мобильное окно 668×590. Проверить фактический фокус и сочетание стилей контейнера и input.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 В фокусе поле поиска имеет единый аккуратный контур без наложения рамок контейнера и input.
- [ ] #2 Клавиатурный фокус остаётся заметным; Escape и возврат фокуса к кнопке поиска работают.
- [ ] #3 Проверены казахский и русский, обе темы и ширины 320/360/desktop.
<!-- AC:END -->
