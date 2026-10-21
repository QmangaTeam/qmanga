---
id: TASK-108
title: Adaptive homepage shelves and continuation priority
status: In Progress
assignee:
  - '@codex'
created_date: '2026-10-21 05:46'
updated_date: '2026-10-21 05:47'
labels: []
dependencies: []
priority: high
type: enhancement
ordinal: 50005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Small real catalogues currently repeat the same titles across several spacious shelves. Owner approved adaptive density and returning-reader priority, preserving honest rankings, updates and empty/error states.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Returning readers with resumable items reach continuation before discovery; guests retain useful discovery and honest empty/error feedback.
- [ ] #2 Default type-top shelf avoids repeating titles already on the main ranking; explicit type selection remains reachable and shows its real results/empty/error state.
- [ ] #3 Sparse and populated data fit320/360/768/1280px in kk/ru and both themes; meaningful regressions pass and owner can inspect locally.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Preserve continuation loader lifecycle while exposing presentation state; render returning continuation first, deduplicate default type shelf against most-read while keeping explicit kind navigation and error/empty states; compact sparse spacing/update grid. Add sparse/populated/returning SSR and browser regressions; inspect local matrix.
<!-- SECTION:PLAN:END -->
