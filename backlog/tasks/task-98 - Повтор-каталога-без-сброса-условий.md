---
id: TASK-98
title: Повтор каталога без сброса условий
status: To Do
assignee: []
created_date: '2026-10-20 15:13'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/catalog/presentation/Catalogue.svelte
priority: low
type: bug
ordinal: 44030
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №30. Владелец: «добавь таски и начни исполнять».

В ветке unavailable ссылка Повторить ведёт на /catalog и теряет query/filter/sort/cursor. Утверждённое устройство фильтров не меняется.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Retry после ошибки повторяет исходную выдачу с q, фильтрами, сортировкой и cursor.
- [ ] #2 Browser-регрессия воспроизводит сбой/восстановление; пустой результат, история и локаль остаются корректными.
<!-- AC:END -->
