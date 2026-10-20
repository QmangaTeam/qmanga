---
id: TASK-106
title: Auth branding and center-logo mobile Home navigation
status: Review
assignee:
  - '@codex-branding'
created_date: '2026-10-20 21:42'
updated_date: '2026-10-20 21:53'
labels: []
dependencies: []
type: enhancement
ordinal: 48005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Owner asks Qmanga logo on login/register, Home moved to center mobile item using Qlogo instead ofhomeicon. Preserve all5destinations,auth flows,colors,locales andtouchsizes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Login/register show accessible single-Q branding inboth themes; existingforms/returnTo preserved.
- [x] #2 Mobile Home is centered,Qlogo shown instead ofhomeicon,no duplicate Home item;all5destinations/badges/focus work320px.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Inspect sharedauth/navigationcomponents, implement boundedpresentation changes,meaningfulSSR/browserchecks; localownerreview,no push.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented shared auth wordmark using existing single-Q favicon and theme text/accent tokens. MobileNavigation centers unique Home as third of five destinations using Q mark; catalogue/library/notifications/menu and badge/current-link/focus preserved. No auth request/returnTo logic or desktop navigation order changed. Verified: branding UI SSR 2/2 tests (17 assertions); bun run check 0 errors/warnings; Chromium read-only auth-branding browser test 1/1 covering login/register at 320/360/1280, kk/ru, dark/light, navigation order/targets/focus/centering; git diff --check clean. Visually inspected 320px login/catalog screenshots in tmp/task-106. Local review only; no push/release. Root coordinates commit SHA, independent review and production acceptance.

Committed source/tests as7905513. Independent reviewer reran auth-branding anddensity tests2pass; centeredHome,all5destinations,44pxtargets andauthbrand bothlocales/themes verified. Rootcombinedbuild/type/UI/contrast/boundary checks pass. Localacceptanceonly,Reviewpendingowner; notproductionrelease.

Final combined verification:10/10 browser tests pass,0skips;51UItests806assertions pass;Svelte0errors0warnings;frontendbuild,boundaries andboth-themecontrast pass. Independentreviews no blockers. Productcommits c6a0254/7905513/be7491c; a37952a recorded docs/readiness beforethisfinalnote. MainfrontendHMR18080 andsyntheticreplypreview3002 remainrunning. Awaitownerlocalacceptance; no productionrelease.
<!-- SECTION:NOTES:END -->
