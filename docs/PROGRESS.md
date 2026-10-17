# Qmanga implementation progress

Plan: docs/IMPLEMENTATION.md

Executor and reviewer: Codex, directly, per owner's latest instruction. AGY is stopped; its work is preserved.

Base: 5cafa1d5a1e5b2e33252c4d86f98ce49c47d583d.

- Setup: repository cloned; Docker Linux engine and Go available; AGY model verified.
- Task 1: complete. Alpine Compose builds and all five services are healthy. HTTP homepage returns 200; Go readiness confirms PostgreSQL and both MinIO buckets. API/MinIO uid100 on Alpine3.21.8, Bun uid1000 on Alpine3.24.2. Svelte check: 0 errors/warnings; production build and frontend/Go dependency checks pass. Real PostgreSQL integration tests pass, including concurrent migration startup and cancellation while another connection holds the lock. Main checkout contains all code.
- Task 2: accounts implemented and running; team roles/invitations remain next. Go domain/application ports, PostgreSQL/Argon2id/SMTP adapters and Gin transport are separate. Kazakh login/register/recovery/verification forms use a TypeScript domain, application gateway port, HTTP adapter and browser/server composition. Database integration proves verification/reset single-use, CSRF/origin rejection, password verification, expiry and session revocation. scripts/smoke-auth.ps1 passes against the actual gateway/API/DB/SMTP stack. Form widths320/360 checked without horizontal overflow; inputs48px, submit46px, secondary action44px.
- Task 3: pending.
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
