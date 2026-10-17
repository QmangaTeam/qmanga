# Qmanga reliability pass — 18 October 2026

User-authorized scope: finish the Senkuro title page first, then implement the actionable recommendations from the supplied GPT review, thoroughly without unnecessary complexity. The review describes `58ce043`; current work starts from `ef14809` plus verified local fixes and the title page.

## Completed starting point

- Library/project error/retry, concurrent editing and pending alias results corrected.
- Title page: responsive cover/actions/sidebar, one facts panel, URL-backed sections, publication metadata editable through admin and persisted in PostgreSQL.
- Owner explicitly replaced green with Senkuro orange/dark palette. Dark samples: `#101014`, `#231A17`, `#FF873D`. Light counterpart retains readable contrast.
- Title-pass verification: 60 Go test/subtest passes, no actual skips; 61 frontend tests; 25 SSR UI tests; app/preview diagnostics zero; architecture/contrast/build pass. Desktop 1280, kk dark 320 and ru light 360 inspected; no document overflow. Guest library action, chapters, comments and statistics remain accessible. No physical-device claim.

## Execution plan and ownership

### A — Publication workflow (publishing agent)
- [ ] Independent cover lifecycle: publication must not silently replace a chosen title cover; explicit authorized management and safe derived images.
- [ ] Page thumbnails, original filenames, staged positions, clear failures/retry.
- [ ] New revision clones published pages/assets/credits; replacing one page preserves current publication until acceptance.
- [ ] Coherent version handling, save feedback, navigation protection and retryable upload queue.
- [ ] Separate release availability and revision state; authorized restoration with reason/audit.
- [ ] Preparation method and human review metadata; truthful public display.
- [ ] Documented idempotent import of prepared pages through existing authorized interfaces; defer ZIP/OCR editor.

### B — Offline reading (offline agent)
- [ ] Cold offline entry opens a list of complete saved chapters, not an arbitrary last chapter.
- [ ] Two saved chapters selectable; missing/evicted/incomplete files never shown as complete.
- [ ] Recoverable download failures and safe service-worker updates, with regression tests.

### C — Moderation/account audit (moderation agent)
- [ ] Reports contain title/team/comment/page context and useful links; technical IDs secondary.
- [ ] Pagination cursor independent of removed/resolved rows.
- [ ] Audit shows actor, target, before/after, reason, time and relevant destination, with authorized visibility.

### D — Integration, reader/team usability and CI (primary)
- [ ] CI separates Docker diagnostics from Go JSON and fails honestly on tests/skips/no events.
- [ ] Resume/guest-cloud conflicts show concrete positions without automatically choosing furthest progress.
- [ ] Mobile labels avoid word splitting; catalogue shows active conditions and reset.
- [ ] Team portfolio: structured Telegram/public link, clear solo-team wording, useful publish/share copy; invitation lifecycle if currently incomplete.
- [ ] Reader settings distinguish frequent and advanced options.
- [ ] Practical performance/bandwidth evidence and repeatable pilot instructions.
- [ ] Integrated test/build, fresh review and one corrective pass; update state and report remaining limitations accurately.

## Constraints and rulings

- Preserve modular domain/application/adapter boundaries, access controls, immutable public revisions, existing data, kk/ru and both themes. Never expose original private image objects or secrets.
- Work in the current owner-selected checkout; preserve pre-existing changes. Shared files require coordination. Reserve migration 000019 for publishing, 000020 for moderation, 000021+ for primary. 000018 already adds title publication metadata.
- No concurrent full Docker builds or service restarts; primary coordinates final builds. Agents may run focused tests, not destructive cleanup.
- Ruling: human research, physical-device testing, translation-quality judgement and inviting external teams cannot be honestly completed by code. Prepare a reproducible pilot and record these as unperformed, without fabricating results or sending unsolicited messages.
- Ruling: do not add speculative team subscriptions, ownership transfer, duplicate merging, ZIP import, messaging, social feeds, OCR/typesetting editor, native apps or microservices: the supplied review explicitly defers or conditions them on user research.
- Ruling: no production deployment, purchases or destructive volume operations. Local integration/acceptance is authorized; any remote CI result must be observed, never inferred from local tests.

## Evidence ledger

Title pass logs: `tmp/title-full-go.log`, `tmp/title-web-final.log` (ignored). Additional task evidence will be appended as results arrive.
