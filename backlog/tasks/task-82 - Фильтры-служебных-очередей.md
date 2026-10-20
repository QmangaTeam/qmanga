---
id: TASK-82
title: Фильтры служебных очередей
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
  - apps/web/src/lib/features/publishing/presentation/Moderation.svelte
priority: medium
type: enhancement
ordinal: 44014
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №14. Владелец: «добавь таски и начни исполнять».

По мере роста команд модератору нужно сузить очередь по команде и типу работы. Потребуется сверить/расширить API без клиентской фильтрации одной страницы.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Команда и тип обращения фильтруют всю доступную очередь серверно и видны в URL/контексте.
- [x] #2 Фильтрация совместима с cursor, правами, пустыми результатами и retry; не скрывает невидимые ограничения.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Проверить текущие контракты и написать регрессии. 2. Реализовать карточку с сохранением прав и контекста. 3. Целевые unit проверки; общий browser/build выполняет root.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Ответственный: admin_finish/root. 6359b65: server-фильтры полной очереди публикаций (команда + способ подготовки), жалоб (команда + тип объекта), метаданных (команда + данные/оформление); параметры сохраняются в URL. Возврат из detail/preview содержит критерии, до 20 загруженных страниц и anchor объекта, без сохранения приватных записей; решение публикации удаляет только обработанную запись, не сбрасывая очередь. ProjectRequests дорабатывает root. Локально: 9 целевых unit tests pass; 46 UI SSR tests pass; Go catalog/community/publishing(application+postgres)/HTTP packages pass без настроенной БД, DB regressions подготовлены. Общая DB/browser приёмка и выпуск — следующий шаг root; Done не подтверждён.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.

21 октября, admin_acceptance_remaining: сверены server gateway/UI параметры team/kind для публикаций, жалоб, изменений тайтла и project requests. URL сохраняет критерии; queue-context unit PASS, browser report recovery и form-workflow PASS в текущем административном наборе15/15 на Docker. Нового пробела реализации не найдено; полная серверная фильтрация ранее проверена интеграционными регрессиями и CI, не подменяется локальной фильтрацией первой страницы. Приватная production приёмка отдельно.

Final release acceptance 2026-10-21: production and staging observed blue797c810725aa6e880cabd477a8a2a2570e4b49a0; public HTTPS/browser footer confirms exactSHA, health200 with PostgreSQL/S3ok. Verify39480229990 and FastRelease39480229999 SUCCESS;89browsercases43+46 with0failures,158module/46UI, Go/PostgreSQL/codecs and independent review. Temporary synthetic staging fixtures only; personal production accounts/data not used. Existing focused UI/domain/Go/browser evidence in this card verified. TASK86 missing focus behavior fixed ina153498, red→green regression and independent review; actual staging edit→focus→cancel→origin-focus passed. This release contains the reviewed administrative UI package. PhysicalAndroid separatelyTASK97; no production fixtures or owner-login approval required.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Administrative UX criteria verified by focused checks and full CI; reviewed package released to production797c810. TASK86 focus/cancel additionally passed actual staging browser.
<!-- SECTION:FINAL_SUMMARY:END -->
