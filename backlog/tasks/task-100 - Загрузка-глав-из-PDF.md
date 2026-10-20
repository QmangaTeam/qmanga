---
id: TASK-100
title: Загрузка глав из PDF
status: Review
assignee:
  - '@codex-pdf'
created_date: '2026-10-20 16:49'
updated_date: '2026-10-20 17:27'
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
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Реализовано локальное bounded PDF→PNG преобразование перед существующим preview/upload. Unit и Chromium codec прошли; Review до общей приёмки и выпуска. Следующий шаг: общий build/browser пакет, независимая итоговая проверка и наблюдаемый production SHA; hard memory cap браузером не обеспечивается.
<!-- SECTION:FINAL_SUMMARY:END -->
