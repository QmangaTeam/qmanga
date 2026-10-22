---
id: TASK-121
title: Telegram notifications for all moderation submissions
status: Done
assignee:
  - codex
created_date: '2026-10-22 12:49'
updated_date: '2026-10-22 13:25'
labels: []
dependencies: []
ordinal: 63005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Docker moderation notifier to owner's Telegram. Cover submitted chapter revisions, team applications/resubmissions, project requests, title metadata/cover changes and reports. Durable transactional queue, retry without blocking submissions, secret-only token/chat configuration, authenticated admin links, no review actions via Telegram. Integrate normal immutable-image staging/production lifecycle and verify delivery. Owner authorizes deployment of all current changes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every moderation submission is durably queued and delivered with its admin link; retries survive restart without blocking product writes.
- [x] #2 Bot runs in Docker with secrets outside Git, production-only recipient configuration, tests and independent review.
- [x] #3 All authorized changes released via normal pipeline and exact production SHA plus real Telegram delivery verified.
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation: reuse Docker worker lifecycle; separate transactional moderation_telegram_queue and five source triggers/backfill. Bounded plain-text messages and explicit queue-root versus direct-review links; retry30s exponential up to1h, at-least-once semantics,1message/sec. Token/chat stored only in private local and production config; staging blank. Real Telegram setup message accepted. Full Go suite including PostgreSQL passed; focused queue source/rollback/resubmit/retry/no-replay test passed; sender tests pass. Compose local/production interpolation quiet checks pass. Independent review resolved misleading paginated deep links by explicit queue buttons; re-review no findings. Required CI-gate stale browser selectors repaired separately and all4 failing scenarios pass locally. Next: commit, push, await full Verify and Fast Release, verify exact production SHA and real durable delivery. No secret values in notes.

Code c6bee6ad3d16896607d6ed9b4fc8335721e8301a pushed with selector074fc2e and CI-selector repairs36d7a85. Verify39782687712 and FastRelease39782687750 running through mandatory migration gate. Real Docker integration in isolated schema passed: inserted synthetic team -> trigger -> durable queue -> real Telegram acceptance -> delivered_at persisted; schema dropped. Owner received setup and clearly marked test messages. No production test data created. Production private env configured, no values logged. Final independent review no findings, including1/sec throttle and repaired browser selectors.

Observed release22Oct13:25UTC: staging and production blue c6bee6ad3d16896607d6ed9b4fc8335721e8301a. Verify39782687712 success:117 browser cases,0 failures, Go/PostgreSQL/codecs and all3 ARM64 image jobs success. FastRelease39782687750 success. Staging health/storage/Postgres/full homepage SHA pass; migration000036 applied, Telegram disabled. Production external HTTPS health and homepage exactSHA pass; active Docker worker logs notifier enabled. Production queue:1 queued,1 delivered,0 waiting (real pending submission delivered, no test production data created). Token/chat only private production config. Local isolated Docker real Telegram test also passed. Independent final review no findings. No user action remains for this bot release.
<!-- SECTION:NOTES:END -->
