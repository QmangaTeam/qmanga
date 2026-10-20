---
id: TASK-81
title: Рабочие очереди в обзоре администратора
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
  - apps/web/src/lib/features/metrics/presentation/Metrics.svelte
priority: medium
type: enhancement
ordinal: 44013
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №13. Владелец: «добавь таски и начни исполнять».

Обзор сейчас показывает метрики; администратору также нужны ожидающие решения действия, не новая трактовка статистики TASK-36.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Из обзора видны доступные ожидающие решения очереди и переходы к ним.
- [ ] #2 Счётчики, если добавлены, подтверждены авторизованным read model; ошибка не подменяется нулём, метрики сохраняют прежний смысл.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Проверить текущие контракты и написать регрессии. 2. Реализовать карточку с сохранением прав и контекста. 3. Целевые unit проверки; общий browser/build выполняет root.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Ответственный: admin_finish. Реализовано локально: очередь администратора (531a2c4), безопасный текстовый diff (fbf2d13), точный comment context (cd01a49), inline account editor и taxonomy search (531a2c4), полный server title search (e66ae39), server audit filters (911c0e5). Svelte check: 0 errors/0 warnings; targeted TS unit/SSR tests pass; Go catalog/community/identity/HTTP packages pass (DB tests без TEST_DATABASE_URL пропускаются). Финальная DB/browser проверка, независимый review и наблюдаемый production SHA выполняются root; выпуск пока не подтверждён.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.

21 октября, admin_acceptance_remaining: повторно подтверждён AdminQueues:5 штатных переходов к решениям без выдуманных счётчиков; UI regression admin-queues PASS в текущем наборе46/46. Общие158 module PASS. Дополнительной реализации по карточке не требуется; приватный production walkthrough этим локальным SSR-тестом не подтверждён.
<!-- SECTION:NOTES:END -->
