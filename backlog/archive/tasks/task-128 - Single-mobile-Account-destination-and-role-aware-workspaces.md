---
id: TASK-128
title: Single mobile Account destination and role-aware workspaces
status: Review
assignee:
  - '@codex'
created_date: '2026-10-22 20:16'
updated_date: '2026-10-22 20:22'
labels: []
dependencies: []
ordinal: 67005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Owner approved replacing bottom Menu with Account, removing mobile header avatar, and consolidating profile/settings/Studio/Admin links. Preserve desktop dropdown and existing route compatibility. Local preview only; no push or release requested.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Mobile has one Account destination, no top avatar, and five existing navigation positions.
- [x] #2 Account supports guest, reader, approved team member, moderator and admin with appropriate links, settings and logout.
- [x] #3 Responsive localized verification, regression checks and independent review pass; local preview available.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Update shared navigation and Account screen; verify role states and responsive themes; independent review; rebuild local web and gateway.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented in d6f2080; CI browser regressions in b58c335. Mobile public header account controls hidden, bottom Account uses user icon/avatar, /menu retained as unified Account screen. Profile/settings/downloads/reports/help/logout preserved; verified approved team membership enables Studio; verified admin/moderator enables management. Desktop dropdown preserved; mobile workspace header links to Account. Footer and duplicate main-navigation tiles removed from Account. Account identity changes remount permission state; avatar cache version tracked. Independent review passed after lifecycle and legacy-test corrections. Local checks PASS: 64 UI tests/860 assertions; 5 locale/profile adapter tests; Svelte app and UI-preview diagnostics 0 errors/0 warnings; architecture/contrast; Docker production web build; backlog doctor; diff check. Integrated Docker browser mobile-account regression PASS: guest, owner, ordinary reader (empty membership response), admin; 48 responsive kk/ru dark/light views at 320/360/1280, no horizontal overflow; team failure/retry, desktop dialog and logout. Screenshots under tmp/recovered-local/account-screens inspected; owner browser refreshed to Account. Existing registration-import/library-cards/cover-statuses locators updated but these data-mutating suites not rerun against preserved local data; full CI remains pending. No backend/schema changes. Local web/gateway healthy at localhost:18080 with existing recovered data. No push, remote verification or production deployment performed; production SHA not re-observed in this local-only task. Next: owner accept local navigation; authorize push/release if desired. Physical mobile device testing remains unperformed.

Owner rejected the new navigation during local acceptance and explicitly requested restoration of the previous UI. Implementation and tests reverted; task canceled, no release requested. Preserve existing local manga/account data.
<!-- SECTION:NOTES:END -->
