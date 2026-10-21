---
id: TASK-111
title: Improve title metadata overflow cues and secondary-tab density
status: Review
assignee:
  - '@codex'
created_date: '2026-10-21 05:46'
updated_date: '2026-10-21 06:49'
labels: []
dependencies: []
priority: medium
type: enhancement
ordinal: 53005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Hidden horizontal scrollbars are an intentional owner preference, but extra facts are not obvious and repeated facts delay comments/chapters. Queue for later; preserve existing bounded rails and contributor chips.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Overflow has a subtle edge/partial-item cue only when more content exists; swipe/keyboard and hidden scrollbar remain.
- [x] #2 Evaluate a compact secondary-tab summary without losing access to original metadata; owner reviews locally before release.
- [x] #3 Metadata and contributor information appears only in About; chapter/comment views show their own content without that panel.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Show metadata/contributors only for About on every viewport; preserve grid flow, tab scroll and statistics access. Add real-overflow cue on existing hidden-scrollbar rail. Verify metadata absent in Chapters/Comments and present in About.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-10-21: 0c28667 makes full metadata/contributors exclusive to About, preserving stats there. Existing hidden-scrollbar rail adds real-overflow edge gradients. Mobile reflow/theme matrix and metadata/query/sort browser tests passed; tab scroll regression updated and passed for natural shorter-page clamping. SSR exclusivity regression passed. Owner chose no secondary-tab summary. Independent review: no blocking defects. Local only; owner acceptance pending.
<!-- SECTION:NOTES:END -->
