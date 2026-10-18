---
id: TASK-30
title: Изолировать аккаунты браузерных тестов библиотеки и тайтла
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:33'
updated_date: '2026-10-18 22:32'
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

Исправление тестовой изоляции: 4b1b9af. Два полных успешных browser run на одинаковой fixture подтверждают повторяемость; pipeline использует тот же glob.

Ревью владельца завершено для23 карточек на базе3302515: проверены реализация, права доступа, границы API/хранилищ и интеграция. Итоговые локальные проверки:106 Go-тестов/подтестов,114 прикладных и29 SSR-тестов,40/40 браузерных сценариев;0 ошибок и0 пропусков, Svelte0 ошибок/предупреждений. Логи tmp/review-complete-go.jsonl, tmp/review-complete-web-build.log, tmp/review-final-browser-acceptance.log. Точные auth/publishing/reading CI smoke пройдены на настоящих PostgreSQL/MinIO/Mailpit. Это локальная приёмка; Review сохранён до удалённого CI, staging и наблюдаемого выпуска того же SHA. Для TASK17 дополнительно нужны утверждение правовых текстов владельцем и действующий ящик поддержки; для TASK26 остаётся предрелизная проверка consent/contact.

Final current-code browser acceptance40/40 in78.999s; gateway audit32 limited authPOSTs, fixed60second windows28+4, zeroHTTP429 with unchanged production limit30. No extra login during audit. Individual user/team fixtures remain isolated; root8080 owner account untouched by disposable18080 fixtures. Evidence tmp/review-final-auth-audit.json.
<!-- SECTION:NOTES:END -->
