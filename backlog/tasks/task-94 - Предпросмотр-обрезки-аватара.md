---
id: TASK-94
title: Предпросмотр обрезки аватара
status: Review
assignee:
  - '@reader-agent'
created_date: '2026-10-20 15:13'
updated_date: '2026-10-20 16:52'
labels:
  - ux-top30-2026-10-20
  - web
dependencies: []
references:
  - apps/web/src/lib/features/identity/presentation/AvatarEditor.svelte
priority: medium
type: enhancement
ordinal: 44026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: согласованный топ-30 UX/UI от 20 октября 2026, пункт №26. Владелец: «добавь таски и начни исполнять».

Сейчас загрузка автоматически обрезает изображение по центру; результат нужно понимать до отправки.

Границы: сохранить палитру, локали, архитектуру, роли, утверждённые фильтры и порядок главной. Эта карточка описывает отдельный результат; соседние возможности не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 До загрузки видна фактическая квадратная и круглая область будущего аватара с возможностью отмены.
- [x] #2 Подтверждение использует существующие лимиты и безопасную серверную обработку, не отправляет файл при отмене.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Сверить контракт и добавить регрессию. 2. Узкое изменение с account scope и локалями. 3. Unit/UI/browser проверки и SHA; выпуск отдельно.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
4b8499b: квадратный и круглый center-cover preview до upload, отмена освобождает objectURL без отправки, account/owner смена сбрасывает pending. Существующая обработка и лимиты сохранены. Browser cancel + confirmation/avatar-switch regressions authored; runtime pending root. Локально текущие svelte-check 0 errors/warnings, frontend boundaries pass, UI45 pass; выпуска не подтверждают. Следующий шаг: root serial browser/runtime, independent review и authorised release acceptance.

Independent review P2 исправлен в 891f3c9: JPEG EXIF orientation 1–8 применяется до center crop, включая зеркальные варианты; сохранены лимиты bytes/dimensions/pixels и stripping метаданных. RED: orientations2–8 failing в обоих endian; GREEN: 16 EXIF/corner-color cases плюс malformed/truncated metadata. go test ./internal/identity/... ./internal/transport/http/... -count=1 pass; fuzz111918 executions/2s pass. Codec fixtures генерируются orientedJPEG в processor_test.go (400x200 four-color JPEG + EXIF APP1). Docker/browser не запускались; последующая runtime/release acceptance у root.

bce8ee8 закрывает PNG eXIf preview mismatch: bounded container scanner+CRC и общий TIFF orientation до crop. RED на PNG orientation6 colors → GREEN; malformed length/CRC/truncation покрыты. WebP эмпирически Chrome сохраняет исходные pixels400x200 даже с EXIF6 (raw TIFF и Exif-prefix): сервер сохраняет то же поведение, тест на own86byte lossless quadrants.webp. Без dead WebP parser. go test avatarimage+HTTP -count=1 pass; fuzz104784execs/2s pass. Browser dataURL proof без аккаунтов/сети: tmp/task72/check-webp-exif*.mjs, PNG proof root1200x800. Docker/runtime rebuild не выполнялся агентом.

Локальная общая приёмка: notifications/offline/navigation8/8, avatar/library/metrics4/4, reader-improvements3/3 PASS; global Escape исправлен83af48a, EXIF JPEG891f3c9 и PNG bce8ee8 согласованы с фактическим Chrome. Полный Go/PostgreSQL PASS, identity/HTTP послеPNG PASS;146module+46UI,Svelte0/0,architecture/contrast/preview PASS. Повторное independent review без замечаний; physicalAndroid не проводился. Выпуск ещё не подтверждён.

Итоговая локальная приёмка пакета bd9654c579a6bd1cd3465c58f74f2fdeec7ee24a: 75 уникальных browser-сценариев PASS (35 файлов; два старых селектора обновлены и перепроверены, отдельный API pause/unpause recovery PASS),146 module и46 UI PASS; Svelte0 errors/0 warnings, architecture/contrast/ui-preview PASS, Docker web/API build+health PASS. Полный Go/PostgreSQL ./... выполнен последовательно -p1 PASS; identity/HTTP после PNG EXIF также PASS. Независимый review трёх областей и повтор после исправлений: замечаний, препятствующих push, нет. Невыполненная физическая приёмка отдельно TASK97. Выпуск до push не подтверждён; статус Review.
<!-- SECTION:NOTES:END -->
