---
id: TASK-110
title: Collapsed comment composer and collapsible reply threads
status: In Progress
assignee:
  - '@codex'
created_date: '2026-10-21 05:46'
updated_date: '2026-10-21 05:47'
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
- [ ] #1 Signed-in root composer starts collapsed, opens with focus, closes without losing a draft, and stays open with text/spoiler on failed send; successful send returns to compact prompt.
- [ ] #2 Reply threads can expand/collapse with count and accessible state; loaded replies are retained, remaining-page fetching works, focused reply links reveal their thread.
- [ ] #3 Guest prompt, reply/edit/cancel/format/preview/report/paging remain correct;320/360/1280px kk/ru dark/light and keyboard/reduced-motion checks pass locally.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Add root composer disclosure and separate saved root draft when replying/editing; accessible close/focus return and failure retention. Track expanded root threads, auto-reveal focused/active replies and retain loaded paging. Update isolated synthetic fixture regressions; inspect local preview without account-data writes.
<!-- SECTION:PLAN:END -->
