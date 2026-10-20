---
id: TASK-104
title: Mobile title reflow and single-Q wordmark in both themes
status: Review
assignee:
  - '@codex'
created_date: '2026-10-20 19:50'
updated_date: '2026-10-20 20:17'
labels: []
dependencies: []
priority: high
type: bug
ordinal: 46005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Owner requests mobile title-page responsiveness using attached second image only as responsive reference, retaining Qmanga style/functions; fix duplicated Qqmanga wordmark to one Qmanga in dark/light/system mode. Leave full app running locally for owner review before release.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 At 320/360/487px title cover, heading, actions and facts fit without clipped content; existing desktop structure and actions preserved.
- [x] #2 Wordmark reads Qmanga once and remains legible in dark, light and OS system themes.
- [x] #3 Focused browser regression, responsive theme screenshots and independent review pass; local Docker URL and restart commands provided to owner.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Inspect existing rendering; adjust presentation only, compose existing Q mark with manga text; add regressions, rebuild existing local Docker stack, inspect mobile/desktop kk/ru/theme matrix, independent review and local handoff.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented mobile title heading→large centered cover→full-width Read/library→wrapping utility actions; facts reflow two columns and all tabs fit at320/360/487. Heading precedes cover in DOM; existing desktop composition retained. Mobile duplicate cover-banner fallback hidden; explicit artwork retained and failed-banner fallback follows actual selected source. ExistingQfavicon + manga. native theme-token text replaces duplicatedQ asset, legible dark/light/OSsystem. RED: newSSRwordmark test failed old2images; browser failed old100pxcover. GREEN: finalrealDocker9browserPASS (new8locale/theme matrix x3phonewidths+desktop and existingtitle/navigation);UI46PASS;Svelte0errors0warnings,architecture/contrastPASS. Independent mobile_logo_review found bannerfallback edge, fixed/re-reviewedPASS. Screenshots under ignoredtmp/task104/screenshots inspected at360dark/1280light. Full localDockerstack remainshealthy at http://localhost:18080; title /titles/synthetic-29e5596f. Restart fromrepo: docker compose --env-file tmp/task68/local.env -p qmanga-stage68-fixes up -d --wait. No secrets copied or volumes reset. User requests localvisualreview; no push/production rollout in this task yet. StatusReview pending owner localacceptance and subsequent authorisedrelease.

Owner follow-up: equal column spacing and aligned labels/values in title facts at885px; inspect row heights/negative link margins and ensure320/360/885/1280 both themes.

Follow-up: equal five-column facts grid with shared label/value rows; two equal columns on phones. Local Svelte check0errors0warnings;46UItests pass; Docker webbuild passes; read-only browser regression passes kk/ru × dark/light/system-dark/system-light at320/360/487 plus885/1280 spacing assertions. Independent review confirms alignment at320/360/768/885/960/1024/1280. Awaiting owner local acceptance, no push/release.
<!-- SECTION:NOTES:END -->
