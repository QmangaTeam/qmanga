---
id: TASK-69
title: Рабочая глава первой на мобильном
status: Review
assignee:
  - '@codex'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 15:19'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/publishing/presentation/PublishingStudio.svelte
priority: high
type: enhancement
ordinal: 44001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №1. Владелец: «добавь таски и начни исполнять».

Переводчик открывает проект ради текущей главы, но служебные блоки и выбор главы занимают первый экран. Развитие TASK-3 и TASK-56, а не повтор пошагового редактора.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 При наличии версии на 320/360px сначала доступны текущая глава, её состояние и рабочий шаг; метаданные тайтла вторичны.
- [x] #2 Выбор или создание другой главы остаётся явным и доступным с клавиатуры; пустое состояние ведёт к созданию главы.
- [x] #3 Сохраняются несохранённый ввод, права ролей и одновременные секции на desktop; проверены kk/ru, обе темы и browser-регрессия.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Browser-регрессия: открытая глава и шаг выше выбора/метаданных на 320/360; пустой проект ведёт к созданию.
2. Переставить DOM PublishingStudio: текущие шаги/версия перед выбором главы, метаданные вторичны; явный переход к выбору с фокусом. Не менять API, draft rules и desktop-секции.
3. Svelte/UI/application checks, Docker web и browser publishing-steps/cover; review и локальный коммит, Review до выпуска.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Локально: текущая версия и шаги стоят первыми в DOM на мобильном, выбор главы ниже с явной кнопкой и переносом фокуса; данные тайтла после работы. Пустой проект открывает создание. Svelte 0 errors/0 warnings, architecture pass, Docker web build/health pass, browser studio-priority + publishing-steps 5/5 pass (320/360/1280, kk/ru, обе темы, keyboard, empty, save failure/dirty guards). Выпуск не подтверждён; карточка остаётся Review. Независимый review будет выполнен по завершении пакета.
<!-- SECTION:NOTES:END -->
