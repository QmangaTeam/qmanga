---
id: TASK-77
title: Следующее действие на карточке проекта
status: In Progress
assignee:
  - '@studio-agent'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 16:20'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/publishing/presentation/ProjectCard.svelte
priority: high
type: enhancement
ordinal: 44009
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №9. Владелец: «добавь таски и начни исполнять».

Карточки TASK-33 уже содержат обложку, команду, число глав и оценку; дополнение должно отражать текущую работу, а не выводить её из счётчиков.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Состояние рабочей версии или ожидающее решения действие получено из авторизованного API-контракта.
- [ ] #2 Карточка ведёт к нужной работе; неизвестное состояние не изображается как готовое, текущие данные карточки сохранены.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Проверить контракт и добавить регрессии. 2. Реализовать согласованное изменение без новых правил и прав. 3. Выполнить проверки, записать SHA и ограничения приёмки.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
c1a491a,20f1853: авторизованный StudioProjectPage выбирает рабочую ревизию в той же scoped SQL/repeatable-read выборке; аддитивные next_revision_id/state и карточка с deep-link. Ready подпись не обещает право отправки. Добавлен PostgreSQL integration сценарий scoped next work/foreign/no work. Локально: svelte-check 0 ошибок/0 предупреждений; test:ui 46 pass; 9 domain/application тестов pass; frontend architecture pass; node --check studio-workflow.test.mjs pass. Browser studio-workflow.test.mjs/publishing-steps.test.mjs/project-cards.test.mjs и PostgreSQL integration на итоговом image ещё не выполнены; выпуск/независимая приёмка не подтверждены. Следующий шаг: общий rebuild и реальные browser/Go проверки координатором, затем Review; production SHA проверяет координатор.
<!-- SECTION:NOTES:END -->
