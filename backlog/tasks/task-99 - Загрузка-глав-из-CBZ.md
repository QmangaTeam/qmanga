---
id: TASK-99
title: Загрузка глав из CBZ
status: Review
assignee:
  - '@codex-cbz'
created_date: '2026-10-20 16:49'
updated_date: '2026-10-20 17:33'
labels:
  - web
  - publishing
  - imports
dependencies: []
priority: high
type: feature
ordinal: 45001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Владелец 20.10.2026 поручил загрузку CBZ. Браузер извлекает и проверяет ограниченный ZIP/CBZ, показывает список и порядок перед подтверждением; страницы идут в существующую модель черновика и обработки, без отдельной публикации.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Выбор CBZ извлекает поддерживаемые изображения из ZIP, показывает порядок и число страниц до принятия в главу; папки/имена сортируются естественно, служебные файлы пропускаются явно.
- [x] #2 Архив проверяется в ограниченных ресурсах: размер/число файлов/распакованный объём, zip-bomb, traversal, ссылки, повреждённые или зашифрованные записи; данные не выходят за временную область и не обходят права проекта.
- [x] #3 Страницы попадают в существующий upload/processing pipeline с понятным прогрессом, отменой, ошибками и повтором без дубликатов; действуют лимиты главы, версии и dirty guards. Browser/HTTP regression и узкий экран подтверждены.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Ограниченный браузерный ZIP/CBZ адаптер с проверкой путей и ресурсов. 2. Общий импорт CBZ/PDF с предпросмотром порядка и подтверждением через существующую очередь. 3. Регрессии повреждений, лимитов и повторов; локальные проверки, Review до выпуска.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Реализованы CBZ адаптер и общий CBZ/PDF предпросмотр с явным принятием, отменой и существующей очередью. Лимиты: 100 MiB архив, 1000 записей, 300 изображений, 25 MiB/изображение, 250 MiB суммарно; проверяются центральные/локальные записи, CRC, шифрование, ссылки, пути, фактический deflate output. CRC вычисляется с уступкой event loop. bun CBZ tests: 6 pass; queue tests: 4 pass; svelte-check 0/0; архитектурные границы pass. Chromium реальный deflate/order/mid-import cancel pass; UI/HTTP browser regression подготовлен, ожидает общего stack build. Production не проверен.

Chromium UI/HTTP regression 2/2 pass на disposable localhost:18080: реальный deflate, естественный порядок, до принятия нет PUT, отмена оставляет ноль страниц; обрыв второго PUT оставляет первую страницу, повтор завершает две страницы без дубликатов; повторный выбор архива не добавляет страницы. 320/360/1280 без горизонтального overflow. Субагент PDF независимо проверил CBZ/UI: исправлена привязка import preview к аккаунту/ревизии и отмена при смене аккаунта. Root проверил 12 сочетаний размеров/языков/тем. Последние изменения guard/help требуют общего rebuild/root acceptance. Осталось: независимый итоговый review пакета, общий CI и наблюдаемый выпуск того же SHA.

Основная реализация SHA02fbab4. Дополнительная совместимость: принимаются стандартные deflate compression-hint flags1/2 без разрешения encryption; реальный Chromium fixture теперь flags2. Unit6pass и isolatedcodecpass. Общий старый bundle закономерно отвергает новый fixture: UI regression повторить после следующего rebuild. Это ограничение относится к последней правке parser, ранее UI/HTTP2pass.

Root integration verification 2026-10-20: rebuilt Docker API/web/gateway healthy; latest CBZ compression flags48e1802 accepted in real browser. cbz-import.test.mjs 2/2 PASS (compressed CRC/order/cancel and real upload preview/partial failure/resume/dedup); PDF shared upload regression also PASS. Root layout 12 kk/ru x dark/light x 320/360/1280 combinations PASS; mobile/desktop screenshots inspected under ignored tmp/task99. Full frontend157 module +46UI, Svelte0/0, architecture/contrast/preview PASS; full Go/PostgreSQL -p1 ./... PASS. Independent PDF-agent review of CBZ/shared UI found account-binding issue, fixed and re-reviewed no remaining blockers. Release remains unconfirmed due TCP22 access timeout; Review.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Ограниченный локальный CBZ импорт с естественным порядком и подтверждаемым списком страниц подключён к существующей очереди обработки. Проверки ZIP, CRC и фактического распакованного объёма защищают ресурсы; отмена и content-based ключи сохраняют безопасный повтор. Локальные unit, architecture, svelte-check и Chromium HTTP regressions пройдены; статус Review до общей приёмки и production.
<!-- SECTION:FINAL_SUMMARY:END -->
