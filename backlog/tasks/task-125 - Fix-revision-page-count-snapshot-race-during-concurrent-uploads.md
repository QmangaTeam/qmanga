---
id: TASK-125
title: Fix revision page-count snapshot race during concurrent uploads
status: Review
assignee:
  - codex
created_date: '2026-10-22 21:26'
updated_date: '2026-10-22 21:28'
labels: []
dependencies: []
ordinal: 67005
---

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Deterministic PostgreSQL regression reproduces stale counts while waiting for revision lock and passes after fix
- [ ] #2 Upload and processing use fresh page counts after revision row lock, preserving immutable publication checks
- [ ] #3 Affected unpublished production revision restored without losing ready pages; independent review and release verified
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Reproduce row-lock snapshot race; split locking and aggregate reads; verify real PostgreSQL regression and publishing tests; independently review; restore affected revision under guarded transaction and release.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Root cause reproduced in real PostgreSQL: blocked SELECT FOR UPDATE returned new revision version3 but page count1 instead of2 from its pre-wait snapshot. Split revision row lock and page aggregation into sequential statements under READ COMMITTED. Deterministic regression waits for pg_blocking_pids, fails before fix and passes after. Full publishing domain/application/PG/codecs plus HTTP tests passed. Independent review approved code and regression. Production recovery pending fresh backup and expired upload lease; retains19 ready pages and does not publish chapter.
<!-- SECTION:NOTES:END -->
