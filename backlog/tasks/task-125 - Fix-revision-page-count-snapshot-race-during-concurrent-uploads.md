---
id: TASK-125
title: Fix revision page-count snapshot race during concurrent uploads
status: Done
assignee:
  - codex
created_date: '2026-10-22 21:26'
updated_date: '2026-10-22 21:41'
labels: []
dependencies: []
ordinal: 67005
---

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Deterministic PostgreSQL regression reproduces stale counts while waiting for revision lock and passes after fix
- [x] #2 Upload and processing use fresh page counts after revision row lock, preserving immutable publication checks
- [x] #3 Affected unpublished production revision restored without losing ready pages; independent review and release verified
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Reproduce row-lock snapshot race; split locking and aggregate reads; verify real PostgreSQL regression and publishing tests; independently review; restore affected revision under guarded transaction and release.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Root cause reproduced in real PostgreSQL: blocked SELECT FOR UPDATE returned new revision version3 but page count1 instead of2 from its pre-wait snapshot. Split revision row lock and page aggregation into sequential statements under READ COMMITTED. Deterministic regression waits for pg_blocking_pids, fails before fix and passes after. Full publishing domain/application/PG/codecs plus HTTP tests passed. Independent review approved code and regression. Production recovery pending fresh backup and expired upload lease; retains19 ready pages and does not publish chapter.

Production backup20261022T212722Z verified locally and off-host. Independently reviewed guarded SQL passed actual rollback dry-run then committed revision-only repair ready59->processing60. All20 pages retained including19 ready; expired uploadingpage20 untouched/retriable, no publication. Fix438eb328808ee3637c901f4f129b75db6150d0f8 pushed; Verify39857591662 and FastRelease39857591695 running.

Automatic staging and production release confirmed on438eb328808ee3637c901f4f129b75db6150d0f8 (production green,2026-10-22 21:41UTC). External HTTPS homepage exactSHA and PostgreSQL/S3 health200 verified; revision remains processing60 with20 pages/19ready and no publication. FastRelease39857591695 success. Verify39857591662 shard0 success; shard1 still installing browser dependencies at handoff, full CI not claimed successful. Local affected publishing and HTTP suites passed; deterministic regression and independent review passed. User can retry expired page20 with same upload key and continue21-99 on same revision, fetching current version before versioned operations.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Fixed PostgreSQL READ COMMITTED snapshot race by reading page counts only after revision lock. Reproducing integration regression fails before and passes after; affected publishing/HTTP suites and independent review passed. Released438eb328808ee3637c901f4f129b75db6150d0f8. Production revision repaired without page loss or publication; current processing60,19ready+1expired uploading. Full remote Verify still running at handoff.
<!-- SECTION:FINAL_SUMMARY:END -->
