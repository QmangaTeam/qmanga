---
id: TASK-88
title: Поиск в редакторе таксономии
status: To Do
assignee: []
created_date: '2026-10-20 15:13'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/catalog/presentation/AdminTaxonomy.svelte
priority: medium
type: enhancement
ordinal: 44020
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №20. Владелец: «добавь таски и начни исполнять».

В AdminTaxonomy длинный select; это не замена уже работающего TagSearch в каталоге.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Тег находится по казахскому и русскому названию, различимые группы помогают выбрать одноимённые метки.
- [ ] #2 Выбор/очистка поиска не меняют ID или данные без сохранения; CRUD и локали остаются совместимыми.
<!-- AC:END -->
