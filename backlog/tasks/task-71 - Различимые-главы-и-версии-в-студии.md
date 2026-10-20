---
id: TASK-71
title: Различимые главы и версии в студии
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
  - apps/web/src/lib/features/publishing/presentation/PublishingStudio.svelte
priority: high
type: enhancement
ordinal: 44003
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №3. Владелец: «добавь таски и начни исполнять».

Плоские кнопки глава:статус затрудняют различение логической главы, опубликованной версии и черновика. Показать фактическую связь без выдуманной нумерации редакций.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Версии сгруппированы по главам и имеют читаемые состояния; открытая версия явно выделена без зависимости только от цвета.
- [x] #2 Переключение открывает выбранную версию без создания черновика и синхронизирует контекст главы; dirty/upload guards сохраняются.
- [x] #3 Список сохраняет все полученные версии, дозагрузку и роль; несколько глав/версий, kk/ru, темы и клавиатура проверены.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Browser-регрессия нескольких глав/версий и дозагрузки, переключения без POST и dirty guard.
2. Группы по chapterId с именем главы, реальными состояниями/количеством страниц и явной отметкой открытой версии. Сохранить все версии и API order.
3. Проверить локали/темы/keyboard и связанные publishing тесты; независимый review пакета, отдельный коммит и Review до выпуска.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Реализовано: группы по chapterId, имя главы при наличии, состояние и текстовая отметка Открыта/Ашық. Количество страниц показано только у открытой версии из полных active.pages: list DTO намеренно содержит pages=[], недостоверный ноль не выводится. GET-переключение синхронизирует chapterId, сохраняет прежнюю версию при 500 и позволяет повтор. Сохранены upload/order/attribution guards, добавлена защита незавершённого ввода новой главы. Дозагрузка и fallback глав вне chapters page сохранены. Browser-regressions включены в CI через tests/browser/*.test.mjs. Итог: studio-priority/publishing-steps/publishing-cover/workspace 11/11 pass, включая 101 synthetic revision, 500/retry, без создающего POST, dirty guard, kk/ru/light/dark/320/360/1280. Svelte 0 errors/0 warnings, architecture, contrast, ui-preview pass; 127 module и 41 UI pass; Docker production web build/health pass. Независимый review diff и повтор после уточнения счётчика: новых замечаний нет; reviewer browser не запускал. Физический Android не проверен (TASK-97). Не выпущено, Review. Следующий шаг очереди TASK-72; для выпуска нужен авторизованный push и наблюдаемый тот же SHA.

SHA результата и локально проверенного пакета: 05725000545f23f1ad1d9eaa063cd514aa9ca2b9. Внешняя проверка 20.10.2026: production health 200, фактический SHA 36b3a5f455eb37830408af44c2e2f00b48eabd51. Пакет TASK-69–71 не запушен и не выпущен. Следующий шаг реализации: TASK-72 (To Do); выпуск требует push и подтверждения production SHA.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.

Приёмка и выпуск 21.10.2026: фактический production SHA e2c9c8910d6dedbe74eddbafa028b8bb67180340 повторно наблюдён во внешней HTTPS главной; /api/v1/health =200, postgres/s3 ok. Verify Qmanga https://github.com/maqsatto/qmanga/actions/runs/39474894775 SUCCESS на этом SHA: обе acceptance-половины,88 браузерных сценариев без failures, frontend/Go/PostgreSQL и ARM64 images. Актуальные CI-логи сверены с относящимися browser-регрессиями; независимый review и исправления описаны выше. Прежние записи «не запушено/не выпущено» исторические и заменены этим наблюдаемым выпуском. Критерии этой карточки выполнены; физический Android отдельно TASK-97 и здесь не заявляется.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Реализация и регрессии проверены, выпущены в production e2c9c89; фактический SHA/health и полный CI подтверждены. Физическая Android-приёмка отслеживается отдельно TASK-97.
<!-- SECTION:FINAL_SUMMARY:END -->
