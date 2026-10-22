---
id: TASK-126
title: Close title library picker after a saved status choice
status: Review
assignee:
  - '@codex'
created_date: '2026-10-22 21:01'
updated_date: '2026-10-22 21:04'
labels: []
dependencies: []
ordinal: 67005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Owner reports title Add to list dialog stays open after choosing a reading status. Close after successful status save, preserve selected label/state and focus. Keep dialog open on failure or pending write. No navigation redesign.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Choosing a status closes the dialog after successful save and restores trigger focus; re-opening retains the selection.
- [x] #2 Failed save keeps dialog open with feedback; regression and local Docker preview verified.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Reproduce with browser regression; add success callback at status-choice boundary; verify failure/success/focus; independent review and local Docker rebuild.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented at 2659123. Status-choice success callback closes the title-owned dialog; native focus returns to trigger and selected label persists. Pending/offline/failed writes leave dialog and sync feedback visible; advanced preferences Save unchanged. Browser regression reproduced old behavior (timeout waiting for dialog to hide), then passed against rebuilt integrated Docker web: failed write, delayed write, success close, focus, updated label and retained selection. Regression intercepts writes to preserve owner data. Actual owner browser kk flow also verified: choosing the existing completed status closes popup and restores focus without changing the choice. Independent subagent review approved. PASS: 64 UI tests/864 assertions, 2 library-sync tests/13 assertions, app and UI-preview Svelte diagnostics 0 errors/0 warnings, architecture/contrast, Docker web build and web/gateway health, backlog doctor and diff check. Updated existing title-actions/cover-statuses expectations; their full data-mutating suites not rerun against preserved local data. Local preview active at localhost:18080; no push, CI or production deployment requested/performed, production SHA not re-observed. Next: owner local acceptance; authorize push/release if desired.
<!-- SECTION:NOTES:END -->
