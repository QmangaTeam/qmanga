---
id: TASK-104
title: Mobile title reflow and single-Q wordmark in both themes
status: Review
assignee:
  - '@codex'
created_date: '2026-10-20 19:50'
updated_date: '2026-10-20 21:26'
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

Owner reference layout: replace separate horizontal facts and metadata cards with one information panel. Desktop under cover/actions; mobile below tabs before active content. Preserve title routes/actions/stats, actual metadata only, current palette. Verify320/360/768/885/960/1080/1440, both locales/themes, stats interaction and one metadata instance; independent review.

Owner refinement: horizontal native swipe/keyboard metadata below960px, desktop sidebar unchanged; description3line preview + localized expand/collapse button on all widths. Preserve44pxtargets and page reflow. Verify keyboardscroll and repeated expansion in both locales/themes, independent review.

Owner refinement: tabs noScroll links withoutfragment, sliding active indicator/content with reducedmotion optout. Compact horizontal facts rail preserving all fields; author/publisher/team credit chips below rail. Verify tab scrollposition and URL/back history pluscompact rail/actions/themes; independentreview.

Owner: remove chapter count only fromfacts,add fieldicons/statusdots and actualteamavatar withfallback. Populate existing synthetic local title with full clearlysynthetic metadata/ownedgeneratedartwork,longtext,multipleteams/avatar to review dense realshape. Keep realCBZ titles untouched; verify fallback,allfields,responsive,themes andindependentreview.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented mobile title heading→large centered cover→full-width Read/library→wrapping utility actions; facts reflow two columns and all tabs fit at320/360/487. Heading precedes cover in DOM; existing desktop composition retained. Mobile duplicate cover-banner fallback hidden; explicit artwork retained and failed-banner fallback follows actual selected source. ExistingQfavicon + manga. native theme-token text replaces duplicatedQ asset, legible dark/light/OSsystem. RED: newSSRwordmark test failed old2images; browser failed old100pxcover. GREEN: finalrealDocker9browserPASS (new8locale/theme matrix x3phonewidths+desktop and existingtitle/navigation);UI46PASS;Svelte0errors0warnings,architecture/contrastPASS. Independent mobile_logo_review found bannerfallback edge, fixed/re-reviewedPASS. Screenshots under ignoredtmp/task104/screenshots inspected at360dark/1280light. Full localDockerstack remainshealthy at http://localhost:18080; title /titles/synthetic-29e5596f. Restart fromrepo: docker compose --env-file tmp/task68/local.env -p qmanga-stage68-fixes up -d --wait. No secrets copied or volumes reset. User requests localvisualreview; no push/production rollout in this task yet. StatusReview pending owner localacceptance and subsequent authorisedrelease.

Owner follow-up: equal column spacing and aligned labels/values in title facts at885px; inspect row heights/negative link margins and ensure320/360/885/1280 both themes.

Follow-up: equal five-column facts grid with shared label/value rows; two equal columns on phones. Local Svelte check0errors0warnings;46UItests pass; Docker webbuild passes; read-only browser regression passes kk/ru × dark/light/system-dark/system-light at320/360/487 plus885/1280 spacing assertions. Independent review confirms alignment at320/360/768/885/960/1024/1280. Awaiting owner local acceptance, no push/release.

Owner follow-up: compact visible label/value spacing to4px while preserving44px link/button targets and equal columns. Scope presentation only, both themes and phone/885/desktop checks.

Compact follow-up verified: visible label/value CSSgap4px, label baselines aligned, links/statistics targets retain44px. Sveltecheck0errors0warnings,46UItests pass, Dockerwebbuild healthy; browser regression kk/ru × explicit/system dark/light at320/360/487/885/960/1280 passes. Independent review raised960px boundary coverage, now included and passing. Local browser screenshot tmp/task104/compact-facts.png. Await owner local review; no push/production release.

Owner requested frontend hot reload and removal of icon magic margin. Keep localhost18080 backend/data; private Compose dev override, source bind mount and gateway WebSocket forwarding. Align icon in value grid row. Verify actual HMR and themes.

Live dev frontend running on existinglocalhost18080 via private Compose override/sourcebind and local gateway WS forwarding; same backend and manga volumes. Actual HMR probe changed a temporary CSS custom property, observed browser update with0page reloads, restored source and observed restoration. Icon now explicit grid-column1/row2; dt/dd occupy column2 rows1/2, displaycontents inner wrapper, no marginoffset. Sveltecheck0errors0warnings;46UItests pass; livebrowser matrix kk/ru×explicit/systemlight/dark320–1280 passes incl960. Independent review passed. docs/LOCAL-DEVELOPMENT.md includes dev startup/revert. Localonly; awaits owner acceptance, no production release.

Reference-layout follow-up implemented locally: single TitleMetadata panel now contains kind, statuses, chapters, format, year where present, statistics, publishers/source/country/authors, aliases, links and teams. Removed separate facts strip and duplicate metadata rendering; no invented data/functions. Desktop>=960 sidebar below cover/actions; smaller screens below tabs,2columns phone/3tablet. Old obsolete facts CSS removed. Verified Svelte0errors0warnings,46UItests764assertions, architecture+both-theme contrast pass, production frontendbuild pass,2read-only browser regressions pass preserving chapter query/sort andstatsdialog. Matrix kk/ru × explicit/system dark/light at320/360/487/768/885/960/1080/1440; independent review confirmed44pxtargets,nooverflow,onepanel and placement. Minor nonblocking reviewer note: mobile keyboard sequence follows DOM metadata before tabs whereas visual tabs precede metadata. DevHMR remains active localhost18080; no productionpush/release. Await owner visual acceptance.

Final refinement: removed negative metadata action margins; label/control4px gap,44px targets and nooverflow rechecked independently320/360/885/960/1080. Final rerun0Svelteerrors/warnings,46UItests764assertions,2browserregressions pass. Screenshots under ignoredtmp/task104/unified-screenshots. Source dev mode remains live. Review pending owner visual acceptance.

Owner swipe/disclosure refinement complete locally. Below960px the entire combined metadata section is native horizontal flex rail with snap proximity, visible thin scrollbar and keyboard focus; desktop sidebar preserved. Description allwidths:3line preview, accessible44px button aria-expanded/controls, localized Толығырақ/Жию and Подробнее/Свернуть; resets on title change, no toggle on empty description. ActualCua expand/collapse observed; reviewer verified Enter/Space plus ArrowRight scrolling688px without pageoverflow. Final checks Svelte0errors0warnings,47UItests774assertions,2browserregressions pass incl mobile scroll and repeated description toggle320/688/1440 inkk/ru×explicit/systemthemes; architecture/contrast and frontendbuild pass. Early browser assertion raced hydration; final test waits hydration/networkidle and reactive state. Independent review no blockers. Screenshot tmp/task104/swipe-description.png. DevHMR remains running; no data or production changes. Await owner visual acceptance.

Owner compact-tabs/credits follow-up implemented locally: tablinks no fragment + data-sveltekit-noscroll;180ms slidingindicator and160ms directional content transition,disabled reducedmotion. Compact mobile facts use content-based88–156px widths,smallerpadding/font; all original fields retained. Authors/publishers/team labeled44px chips now OUTSIDE metadata container per latestownercomment, linkedteams preserved. Final checks Svelte0errors0warnings,48UItests784assertions,3browserregressions pass. New scroll regression checks About→Chapters andBack360/688/1280 withnormal/reducedmotion, no scrolljump and correct URL/activeindicator. Existing kk/ru explicit/systemthemes responsive/disclosure/filter/sort tests pass. Build/architecture/contrast pass. Independent reviewer confirmed607px exactpreservedscroll,creditsoutside/nooverflow andnodataloss. Screenshot tmp/task104/compact-tabs-credits.png. Minor test fixes wait for reactive backnavigation and allow Svelte-scoped animationname; final fullrun3pass0fail. Localdev running, no push/productionchange. Await owner acceptance.

Owner icons/avatar/dense-data follow-up: removed chapter fact only (tab/stats retained); all8populated fieldlabels haveicons, original/translationstatuses showdecorativecoloreddots withtext includingpaused/stopped enums. Teamchips use existing/media/avatars/team/id withlazyimage,errorfallback andtitle-resetstate. Horizontal scrollbar hidden inFirefox/WebKit whilekeyboard/swipe preserved. Dense data exposed absolute sr-onlyexternal-link text escapingrail; positionrelative onlinks fixespageoverflow. Longaliaslist nowhorizontalmobile, preservingallnames with~136pxrail. Checks finalSvelte0errors0warnings,49UItests789assertions, frontendbuild/architecture/contrast pass;3genericbrowserregressions pluslocaldenseacceptance pass onfullfixture, targeted dense rerunafteraliascompactness passes kk/ru dark/light widths320/360/730/960/1280, actualavatar+fallback,hiddenbarArrowRight,longtextexpansion. Independentreview confirmed nooverflow320/730/1280,allfields/statusdots/avatar andcompactrail. Denseverification script/assets/screenshots ignoredtmp/task104/dense-demo; sourceSSR regressions retained inCI. Localdevlive, no push/productionchange. Await ownervisualacceptance.
<!-- SECTION:NOTES:END -->
