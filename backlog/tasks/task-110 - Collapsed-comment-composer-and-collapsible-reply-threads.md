---
id: TASK-110
title: Collapsed comment composer and collapsible reply threads
status: Review
assignee:
  - '@codex'
created_date: '2026-10-21 05:46'
updated_date: '2026-10-21 06:04'
labels: []
dependencies: []
priority: high
type: enhancement
ordinal: 52005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Owner approved progressive discussion UI: signed-in users should see a compact writing prompt, and long loaded reply threads should be collapsible. Existing drafts, failures, deep links and mutation scope must survive.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Signed-in root composer starts collapsed, opens with focus, closes without losing a draft, and stays open with text/spoiler on failed send; successful send returns to compact prompt.
- [x] #2 Reply threads can expand/collapse with count and accessible state; loaded replies are retained, remaining-page fetching works, focused reply links reveal their thread.
- [x] #3 Guest prompt, reply/edit/cancel/format/preview/report/paging remain correct;320/360/1280px kk/ru dark/light and keyboard/reduced-motion checks pass locally.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Add root composer disclosure and separate saved root draft when replying/editing; accessible close/focus return and failure retention. Track expanded root threads, auto-reveal focused/active replies and retain loaded paging. Update isolated synthetic fixture regressions; inspect local preview without account-data writes.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented signed-in compact root prompt; opening focuses editor, closing preserves root body/spoiler and returns focus. Inline reply/edit restore root draft on cancellation. Success collapses root; failed send retains editor/draft. Loaded threads expand/collapse without refetch or destruction; remaining paging and focusedreply auto-reveal preserved. Busy actions protected; short composer entrance respects reduced motion. Guest flow unchanged. Isolated synthetic preview available3002. Final verified local source/test/docs SHA17af6d8e834bb55b2752ba188507bf4888aa66ed (product commitsb07df75/f0aa9ab/4baebd9). Local Svelte0errors0warnings;55UItests813assertions;productionbuild,boundaries,both-themecontrast pass.14unique browser checks pass (5titlematrix,1actualguesthome,8isolatedcomments/home);0fail0skip in full selected runs. Sparse/populated32theme-locale-size-width combinations,320/360/768/1024/1280/1440 title checks, keyboard/deep links/reduced motion/drafts/errors covered. Independent product and incremental fixture/docs reviews no blockers. Screenshots/logs ignoredtmp/task108. MainHMRhttp://localhost:18080 and syntheticpreviewhttp://localhost:3002 running. Owner requestedLOCAL reviewfirst: no push/deploy; production still1cdb00a11d9ef8a24c296a9a117501f44ebae203 confirmedpublicSHA. Full mutating acceptance suite not run on persistent owner stack; existing authenticated integration cases updated for isolatedCI. PhysicalAndroid not tested. Next: owner checks local UI; release only after authorization.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented signed-in compact root prompt; opening focuses editor, closing preserves root body/spoiler and returns focus. Inline reply/edit restore root draft on cancellation. Success collapses root; failed send retains editor/draft. Loaded threads expand/collapse without refetch or destruction; remaining paging and focusedreply auto-reveal preserved. Busy actions protected; short composer entrance respects reduced motion. Guest flow unchanged. Isolated synthetic preview available3002. Local checks/review pass; waiting owner local acceptance, no release.
<!-- SECTION:FINAL_SUMMARY:END -->
