---
id: TASK-74
title: Возврат из предпросмотра в рабочее место
status: To Do
assignee: []
created_date: '2026-10-20 15:13'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/publishing/presentation/PublishingStudio.svelte
priority: high
type: enhancement
ordinal: 44006
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №6. Владелец: «добавь таски и начни исполнять».

После проверки главы переводчику нужно продолжить ту же версию и шаг без поиска позиции заново.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Возврат из предпросмотра восстанавливает проект, версию, рабочую секцию и позицию.
- [ ] #2 Прямые ссылки, Back/reload и недоступная версия имеют честный fallback; приватность preview и account isolation сохраняются.
<!-- AC:END -->
