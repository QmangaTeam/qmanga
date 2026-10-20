---
id: TASK-95
title: Редактирование текущего глобального поиска
status: Done
assignee:
  - '@reader-agent'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 19:01'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/shared/presentation/ui/AppHeader.svelte
priority: low
type: bug
ordinal: 44027
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №27. Владелец: «добавь таски и начни исполнять».

При прямом открытии каталога с q глобальное поле не восстанавливает запрос, хотя активный фильтр его показывает.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Открытие глобального поиска подставляет актуальный q из URL каталога.
- [x] #2 Изменение/очистка/повторное открытие и Escape работают предсказуемо без сброса локали и без устаревшего запроса.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Сверить контракт и добавить регрессию. 2. Узкое изменение с account scope и локалями. 3. Unit/UI/browser проверки и SHA; выпуск отдельно.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
4b8499b: каждое открытие global search берёт актуальный q из каталога URL; Escape discards unsubmitted values on next open. Browser regression authored, runtime pending root. Локально текущие svelte-check 0 errors/warnings, frontend boundaries pass, UI45 pass; выпуска не подтверждают. Следующий шаг: root serial browser/runtime, independent review и authorised release acceptance.

83af48a исправляет найденный browser-регрессией настоящий дефект: native input type=search перехватывал первый Escape, очищая ввод и оставляя модальный поиск открытым. Явный Escape handler закрывает dialog с preventDefault. RED browser reproduced; svelte-check 0 errors/warnings. GREEN browser требует обновлённого image, root выполняет rebuild+rerun.

Локальная общая приёмка: notifications/offline/navigation8/8, avatar/library/metrics4/4, reader-improvements3/3 PASS; global Escape исправлен83af48a, EXIF JPEG891f3c9 и PNG bce8ee8 согласованы с фактическим Chrome. Полный Go/PostgreSQL PASS, identity/HTTP послеPNG PASS;146module+46UI,Svelte0/0,architecture/contrast/preview PASS. Повторное independent review без замечаний; physicalAndroid не проводился. Выпуск ещё не подтверждён.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.

Приёмка и выпуск 21.10.2026: фактический production SHA e2c9c8910d6dedbe74eddbafa028b8bb67180340 повторно наблюдён во внешней HTTPS главной; /api/v1/health =200, postgres/s3 ok. Verify Qmanga https://github.com/maqsatto/qmanga/actions/runs/39474894775 SUCCESS на этом SHA: обе acceptance-половины,88 браузерных сценариев без failures, frontend/Go/PostgreSQL и ARM64 images. Актуальные CI-логи сверены с относящимися browser-регрессиями; независимый review и исправления описаны выше. Прежние записи «не запушено/не выпущено» исторические и заменены этим наблюдаемым выпуском. Критерии этой карточки выполнены; физический Android отдельно TASK-97 и здесь не заявляется. Дополнительно root на production реально выполнил: глобальный поиск → Ажал → Enter → catalog?q=Ажал с совпавшим тайтлом; повторное открытие сохранило q и фокус; Escape вернул фокус исходной кнопке и сохранил URL, без записи данных.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Реализация и регрессии проверены, выпущены в production e2c9c89; фактический SHA/health и полный CI подтверждены. Физическая Android-приёмка отслеживается отдельно TASK-97.
<!-- SECTION:FINAL_SUMMARY:END -->
