---
id: TASK-119
title: Expandable title genres and tags with dense local examples
status: Review
assignee:
  - '@codex'
created_date: '2026-10-21 06:40'
updated_date: '2026-10-21 06:49'
labels: []
dependencies: []
priority: high
type: enhancement
ordinal: 61005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Owner requests many genres/tags to remain compact with conditional More/Collapse instead of an endless horizontal header. Preserve every taxonomy link and current palette; enrich only clearly synthetic local data.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Genres/tags wrap into a bounded preview with disclosure only on actual overflow; expand/collapse, resize, keyboard and both locales/themes work.
- [x] #2 All category/collection/genre destinations remain available; a dense local synthetic example is provided without changing real manga metadata.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Build measured taxonomy disclosure with responsive preview height and preserved links; test short/empty/dense content and resize. Add more real taxonomy IDs to local synthetic demo through existing API.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-10-21: 0c28667 adds measured responsive tag preview, conditional More/Collapse and inert clipped links; native taxonomy destinations retained. Short/dense lists, expansion, resize, both locales/themes and viewport fit passed isolated browser tests. Local fictional synthetic-29e5596f enriched through admin API with 12 existing tags (API maximum), preserving other metadata; isolated preview adds 30 tags and up to 50 release editions without backend mutation. Check 0/0, 56 UI tests/816 assertions, build/boundaries/contrast pass; independent review no blockers. Local only; owner review pending. Preview: http://localhost:3002/?view=title&tags=30&editions=10&section=chapters . No production changes.
<!-- SECTION:NOTES:END -->
