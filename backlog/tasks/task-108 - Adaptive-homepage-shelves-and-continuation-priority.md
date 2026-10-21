---
id: TASK-108
title: Adaptive homepage shelves and continuation priority
status: Review
assignee:
  - '@codex'
created_date: '2026-10-21 05:46'
updated_date: '2026-10-21 06:04'
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
- [x] #1 Returning readers with resumable items reach continuation before discovery; guests retain useful discovery and honest empty/error feedback.
- [x] #2 Default type-top shelf avoids repeating titles already on the main ranking; explicit type selection remains reachable and shows its real results/empty/error state.
- [x] #3 Sparse and populated data fit320/360/768/1280px in kk/ru and both themes; meaningful regressions pass and owner can inspect locally.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Preserve continuation loader lifecycle while exposing presentation state; render returning continuation first, deduplicate default type shelf against most-read while keeping explicit kind navigation and error/empty states; compact sparse spacing/update grid. Add sparse/populated/returning SSR and browser regressions; inspect local matrix.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented continuation presentation snippet without loader remount; returning readers see it first. Default type-top excludes main-ranking duplicates; explicit top query keeps real results/empty/error and all type links. Sparse spacing/update columns adapt. Regressions cover new/returning order, default vs explicit top, populated8-title eager budget and real guest local progress with exactly1referencesrequest. Final verified local source/test/docs SHA17af6d8e834bb55b2752ba188507bf4888aa66ed (product commitsb07df75/f0aa9ab/4baebd9). Local Svelte0errors0warnings;55UItests813assertions;productionbuild,boundaries,both-themecontrast pass.14unique browser checks pass (5titlematrix,1actualguesthome,8isolatedcomments/home);0fail0skip in full selected runs. Sparse/populated32theme-locale-size-width combinations,320/360/768/1024/1280/1440 title checks, keyboard/deep links/reduced motion/drafts/errors covered. Independent product and incremental fixture/docs reviews no blockers. Screenshots/logs ignoredtmp/task108. MainHMRhttp://localhost:18080 and syntheticpreviewhttp://localhost:3002 running. Owner requestedLOCAL reviewfirst: no push/deploy; production still1cdb00a11d9ef8a24c296a9a117501f44ebae203 confirmedpublicSHA. Full mutating acceptance suite not run on persistent owner stack; existing authenticated integration cases updated for isolatedCI. PhysicalAndroid not tested. Next: owner checks local UI; release only after authorization.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented continuation presentation snippet without loader remount; returning readers see it first. Default type-top excludes main-ranking duplicates; explicit top query keeps real results/empty/error and all type links. Sparse spacing/update columns adapt. Regressions cover new/returning order, default vs explicit top, populated8-title eager budget and real guest local progress with exactly1referencesrequest. Local checks/review pass; waiting owner local acceptance, no release.
<!-- SECTION:FINAL_SUMMARY:END -->
