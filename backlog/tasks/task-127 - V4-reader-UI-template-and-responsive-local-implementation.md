---
id: TASK-127
title: V4 reader UI template and responsive local implementation
status: Review
assignee:
  - '@codex'
created_date: '2026-10-22 22:25'
updated_date: '2026-10-22 23:12'
labels: []
dependencies: []
ordinal: 68005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Owner supplied V4 HTML/UX notes and drawer reference. UI only; no backend/database changes and no admin redesign. First build editable Figma template, then adapt public/reader screens to light/dark, kk/ru and PC/laptop/tablet/phone. Latest owner direction overrides old V4 bottom menu ordering: main navigation drawer opens from bottom-left, separate top-right account menu. Preserve real actions/data, no mock product flows. New package local only.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Editable Figma template covers light/dark and kk/ru at phone/tablet/laptop/PC widths with main navigation drawer and separate account menu.
- [x] #2 Local reader UI implements supplied V4 direction with real existing application contracts; backend/database and admin design untouched.
- [x] #3 Responsive local browser verification and independent review completed; limitations and remaining acceptance recorded.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Inspect supplied reference and current public views; create Figma foundation/screens; implement public navigation and V4 reader layouts; validate themes/locales/responsive interactions; independent review and rebuild local web.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Owner explicitly reiterated NEW UI MUST NOT BE PUSHED before local testing. Earlier popup fix was rebased onto new remote main and pushed as d312715; duplicate archived task125 was repaired by CLI to128 and pushed in dbdb4b0. All subsequent V4 commits are LOCAL ONLY. Implemented UI-only main mobile left drawer (bottom-left trigger; native dialog/focus/Escape), retained top-right account menu, separate contribution +, global header theme, title About taxonomy/metadata/inline actual statistics, desktop title actions, mobile one-row catalog controls with grid/list view, simplified spoiler/emoji composer and settings category index/detail/sidebar. Existing real auth/library/notifications/profile flows reused. No backend/schema/API/admin-design changes; excluded paths compared against origin/main. Figma new file ITJJNYDpt02v5Rn9Msnr6T created but Starter quota blocked native authoring immediately afterward; native UI/screens/components/AutoLayout NOT delivered. Fallback208 editable SVG viewport boards at tmp/v4/template (13 states x kk/ru x light/dark x390/834/1280/1440),7219editable text nodes,Onest embedded; XML validated and4representative render proofs inspected. SVG shadows/native controls approximate; full-page content and native Figma import unverified. ZIP tmp/v4/Qmanga-V4-Editable-Template.zip, gallery localhost18090/template/index.html. Acceptance1 remains open for nativeFigma when quota/access available. Local verification PASS: deterministic240view matrix(12routes x2locales x2themes x320/390/834/1280/1440) asserts actual theme,no overflow; drawer/list/filter/composer/settings/theme interactions; auth-branding and title-mobile-brand browser regressions;10 isolated community tests; library-picker regression;64UI tests/866assertions; app and preview Svelte0errors0warnings; architecture/contrast; Dockerweb build; backlogdoctor/diffcheck. Actual local drawer visually inspected. API/data/worker unchanged; web+gateway healthy localhost18080. No physical-device/full accessibility audit. RemoteCLI inspection unavailable(gh not authenticated); prior fix pushed, productionSHA not re-observed. Next: owner test local V4; nativeFigma completion blocked by account quota; no push/release without explicit approval.

Final SVG export refined native switch/checkbox rendering:208XML-validboards,7026editabletextnodes(earlier7219count included raw native-control values). Onest embedded;ZIP refreshed after four representative render checks. Final UI code commits fc1b4e7/bb99d56 are local-only; origin/main remains dbdb4b0. Native Figma acceptance remains open; no user approval of new UI or release claimed.

Final independent static review caught tablet768–1023Library shortcut gap; restored header icon and added direct-child visibility assertion at834across every matrix route/theme/locale. Correction independently approved. Rebuilt web/gateway healthy; repeated deterministic240view matrix and all interactions PASS(tmp/v4/matrix-tablet-final.log). TabletSVGboards regenerated after correction and XML/ZIP refreshed. New UI remains LOCAL ONLY; no push. Owner acceptance and nativeFigma quota/access remain open.

Final refreshed artifact count after tablet recapture:208SVGboards,7090editabletextnodes; embeddedOnest+SILOFL license included in ZIP.
<!-- SECTION:NOTES:END -->
