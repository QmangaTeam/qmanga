---
id: TASK-123
title: Admin-managed API tokens and headless manga publication
status: Review
assignee:
  - codex
created_date: '2026-10-22 19:06'
updated_date: '2026-10-22 19:16'
labels: []
dependencies: []
ordinal: 65005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Owner wants Codex to upload manga without browser. Clarification23Oct: create/revoke API tokens in admin panel, act as token owner account and publish directly without moderation. Reuse existing catalog/upload/processing pipeline and preserve ordinary user moderation. No deployment requested in this package.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Administrator creates and revokes own named tokens; secret displayed only once, database stores hash, expires and current owner role/block status enforced.
- [x] #2 Bearer import can create titles/chapters and upload prepared pages under owner identity; allowed routes/scopes deny unrelated administration and credential management.
- [x] #3 Explicit direct-publish scope publishes ready valid revision without moderation queue, atomically audited under owner/token; ordinary submissions remain moderated.
- [x] #4 Documented CLI uploads folder or CBZ, resumes interrupted pages without duplicate revisions and reports processing and publication failures truthfully.
- [x] #5 Real PostgreSQL/HTTP and local browser tests plus independent review verify tokens, revocation, privilege loss and end-to-end publishing.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1.Extend existing identity flow with hashed owner-bound admin API tokens, narrow import route access, expiry/revocation and live role checks.2.Add atomic explicit direct publication for token scope reusing domain/media validation and audit/outbox.3.Admin UI create/list/revoke once-only secrets within existing workspace.4.Headless upload CLI/manifest with folder/CBZ ordering, stable keys and resumability, documentation.5.Integrate real Docker/PG/HTTP/media end-to-end and browser regressions; independent review; local Review status, no push/deploy unless requested.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Owner authorized push+deployment after completion and requested super short Codex API guide. Implemented hashed adminowner tokens, scoped bearer, token-only atomic direct publication/cover and adminUI. Parent integrates CLI/guide and tests; review found CLI chapter409/name, historicalrevision and lostsubmitresponse issues now fixed with regressions. FullGoPG/codecs PASS,168webunit and64UI tests PASS; focusedpublishingHTTPaftercover PASS. End-to-end localDocker/browser in progress before final independent review and release.

Local implementation verified:65ea31b UI,322dc27 identity,b8ba023 publication,690b146 cover/currentpointer,96d0825 CLI/guide,a55a967 bootstrap+CI/E2E. Full Go realPG/codecs PASS; post-cover publishing+HTTP realPG PASS;168webunit,64UI,11Pythonclientregressions PASS;productionwebbuild PASS;svelte-check0errors0warnings;architecture/contrastPASS. RealDockerE2E adminUItokencreate -> headlessnewtitle/project/chapter -> PNGprocessing -> selectedcover -> directpublish -> anonymousimage200 -> identicalrepeat -> disallowedadminroute403 -> UIrevoke -> bearer401 PASS.12renderedadminviews kk/ru,light/dark,320/360/1280 PASS; inspected screenshot. Independentidentity/UI/publicationreviews clean; CLI4reviewfindings fixed and regressiontested, stalecoverdocs corrected. No real-user token created; secrets only in test process memory. Owner authorized push/release and mini guide docs/CODEX-UPLOAD.md. ReadyforremoteCI/staging/production exactSHA acceptance; stayReview untilobserved.
<!-- SECTION:NOTES:END -->
