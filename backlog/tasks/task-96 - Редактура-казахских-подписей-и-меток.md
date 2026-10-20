---
id: TASK-96
title: Редактура казахских подписей и меток
status: Review
assignee:
  - '@codex'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 16:52'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/catalog/presentation/messages.ts
priority: low
type: enhancement
ordinal: 44028
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №28. Владелец: «добавь таски и начни исполнять».

В интерфейсе и метках нужны последовательные термины; содержимое манги и собственные названия не переводить автоматически.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Подтверждён список непонятных или смешанных UI/таксономических подписей и согласованы корректные варианты.
- [x] #2 Изменения не меняют slug/ID и смысл фильтров; длинные казахские строки проверены в обеих темах на телефоне.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Сверить фактические формы/контракт; воспроизвести недостающий сценарий; выполнить ограниченную реализацию, проверить локали/узкие экраны/ошибки и сохранить результат с SHA.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Редактура по текущим наблюдаемым UI-строкам: статус draft и действия студии приведены к Нобай/нобай (винительный Нобайды), чтобы не путать с Қаралым (просмотры) и проектом; ready label не обещает отправку для роли member. Значение Нобай сверено по Сөздікқор https://sozdikqor.kz/soz?a=NOBAI&id=139344. Это рабочее решение в рамках авторизованной редактуры; не заявляется профессиональная проверка носителем. ID/slug/названия произведений и таксономические данные не менялись. Unit regression messages.test.ts pass. kk/ru обе темы и320/360/desktop покрываются publishing-steps/studio-workflow; финальный browser прогон продолжается. SHA ce0799f. Не выпущено.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.
<!-- SECTION:NOTES:END -->
