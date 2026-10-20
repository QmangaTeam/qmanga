---
id: TASK-98
title: Повтор каталога без сброса условий
status: Review
assignee:
  - '@reader-agent'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 17:18'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/catalog/presentation/Catalogue.svelte
priority: low
type: bug
ordinal: 44030
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №30. Владелец: «добавь таски и начни исполнять».

В ветке unavailable ссылка Повторить ведёт на /catalog и теряет query/filter/sort/cursor. Утверждённое устройство фильтров не меняется.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Retry после ошибки повторяет исходную выдачу с q, фильтрами, сортировкой и cursor.
- [x] #2 Browser-регрессия воспроизводит сбой/восстановление; пустой результат, история и локаль остаются корректными.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Сверить контракт и добавить регрессию. 2. Узкое изменение с account scope и локалями. 3. Unit/UI/browser проверки и SHA; выпуск отдельно.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
18e11b4: SSR load сохраняет точный pathname+search для retry включая cursor. SSR regression pass. Browser сбой/восстановление на актуальном runtime pending root; не выдаётся за пройденный. Локально текущие svelte-check 0 errors/warnings, frontend boundaries pass, UI45 pass; выпуска не подтверждают. Следующий шаг: root serial browser/runtime, independent review и authorised release acceptance.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.

CI6e2fa97 выявил зависимость теста от названия baseline-тайтла. Тест теперь создаёт два своих listed manga с search_text, получает действительный cursor и удаляет оба в finally. Изолированный API pause→unpause→retry21s local PASS. Продукт не менялся; independent review без замечаний. SHA тестовых исправлений dd66757cb766a5f5809e018b85314174f0af2730. Требуется повторный Verify Qmanga; статус Review, выпуск не подтверждён.

Final remote verification 2026-10-20: SHA 4145910eb7bc649aaca2a2e890a44fbfa5f75575, Verify Qmanga https://github.com/maqsatto/qmanga/actions/runs/39463685921 SUCCESS (both acceptance/browser shards, Go/PostgreSQL/codecs and frontend checks); Fast Release https://github.com/maqsatto/qmanga/actions/runs/39463685942 SUCCESS. Earlier CI follow-up failures are resolved. External production still reports 36b3a5f455eb37830408af44c2e2f00b48eabd51; deployment of this package is not confirmed, status remains Review. Physical Android TASK-97 remains not_run; new imports TASK-99–103 are queued only.
<!-- SECTION:NOTES:END -->
