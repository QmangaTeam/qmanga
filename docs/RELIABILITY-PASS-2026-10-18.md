# Qmanga reliability pass — 18 October 2026

User-authorized scope: finish the Senkuro title page first, then implement the actionable recommendations from the supplied GPT review, thoroughly without unnecessary complexity. The review describes `58ce043`; current work starts from `ef14809` plus verified local fixes and the title page.

## Completed starting point

- Library/project error/retry, concurrent editing and pending alias results corrected.
- Title page: responsive cover/actions/sidebar, one facts panel, URL-backed sections, publication metadata editable through admin and persisted in PostgreSQL.
- Owner explicitly replaced green with Senkuro orange/dark palette. Dark samples: `#101014`, `#231A17`, `#FF873D`. Light counterpart retains readable contrast.
- Title-pass verification: 60 Go test/subtest passes, no actual skips; 61 frontend tests; 25 SSR UI tests; app/preview diagnostics zero; architecture/contrast/build pass. Desktop 1280, kk dark 320 and ru light 360 inspected; no document overflow. Guest library action, chapters, comments and statistics remain accessible. No physical-device claim.

## Execution plan and ownership

### A — Publication workflow (publishing agent)
- [x] Independent cover lifecycle: publication must not silently replace a chosen title cover; explicit authorized management and safe derived images.
- [x] Page thumbnails, original filenames, staged positions, clear failures/retry.
- [x] New revision clones published pages/assets/credits; replacing one page preserves current publication until acceptance.
- [x] Coherent version handling, save feedback, navigation protection and retryable upload queue.
- [x] Separate release availability and revision state; authorized restoration with reason/audit.
- [x] Preparation method and human review metadata; truthful public display.
- [x] Documented idempotent import of prepared pages through existing authorized interfaces; defer ZIP/OCR editor.

### B — Offline reading (offline agent)
- [x] Cold offline entry opens a list of complete saved chapters, not an arbitrary last chapter.
- [x] Two saved chapters selectable; missing/evicted/incomplete files never shown as complete.
- [x] Recoverable download failures and safe service-worker updates, with regression tests.

### C — Moderation/account audit (moderation agent)
- [x] Reports contain title/team/comment/page context and useful links; technical IDs secondary.
- [x] Pagination cursor independent of removed/resolved rows.
- [x] Audit shows actor, target, before/after, reason, time and relevant destination, with authorized visibility.

### D — Integration, reader/team usability and CI (primary)
- [x] CI separates Docker diagnostics from Go JSON and fails honestly on tests/skips/no events.
- [x] Resume/guest-cloud conflicts show concrete positions without automatically choosing furthest progress.
- [x] Mobile labels avoid word splitting; catalogue shows active conditions and reset.
- [x] Team portfolio: structured Telegram/public link, clear solo-team wording, useful publish/share copy; invitation lifecycle if currently incomplete.
- [x] Reader settings distinguish frequent and advanced options.
- [x] Practical performance/bandwidth evidence and repeatable pilot instructions.
- [x] Integrated test/build, fresh review and one corrective pass; update state and report remaining limitations accurately.

## Constraints and rulings

- Preserve modular domain/application/adapter boundaries, access controls, immutable public revisions, existing data, kk/ru and both themes. Never expose original private image objects or secrets.
- Work in the current owner-selected checkout; preserve pre-existing changes. Shared files require coordination. Reserve migration 000019 for publishing, 000020 for moderation, 000021+ for primary. 000018 already adds title publication metadata.
- No concurrent full Docker builds or service restarts; primary coordinates final builds. Agents may run focused tests, not destructive cleanup.
- Ruling: human research, physical-device testing, translation-quality judgement and inviting external teams cannot be honestly completed by code. Prepare a reproducible pilot and record these as unperformed, without fabricating results or sending unsolicited messages.
- Ruling: do not add speculative team subscriptions, ownership transfer, duplicate merging, ZIP import, messaging, social feeds, OCR/typesetting editor, native apps or microservices: the supplied review explicitly defers or conditions them on user research.
- Ruling: no production deployment, purchases or destructive volume operations. Local integration/acceptance is authorized; any remote CI result must be observed, never inferred from local tests.

## Evidence ledger

Title pass logs: `tmp/title-full-go.log`, `tmp/title-web-final.log` (ignored). Additional task evidence will be appended as results arrive.

## Completed verification and remaining limits

- Final full backend run: `tmp/reliability-go-release.jsonl`, 66 test/subtest passes, 21 packages, no actual skips. PostgreSQL and real image codec tests executed.
- Integrated auth/SMTP/password reset, publishing/processing/moderation and reading/community scripts passed: `tmp/reliability-auth-smoke.log`, `tmp/reliability-publishing-smoke.log`, `tmp/reliability-reading-smoke.log`.
- Publishing fixture: 45 synthetic pages, three actual WebP tiers. Browser correction cloned all 45 pages; replacement of page 23 preserved the public revision and its original page. Explicit cover selection and unsaved-edit navigation protection worked.
- Browser cold offline acceptance: downloaded two chapters, stopped the local gateway, closed tabs, reopened `/saved`, read both chapters and reloaded page 23. Removed one cached image: incomplete chapter lost its read action; restarted gateway and resumed successfully. This found and fixed a real Svelte proxy/IndexedDB structured-clone error. Cache verification then passed. Service-worker update waited safely for old tabs to close.
- Browser moderation acceptance: page-23 report displayed title/team/chapter/page, opened the exact preview position, resolved with a reason and appeared in audit. Follow-up review fixed localized cover audit states and unnamed actor links.
- Responsive inspection: kk dark at 360 px, ru light at 320 px and desktop at 1280 px; no document overflow in the inspected title views. Screenshots: `tmp/reliability-screenshots/` (local evidence, ignored).
- Independent cross-review covered reader position races, studio polling versus dirty edits, report pagination, safe audit fields and offline state handling. Corrections are included.
- Browser regression file `apps/web/tests/browser/offline.test.mjs` was syntax checked; local acceptance used the managed browser. The complete browser runner is a separate CI gate, not claimed to have run locally.

Scope limits: cover selection uses a ready project image and an independent safe derivative snapshot; dedicated file upload/cropping is not implemented. Upload-first replacement needs one free page slot and temporary byte quota; at the 300-page/500-MiB limit an explicit draft edit is required. File retry survives the current visit; after reload files must be selected again. Team avatar uses initials, not an image uploader. Historical absent audit fields cannot be reconstructed; complaint text reflects current content. Human review is an attributed team attestation, not proof of translation quality.

Physical-device endurance, representative mobile performance, human usability and Kazakh translation review remain unperformed pilot gates. No production deployment or external invitations were performed. See `PILOT-AND-PERFORMANCE-2026-10-18.md` and `PREPARED-PUBLICATION-IMPORT.md`.
Final production build: `tmp/reliability-final-build.log` — 84 frontend tests (380 assertions), 26 SSR UI tests (650 assertions), zero failures; app and preview each zero errors/warnings; architecture and both-theme contrast checks pass. API, worker and web rebuilt; all seven local services healthy. Final browser recheck confirmed cover audit before/after and actor links, decoded offline page 23 from blob storage, preserved online `auto` quality, and offline chapter panel linking to saved chapters without online directory errors.

## Remote CI confirmation

Code commit `eef58597922ea5a4d4043bb2cb52de4eee988d82` passed the complete GitHub Actions run [39082237199](https://github.com/maqsatto/qmanga/actions/runs/39082237199): clean build, setup checks, full Go/PostgreSQL/codec suite, auth/publishing/reading smoke, and all `tests/browser/*.test.mjs` including the new cold-offline/eviction/retry regression. The earlier JSON-log parsing failure no longer blocks downstream tests. This remote result was observed as `completed / success`, not inferred from local checks.
