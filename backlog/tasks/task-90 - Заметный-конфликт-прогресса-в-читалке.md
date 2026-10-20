---
id: TASK-90
title: Заметный конфликт прогресса в читалке
status: Review
assignee:
  - '@reader-agent'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 17:18'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/reading/presentation/Reader.svelte
priority: medium
type: enhancement
ordinal: 44022
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №22. Владелец: «добавь таски и начни исполнять».

Разрешение конфликта уже есть в настройках, но о необходимости выбора должно быть понятно до их открытия.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Ненавязчивый видимый индикатор конфликта ведёт к существующему сравнению позиций.
- [x] #2 Чтение не блокируется и выбор не делается автоматически; разрешение конфликта и смена аккаунта корректно обновляют индикатор.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Сверить контракт и добавить регрессию. 2. Узкое изменение с account scope и локалями. 3. Unit/UI/browser проверки и SHA; выпуск отдельно.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
4b8499b: видимый неблокирующий индикатор конфликта ведёт к существующему сравнению. Account-change очищает session/индикатор. Browser regression authored в reader-improvements; runtime pending root. Локально текущие svelte-check 0 errors/warnings, frontend boundaries pass, UI45 pass; выпуска не подтверждают. Следующий шаг: root serial browser/runtime, independent review и authorised release acceptance.

83af48a: browser regression исправлена на естественный HTTP409 progress conflict (мок outcome, актуальная cloud progress из авторизованного GET), сравнение открывается без блокировки чтения, account-change убирает индикатор. Прогон реального browser на disposable local stack PASS 3.8s, tmp/task72/reader-conflict-http-fixed.log. Предыдущая IDB fixture была нестабильна и удалена.

Локальная общая приёмка: notifications/offline/navigation8/8, avatar/library/metrics4/4, reader-improvements3/3 PASS; global Escape исправлен83af48a, EXIF JPEG891f3c9 и PNG bce8ee8 согласованы с фактическим Chrome. Полный Go/PostgreSQL PASS, identity/HTTP послеPNG PASS;146module+46UI,Svelte0/0,architecture/contrast/preview PASS. Повторное independent review без замечаний; physicalAndroid не проводился. Выпуск ещё не подтверждён.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.

CI follow-up 2026-10-20: Verify 39462633274 passed shard 0, shard 1 exposed a test timeout shorter than the normal 2s local save + 15s cloud flush. Regression now awaits the real POST 409 (bounded 25s), preserving conflict visibility, nonblocking pages and account-change assertions. Local reader-improvements 3/3 PASS. Independent review_reader review: no findings. Product code unchanged; full CI rerun required after push.

Final remote verification 2026-10-20: SHA 4145910eb7bc649aaca2a2e890a44fbfa5f75575, Verify Qmanga https://github.com/maqsatto/qmanga/actions/runs/39463685921 SUCCESS (both acceptance/browser shards, Go/PostgreSQL/codecs and frontend checks); Fast Release https://github.com/maqsatto/qmanga/actions/runs/39463685942 SUCCESS. Earlier CI follow-up failures are resolved. External production still reports 36b3a5f455eb37830408af44c2e2f00b48eabd51; deployment of this package is not confirmed, status remains Review. Physical Android TASK-97 remains not_run; new imports TASK-99–103 are queued only.
<!-- SECTION:NOTES:END -->
