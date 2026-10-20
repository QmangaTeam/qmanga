---
id: TASK-72
title: Защита остальных длинных служебных форм
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
  - apps/web/src/lib/features/publishing/presentation/ProjectRequests.svelte
priority: high
type: bug
ordinal: 44004
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №4. Владелец: «добавь таски и начни исполнять».

После заполнения команды, заявки или административной формы уход и смена объекта могут обесценить выполненную работу. Не дублирует уже исправленную заявку метаданных TASK-68.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Перечень грязных форм команд, заявок и админки подтверждён текущим кодом; выход/смена объекта требуют явного согласия на потерю несохранённого.
- [x] #2 Отмена сохраняет текст и файлы; успешное сохранение снимает предупреждение; смена аккаунта не переносит черновик другому пользователю.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Сверить фактические формы/контракт; воспроизвести недостающий сценарий; выполнить ограниченную реализацию, проверить локали/узкие экраны/ошибки и сохранить результат с SHA.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Подключена защита несохранённого в Studio команд (профиль/приглашение), ProjectRequests (новая/повторная заявка/причина решения), AdminCatalogue/Taxonomy/AccountManagement, TeamApprovals, Moderation (возврат/доступность), Reports (причины по объектам), TitleChangeDetail (причина, включая refresh после ошибки). Metadata TitleChangeForm сохраняет прежнюю защиту текстов/файлов. Browser form-workflow 3/3 pass: отмена close/navigation/switch сохраняет ввод, админская причина остаётся, независимый review исправил bypass повторного load. Успешные операции снимают соответствующий dirty baseline, root identity key изолирует аккаунты. Реализация 9edb620/3c9248e/f04921a +531a2c4/604cf77. Локальная приёмка; выпуск ещё не подтверждён.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.

Приёмка и выпуск 21.10.2026: фактический production SHA e2c9c8910d6dedbe74eddbafa028b8bb67180340 повторно наблюдён во внешней HTTPS главной; /api/v1/health =200, postgres/s3 ok. Verify Qmanga https://github.com/maqsatto/qmanga/actions/runs/39474894775 SUCCESS на этом SHA: обе acceptance-половины,88 браузерных сценариев без failures, frontend/Go/PostgreSQL и ARM64 images. Актуальные CI-логи сверены с относящимися browser-регрессиями; независимый review и исправления описаны выше. Прежние записи «не запушено/не выпущено» исторические и заменены этим наблюдаемым выпуском. Критерии этой карточки выполнены; физический Android отдельно TASK-97 и здесь не заявляется.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-10-20 16:23

---

Admin contribution: 531a2c4/e66ae39/6359b65 добавляют UnsavedChanges в AdminTaxonomy, AdminCatalogue и AccountManagement; смена выбранного объекта требует согласия, cancel не меняет drafts, сохранение сбрасывает baseline; user prop change очищает account/catalog drafts. Финальная browser проверка у root.

---

created: 2026-10-20 16:31

---

Независимый review:604cf77 создание проекта обновляет только detail.projects и не вызывает choose(selected), сохраняя input, aliases, authors, categoryIds и dirty baseline. Browser регрессия draft→createProject→preserve добавлена в admin.test.mjs, общий запуск выполняет root.

---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Реализация и регрессии проверены, выпущены в production e2c9c89; фактический SHA/health и полный CI подтверждены. Физическая Android-приёмка отслеживается отдельно TASK-97.
<!-- SECTION:FINAL_SUMMARY:END -->
