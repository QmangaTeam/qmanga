---
id: TASK-109
title: Compact responsive title header and shallower banner
status: In Progress
assignee:
  - '@codex'
created_date: '2026-10-21 05:46'
updated_date: '2026-10-21 05:47'
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
- [ ] #1 At320/360px cover and title share a compact header, reading actions and tabs appear earlier without clipped long titles or inaccessible controls.
- [ ] #2 Desktop and laptop retain sidebar actions with a shallower banner and no excessive blank header space; absent/broken images remain usable.
- [ ] #3 Title dialogs, tab URLs/no-scroll, conditional description and metadata behaviour remain; local theme/locale/responsive checks and independent review pass.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Refine existing grid: cover and title/rating side by side on mobile, actions below; shallower banner and reduced desktop overlap/padding. Preserve all dialogs, metadata and tab contracts. Verify long/sparse/broken-art titles320–1440 in themes/locales.
<!-- SECTION:PLAN:END -->
