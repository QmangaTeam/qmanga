---
id: TASK-121
title: Telegram notifications for all moderation submissions
status: In Progress
assignee:
  - codex
created_date: '2026-10-22 12:49'
updated_date: '2026-10-22 12:57'
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
- [ ] #1 Every moderation submission is durably queued and delivered with its admin link; retries survive restart without blocking product writes.
- [ ] #2 Bot runs in Docker with secrets outside Git, production-only recipient configuration, tests and independent review.
- [ ] #3 All authorized changes released via normal pipeline and exact production SHA plus real Telegram delivery verified.
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation: reuse Docker worker lifecycle; separate transactional moderation_telegram_queue and five source triggers/backfill. Bounded plain-text messages and explicit queue-root versus direct-review links; retry30s exponential up to1h, at-least-once semantics,1message/sec. Token/chat stored only in private local and production config; staging blank. Real Telegram setup message accepted. Full Go suite including PostgreSQL passed; focused queue source/rollback/resubmit/retry/no-replay test passed; sender tests pass. Compose local/production interpolation quiet checks pass. Independent review resolved misleading paginated deep links by explicit queue buttons; re-review no findings. Required CI-gate stale browser selectors repaired separately and all4 failing scenarios pass locally. Next: commit, push, await full Verify and Fast Release, verify exact production SHA and real durable delivery. No secret values in notes.
<!-- SECTION:NOTES:END -->
