---
id: TASK-27
title: Исправить подсветку поля в диалоге поиска
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:21'
updated_date: '2026-10-18 21:35'
labels:
  - ui
  - bug
dependencies: []
references:
  - apps/web/src/lib/shared/presentation/ui/AppHeader.svelte
ordinal: 1026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
В диалоге поиска пользователь заметил лишний оранжевый контур вокруг вложенного поля. Источник: замечание владельца от 19 октября 2026 года на /menu, мобильное окно 668×590. Проверить фактический фокус и сочетание стилей контейнера и input.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 В фокусе поле поиска имеет единый аккуратный контур без наложения рамок контейнера и input.
- [x] #2 Клавиатурный фокус остаётся заметным; Escape и возврат фокуса к кнопке поиска работают.
- [x] #3 Проверены казахский и русский, обе темы и ширины 320/360/desktop.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Воспроизвести отдельный контур на input внутри составного поля. 2. Регрессионным тестом проверить единственную рамку фокуса, Escape/возврат фокуса и поиск через реальный маршрут. 3. Применить единый focus-within контур контейнера без потери доступности кнопок. 4. Проверить обе локали/темы/размеры, полный frontend build и доступные интеграционные проверки.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Причина: глобальный :focus-visible рисовал отдельный контур input внутри составного search-field. У input контур убран, focus-within на контейнере показывает единую рамку. Новый тест до исправления падал во всех kk/ru × light/dark сочетаниях. На обновлённом локальном Docker: 90 frontend + 26 SSR тестов, app/preview 0 errors/warnings, architecture/contrast/build прошли; 4 поисковые браузерные проверки прошли, включая 320/360/1280, Tab, Escape, кнопку закрытия и реальный GET /catalog?q= с кодированием строки. Во время первого запуска одновременно с пересозданием web был один timeout; повтор после полной готовности сервисов прошёл. Полная интеграция выполняется отдельно на qmanga-stage4-integration; до её результата статус In Progress.

Итоговый код: d6f9a0b. Полный Go/PostgreSQL/codec прогон прошёл (73 test/subtest, 23 packages, actual skips 0). Полный browser suite на финальном Docker образе со шрифтом Onest прошёл дважды: 21/21, без ошибок и skips; поисковые 4 сочетания включены. Обнаруженная при интеграции изоляция fixture исправлена в TASK-30 (4b1b9af), продуктовая библиотека не менялась. Финальные логи tmp/integration/browser-repeat-{1,2}.log; скриншоты search-*.png в tmp/integration/final-screenshots. Review до remote CI и выпуска.
<!-- SECTION:NOTES:END -->
