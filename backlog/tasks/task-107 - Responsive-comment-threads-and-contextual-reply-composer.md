---
id: TASK-107
title: Responsive comment threads and contextual reply composer
status: Review
assignee:
  - '@codex-comments'
created_date: '2026-10-20 21:42'
updated_date: '2026-10-20 21:53'
labels: []
dependencies: []
type: enhancement
ordinal: 49005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Owner reference for comment/replyUI flexibility, keepQmanga colors and existingfunctions. Improve author/avatar hierarchy, threadindentation, inline replycontext/composer/toolbaractions onphone anddesktop.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Comments/replies/composer use existingpalette andsupport320px/desktop,44pxtargets andlongtextwithoutoverflow.
- [x] #2 Reply/edit/cancel/format/spoiler/preview/report/pagination/drafts behavior preserved, meaningfulchecks andindependentreview.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Read currentdiscussioncomponent/contracts, refineUIonly, addscopedchecks and localpreview; no newbackendbehavior.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented scoped Discussion.svelte refinement: existing theme tokens retained; avatar via established /media/avatars/user/:id with initials fallback, aligned author/date headers on roots and replies, connected thread line, contextual reply/edit composer with close and focus return, compact flexible format/preview/cancel/send footer. Existing gateway methods and comment/report/paging contracts retained. Isolated synthetic gateway fixture and self-starting browser test added; no production/account/library mutations. Checks: browser community-discussion 4/4 passed (320/360/1280, kk/ru, light/dark, reduced motion, 44px targets, long text, reply/edit/cancel/bold/spoiler/preview/remove/reports/paging/focused link/guest/empty/retry); svelte-check 0 errors/warnings. Screenshots retained only in ignored tmp/task107. Next: root independent review, record final commit SHA and combined checks; owner acceptance/release remains pending. No production release performed.

Final verification after compact-composer feedback: community browser tests now 5/5 pass, including failed-submit draft/recipient/spoiler retention and successful retry; svelte-check again 0 errors/0 warnings; frontend architecture boundaries pass; git diff --check clean. Persistent isolated preview on http://127.0.0.1:3002/?locale=ru uses tests/community-preview/community-preview.config.ts; browser test starts and stops its own separate random-port Vite process.

Independentreview completed:4browsercasespassedpluslive320px nestedcomposerinspection; rootcombinedrunpassesall5communitycases includingfailedreplydraft/recipient/spoilerretention andexactmutationtargetIDs. RootfinalSvelte0/0,51UItests806assertions,frontendbuild+boundaries+contrastpass. Palette/gatewaycontracts unchanged; no datawrites inpreview. Compactinlinecomposer observed~225pxdesktop; fulltargettext retainedintitle/DOM. Awaitownervisualacceptance, localonly.

Committed source/isolatedpreview/tests asbe7491c. Independentreviewpassed UI/code and4browsercases; finalsource testedwith5casesincludingfailure-draftretention. Manualsyntheticpreview localhost3002 runningseparately; normalapplication remainslocalhost18080. Root observedcompactinlinecomposer inCua andsavedtmp/task104/discussion-refinement-live.png. DocsLOCALrecordsmanualpreviewcommand. Awaitownervisualacceptance; no releaseclaim.

Final combined verification:10/10 browser tests pass,0skips;51UItests806assertions pass;Svelte0errors0warnings;frontendbuild,boundaries andboth-themecontrast pass. Independentreviews no blockers. Productcommits c6a0254/7905513/be7491c; a37952a recorded docs/readiness beforethisfinalnote. MainfrontendHMR18080 andsyntheticreplypreview3002 remainrunning. Awaitownerlocalacceptance; no productionrelease.
<!-- SECTION:NOTES:END -->
