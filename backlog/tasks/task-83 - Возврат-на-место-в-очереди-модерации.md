---
id: TASK-83
title: Возврат на место в очереди модерации
status: To Do
assignee: []
created_date: '2026-10-20 15:13'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/catalog/presentation/TitleChangeQueue.svelte
priority: medium
type: enhancement
ordinal: 44015
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №15. Владелец: «добавь таски и начни исполнять».

Проверка объекта и решение не должны заставлять модератора искать прежнюю позицию заново.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Возврат из detail/preview восстанавливает очередь, критерии и позицию.
- [ ] #2 После решения показан результат и понятный следующий объект; stale/уже обработанные записи не приводят к повторному решению.
<!-- AC:END -->
