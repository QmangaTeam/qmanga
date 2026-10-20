---
id: TASK-79
title: Заявка из выбранной команды
status: Review
assignee:
  - '@codex'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 16:52'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/teams/presentation/Studio.svelte
priority: medium
type: enhancement
ordinal: 44011
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №11. Владелец: «добавь таски и начни исполнять».

Ссылка Запросить проект со страницы команды не должна молча выбирать другую первую доступную команду.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Переход из команды сохраняет её как контекст заявки при наличии прав.
- [x] #2 Недоступная или неутверждённая команда объяснена; сервер остаётся авторитетным, нет скрытой подмены другой командой.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Сверить фактические формы/контракт; воспроизвести недостающий сценарий; выполнить ограниченную реализацию, проверить локали/узкие экраны/ошибки и сохранить результат с SHA.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Реализовано 3c9248e: ссылка из команды передаёт team ID, ProjectRequests ищет выбранную команду по всей пагинации; выбирает только approved owner/editor. При недоступности — явное объяснение и пустой выбор без подмены первой командой. Browser подтвердил ID за первой сотней, недоступный team и сохранение текста после ошибки/отмены. form-workflow 3/3 pass, Svelte/architecture pass. Сервер авторитетен; не выпущено.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.
<!-- SECTION:NOTES:END -->
