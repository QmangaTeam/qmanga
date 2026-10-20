---
id: TASK-70
title: Компактная оболочка кабинета переводчика
status: Review
assignee:
  - '@codex'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 18:04'
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

CI hydration regression: reproduce delayed application JS and lost sidebar keyboard activation; disable disclosure until onMount and wait enabled in browser keyboard test; verify actual built app and independent review.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Локально: отдельный StudioHeader с аккаунтом/возвратом, пять направлений в native dialog; public chrome остаётся на публичных маршрутах. Svelte 0 errors/0 warnings, architecture pass, UI 41/41, Docker web build/health pass, browser studio-priority/workspace/notification-isolation 6/6 (kk/ru, темы, 320/360/1280, keyboard/Escape/focus, переход в каталог, account isolation). Визуально проверены synthetic 360/1280. Не выпущено, независимый review по завершении пакета.

Результат: 5dd3ace; итоговый проверенный пакет 05725000545f23f1ad1d9eaa063cd514aa9ca2b9. Независимый review: новых замечаний нет. Дополнительно browser notification-isolation pass; итоговые browser publishing/workspace 11/11 и Svelte/architecture/UI/contrast/ui-preview/Docker web pass. Review до выпуска. Следующий шаг очереди TASK-72.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.

CI regression 39468803806: deterministic delayed /_app JS reproduced enabled SSR disclosure swallowing Enter (navigation stayed closed after hydration; second Enter opened it). WorkspaceNavigation now disables native sidebar button until onMount attaches handlers; keyboard test waits enabled, new held-JS regression verifies SSR disabled then hydrated Enter opens navigation. Root rebuilt actual local Docker bundle: workspace browser 3/3 PASS (tmp/task99/navigation-browser-green.log), UI46/46 PASS, Svelte0 errors/0 warnings; frontend boundaries PASS. Independent CBZ agent review: no blockers. Review remains pending observed release of this SHA.

Remote release readiness 2026-10-20: tested product SHA842a23f2c9bd55da1ad9447def70f428bda0b3b1. Verify Qmanga https://github.com/maqsatto/qmanga/actions/runs/39470308246 SUCCESS: both acceptance shards,88 browser scenarios (42+46),0 failures, full Go/PostgreSQL/codecs, frontend and all3 ARM64 images. Fast Release https://github.com/maqsatto/qmanga/actions/runs/39470308114 SUCCESS; refs/deploy-ready confirmed842a23f. Prior39468803806 failure was a real pre-hydration sidebar lost-Enter race, reproduced and fixed with native disabled-until-mounted control plus cold-JS regression; independent review passed, local workspace3/3/UI46/Svelte0/0 and repeated full CI green. Live Shikimori/AniList DTO→mapper→merge→metadataError/validPublication also validated (no titles created/changed). Public production still reports36b3a5f455eb37830408af44c2e2f00b48eabd51, health200; strict SSH with existing key and trusted known_hosts still times out before authentication. Rollout/production acceptance remains unconfirmed; keep Review. Next step: restore authorized SSH network access and inspect existing poller journal/state, preserve locks/backup/restore gates and release the ready immutable SHA through the single standard path.
<!-- SECTION:NOTES:END -->
