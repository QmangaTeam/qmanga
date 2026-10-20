---
id: TASK-85
title: Ссылки на конкретный комментарий из уведомлений и жалоб
status: Review
assignee:
  - '@admin-agent'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 19:03'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/community/presentation/Discussion.svelte
priority: medium
type: enhancement
ordinal: 44017
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №17. Владелец: «добавь таски и начни исполнять».

Переход только на тайтл/главу теряет контекст ответа или жалобы. Нужен сквозной контракт объекта, загрузки нужной ветки и якоря.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Ссылки из уведомления и модерации открывают конкретное сообщение и его контекст, включая запись вне первой страницы.
- [ ] #2 Удалённый/скрытый/недоступный комментарий имеет понятное состояние; auth return сохраняет назначение, приватные данные не раскрываются.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Проверить текущие контракты и написать регрессии. 2. Реализовать карточку с сохранением прав и контекста. 3. Целевые unit проверки; общий browser/build выполняет root.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Ответственный: admin_finish. Реализовано локально: очередь администратора (531a2c4), безопасный текстовый diff (fbf2d13), точный comment context (cd01a49), inline account editor и taxonomy search (531a2c4), полный server title search (e66ae39), server audit filters (911c0e5). Svelte check: 0 errors/0 warnings; targeted TS unit/SSR tests pass; Go catalog/community/identity/HTTP packages pass (DB tests без TEST_DATABASE_URL пропускаются). Финальная DB/browser проверка, независимый review и наблюдаемый production SHA выполняются root; выпуск пока не подтверждён.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.

21 октября, admin_acceptance_remaining: focused reply browser regression PASS: focus25 не заменяет contiguous cursor10; load more показывает каждый ответ1–30 ровно один раз. Проверка выполнена на текущем локальном Docker web/API в административном наборе15/15; HTTP root/reply authorization и hidden/deleted контракт сохранены. Нового пробела реализации не найдено; приватная production приёмка отдельно.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-10-20 16:31

---

Независимый review: fdbf9bb отделяет root.reply_cursor (последний из первых 10 последовательных ответов) от дополнительно загруженного focus reply. Merge обычной страницы удаляет дубли и сортирует ответы. Unit регрессия RED: старый cursor=25 вместо10; GREEN:4 assertions. DB регрессия30 ответов focus25 + load more подготовлена; browser регрессия в604cf77 ожидает общего прогона root.

---
<!-- COMMENTS:END -->
