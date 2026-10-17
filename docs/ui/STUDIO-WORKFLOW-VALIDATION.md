# Translator workflow repair: verification

17 October 2026. Baseline: `e6bbab2`. User acceptance exposed gaps that the previous functional checklist missed: a translator could not request a project or add its logical chapters without an administrator pre-populating them. The earlier claim that the full translator path was ready overstated what had actually been exercised.

## Delivered path

- `/menu` makes reader, translator and authorized management destinations discoverable. The mobile Menu destination now opens that menu; account preferences remain at `/settings`. Header login links preserve the original return target even from an authentication page.
- `/studio` opens projects; persistent Projects / Requests / Teams navigation is shared across the workspace.
- `/studio/teams`: actual team application, visible pending/rejected/approved state, owner correction and resubmission, invitations and membership management. Removal requires an explicit confirmation within the interface. Generated default team addresses remove a mandatory technical naming step.
- `/studio/requests`: choose an existing public title or propose a new one, select an eligible approved team, explain the plan, see history/decisions, cancel pending requests or correct and resubmit rejected ones. Five concurrent requests per team and twenty per day; duplicate pending targets are rejected.
- `/admin/requests`: approved-admin queue, rejection with reason, atomic title/project creation on approval, replay without duplicate projects. New titles remain private until normal publication. Decisions are visible to the team and create notifications.
- `/studio/projects/[id]`: owners/editors add logical chapters within their approved active project; existing logical chapters are reused only when supplied metadata agrees. Members retain upload permissions. Published/submitted versions are shown read-only; explicit correction opens a new revision, requiring chapter files again. File order changes remain staged and versioned.
- Preparation follows Files → Source/credits → Preview/submit. Unsaved attribution prevents submission. Processing results refresh the version list and do not retain a contradictory waiting message after readiness.
- Moderation uses a reason on the selected publication, shows successful publication with a reading link, and accepts a same-site public chapter link for withdrawal instead of asking users to find an internal identifier.

## Executed browser journey

At 360px, using existing local test roles:

1. Owner submitted a new-title request, saw pending history.
2. Admin approved it through `/admin/requests`; owner saw approval and opened the issued project.
3. Owner created chapter 1 without catalogue-admin access and uploaded a synthetic 800×1200 PNG generated specifically for this test, containing no manga artwork.
4. Worker completed the WebP reading variants. Source/credits for that own synthetic material were saved, preview decoded at 800×1200, and the revision was submitted.
5. Admin published it; the ordinary reader URL decoded the image successfully.
6. Owner submitted a separate team application. Admin returned it with a reason; owner saw that exact reason, corrected the description, and resubmitted. The test application remains pending for manual testing.
7. The published editor state showed source/credits and a dedicated correction action. That action opened a separate empty draft; the published revision stayed selected for ordinary reading. Empty/unfinished/unsaved drafts do not expose an active preview action.

The local project is named **Студия сынағы · QA**. It is a clearly labelled synthetic test, not a claim of a real authorized manga catalogue. Browser actions used the local application only. No credentials, cookies or private sample artwork are included in this repository.

Russian menu/projects/editor were inspected at 320px, Kazakh requests at 360px, and editor at 1280px. No horizontal overflow was observed at 320px; visible main editor controls met 44px height (native checkboxes are inside larger labels). Screenshots are outside Git in the sibling `outputs/ui-review` directory. This is desktop emulation, not a physical Android test.

## Automated and independent checks

- Full `go test ./... -count=1`, including PostgreSQL and real image processing, passed.
- Domain/integration checks cover request validation, team/role isolation, private history, approval replay, rejection/cancellation, chapter creation/reuse, and team correction/resubmission.
- A read-only review found stale team approval overwriting rejection, and chapter reuse silently ignoring supplied order/name. Both were reproduced by failing tests and fixed; the relevant suites passed afterward. The final review reported no additional material findings in requests, menu, moderation or link parsing.
- Frontend checks: 39 tests / 191 assertions; 13 UI tests / 578 assertions; Svelte zero errors/warnings; architecture, theme contrast, preview checks and Docker builds passed. Local services are healthy.

Production deployment, hosting, external mail delivery and physical-device/language/content acceptance remain outside this repair. This pass fixes the connected translator workflow and shared navigation; it does not certify every screen's UX as finished.
