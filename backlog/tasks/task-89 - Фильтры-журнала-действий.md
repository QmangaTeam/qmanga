---
id: TASK-89
title: Фильтры журнала действий
status: To Do
assignee: []
created_date: '2026-10-20 15:13'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/identity/presentation/AuditLog.svelte
priority: medium
type: enhancement
ordinal: 44021
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №21. Владелец: «добавь таски и начни исполнять».

Для расследования изменения нужно искать по исполнителю, объекту, действию и времени, а не читать всю ленту.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Фильтры охватывают весь доступный журнал и совместимы с server cursor и стабильным порядком.
- [ ] #2 Права, приватность и читаемые before/after сохранены; запрос видим, пустое состояние отделено от ошибки.
<!-- AC:END -->
