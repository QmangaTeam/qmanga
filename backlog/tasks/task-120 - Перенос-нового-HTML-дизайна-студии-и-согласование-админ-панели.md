---
id: TASK-120
title: Перенос нового HTML-дизайна студии и согласование админ-панели
status: In Progress
assignee:
  - '@codex'
created_date: '2026-10-22 08:22'
updated_date: '2026-10-22 08:45'
labels: []
dependencies: []
priority: high
ordinal: 62005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Владелец предоставил qmanga-studio.html и Figma node 2:2. Перенести доступный дизайн на действующие страницы, сохранив права, загрузку/импорт, версии и модерацию. Figma MCP ограничен Starter quota; HTML содержит только студию, образец админ-панели запрошен отдельно.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Студия использует боковую навигацию, верхнюю строку и плотный список проектов из HTML, сохраняя реальные данные и ссылки.
- [ ] #2 Редактор использует сетку страниц и компактные панели из HTML с рабочими загрузкой, порядком, сведениями, предпросмотром и отправкой.
- [ ] #3 Админ-панель перенесена по доступному согласованному образцу без потери существующих функций.
- [ ] #4 Проверены 320/360px, desktop, обе локали/темы, значимые сценарии и независимый review; результат доступен локально.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Сверить HTML, реальные маршруты и Figma. 2. Перенести оболочку и список проектов. 3. Перенести редактор, сохранив действия и guard. 4. Применить образец админки после его получения. 5. Проверки, независимый review, локальные коммиты; без push/deploy.

Admin continuation: read local exported PDF and SVG (Downloads/QMANGA · Админ-панель · UX → UI.pdf/.fig, 01 · Рабочая очередь.svg); map existing queues/tables/review forms to Figma layouts, preserving backend guards and operations; repeat browser matrix and independent review before local handoff.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-10-22: studio implementation committed a35ed6c (shell/directory), aa3cc39 (canvas/inspector). Fresh check 0 errors/0 warnings; 57 UI tests/824 assertions; 5 browser regressions (preflight, mocked reorder/failed attribution retry, mobile priority/empty project/shell) passed; separate 4-case sidebar matrix and admin all-nine-route 320/360/1280 both locale/theme matrix passed. Build, boundaries and palette contrast passed. Independent review found unnamed compact links; fixed and re-reviewed; final density CSS reviewed without findings. No push/deploy. Owner then enabled Figma account: exported full .fig (506994 bytes), 50-page PDF and queue SVG into Downloads through browser UI. Exact admin reference now available; proceed with admin implementation rather than await export. No credentials saved in repo.
<!-- SECTION:NOTES:END -->
