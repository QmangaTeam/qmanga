---
id: TASK-12
title: Управление и фильтрация категорий и подборок
status: In Progress
assignee:
  - '@codex'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 21:57'
labels:
  - stage-3
dependencies:
  - TASK-7
  - TASK-11
references:
  - UPDATE.md
ordinal: 12
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-3. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Администратор создаёт/редактирует жанры, категории и подборки, назначает их тайтлу. Публичный каталог фильтрует по ним и возрасту; сочетания фильтров отражаются в URL и доступны по прямой ссылке. На карточке метки компактны, без «ещё 31» как единственного способа понять содержание.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add authorized taxonomy editing and title assignment. 2. Expose public URL filters and compact title labels. 3. Verify Go, frontend and browser paths.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented admin /admin/taxonomy CRUD, ordered collection title selection, category assignment in title editor, combined category/collection/genre/age URL filters, and compact linked title labels. Commits 433aa31/e85e4e5/b9058a8. bun test catalog: 10 pass, 47 assertions; frontend dependency boundaries pass; Go catalog and HTTP pass. Initial web check clean; latest shared check blocked only by pending metrics/avatar composition wiring from concurrent work. Browser regression taxonomy.test.mjs added to existing glob; awaiting coordinated API/web rebuild for execution. No release claimed.
<!-- SECTION:NOTES:END -->
