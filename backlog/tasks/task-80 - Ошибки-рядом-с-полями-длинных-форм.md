---
id: TASK-80
title: Ошибки рядом с полями длинных форм
status: Done
assignee:
  - '@codex'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 19:01'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/catalog/presentation/TitleChangeForm.svelte
priority: medium
type: enhancement
ordinal: 44012
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №12. Владелец: «добавь таски и начни исполнять».

Общая ошибка в начале длинной формы не всегда помогает понять конкретное исправление.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Подтверждённые ошибки ввода связаны с полями, а сводка переводит к ним фокус.
- [x] #2 Сетевые и permission ошибки отделены от невалидного ввода; черновик и файлы сохраняются, исправление убирает устаревшую ошибку.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Сверить фактические формы/контракт; воспроизвести недостающий сценарий; выполнить ограниченную реализацию, проверить локали/узкие экраны/ошибки и сохранить результат с SHA.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Реализовано 9edb620/3c9248e/f04921a/531a2c4: native constraint errors локализованы и связаны aria-invalid/describedby с полями, summary ведёт к полю/раскрывает details. Ошибка вне label сохраняет точное доступное имя поля. Сетевые/permission ошибки остаются сообщением формы без ложной привязки к полю. Input удаляет исправленную ошибку, ошибка POST сохраняет draft. Browser form-workflow проверил invalid/summary/focus/recovery и сохранение note/name при500; unit3pass, общие46UIpass. Локально, выпуск не подтверждён.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.

Приёмка и выпуск 21.10.2026: фактический production SHA e2c9c8910d6dedbe74eddbafa028b8bb67180340 повторно наблюдён во внешней HTTPS главной; /api/v1/health =200, postgres/s3 ok. Verify Qmanga https://github.com/maqsatto/qmanga/actions/runs/39474894775 SUCCESS на этом SHA: обе acceptance-половины,88 браузерных сценариев без failures, frontend/Go/PostgreSQL и ARM64 images. Актуальные CI-логи сверены с относящимися browser-регрессиями; независимый review и исправления описаны выше. Прежние записи «не запушено/не выпущено» исторические и заменены этим наблюдаемым выпуском. Критерии этой карточки выполнены; физический Android отдельно TASK-97 и здесь не заявляется.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-10-20 16:23

---

Admin contribution: fieldValidation action подключён ко всем формам AdminTaxonomy/AdminCatalogue и inline AccountManagement edit; native validity отображается рядом с полем через shared implementation root. Общий Svelte check и 46 SSR tests pass; browser приёмка root.

---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Реализация и регрессии проверены, выпущены в production e2c9c89; фактический SHA/health и полный CI подтверждены. Физическая Android-приёмка отслеживается отдельно TASK-97.
<!-- SECTION:FINAL_SUMMARY:END -->
