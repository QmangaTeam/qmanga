---
id: TASK-116
title: Scale chapter lists for volumes and alternate releases
status: Review
assignee:
  - codex
created_date: '2026-10-21 05:46'
updated_date: '2026-10-22 13:25'
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

Owner follow-up22Oct: add title-wide translation-team selection above chapter list, matching supplied reference with team avatar/name and selected state. Preserve All translations, URL query/sort/pagination, exact release IDs, both locales/themes and320px. Implement server-side team filter before pagination, UI, targeted regressions, independent review; local only.

074fc2e: title-wide team chips with real avatar/fallback, selected state and All translations. Server team slug filter applies to visible logical chapters before pagination and to returned releases. Search/sort/next URLs preserve team; switching resets cursor; main Read respects team. Single-team unfiltered titles omit redundant selector. Verification: Go application/HTTP pass; actual isolated PostgreSQL mixed-team/search/pagination visibility regression pass; gateway4 tests pass; UI62/854 pass; Svelte0 errors0 warnings; build,boundaries,contrast,diff checks pass. Permanent browser regression1/1 pass; actual synthetic two-team fixture switching/sort/reset/exact reading links across320/360/668/1280 x kk/ru x dark/light passed, screenshots inspected, zero page errors/root overflow. Independent review found Read mismatch, fixed and re-reviewed with no remaining findings. Local dev API rebuilt and restarted, web HMR remains active. No push/deploy. AC1 broader volume navigation remains outside this requested package; next: owner local acceptance.

Owner-authorized deployment confirmed22Oct13:25UTC production c6bee6ad3d16896607d6ed9b4fc8335721e8301a includes team selector074fc2e. Full Verify39782687712 and FastRelease39782687750 success;117 browser regressions,0 failures; exact public SHA/health verified. Broader volume-navigation AC1 remains outside released selector scope, so Review retained.
<!-- SECTION:NOTES:END -->
