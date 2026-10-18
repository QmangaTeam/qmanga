---
id: TASK-30
title: Изолировать аккаунты браузерных тестов библиотеки и тайтла
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:33'
updated_date: '2026-10-18 21:35'
labels:
  - ci
  - bug
dependencies: []
ordinal: 4026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Полный повторный browser suite на отдельном qmanga-stage4-integration выявил загрязнение fixture: title-interactions использует тот же owner, что reading.test, и восстанавливает removed:false даже для удалённой исходной записи. Это меняет ожидаемые 110 записей и создаёт гонки между файлами тестов.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Тесты карточки и массовой библиотеки не изменяют библиотеку одного аккаунта параллельно.
- [x] #2 Cleanup восстанавливает исходный removed и проверяет ответ API; повторный полный browser suite проходит на той же локальной fixture.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Использовать отдельный проверенный аккаунт fixture для rating/library действий карточки. 2. Восстанавливать исходный removed и проверять cleanup API. 3. Восстановить только известную синтетическую source-entry отдельной среды после упавшего прогона и прогнать весь набор дважды без пересоздания данных.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Исправлен title-interactions.test.mjs: действия rating/library используют отдельный verified admin fixture, а массовая библиотека остаётся у owner. Cleanup восстанавливает исходный removed и проверяет HTTP 204 и прочитанный результат вместо проглатывания ошибок. Исходный дефект подтверждён полным повторным прогоном: source-entry owner была removed:false, ожидание 110 записей падало. Восстановлена только эта известная синтетическая запись в отдельной среде qmanga-stage4-integration. Затем весь browser/*.test.mjs выполнен дважды подряд на тех же данных: 21/21 pass, 0 fail, 0 skip, без пересоздания fixture между прогонами. Логи tmp/integration/browser-repeat-1.log и browser-repeat-2.log. Product/backend behavior не менялись. Review до remote CI нового коммита.
<!-- SECTION:NOTES:END -->
