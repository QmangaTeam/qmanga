---
id: TASK-10
title: Боковые фильтры каталога и мобильная панель
status: Review
assignee:
  - '@codex'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 21:52'
labels:
  - stage-3
dependencies: []
references:
  - UPDATE.md
ordinal: 10
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-3. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 На компьютере фильтры — постоянная правая колонка рядом с сеткой; на телефоне кнопка открывает панель справа налево, с затемнением, закрытием и фиксированной кнопкой «Применить». Текущие поисковые параметры и URL остаются источником истины; при открытии видны выбранные значения. Сетка не прыгает и не прячется под нижней навигацией.
- [x] #2 Для примеров из скриншотов 5–7 сохранить общую плотность, контраст и плавное направление панели, но не переносить чужие названия, обложки или данные.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Browser regression: desktop right sidebar and mobile drawer preserving URL values, cancellation, Escape/backdrop/focus, Apply and query/cursor behavior. 2. Extract one filter-fields snippet into a presentation component with a persistent desktop form and native mobile dialog; retain current parameters including age. 3. URL remains committed state; opening resets only unsubmitted draft, Apply performs real GET, responsive resize closes modal. 4. Check both locales/themes, 320/360/desktop, reduced motion, build and full browser integration; Review before remote acceptance/release.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented persistent right sidebar at desktop and native right-hand modal drawer on mobile, sharing fields and GET contract. Committed URL values are restored on opening; canceled edits do not leak. Search preserves applied filters; Apply clears pagination cursor. Escape, close button, backdrop, focus return, desktop resize, reduced motion and fixed Apply covered. Visual inspection: 320px light drawer and 1280px dark sidebar. Docker build passes: 91 application + 28 SSR tests, Svelte checks 0 errors/warnings, architecture and contrast checks. Full isolated browser integration: 26/26 passing, 0 skipped (tmp/filter-full-browser.log), including age-rating compatibility. Root and isolated sandbox rebuilt successfully. Existing verify workflow discovers the new regression automatically. Local acceptance only: remote CI, staging and release of this SHA remain required; do not mark Done.

Implementation commit: 7958579. Age classification commits: 070824b and 91d2ba9. All 26 browser scenarios passed together against the rebuilt Docker integration stack.
<!-- SECTION:NOTES:END -->
