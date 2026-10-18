---
id: TASK-11
title: 'Модель жанров, категорий и подборок'
status: In Progress
assignee:
  - '@codex'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 21:55'
labels:
  - stage-3
dependencies: []
references:
  - UPDATE.md
ordinal: 11
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-3. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Развести **жанры** (уже существуют), **метки/категории содержания** и **редакционные подборки**. Подборка имеет заголовок, описание, порядок, видимость и выбранные тайтлы; жанр — таксономия для фильтра; возраст — отдельное поле, не жанр. Не смешивать эти сущности в одну таблицу.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add separate category and editorial collection entities, persistence and invariants. 2. Verify migration and repository behavior.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented separate categories and ordered, visible editorial collections in migration 000024; genres and age remain independent. Added domain and isolated PostgreSQL regression. Go catalog/HTTP unit tests pass; PostgreSQL run pending harness URL.
<!-- SECTION:NOTES:END -->
