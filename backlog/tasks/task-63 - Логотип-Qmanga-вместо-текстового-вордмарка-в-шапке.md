---
id: TASK-63
title: Логотип Qmanga вместо текстового вордмарка в шапке
status: To Do
assignee:
  - codex
created_date: '2026-10-19 04:58'
updated_date: '2026-10-19 04:58'
labels:
  - owner-feedback-2026-10-19
dependencies: []
ordinal: 38026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Источник: поручение владельца 19 октября 2026 года (логотип приложен к сообщению, исходники в Downloads: «Pristine qmanga logo on transparent background-2.png»).

## Формулировка владельца — дословно

вот лого его вставь в проект вместо текста qmanga

## Дополнительное описание (пояснение агента)

Заменить текстовый вордмарк «qmanga.» в десктопной шапке (AppHeader.svelte) изображением логотипа: прозрачный PNG 288×96 в static/brand/, высота отображения 30px, alt пустой, доступное имя остаётся у ссылки (aria-label=homeLabel). TDD-регрессия в tests/ui/discovery.test.ts проверяет img + aria-label ссылки. Мобильная навигация без вордмарка; бренд в админ-сайдбаре остаётся текстовым (вне рамки поручения, если владелец не попросит отдельно). Логотип сохраняет фирменный светло-серый текст — проверено в светлой и тёмной теме.

## Проверки

svelte-check 0/0; discovery-тесты 9/9 (новая регрессия зелёная после RED-фазы); architecture/contrast; web build; скриншоты шапки на стенде — tmp/logo-header-*.png.
<!-- SECTION:DESCRIPTION:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Реализовано: ассет static/brand/qmanga.png (magick strip+resize 288x96, 15КБ), img в .wordmark, CSS высота 30px, RED/GREEN регрессия в discovery.test.ts. Полные локальные проверки зелёные; скриншот-проверка выполняется на стенде перед выпуском.
<!-- SECTION:NOTES:END -->
