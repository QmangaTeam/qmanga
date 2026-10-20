---
id: TASK-84
title: Подсветка фрагментов в сравнении изменений тайтла
status: To Do
assignee: []
created_date: '2026-10-20 15:13'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/catalog/presentation/TitleChangeDetail.svelte
priority: medium
type: enhancement
ordinal: 44016
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №16. Владелец: «добавь таски и начни исполнять».

Сравнение уже выделяет изменившиеся поля, но длинные описания приходится сопоставлять целиком.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 В изменённых длинных текстах различимы добавленные и удалённые фрагменты при сохранении before/current/proposed.
- [ ] #2 Подсветка безопасна для HTML, читается без одного цвета и не меняет исходный текст или версионные проверки.
<!-- AC:END -->
