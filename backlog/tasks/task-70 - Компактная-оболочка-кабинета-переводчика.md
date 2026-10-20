---
id: TASK-70
title: Компактная оболочка кабинета переводчика
status: Review
assignee:
  - '@codex'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 15:21'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/routes/studio/+layout.svelte
priority: high
type: enhancement
ordinal: 44002
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №2. Владелец: «добавь таски и начни исполнять».

В студии читательская шапка, нижняя панель и большой футер конкурируют с рабочими разделами. Уже существующее боковое меню TASK-31 сохраняется.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 У студии компактная рабочая оболочка с аккаунтом и явным возвратом на сайт; читательские панели не занимают рабочую область.
- [x] #2 Все пять основных направлений сайта и разделы студии остаются доступны; публичные страницы сохраняют прежнюю навигацию.
- [x] #3 Клавиатура, Escape/фокус, переходы, kk/ru, light/dark и 320/360/desktop проверены; account isolation сохранена.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Browser-регрессия для компактной шапки студии и доступности пяти направлений сайта.
2. Presentation-only StudioHeader с аккаунтом, ссылкой на сайт и native dialog навигации; reader chrome исключить только для /studio.
3. Проверить keyboard/Escape/focus, kk/ru, темы, 320/360/desktop, public chrome и account isolation; локальный коммит и Review до выпуска.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Локально: отдельный StudioHeader с аккаунтом/возвратом, пять направлений в native dialog; public chrome остаётся на публичных маршрутах. Svelte 0 errors/0 warnings, architecture pass, UI 41/41, Docker web build/health pass, browser studio-priority/workspace/notification-isolation 6/6 (kk/ru, темы, 320/360/1280, keyboard/Escape/focus, переход в каталог, account isolation). Визуально проверены synthetic 360/1280. Не выпущено, независимый review по завершении пакета.
<!-- SECTION:NOTES:END -->
