---
id: TASK-67
title: 'Каталог: панель фильтров на ПК не скроллится'
status: Review
assignee: []
created_date: '2026-10-20 03:29'
updated_date: '2026-10-20 03:41'
labels:
  - catalog
  - frontend
dependencies: []
type: bug
ordinal: 42026
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Дефект к TASK-65. На desktop (>=1024px) правая панель фильтров обрезает содержимое: .panel-clip имеет только max-height, поэтому .panel-scroll с height:100% разрешается в auto и не становится скролл-контейнером; нижние секции (доступность и др.) недостижимы. Мобильный полноэкранный диалог скроллится нормально. Исправление: задать .panel-clip определённую высоту height:calc(100dvh - 166px).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Desktop 1280: .panel-scroll имеет overflow-y:auto и scrollHeight>clientHeight, скролл доходит до последней секции; контент не обрезается.
- [x] #2 Регрессия в apps/web/tests/browser/catalogue-filters.test.mjs падает до фикса и проходит после.
- [x] #3 Мобильная панель 320/360 скроллится как раньше; kk/ru, обе темы, нет горизонтального переполнения.
- [x] #4 Выпущено на production, внешняя проверка того же SHA.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1) Фикс CSS .panel-clip. 2) Регрессия в браузерном тесте. 3) Локальная проверка через браузер. 4) Fast Release и внешняя проверка скролла.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Корневая причина: .panel-clip в desktop-медиа имел только max-height; содержащий блок .panel-scroll (.panel-body, растянутый flex-элемент) не имел определённой высоты, height:100% разрешался в auto, .panel-scroll получал высоту контента 4665px, overflow-y:auto не активировался, а .panel-clip{overflow:hidden} обрезал нижние секции. Исправление (коммит 36b3a5f): в media >=1024px .panel-clip height:calc(100dvh - 166px) вместо max-height. Регрессия в apps/web/tests/browser/catalogue-filters.test.mjs: overflow-y, scrollHeight>clientHeight, scrollTop доходит до конца, последний заголовок внутри .panel-clip. Мобильный <dialog> не затронут.

Проверки: Verify Qmanga 39359422647 success (обе acceptance-половины + браузерные шарды с новой регрессией); Fast Release 39359422603 success (web-only, build 34s). Production 36b3a5f (push->prod 608s при старом 5-мин таймере). Внешне на https://qmanga.maqsatto.dev/catalog: desktop 1280 kk/ru — .panel-scroll 4665/570, scrollTop 4095, последняя секция 'Қолжетімділік'/'Доступность' внутри clip, горизонтального переполнения нет; mobile 360/320 — диалог скроллится (4665/675), нижняя секция достижима, переполнения нет. Осталось: независимый review субагентом (запущен) и, при чистом результате, статус Done.
<!-- SECTION:NOTES:END -->
