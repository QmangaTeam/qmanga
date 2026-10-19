---
id: TASK-63
title: Логотип Qmanga вместо текстового вордмарка в шапке
status: Done
assignee:
  - codex
created_date: '2026-10-19 04:58'
updated_date: '2026-10-19 05:13'
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

Владелец прислал логотип; светлая тема с исходным светлым текстом была нечитаема (скриншоты tmp/logo-header-*-v2). Добавлен светлый вариант static/brand/qmanga-light.png (текст #211b18, точка/пузырь оранжевые #FD6C33, magick -fuzz 5% -opaque), в шапке два img (.dark/.light) с переключением: data-theme=light и prefers-color-scheme light без атрибута. Скриншот-проверка на стенде: light показывает qmanga-light.png, dark — qmanga.png; обе темы читаемы. svelte-check 0/0, discovery 9/9, architecture/contrast, build — зелёные.

Выпуск: в пакете SHA f4d73ef9a047810e766f10be930e1089b2e01085. Verify Qmanga 39202510720 — acceptance и три ARM64 images success. Staging поднят poller-ом, production — release.sh (green, blue тёплый). Внешне: health 200, SHA на главной, оба ассета /brand/qmanga{,-light}.png 200 на staging и production. Статус Done — взятка подтверждена выпуском наблюдаемого SHA.
<!-- SECTION:NOTES:END -->
