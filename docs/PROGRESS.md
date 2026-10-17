# Qmanga implementation progress

Plan: docs/IMPLEMENTATION.md

Executor and reviewer: Codex, directly, per owner's latest instruction. AGY is stopped; its work is preserved.

Base: 5cafa1d5a1e5b2e33252c4d86f98ce49c47d583d.

- Setup: repository cloned; Docker Linux engine and Go available; AGY model verified.
- Task 1: complete. Alpine Compose builds and all five services are healthy. HTTP homepage returns 200; Go readiness confirms PostgreSQL and both MinIO buckets. API/MinIO uid100 on Alpine3.21.8, Bun uid1000 on Alpine3.24.2. Svelte check: 0 errors/warnings; production build and frontend/Go dependency checks pass. Real PostgreSQL integration tests pass, including concurrent migration startup and cancellation while another connection holds the lock. Main checkout contains all code.
- Task 2: implemented. Accounts, verified-email team applications, admin approval, owner/editor/member permissions, scoped invitations, exact-recipient acceptance, owner-only member management and immediate revocation. Gin transport injects application services; domain policies have no transport/persistence dependencies. Real PostgreSQL tests cover cross-team denial, unapproved uploads, self-approval denial, wrong-recipient and reused invitations, last-owner protection and revoked access. Studio, approvals, invitation, account and theme pages are implemented. Auth smoke with real SMTP passes. Form widths320/360 checked without horizontal overflow; inputs48px, submit46px, secondary action44px.
- Task 3: started. Catalogue metadata and chapter-label domain validation, application ports/read models, and search/index migration are present. SQL adapters, catalogue HTTP endpoints, actual title/project flows and catalogue UI are not implemented yet.
- Task 4: pending.
- Task 5: pending.
- Task 6: pending.
- Task 7: pending.
- Task 8: pending.
- Task 9: pending.

Decision: implementation is explicitly authorized; planning review gates are satisfied by the user's instruction to execute the prepared MVP with the named executor.
Decision: infrastructure sizing remains out of scope, but Alpine Docker execution is now a required deliverable.

Decision: every image and upload goes through Go; see MEDIA-GATEWAY.md. MinIO remains private. Reader Original is a sanitized full-resolution published copy, not the private raw upload.

Initial review: reject nil dependencies reported healthy; reject swallowed advisory-lock errors; do not expose raw database/storage errors publicly; do not silently accept invalid APP_ENV/S3_USE_SSL or default database passwords. No runtime is deployed from the WIP checkpoint.

Owner amendment: Clean Architecture + DDD on frontend and backend. Two pending AGY runs were cancelled before either changed files, with worktrees preserved. See ARCHITECTURE.md. Restarted assignments must separate domain, application ports/use cases, adapters and composition; do not integrate a flat DB-coupled identity service.

Execution location: all active development is now directly in outputs/qmanga, as requested. The prior isolated worktree was fast-forwarded into main; no files were discarded.

Architecture correction: Gin 1.12 is the HTTP transport in internal/transport/http. Removed the artificial health context and the duplicate platform/httpserver router. Technical readiness is platform-only. Go publication domain tests cover immutable submission, failed replacement preserving the current publication, upload/pixel limits and bounded fragments. TypeScript reader domain tests cover explicit quality, low-memory concurrency, original-page positions and incomplete/off-revision downloads; 3 tests/13 assertions pass without browser APIs.

Deployment fix: nginx resolves Docker service addresses dynamically, so replacing API/web containers does not leave stale upstream addresses. All six default services healthy; Mailpit is Alpine3.24.2 uid1000, UI loopback8025. The browser viewport override was reset after mobile checks.

Frontend DI correction: presentation components receive application gateway ports through props. SvelteKit route entry points construct adapters; SSR composition is per request. Architecture checker now inspects Svelte script imports too. It failed on the old outward composition imports, then passed after injection. Domain/application cannot access browser APIs; presentation cannot fetch, access storage or import infrastructure/composition. Themes use an application port and browser storage adapter.

Next: Task3 catalogue/admin metadata and translation projects, then Task4 uploads, worker, image variants and Go-only media delivery. A complete reading MVP is still in progress; do not claim the current shell/accounts are the whole product.

Owner UI amendment: keep green; the rejected part is the layout/composition, not its hue. Complete functional MVP before the final design pass. The reusable design-product-ui skill is stored in .agents/skills and installed in the personal skills directory. Frontend AGENTS.md links the skill and docs/ui/DESIGN-REVIEW.md. No UI restyling is included in this checkpoint.
