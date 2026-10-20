---
id: TASK-100
title: Загрузка глав из PDF
status: Review
assignee:
  - '@codex-pdf'
created_date: '2026-10-20 16:49'
updated_date: '2026-10-20 18:04'
labels:
  - web
  - publishing
  - imports
dependencies: []
priority: high
type: feature
ordinal: 45002
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Реализуется локальное преобразование PDF в упорядоченные PNG перед подтверждением выбора и обычной загрузкой изображений. Исходный PDF остаётся в браузере; внешний JavaScript/ссылки/вложения документа не выполняются. Browser API ограничивает вход/страницы/пиксели/выход и время, но не предоставляет жёсткий OS-level memory cap.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Выбор PDF показывает фактическое число и порядок страниц; безопасная растеризация с ограниченными разрешением, временем, памятью и размером сохраняет читаемость текста/изображений.
- [ ] #2 Шифрованный, повреждённый или превышающий лимиты PDF даёт понятную причину без потери открытого черновика; внешние ссылки/вложения/действия PDF не исполняются.
- [ ] #3 Страницы совместимы с обычной обработкой/предпросмотром/повтором и лимитами главы, прогресс отличает загрузку от растеризации и готовности; проверены cancel/partial failure/retry/права и mobile.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Bounded browser PDF.js adapter and self-hosted worker, no document upload endpoint. 2. Return ordered PNG files into existing selection/upload pipeline, reject encrypted/corrupt/oversized input and support cancel. 3. Unit and real-codec browser checks; independent review and production acceptance handled by root.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Локально реализован PDF.js 6.4.299 decoder с отдельным self-hosted Worker, PNG в порядке документа, 50MiB/100 страниц/1600px/7500px/120M суммарных пикселей/250MiB output/60s; cancel завершает worker и render, очищает canvas. Шифрованные/повреждённые документы отклоняются; исходный PDF не отправляется. Локальные fonts/cmaps/wasm и их лицензии поставляет Vite plugin, QuickJS/fallback JS исключены. Проверено: 3 unit (15 asserts), реальный Chromium codec (2 цветные страницы и порядок, generated Standard R2 password PDF, corrupt, pre-abort, cancel после 1 страницы, forced timeout, retry, локальные assets/license/no external requests); frontend boundaries pass. CBZ agent independently reviewed adapter/assets; shared UI account-binding review finding исправлен. Ограничения: браузер не даёт hard OS memory cap; mobile physical device, полный shared UI/release и фактический production SHA ожидаются от root.

Commit реализации 049c929. Root проверил общий upload UI на actual Docker app: 12 сочетаний kk/ru × dark/light × 320/360/1280 прошли, screenshot tmp/task99/import-kk-dark-320.png просмотрен. Это viewport emulation, не физическое Android-устройство. Ожидается финальный rebuild с assets plugin и внешняя проверка выпуска.

Root integrated production-build verification: added PDF UI→preview(noPUT)→confirm→two ordered authorized API uploads regression64b9915, cleanupd3cab1e, resource assertions0941728. Actual Docker build serves LICENSE, fallbackfont, CMap and OpenJPEGWASM HTTP200; PDF production bundle test1/1 PASS. Standalone real codec test1/1 PASS repeated by root. Root12 responsive locale/theme combinations PASS, screenshots inspected; full Go/PostgreSQL and157module/46UI/Svelte0/0/architecture/contrast/preview pass. Independent PDF-agent review of root regression no blocking findings. Production acceptance pending SSH TCP22 timeout; no release claim.

Cold decoder chunk cancellation regression: delayed pdfjs-dist.js route proved old code constructed 1 worker after cancellation; added post-import abort guard, new Chromium test proves 0 workers. Browser codec suite now 2/2 PASS, previous error/cancel/encryption/order cases retained.

Final integration 2026-10-20 at833f606: Docker API/web/gateway rebuilt and healthy. 158 frontend module +46 UI PASS; Svelte0errors/0warnings; architecture, both-theme contrast and preview checks PASS; Backlog doctor clean. Full Go/PostgreSQL -p1 ./... passed, latest catalog/source/HTTP/bootstrap real-DB checks repeated after endpoint/pacing changes PASS. Latest CBZ/PDF browser7/7 (including immediate cold decoder cancellation); metadata browser5/5. Root additionally verified12 responsive kk/ru x themes x320/360/1280 combinations for import forms and44px targets, and actual local API→live Shikimori.io/AniList search/detail/reprocessed WebPcover. Independent CBZ/PDF/metadata and root docs/test reviews have no remaining blocking findings. All12 new browser scenarios enter existing CI glob. Product tasks remain Review until full remote Verify and observed release of the same SHA. Production access blocker remains TCP22 timeout before authentication; no lock/hold/firewall/deployment bypass used.

Remote release readiness 2026-10-20: tested product SHA842a23f2c9bd55da1ad9447def70f428bda0b3b1. Verify Qmanga https://github.com/maqsatto/qmanga/actions/runs/39470308246 SUCCESS: both acceptance shards,88 browser scenarios (42+46),0 failures, full Go/PostgreSQL/codecs, frontend and all3 ARM64 images. Fast Release https://github.com/maqsatto/qmanga/actions/runs/39470308114 SUCCESS; refs/deploy-ready confirmed842a23f. Prior39468803806 failure was a real pre-hydration sidebar lost-Enter race, reproduced and fixed with native disabled-until-mounted control plus cold-JS regression; independent review passed, local workspace3/3/UI46/Svelte0/0 and repeated full CI green. Live Shikimori/AniList DTO→mapper→merge→metadataError/validPublication also validated (no titles created/changed). Public production still reports36b3a5f455eb37830408af44c2e2f00b48eabd51, health200; strict SSH with existing key and trusted known_hosts still times out before authentication. Rollout/production acceptance remains unconfirmed; keep Review. Next step: restore authorized SSH network access and inspect existing poller journal/state, preserve locks/backup/restore gates and release the ready immutable SHA through the single standard path.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Реализовано локальное bounded PDF→PNG преобразование перед существующим preview/upload. Unit и Chromium codec прошли; Review до общей приёмки и выпуска. Следующий шаг: общий build/browser пакет, независимая итоговая проверка и наблюдаемый production SHA; hard memory cap браузером не обеспечивается.
<!-- SECTION:FINAL_SUMMARY:END -->
