---
id: TASK-86
title: Редактирование аккаунта рядом с выбранной строкой
status: Review
assignee:
  - '@admin-agent'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 16:22'
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
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Ответственный: admin_finish. Реализовано локально: очередь администратора (531a2c4), безопасный текстовый diff (fbf2d13), точный comment context (cd01a49), inline account editor и taxonomy search (531a2c4), полный server title search (e66ae39), server audit filters (911c0e5). Svelte check: 0 errors/0 warnings; targeted TS unit/SSR tests pass; Go catalog/community/identity/HTTP packages pass (DB tests без TEST_DATABASE_URL пропускаются). Финальная DB/browser проверка, независимый review и наблюдаемый production SHA выполняются root; выпуск пока не подтверждён.
<!-- SECTION:NOTES:END -->
