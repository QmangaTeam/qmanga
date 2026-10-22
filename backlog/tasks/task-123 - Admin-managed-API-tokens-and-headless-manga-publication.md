---
id: TASK-123
title: Admin-managed API tokens and headless manga publication
status: Done
assignee:
  - codex
created_date: '2026-10-22 19:06'
updated_date: '2026-10-22 19:48'
labels: []
dependencies: []
ordinal: 65005
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Headless manga upload through admin-managed owner-bound API tokens, with explicit direct publication scope and ordinary moderation preserved. Owner authorized push/deployment and requested a short Codex guide.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Administrator creates and revokes own named tokens; secret displayed only once, database stores hash, expires and current owner role/block status enforced.
- [x] #2 Bearer import can create titles/chapters and upload prepared pages under owner identity; allowed routes/scopes deny unrelated administration and credential management.
- [x] #3 Explicit direct-publish scope publishes ready valid revision without moderation queue, atomically audited under owner/token; ordinary submissions remain moderated.
- [x] #4 Documented CLI uploads folder or CBZ, resumes interrupted pages without duplicate revisions and reports processing and publication failures truthfully.
- [x] #5 Real PostgreSQL/HTTP and local browser tests plus independent review verify tokens, revocation, privilege loss and end-to-end publishing.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1.Extend existing identity flow with hashed owner-bound admin API tokens, narrow import route access, expiry/revocation and live role checks.2.Add atomic explicit direct publication for token scope reusing domain/media validation and audit/outbox.3.Admin UI create/list/revoke once-only secrets within existing workspace.4.Headless upload CLI/manifest with folder/CBZ ordering, stable keys and resumability, documentation.5.Integrate real Docker/PG/HTTP/media end-to-end and browser regressions; independent review; local Review status, no push/deploy unless requested.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented owner-bound hashed admin API tokens with expiry/revocation/live role checks, scoped bearer imports, atomic privileged publication and cover selection, and once-only admin token UI. Upload CLI supports folder/CBZ, stable page keys, private resumable state and truthful publication checks. Short guide: docs/CODEX-UPLOAD.md.

Verified locally: full Go/PostgreSQL/codecs; 168 web unit tests, 64 UI tests, 11 Python regressions; production build; svelte-check zero errors/warnings; architecture and contrast checks. Real Docker browser flow verifies token creation, headless title/project/chapter import, image processing, cover selection, direct publication, idempotent repeat, scope denial, revocation and anonymous image access. Twelve admin views cover kk/ru, dark/light and 320/360/1280 widths. Independent backend/UI/publication/client reviews completed without remaining blockers.

Owner authorized push and deployment. Candidate b7c5578403a7bf0bce27a4860bbc22accb68c773: Verify39840064775 and FastRelease39840064908 succeeded, including 120 browser cases (54+66), zero failures and ARM64 images. Staging runs exact SHA; automatic production promotion pending. No workflow privilege increase or deployment gate bypass.

Explicitly authorized production maintenance completed after fresh off-host verified backup, rollback dry-run and independent SQL review: removed four test accounts, two unwanted titles and three unwanted teams. Retained four accounts (three exceptions and one recent), Haimiya with 12 published chapters, and Qmanga. Team ownership transferred to protected existing administrator. All 206 retained page IDs/checksums/statuses and 12 publication pointers unchanged; public chapter count and sample image200 verified, deleted title routes404. No personal identifiers or secrets recorded here.

Production automatically promoted to blue b7c5578403a7bf0bce27a4860bbc22accb68c773 at 2026-10-22 19:47 UTC after mandatory migration backup/off-host verification. External HTTPS homepage confirms exact SHA; PostgreSQL/S3 health ok; token API guest401 and admin page login redirect verified. Catalog retains exactly Haimiya with12 chapters; Telegram notifier enabled. No owner token created; owner creates secret once in admin panel.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Released admin-managed owner API tokens, scoped headless uploads and direct publication on production b7c5578403a7bf0bce27a4860bbc22accb68c773. Full CI and independent reviews passed;120 remote browser cases. Short guide docs/CODEX-UPLOAD.md. Authorized test-data cleanup completed and verified:4 users,1 title/12 chapters,1 team.
<!-- SECTION:FINAL_SUMMARY:END -->
