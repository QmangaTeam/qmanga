# Functional MVP gap audit

Baseline: f04c1b1. Scope: original QMANGA-PLAN / QMANGA-READER / QMANGA-ROADMAP, including the later Kazakh-first/Russian and backend-media amendments. User explicitly authorizes closing functional gaps and excludes production work.

The previous redesign checkpoint was not complete parity with the original product specification. This audit does not relabel missing functions as production tasks or substitute a green build for product acceptance.

## Existing core

Guest catalogue/search and reading; source-page navigation and quality tiers; local/cloud progress with explicit conflict resolution; saved-chapter integrity/offline reader; independent approved-team publishing and atomic revisions; private Go media gateway; comments/replies/spoilers/ratings/reports; accounts and Kazakh email flows are implemented. Physical-device/language/content acceptance remains separate.

## Execution checklist

- [ ] Publication decision notifications; deduplicate alternative translations of one logical chapter when no preferred team is selected.
- [ ] Real latest-release feed and explicit editorial picks, not an updated-title list presented as new chapters.
- [ ] Catalogue metadata: authors, new-title ordering, visible source/credits and publication rights confirmation.
- [ ] Rights contact without registration, private moderator handling and reporter-visible report outcomes.
- [ ] Minimal editable public display name and public team roster without exposing email.
- [ ] Safe return to the original page after login.
- [ ] Remove inaccessible first-page-only limits from essential studio/moderation/admin workflows.
- [ ] Inspect remaining library-offline, moderation/audit and reader navigation contracts; implement uncovered functional gaps rather than claiming them tested.
- [ ] Fresh full tests, local Docker smoke, independent review, commits and updated requirement matrix.

## Rules

Implement directly in the existing main checkout, domain/application/adapters and presentation boundaries unchanged. Add focused failing tests before behavior; verify relevant PostgreSQL/security cases and UI contracts. Preserve existing data through additive migrations. Do not deploy publicly or change production TLS, hosting, infrastructure sizing or external mail.

The 13 acceptance criteria are not all software-only: physical weak Android, native-language quality review and actual distribution rights cannot be established by implementation or emulation. Keep them explicit at completion.
