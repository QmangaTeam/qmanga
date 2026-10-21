---
id: TASK-106
title: Auth branding and center-logo mobile Home navigation
status: Done
assignee:
  - '@codex-branding'
created_date: '2026-10-20 21:42'
updated_date: '2026-10-21 05:01'
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

Owner-authorized release accepted2026-10-21: final source1cdb00a11d9ef8a24c296a9a117501f44ebae203 pushed to GitHub main; Fast Release39546480513 success (https://github.com/maqsatto/qmanga/actions/runs/39546480513). Full Verify39546480550 success,98browser regressions45+53,0fail0skip (https://github.com/maqsatto/qmanga/actions/runs/39546480550). Cold CI test config/cache and statistics accessible-label regressions fixed; title interactions wait for hydration. Local Svelte0errors0warnings,51UItests802assertions,frontendbuild pass; isolated codec4/4 and clean community5/5, read-only title interaction1/1 pass. Independent final review no blockers. VM poller automatically released same SHA staging blue then production blue05:00UTC; oldgreen retained for rollback. Public health/PostgreSQL/S3 and exact homepageSHA confirmed; actual public kk/ru x dark/light browser matrix320/360/1280 passes login/register singleQbrand, centeredmobileHomeQ, titlemetadata/icons/status/contributors, hiddenrails, conditionaldescription and tabs/nooverflow. No local private fixture data or accounts uploaded; no migration/backend changes. Physical Android is not claimed. No remaining release action for these cards.
<!-- SECTION:NOTES:END -->
