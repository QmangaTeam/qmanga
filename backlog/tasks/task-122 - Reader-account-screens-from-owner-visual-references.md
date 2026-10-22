---
id: TASK-122
title: Reader account screens from owner visual references
status: Done
assignee:
  - codex
created_date: '2026-10-22 13:23'
updated_date: '2026-10-22 14:51'
labels: []
dependencies: []
ordinal: 64005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Owner22Oct screenshots: compact account menu, shared login/register card, timeline notifications, tabbed settings and profile presentation. Owner explicitly approved implementing missing functions: persisted bio/public profile/following/privacy, delivery preferences and working reader preferences. Existing accounts private by default; preserve real authentication, Google linking, avatars and account boundaries. Match orange accents and reference hierarchy,320/360px,kk/ru,both themes. New UI local review first; no new production authorization for this package.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Account menu and auth screens follow supplied layout while preserving authentication, safe return URLs, roles, avatars and keyboard interaction.
- [x] #2 Notifications and settings follow supplied timeline/tab layout using real supported state and honest unsupported states; responsive, localized and both themes verified.
- [x] #3 Agreed profile/new-setting scope implemented or explicitly deferred; independent review and local owner preview available.
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Owner explicitly approved missing functions via async response: full persisted public profile/bio/following/privacy and notification/reading settings, not only visual skin. Plan:1.finish shared auth/account presentation;2.identity profile/preferences/follow storage+guarded APIs (existing accounts private by default);3.public profile and settings tabs backed by real ports;4.notification preferences enforced on delivery and supported email digest, reader save/prefetch settings actually applied;5.browser responsive/localized/roles/concurrency checks and independent review. New package local only for owner preview. Telegram+selector prior release separately confirmed c6bee6a.

Local implementation complete at 2a4ef25/af26e07 and browser regression commit. Profiles, bio, privacy and follow APIs persisted with concurrency guards; private defaults and no public emails. Notification category preferences enforced, real weekly digest wired to configured mailer. UI menu/auth/settings/profile/timeline follow owner references. Remember=false uses session cookies and24h server TTL; true30days. Reading saveProgress/prefetchNext and reduced-motion persist and affect runtime. Independent backend and frontend reviews completed; digest public-visibility and frontend race/navigation findings fixed and re-reviewed clean. Checks: full Go suite with actual PostgreSQL/codecs PASS (tmp/account-go-final.log); identity/notification PostgreSQL matrix PASS;21 focused TS tests,62 UI tests/854 assertions PASS;4 targeted browser regressions PASS including registration/import/Google/authbranding/newprofile-flow and disabled-progress no POST;120 responsive rendered views kk/ru,dark/light,320/360/1280 pass; web production build PASS; svelte-check0errors0warnings; architecture/contrast PASS. Screenshots inspected; broken profile avatar now falls back to initials. New migrations000037/000038 applied in localDocker only. Local dev API/worker rebuilt and running; preview http://localhost:18080/settings?section=appearance. No push or deployment of TASK122; owner local acceptance pending. Production last observed release remains c6bee6ad3d16896607d6ed9b4fc8335721e8301a from prior bot/selector release. Next: owner inspect local account screens, then explicitly authorize publishing this package. Email digest verified using integration mailer; physical mobile devices and external real mailbox delivery for this package not exercised.

Owner authorized deployment22Oct. First fullCI exposed legacy settings/notification locators; corrected actual Google callback routing (e6c620b), updated avatar disclosure/tab regressions (96f5c8d), and destination-based notification locator (dfa6096), all independently reviewed; targeted local tests pass. Verify39794946707 success: browser63+55=118 pass0fail, full Go/PostgreSQL/codecs and ARM64 images. Same SHA dfa6096f9a1cbd196885b9ac024ae3277365d327 observed staging green and production green, public homepage SHA and HTTPS health Postgres/S3 ok. Production new API/web healthy and Telegram notifier enabled in green worker. Final read-only production responsive browser smoke running before closing task.

Final external production browser smoke PASS48 views: login/register/settings/notifications × kk/ru × dark/light ×320/360/1280; password reveal and required registration consent present,0horizontaloverflow0runtimeerrors. Public and server production SHA confirmed dfa6096f9a1cbd196885b9ac024ae3277365d327. Previous blue retained. Physical mobile devices and real weekly-mailbox delivery remain outside exercised checks.
<!-- SECTION:NOTES:END -->
