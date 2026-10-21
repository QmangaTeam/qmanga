---
id: TASK-116
title: Scale chapter lists for volumes and alternate releases
status: Review
assignee:
  - '@codex'
created_date: '2026-10-21 05:46'
updated_date: '2026-10-21 06:49'
labels: []
dependencies: []
priority: medium
type: enhancement
ordinal: 58005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
As real series grow, readers need volume structure, release-team distinctions and reading-state cues. Current release selection and progress must remain authoritative.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Large chapter lists support useful volume/release organization without mixing editions or inventing progress.
- [x] #2 Sparse titles remain compact; direct links, filters, sorting and keyboard/mobile navigation retain context.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Use compact logical-chapter rows; multiple releases open one accessible bounded searchable edition chooser, preserving exact release/team destinations and dates. Test1/2/10/50releases with long names; enrich synthetic local review fixture only as needed.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-10-21: 0c28667 completes requested alternate-edition density package: single-release native reading link; multiple editions collapse into one chapter row, bounded scrollable chooser and team search above five releases. Exact read/team URLs and dates retained. Isolated browser stress passed 1/2/10/50 releases, long names and 320/624/1280 widths; API-backed search/sort test passed. Independent review found no blockers. AC1 broader volume-navigation work remains outside this package; current volume labels remain intact. Local owner review pending; no push/deploy.
<!-- SECTION:NOTES:END -->
