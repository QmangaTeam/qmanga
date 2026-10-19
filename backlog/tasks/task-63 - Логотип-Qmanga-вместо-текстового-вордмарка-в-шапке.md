---
id: TASK-63
title: Логотип Qmanga вместо текстового вордмарка в шапке
status: In Progress
assignee:
  - codex
created_date: '2026-10-19 04:58'
updated_date: '2026-10-19 05:39'
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

Дополнение владельца 19.10.2026, дословно: «да я зашел через гугл но не мог бы ты увеличить лого а то слишком мелки, и favicon поставь вот [изображение Q-пузыря]». Также подтверждён реальный вход Google (закрывает проверку TASK-62). Работа: высота вордмарка 30→40px, favicon.png 96×96 и apple-touch-icon.png 180×180 из пузыря (фон #111112 снят в прозрачность, magick fuzz 12% -transparent, trim), link rel=icon/apple-touch-icon в app.html head. TDD: RED-проверки favicon и высоты 40px в navigation-accountbrowser-тесте.

Дополнение владельца 19.10.2026, дословно: «да я зашел через гугл но не мог бы ты увеличить лого а то слишком мелки, и favicon поставь вот [изображение Q-пузыря]». Также подтверждён реальный вход через Google (закрывает проверку TASK-62). Работа: высота вордмарка 30→40px; favicon.png 96x96 и apple-touch-icon.png 180x180 из пузыря (фон #111112 снят в прозрачность, magick fuzz 12% -transparent, trim); link rel=icon и apple-touch-icon в head app.html. TDD: RED-проверки favicon-ссылок и высоты 40px в navigation-account браузерном тесте.

Реализация дополнения: favicon.png 96x96 и apple-touch-icon.png 180x180 (Q-пузырь на прозрачном фоне, apple-версия на #111112), link rel=icon + apple-touch-icon в head app.html; высота вордмарка 40px. Отладка браузерного теста: после гидрации атрибуты href head-ссылок нормализуются в абсолютные, а после width-лупа вьюпорт мобильный (логотип скрыт) — проверки переписаны на rel-селекторы + regex href + снятие высоты при 1280x800. GREEN: navigation-account 5/0 на стенде; favicon.png и apple-touch-icon.png 200 на стенде. Скриншоты шапки light/dark 1280: tmp/logo-header-*-v3.png.
<!-- SECTION:NOTES:END -->
