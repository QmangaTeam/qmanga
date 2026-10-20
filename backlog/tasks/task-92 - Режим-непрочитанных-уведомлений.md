---
id: TASK-92
title: Режим непрочитанных уведомлений
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
  - apps/web/src/lib/features/notifications/presentation/Notifications.svelte
priority: medium
type: enhancement
ordinal: 44024
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №24. Владелец: «добавь таски и начни исполнять».

Новые события нужно отделять от просмотренной истории без ручного перебора; счётчик и чтение одного события уже существуют.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Режим непрочитанных фильтрует полную авторизованную выдачу, а не только загруженную страницу.
- [x] #2 Чтение события обновляет список/счётчик без потери контекста; смена аккаунта, pagination и ошибки проверены.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Сверить контракт и добавить регрессию. 2. Узкое изменение с account scope и локалями. 3. Unit/UI/browser проверки и SHA; выпуск отдельно.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
5bfa055: server-side unread filter across cursor pages; client generation/account isolation, read removes row preserving cursor. Unit 9 feed tests pass. Backend Go package tests pass, реальный PostgreSQL test добавлен, без TEST_DATABASE_URL не запускался. Browser regression authored, runtime pending root. Локально текущие svelte-check 0 errors/warnings, frontend boundaries pass, UI45 pass; выпуска не подтверждают. Следующий шаг: root serial browser/runtime, independent review и authorised release acceptance.

Локальная общая приёмка: notifications/offline/navigation8/8, avatar/library/metrics4/4, reader-improvements3/3 PASS; global Escape исправлен83af48a, EXIF JPEG891f3c9 и PNG bce8ee8 согласованы с фактическим Chrome. Полный Go/PostgreSQL PASS, identity/HTTP послеPNG PASS;146module+46UI,Svelte0/0,architecture/contrast/preview PASS. Повторное independent review без замечаний; physicalAndroid не проводился. Выпуск ещё не подтверждён.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.
<!-- SECTION:NOTES:END -->
