---
id: TASK-91
title: Статусы в смешанной библиотеке
status: Review
assignee:
  - '@reader-agent'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 16:52'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/reading/presentation/LibraryRow.svelte
priority: medium
type: enhancement
ordinal: 44023
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №23. Владелец: «добавь таски и начни исполнять».

Гипотеза: в общих списках трудно сравнивать статусы, спрятанные в меню. Проверить компактные отметки, сохраняя утверждённые сетку и меню ⋮.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 На прототипе с разными статусами проверена польза компактных отметок в общем списке.
- [x] #2 При принятии варианта статус взят из реальных данных, не перекрывает обложку/действия; сетка и меню ⋮ сохранены.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Сверить контракт и добавить регрессию. 2. Узкое изменение с account scope и локалями. 3. Unit/UI/browser проверки и SHA; выпуск отдельно.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
980d876: компактная подпись реального статуса вне меню и обложки. SSR regression mixed-status pass. Это экспертный walkthrough прототипа, не пользовательское исследование; runtime mobile/themes pending root. Локально текущие svelte-check 0 errors/warnings, frontend boundaries pass, UI45 pass; выпуска не подтверждают. Следующий шаг: root serial browser/runtime, independent review и authorised release acceptance.

Локальная общая приёмка: notifications/offline/navigation8/8, avatar/library/metrics4/4, reader-improvements3/3 PASS; global Escape исправлен83af48a, EXIF JPEG891f3c9 и PNG bce8ee8 согласованы с фактическим Chrome. Полный Go/PostgreSQL PASS, identity/HTTP послеPNG PASS;146module+46UI,Svelte0/0,architecture/contrast/preview PASS. Повторное independent review без замечаний; physicalAndroid не проводился. Выпуск ещё не подтверждён.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.
<!-- SECTION:NOTES:END -->
