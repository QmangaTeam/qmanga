---
id: TASK-7
title: Возрастная маркировка тайтлов
status: In Progress
assignee:
  - '@codex'
created_date: '2026-10-18 21:20'
updated_date: '2026-10-18 21:40'
labels:
  - stage-2
dependencies: []
references:
  - UPDATE.md
ordinal: 7
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Перенесено из UPDATE.md от 18 октября 2026 года, stage-2. Общие требования: UPDATE.md. Перед реализацией сверить с текущим кодом; за пределами этой карточки соседние функции не добавлять.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Возрастная маркировка на тайтле и карточках: минимум 12+/16+/18+, 18+ заметно красная; возраст хранится на сервере и фильтруется.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Регрессиями зафиксировать допустимые значения 0/12/16/18 и отклонение остальных. 2. Добавить совместимую миграцию age_rating=0 для прежних тайтлов, запись/чтение и точный серверный фильтр age_rating; старый PUT без поля сохраняет маркировку. 3. Прокинуть поле через DTO/адаптеры, админку, карточку/полки/каталог, URL-фильтр и красную метку 18+. 4. Реальные PostgreSQL/API/browser проверки, полный build и интеграционный набор; коммиты по завершённым частям, Review до CI/выпуска.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Backend slice verified: migration 000023 adds constrained smallint age_rating default 0 and composite age/update/id index. Domain validates 0/12/16/18, summary/detail DTO expose integer, public age_rating filter exact (including 0). Update omission uses nullable command field and SQL COALESCE to preserve classification; explicit 0 clears. Real isolated PostgreSQL tests verify persistence, filtering, DB rejection of 14, legacy insert/update, admin-only changes and clearing. HTTP regression before implementation accepted age_rating=14 with 200; now rejects invalid/malformed values. Full Go/PostgreSQL/codec suite passed 76 test/subtests across 23 packages, no actual skips: tmp/age-full-go.jsonl. Frontend adapter 3 tests and SSR 27 passed; app diagnostics 0 errors/warnings. UI/browser integration and Docker deployment still pending.
<!-- SECTION:NOTES:END -->
