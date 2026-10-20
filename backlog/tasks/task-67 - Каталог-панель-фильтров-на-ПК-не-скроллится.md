---
id: TASK-67
title: 'Каталог: панель фильтров на ПК не скроллится'
status: In Progress
assignee: []
created_date: '2026-10-20 03:29'
labels:
  - catalog
  - frontend
dependencies: []
type: bug
ordinal: 42026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Дефект к TASK-65. На desktop (>=1024px) правая панель фильтров обрезает содержимое: .panel-clip имеет только max-height, поэтому .panel-scroll с height:100% разрешается в auto и не становится скролл-контейнером; нижние секции (доступность и др.) недостижимы. Мобильный полноэкранный диалог скроллится нормально. Исправление: задать .panel-clip определённую высоту height:calc(100dvh - 166px).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Desktop 1280: .panel-scroll имеет overflow-y:auto и scrollHeight>clientHeight, скролл доходит до последней секции; контент не обрезается.
- [ ] #2 Регрессия в apps/web/tests/browser/catalogue-filters.test.mjs падает до фикса и проходит после.
- [ ] #3 Мобильная панель 320/360 скроллится как раньше; kk/ru, обе темы, нет горизонтального переполнения.
- [ ] #4 Выпущено на production, внешняя проверка того же SHA.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1) Фикс CSS .panel-clip. 2) Регрессия в браузерном тесте. 3) Локальная проверка через браузер. 4) Fast Release и внешняя проверка скролла.
<!-- SECTION:PLAN:END -->
