# Translator workflow repair

The owner reports that the working pages do not provide a usable translator journey. Existing approval and publication policies remain: verified users request teams; approved teams request projects; administrators approve catalogue access; editors prepare chapters; moderators publish revisions.

Session authorization: the owner previously instructed autonomous completion and now explicitly asks for real functional testing and UX repair. Execute directly in this checkout; preserve the palette, Kazakh-first/Russian UI, mobile-first and Clean Architecture boundaries. No production work.

## Observed gaps

- Studio splits team/project forms across disconnected pages without an explicit next step.
- Project list says an administrator assigns projects but provides no request action or history.
- Editor cannot add a logical chapter; an administrator must pre-create each chapter.
- Team applications only support approval, with no rejection reason or correction path.
- Requests/errors/permissions and publication stages are difficult to distinguish.

## Execution

1. Catalogue-context project request aggregate: existing-title and new-title requests, team-scoped history, duplicate/rate protection, admin queue, atomic approval creating title/project, rejection with reason, cancellation. Preserve tenant checks, verified membership and row locks.
2. Team request rejection/reason and owner correction/resubmission; retain approval policy and last-owner protections.
3. Editor-owned chapter creation within approved projects, shared logical chapter reuse, explicit conflicting metadata errors; no grant of catalogue administration.
4. Reusable workspace navigation and request-state presentation first. Assemble studio projects, requests and team screens with direct actions, visible loading/errors/empty states and confirmation for destructive membership actions. Improve editor chapter/draft/upload/details/preview/submit sequence.
5. Integration tests of request isolation/replay/approval and chapter rights; browser journey with local test accounts, 320/360px and desktop, both languages. No private sample artwork or credentials in Git.
6. Build/restart local containers, focused review, commit/push existing private repository and leave concrete testing instructions.

Completion is the browser journey and its checks, not merely the presence of routes. Record actual limits in the existing MVP audit.

## Result

Implemented directly; no AGY, branch, production deployment or private sample redistribution. Verification and remaining limits are recorded in `docs/ui/STUDIO-WORKFLOW-VALIDATION.md`. Fresh review findings were fixed with failing/passing regressions. The browser journey reached an actual published synthetic chapter and a returned/corrected team application. Final frontend checks: 39 + 13 passing tests; full Go/PostgreSQL/codec suite passed.
