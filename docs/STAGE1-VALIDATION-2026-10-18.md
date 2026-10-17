# Stage 1 first package — evidence and remaining gates

Date: 18 October 2026. Baseline: `46737a0`. Work is in the owner's requested new clone, not the original Codex output directory. Scope is the first package in [the supplied development plan](DEVELOPMENT-PLAN-2026-10-18.md), not all 35 tasks or a claim of complete beta/public-launch readiness.

## Changes and acceptance

| Task | Implemented / observed | Remaining boundary |
| --- | --- | --- |
| T01 | Shared API/worker build and project-scoped runtime image. Config checks pass for two different project names. New `qmanga-stage1-20261018` project started through the documented setup/up flow; all seven services healthy. Inspect showed identical API/worker runtime image IDs. | Existing BuildKit base/layer cache was deliberately retained. A wholly empty remote runner/cache has not yet been observed. |
| T02 | Old ready-page submission reproduced 409. Updated smoke first expects that negative result, writes attribution, reloads version, submits/moderates and verifies WebP MIME, file signature, bytes and checksums through Go for all three profiles. Final run used three freshly generated geometric pages and explicit rights confirmation. | These are synthetic fixtures, not an editorial/content-rights acceptance exercise. |
| T03 | Admin catalogue link opens an admin-only read-only project card. Actual browser navigation succeeds for an admin outside the team; direct studio API remains 403. Database integration exercises owner/editor/member/admin/moderator/blocked admin/foreign user and a missing project. Member cannot create a chapter. | Directly entering a forbidden studio URL still has pre-existing form remnants after the error; this is not the admin navigation path and grants no authority. Broader workspace UX remains in later tasks. |
| T04 | Opening/selecting does not show Saved. Confirmed write is checked against actual persisted metadata. HTTP 409 retains entered data and no prior success. Reviewer found native validation could block submit while leaving old Saved: reproduced RED, fixed by clearing success on form input, then GREEN. Selection commits its target only after its read succeeds. | General dirty-navigation protection and independent form composition remain later backlog. |
| T05 | Two browser regressions use a fresh profile and real API authentication; both pass, no skips. Added PR/main/manual CI for build, real PostgreSQL/codecs, auth/publish/read and staff browser transitions. Missing browser fixture fails rather than skips. | Remote Actions success is not claimed here. This is a starter suite, not yet the complete reader/offline acceptance matrix. |
| T06 | One local/recovery guide; tested PowerShell and shell setup safeguards, preservation of existing env, Docker inspection failure and project isolation. Real default setup refused to create a missing env over the existing original project's volumes. | No independent second-person onboarding session or recovery of the owner's unknown credentials was performed. |

## Fresh checks

- Full container Go suite: **56 test/subtest pass events**, **0 actual test skips**, **0 failure events**; 14 packages have no test files. Go 1.26 represents those packages as skip events without `Test`, so CI reports them separately rather than pretending they are skipped integration tests.
- Web image after the review fix: 41 application/domain/adapter tests; 23 SSR UI contracts; application and preview Svelte checks each 0 errors / 0 warnings; boundaries, contrast and production build successful.
- Real local auth, publishing and reading smoke all pass against the same isolated project, env, app port 18080 and Mailpit port 18025. Publishing and reading bootstrap now use that same Compose context.
- Browser: 2/2 pass after the review fix. The final run includes the native-validation case in addition to the HTTP-rejection case. Optional screenshot pass exercised the admin card at 320/360/1280, kk/ru and both system color schemes with reduced motion; no document overflow and card actions at least 44px.
- Screenshots inspected: Russian/light at 320 and 1280; Kazakh/dark at 360. The existing narrow Russian bottom-navigation wrapping is visible and is not presented as fixed. These are desktop browser emulations, not physical Android/iOS testing.
- One fresh read-only reviewer checked the package. CI skip classification and the native-validation finding were fixed; no further Critical/Important defect was reported. No second review or delegated implementation.

Local logs, env, passwords, generated images and screenshots remain ignored under `tmp/stage1/`; none belong in Git. Do not distribute fixture JSON files: they contain disposable local account credentials. The original project's data and the separate audit environment were not restarted, overwritten or removed. The verification project is stopped after acceptance; its volumes are retained.

## Decisions and their limits

1. Keep the existing main branch in the explicitly requested new clone; no extra branch/worktree or duplicate implementation session. This prioritizes the owner's working location over isolation via another checkout.
2. Staff project context is **admin-only read-only**, not a new global editing right. Moderators retain the existing moderation workflow. If future staff need another task, design that permission explicitly rather than inheriting team editor rights.
3. Pass project/env into both smoke tools. Otherwise a sandbox test can bootstrap an admin in the wrong environment or fail despite a healthy sandbox.
4. Require explicit test-operator rights confirmation and truthful attribution. This adds a flag to the smoke invocation; it avoids silently claiming rights over arbitrary files. Generated geometric pages need no third-party artwork.
5. Keep later pagination, unsaved-navigation, reader/offline and moderation tasks outside this first package. Those are real remaining beta limitations, not evidence that this package is a finished MVP.
6. Production readiness, backups, physical-device performance, an entirely empty layer cache and remote CI success remain unproven. Local tests do not replace those gates. No production configuration or global WSL/Docker resource policy changed.
