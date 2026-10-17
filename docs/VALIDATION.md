# Local functional validation

The code is a local functional alpha, not a declaration of production or physical-device readiness.

## Repeatable checks

From the repository root:

```powershell
docker compose --profile tools run --rm go-tests
docker compose build api web
docker compose up -d --wait
./scripts/smoke-auth.ps1
./scripts/smoke-publishing.ps1 -SourceDirectory '<local authorized image directory>' -PageLimit 5
./scripts/smoke-reading.ps1 -ReleaseId '<release UUID printed by publishing smoke>'
```

Publishing smoke is localhost-only. It streams files through Go, creates random local test identities/team/title, checks draft privacy, waits for the actual worker and publishes through moderation. Reading smoke tests real cookies/CSRF, library separation, event replay/conflicts/account mismatch, comments/ratings/reports, outbox delivery and public reader HTML privacy. Credentials are random and not saved in source or printed. Test identities and local published test content remain in the development database. Do not run against production or share the local sample content publicly.

Use `go -C backend test ./...` for a fast host check, but it skips PostgreSQL without `TEST_DATABASE_URL` and may skip codec execution if ImageMagick is absent. The tools container exercises real PostgreSQL in isolated schemas and includes the actual JPEG/WebP modules. Web image builds run dependency boundaries, Bun tests, Svelte diagnostics and production compilation.

## Reading workflows

- `/catalog`: find a published title; `/admin/catalog`: manage metadata, chapters and approved translation projects.
- `/studio/projects`: naturally ordered uploads, progress/cancellation, retry with the same files, page ordering, draft preview and submission.
- `/moderation`: preview/reject/publish revisions or withdraw a release; `/moderation/reports`: reasons, visibility decisions and resolution.
- `/read/<release>`: read without an account. The drawer contains quality, mode, fit, theme, direction, size and explicit original/download actions. Chapter selection stays with the selected translation project.
- `/library`: statuses, independent favorites, unread counts, continuation and explicit guest-progress import.
- `/saved`: inspect bytes/readiness, verify and remove selected copies. `/offline-reader?release=<release>` uses the saved manifest and verified cache-only image adapter.
- `/notifications`: new releases and replies; no private inbox or account HTML is included in offline caches.

## Regression evidence

Go tests exercise tenant isolation, revision replacement/withdrawal, upload limits, damaged images, actual JPEG encoding, leased-job fencing, conditional/range media, event replay, completion, cloud conflicts, comment ownership/depth, rating updates, report moderation and deduplicated notification delivery. Frontend tests exercise source anchors, bounded/cancelled image requests, continuous-scroll throttling, account scopes, explicit imports and partial/corrupted/off-revision saved files.

A fresh-context review found and fixed: a duplicate retry rebasing over another device after a lost response; continuous scrolling indefinitely postponing persistence; saved copies still depending on a successful network response; and guest import resetting existing library choices. Dedicated regressions cover each fix. Architecture checks reject cross-context private frontend adapter imports; identity scope is injected and the shared network gate is technical infrastructure, not part of the reader domain.

## Browser checks performed

The local five-page JPEG chapter opens through the Go manifest/media gateway. At320px and360px, the reader and drawers have no horizontal overflow; inspected buttons are44px and selects48px. Light/dark panels, page turns, immediate page-turn persistence across reload, strip mode, eco/high switching, source-page counter and Back closing a drawer were exercised. At1280px the centered stage measures1024px and only a small image window is rendered. A saved eco copy opened, turned to another decoded page and reloaded with the tab's network disabled through browser developer controls; the network was restored afterwards. This is desktop browser emulation, not a physical Android test. Local screenshot proofs are kept outside the Git repository, so sample artwork is not distributed with the code.

## Reader navigation and codec regression follow-up

The owner reported right-edge taps, buttons and arrow keys not advancing. Reproduced missing image tap targets, arrow suppression after button focus, strip mode inheriting title RTL, and a fractional CSS-pixel boundary that left the first visible page/counter on the previous source page. Navigation now has a shared application cursor with latest-request fencing; rounding within1px is accepted at boundaries. Disabled edge controls identify chapter limits. Page-mode RTL remains explicit/title-driven and consistent across all inputs.

Fresh checks:22 Bun tests/113 assertions; frontend boundaries, Svelte0 errors/warnings and production build; full Go/PostgreSQL/real-codec suite. Real JPEG/PNG/WebP inputs produce decodable WebP in all reading profiles and PNG originals; MIME, dimensions, actual byte counts and hashes are checked. A PNG masquerading as WebP is rejected. A fresh one-page upload from the owner's local test folder passed upload → worker → publication → actual RIFF/WEBP bytes, image/webp HTTP headers and SHA-256 for every eco/standard/high fragment through Go.

Browser: the original five-page chapter was exercised at320/360px in both themes, in pages and strip modes. Tested a physical right-edge coordinate click, toolbar previous/next, ArrowLeft/ArrowRight after button focus, two rapid turns from1/5 to3/5, end-of-chapter disabled controls and RTL's consistent reversed sides. No horizontal overflow at320/360px; mobile toolbar targets remain44px. These are desktop-browser checks, not physical Android acceptance.

## Before a public launch

- Test a physical weak Android, real network interruptions, long webtoons, app backgrounding, quota/eviction and service-worker updates across browser restarts. Emulation does not establish device performance.
- Obtain authorized content and have Kazakh UI/content terminology reviewed. No copyrighted sample pages are distributed with the repository.
- Finish the deferred general visual pass. Studio/admin lists are bounded; some do not yet expose cursor navigation beyond their initial limit.
- Define source/orphan retention and garbage collection, database plus both-bucket backup/restore, SMTP delivery monitoring and worker/error monitoring.
- Configure per-client rate limiting at the trusted public gateway before exposing it: the default private Docker API deliberately does not trust forwarded IP headers, so its authentication limiter currently sees the gateway address.
- Add the agreed deployment TLS/domain, security review, release/rollback procedure and CI execution. No public deployment, external mail delivery or production secrets are configured by this task.

Selected downloads have a soft200MiB ceiling and use no more than20% of estimated remaining storage. Automatic browser HTTP caching is browser-managed, not a separately implemented50MiB app LRU; only explicitly selected chapters are guaranteed eligible for offline reading. Decoded reader scheduling uses a40MiB pixel estimate, not a guarantee of total browser process memory. These operational/acceptance limits are not hidden by a passing build.
