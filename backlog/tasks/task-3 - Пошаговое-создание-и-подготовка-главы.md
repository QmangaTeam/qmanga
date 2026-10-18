---
id: TASK-3
title: Пошаговое создание и подготовка главы
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 22:31'
labels:
  - stage-1
dependencies: []
references:
  - UPDATE.md
ordinal: 3
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-1. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Разбить создание главы на понятные действия: выбор существующей или новой главы; после открытия — загрузка страниц; затем права/источник и проверка. Показывать только нужное на данном шаге, прогресс и следующий шаг. Убрать дублирующиеся инструкции, сохранить все существующие поля и проверки.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Проверить текущий редактор и зафиксировать браузерную регрессию на одновременный показ всех разделов. 2. Разделить выбор главы и три раздела черновика с понятными переходами, сохранив ввод, проверки и защиту несохранённых изменений. 3. Проверить ошибки сохранения, возврат между шагами и 320/360/desktop в обеих локалях. 4. Выполнить сборку и регрессии, записать ограничения и оставить Review до выпуска.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
До изменения браузерная регрессия подтверждённо упала: поля источника показывались сразу на шаге файлов. Реализованы отдельные видимые разделы, сохранение ввода при переходах, следующий шаг после сохранения сведений, сворачивание выбора главы после открытия и прогресс 1–3. Svelte diagnostics: 0 errors / 0 warnings. Полная сборка и браузерные проверки ещё выполняются.

Локальный результат 19 октября 2026: видны только поля выбранного шага; выбор главы сворачивается после успешного открытия; есть прогресс и переходы назад/вперёд; сохранение и переход к проверке выполняются только после успешного ответа. Переходы сохраняют введённые сведения, обработку/порядок и существующие проверки. Фокус переводится на заголовок шага; скролл больше не переключает шаги, поскольку разделы теперь показываются по одному (это заменяет прежний scroll-spy из TASK-2). Проверено: финальная Docker-сборка, 90 frontend + 26 SSR тестов, app/preview diagnostics 0 errors / 0 warnings, architecture/contrast. Четыре браузерные регрессии прошли: отдельная обложка, отмена/удаление черновика без изменения публикации, загрузка→сведения→проверка с HTTP 500 и повтором, клавиатура/kk/ru/light/dark/320/360/1280. Новый тест сначала падал на исходном показе сведений. Логи tmp/studio-steps-final-build.log и tmp/steps-final-browser.log; скриншоты tmp/studio-steps-screenshots/. Использованы только синтетические изображения и одноразовые локальные аккаунты. Backend не менялся; полный Go и полный набор browser/*.test.mjs в этой сессии не запускались. Следующий шаг: CI и проверка на тестовом домене, затем выпуск; до этого статус Review. Push и remote deployment не выполнялись.

Последующая общая интеграция: на финальном окружении со шрифтом Onest и фиксом поиска весь browser suite прошёл дважды по 21 проверке; проверки пошагового редактора включены. Функциональный код TASK-3 остаётся 211a90b; приёмка remote CI/домена по-прежнему не заявляется.

Owner-requested review found reproducible unsaved-data gap: newchapter label/volume/name/order were absent from dirty tracking, so AllProjects navigation discarded input without confirmation. Fix and actual browser regression in progress; review cannot be considered clean until rechecked.

Owner review found unsaved new-chapter label/volume/name/order were omitted from leave protection. Fixed baf8025; browser regression confirmed failed-save preservation, cancel-leave preservation, explicit confirmation and successful-save reset. New and existing editor browser scenarios passed on rebuilt web; remote release remains pending.

Ревью владельца завершено для23 карточек на базе3302515: проверены реализация, права доступа, границы API/хранилищ и интеграция. Итоговые локальные проверки:106 Go-тестов/подтестов,114 прикладных и29 SSR-тестов,40/40 браузерных сценариев;0 ошибок и0 пропусков, Svelte0 ошибок/предупреждений. Логи tmp/review-complete-go.jsonl, tmp/review-complete-web-build.log, tmp/review-final-browser-acceptance.log. Точные auth/publishing/reading CI smoke пройдены на настоящих PostgreSQL/MinIO/Mailpit. Это локальная приёмка; Review сохранён до удалённого CI, staging и наблюдаемого выпуска того же SHA. Для TASK17 дополнительно нужны утверждение правовых текстов владельцем и действующий ящик поддержки; для TASK26 остаётся предрелизная проверка consent/contact.
<!-- SECTION:NOTES:END -->
