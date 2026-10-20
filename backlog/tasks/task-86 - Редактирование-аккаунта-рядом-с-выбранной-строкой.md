---
id: TASK-86
title: Редактирование аккаунта рядом с выбранной строкой
status: Review
assignee:
  - '@admin_acceptance_remaining'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 19:02'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/identity/presentation/AccountManagement.svelte
priority: medium
type: enhancement
ordinal: 44018
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №18. Владелец: «добавь таски и начни исполнять».

Форма аккаунта сейчас расположена над списком, из-за чего выбор строки далеко внизу теряет визуальный контекст.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 После выбора видны имя/роль целевого аккаунта, редактируемая форма получает фокус в понятном месте.
- [ ] #2 Отмена возвращает к строке; подтверждение, причины, запрет саморедактирования и серверные права сохранены.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Проверить текущие контракты и написать регрессии. 2. Реализовать карточку с сохранением прав и контекста. 3. Целевые unit проверки; общий browser/build выполняет root.

21 октября: воспроизвести потерю клавиатурного фокуса при inline edit/cancel; переводить фокус в первый доступный контрол формы и возвращать на кнопку исходной строки; browser регрессия и Svelte/UI проверки.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Ответственный: admin_finish. Реализовано локально: очередь администратора (531a2c4), безопасный текстовый diff (fbf2d13), точный comment context (cd01a49), inline account editor и taxonomy search (531a2c4), полный server title search (e66ae39), server audit filters (911c0e5). Svelte check: 0 errors/0 warnings; targeted TS unit/SSR tests pass; Go catalog/community/identity/HTTP packages pass (DB tests без TEST_DATABASE_URL пропускаются). Финальная DB/browser проверка, независимый review и наблюдаемый production SHA выполняются root; выпуск пока не подтверждён.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.

21 октября, admin_acceptance_remaining: повторная приёмка выявила невыполненный фокус в AC1/AC2. Browser regression RED: после Enter фокус оставался на «Изменить». Исправлено: после Svelte tick фокус переходит в первый доступный контрол inline формы; отмена возвращает его на кнопку исходной строки. GREEN: account-search 2/2, общий административный browser набор 15/15; 158 module,46 UI PASS; Svelte0 errors/0 warnings; architecture PASS; Docker web build/health PASS. Новая правка ожидает независимого review, push и наблюдаемого production SHA; Done не подтверждён.
<!-- SECTION:NOTES:END -->
