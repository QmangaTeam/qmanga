---
id: TASK-109
title: Compact responsive title header and shallower banner
status: Review
assignee:
  - '@codex'
created_date: '2026-10-21 05:46'
updated_date: '2026-10-21 06:04'
labels: []
dependencies: []
priority: high
type: enhancement
ordinal: 51005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The audited mobile cover/header delays reading and tabs; the desktop banner consumes320px. Owner approved compact cover beside title and earlier reading actions while keeping Qmanga colours and all metadata/actions.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 At320/360px cover and title share a compact header, reading actions and tabs appear earlier without clipped long titles or inaccessible controls.
- [x] #2 Desktop and laptop retain sidebar actions with a shallower banner and no excessive blank header space; absent/broken images remain usable.
- [x] #3 Title dialogs, tab URLs/no-scroll, conditional description and metadata behaviour remain; local theme/locale/responsive checks and independent review pass.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Refine existing grid: cover and title/rating side by side on mobile, actions below; shallower banner and reduced desktop overlap/padding. Preserve all dialogs, metadata and tab contracts. Verify long/sparse/broken-art titles320–1440 in themes/locales.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented100px mobile cover beside title/rating, reading/actions below; desktop sidebar200px and banner max180px. Read on actual Jupar at360px measured286px fromtop,1280px577px. Long names reflow, missing-art SSR placeholder retainsRead, aborted image requests leave usable320/1280layout. Existing metadata/tab/dialog/description/branding/theme contracts retained. Final verified local source/test/docs SHA17af6d8e834bb55b2752ba188507bf4888aa66ed (product commitsb07df75/f0aa9ab/4baebd9). Local Svelte0errors0warnings;55UItests813assertions;productionbuild,boundaries,both-themecontrast pass.14unique browser checks pass (5titlematrix,1actualguesthome,8isolatedcomments/home);0fail0skip in full selected runs. Sparse/populated32theme-locale-size-width combinations,320/360/768/1024/1280/1440 title checks, keyboard/deep links/reduced motion/drafts/errors covered. Independent product and incremental fixture/docs reviews no blockers. Screenshots/logs ignoredtmp/task108. MainHMRhttp://localhost:18080 and syntheticpreviewhttp://localhost:3002 running. Owner requestedLOCAL reviewfirst: no push/deploy; production still1cdb00a11d9ef8a24c296a9a117501f44ebae203 confirmedpublicSHA. Full mutating acceptance suite not run on persistent owner stack; existing authenticated integration cases updated for isolatedCI. PhysicalAndroid not tested. Next: owner checks local UI; release only after authorization.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented100px mobile cover beside title/rating, reading/actions below; desktop sidebar200px and banner max180px. Read on actual Jupar at360px measured286px fromtop,1280px577px. Long names reflow, missing-art SSR placeholder retainsRead, aborted image requests leave usable320/1280layout. Existing metadata/tab/dialog/description/branding/theme contracts retained. Local checks/review pass; waiting owner local acceptance, no release.
<!-- SECTION:FINAL_SUMMARY:END -->
