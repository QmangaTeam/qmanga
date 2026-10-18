# Pilot and performance acceptance

This is an executable acceptance plan, not evidence that interviews or physical-device tests have happened. The current development fixtures are synthetic geometry; neither their compression ratio nor an automated pass establishes translation quality or production readiness.

## One-hour pilot

Use two readers (including one guest), one translator and one moderator. Use a permitted chapter with 30–45 pages plus a second chapter. Do not provide task-by-task hints: record where the person hesitates, expected result, actual result and completion time. Get permission before recording a session. Do not collect passwords or personal correspondence.

1. Reader: find a title, interpret publication/translation status, select a team, read to page 12, close and reopen. Explain the displayed continuation before opening it.
2. Reader: create a different position in another browser, then choose a cloud/device position using chapter, page, team and timestamp. Re-reading an earlier chapter is legitimate; never choose the furthest progress automatically.
3. Guest: read without logging in; sign in, inspect both positions, explicitly import or retain the cloud position. Verify another account cannot see the first account's local queue.
4. Offline: save two chapters, close every site tab, disable networking, open `/saved` directly, select each chapter and reload. Evict one cached image and verify the chapter needs recovery rather than reporting complete. Restore networking and retry.
5. Translator: create a solo team or use an existing team, invite a collaborator, inspect pending/expired invitation state, revoke one invitation. Prepare pages, inspect filenames/thumbnails and positions, reorder, retry one failed upload, leave with unsaved edits and recover safely.
6. Translator: correct page 23 of an existing 45-page chapter. Verify the published version remains available until the correction is accepted; unchanged pages and credits survive. Set a title cover and publish another chapter without changing the selected cover.
7. Moderator: open a page/comment complaint, identify title/team/chapter/page/author without interpreting UUIDs, follow its context link, review earlier decisions, hide and restore with reasons. Finish a full queue page and continue to the next one. Verify actor/target/before/after in audit.

Record each task as completed unaided / completed with help / failed. Prioritize data loss, wrong publication, inaccessible reading and account leakage before cosmetic issues. Run a two-week pilot with 2–3 teams only after these tasks pass. Invitations to external participants are a human action; no invitations were sent by this implementation pass.

## Performance measurements

Run `./scripts/measure-reader.ps1 -ReleaseId <published-id> -AppUrl http://localhost:18080`. The output includes manifest HTTP time, all chapter bytes, per-tier first-fragment HTTP samples, checksum validation, transfer-only estimates at 1.6/9/30 Mbps, and GiB per 1,000 complete chapter reads. These are local measurements and mathematical estimates, not real mobile network results or browser paint times.

For a browser run, use a clean cache and record viewport, network latency/throughput, CPU throttling and the exact chapter. Observe the first visible decoded image, request counts, bytes, failed-image retry, page navigation and a quality switch. Repeat warm-cache and offline runs. Do not label `load`, HTTP completion or a DOM placeholder as first useful image.

For a physical Android/iOS test: read a representative long chapter for 20 minutes, switch apps for 2 minutes, resume, change orientation, move between chapters and reopen offline. Record crashes/reloads, lost position and memory from the device's supported tooling. The scheduler's decoded-image budget is an allocation policy, **not measured total browser RAM**.

Cost calculation: monthly egress = measured chapter bytes × complete-read equivalents, plus thumbnails/manifests/retries; billable storage = originals + derived tiers + retained revisions + backups. Multiply by the actual provider's current tariff when a provider is selected. No invented hosting bill or vendor pricing is supplied.

## Release gates

- Full Go suite and actual PostgreSQL/codec tests pass with no hidden test skips.
- Web application/unit/UI checks, Svelte diagnostics, architecture and contrast checks pass; production build succeeds.
- Auth → upload/processing → moderation/publication → library/progress/comment/report acceptance passes on the integrated runtime.
- Cold two-chapter offline flow, continuation conflicts and staff context links pass in the browser.
- Physical-device endurance, human usability and Kazakh translation review remain explicit human pilot gates, not automatically checked boxes.

## Observed local sample, 18 October 2026

45 synthetic pages; manifest 40,838 UTF-8 bytes, first loopback request 2,074.67 ms (includes client/connection setup). Whole-chapter derivative bytes: eco 111,570; standard 126,210; high 118,830. Three first-fragment HTTP samples: eco 18.07/6.35/6.00 ms, standard 6.00/6.08/6.25 ms, high 5.22/4.56/4.37 ms. At 1,000 complete reads the image payload is 0.1039/0.1175/0.1107 GiB respectively, excluding other traffic. Repeated geometric images explain both the tiny files and non-monotonic tier sizes; these numbers are not a production bandwidth forecast or browser first-image benchmark. Raw local result: `tmp/reliability-reader-measurement.json`.
