---
id: TASK-87
title: Поиск тайтла при административном выборе
status: Done
assignee:
  - '@admin-agent'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 19:16'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/catalog/presentation/AdminCatalogue.svelte
priority: medium
type: enhancement
ordinal: 44019
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №19. Владелец: «добавь таски и начни исполнять».

Длинный загруженный select тайтлов затрудняет выбор нужного произведения; публичный каталог и библиотечный поиск не меняются.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Администратор находит нужный доступный тайтл за пределами первой страницы по понятному поиску.
- [x] #2 Выбор не сохраняет данные сам по себе; dirty guard, состояние загрузки и существующие права сохранены.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Проверить текущие контракты и написать регрессии. 2. Реализовать карточку с сохранением прав и контекста. 3. Целевые unit проверки; общий browser/build выполняет root.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Ответственный: admin_finish. Реализовано локально: очередь администратора (531a2c4), безопасный текстовый diff (fbf2d13), точный comment context (cd01a49), inline account editor и taxonomy search (531a2c4), полный server title search (e66ae39), server audit filters (911c0e5). Svelte check: 0 errors/0 warnings; targeted TS unit/SSR tests pass; Go catalog/community/identity/HTTP packages pass (DB tests без TEST_DATABASE_URL пропускаются). Финальная DB/browser проверка, независимый review и наблюдаемый production SHA выполняются root; выпуск пока не подтверждён.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.

21 октября, admin_acceptance_remaining: AdminCatalogue применяет server query через gateway.titles, pagination использует appliedTitleQuery; каталоговый DB regression SearchAdminTitles ищет Extra105 за первой сотней. Текущие browser проверки подтверждают выбор без сохранения, rejected write сохраняет draft, создание проекта не сбрасывает metadata/categories; административный набор15/15 PASS. Нового пробела реализации не найдено; приватная production приёмка отдельно.

Final release acceptance 2026-10-21: production and staging observed blue797c810725aa6e880cabd477a8a2a2570e4b49a0; public HTTPS/browser footer confirms exactSHA, health200 with PostgreSQL/S3ok. Verify39480229990 and FastRelease39480229999 SUCCESS;89browsercases43+46 with0failures,158module/46UI, Go/PostgreSQL/codecs and independent review. Temporary synthetic staging fixtures only; personal production accounts/data not used. Existing focused UI/domain/Go/browser evidence in this card verified. TASK86 missing focus behavior fixed ina153498, red→green regression and independent review; actual staging edit→focus→cancel→origin-focus passed. This release contains the reviewed administrative UI package. PhysicalAndroid separatelyTASK97; no production fixtures or owner-login approval required.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Administrative UX criteria verified by focused checks and full CI; reviewed package released to production797c810. TASK86 focus/cancel additionally passed actual staging browser.
<!-- SECTION:FINAL_SUMMARY:END -->
