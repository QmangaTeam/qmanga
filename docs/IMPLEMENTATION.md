# Qmanga MVP Implementation Plan

> **For agentic workers:** Use the AGY delegation workflow for implementation. The owner requested `gemini-3.8-flash-high` as executor and Codex as reviewer. Work task by task, test meaningful behavior, and commit accepted milestones.

**Goal:** Deliver the Kazakh mobile-first manga platform specified in QMANGA-PLAN.md, runnable through Docker Compose with Alpine-based application images.

**Architecture:** SvelteKit SSR on Bun calls one Go REST API. PostgreSQL is authoritative for business data and processing jobs; S3-compatible MinIO stores private source files and processed media. A Go worker prepares images and publication manifests; all externally accessible routes use one origin.

**Tech Stack:** Bun, SvelteKit, TypeScript, Tailwind, Go, pgx, PostgreSQL, MinIO S3 SDK, Docker Compose. Image transformations use a bounded Alpine-compatible image tool in the worker.

**Spec:** ../../QMANGA-PLAN.md, ../../QMANGA-READER.md, ../../QMANGA-ROADMAP.md. Paths in this paragraph are repository-root names; read the root files.

## Global Constraints

- All product copy is Kazakh; development documentation can be English or Russian.
- Mobile first at 360px, functional from 320px, controls at least 44px.
- Anonymous reading, private originals and drafts, multiple approved teams per title.
- Go owns validation and authorization. SvelteKit never reads PostgreSQL directly.
- Use real PostgreSQL and S3 persistence, not an in-memory demo implementation.
- Docker application runtimes are Alpine; PostgreSQL uses its Alpine image. Build the MinIO runtime on Alpine if the upstream image is not Alpine.
- No real passwords or storage keys in git. Runtime secrets are generated locally or supplied through environment variables.
- Preserve source files and browser privacy. Demo content must be original and clearly marked as fictional.
- The user authorizes implementation, delegation, frequent commits and saving to the private GitHub repository. No public deployment is requested.
- Only AGY Gemini 3.8 Flash High implements product code. Codex coordinates, reviews, runs checks, and integrates.
- Do not spawn nested agents. Do not add inline code comments or one-letter identifiers.
- Never silently replace a working feature with a placeholder or claim an untested feature works.

## Review Focus

1. Tenant isolation: changing a team/project/release identifier cannot expose or mutate another team's drafts.
2. Publication integrity: processing failure, retries, concurrent approval and revision changes cannot publish partial or mixed chapters.
3. Reading integrity: offline updates from another device cannot silently overwrite newer progress; different translations remain distinct.
4. Resource limits: malformed images, very tall sources, partial downloads and quota errors do not crash the server or mislabel offline readiness.
5. Deployment: a fresh checkout starts with documented commands, persists restarts, keeps secrets out of responses, and actually runs Alpine images.

## Shared Contracts

Repository layout: `apps/web` for SvelteKit; `backend/cmd/api`, `backend/cmd/worker`, `backend/internal/*`, `backend/migrations`; `infra` for image builds and gateway; `tests` for browser and integration checks; `scripts` for repeatable setup; `docs` for review notes and developer instructions.

Go exposes `/api/v1` and `/media`. Go uses standard `net/http` ServeMux and pgxpool. `internal/config.Load() (Config, error)`, `internal/store.Open(context.Context, string) (*Store, error)`, `internal/httpapi.New(config.Config, *store.Store, *storage.Client) http.Handler` are composition interfaces. Extend them explicitly when needed; avoid globals. Public JSON is lower_snake_case. Error shape is `{error:{code:string,message:string}}`, messages are Kazakh. List responses are `{items:[],next_cursor:string|null}`. IDs are strings. UUIDs are generated server-side.

Core entities: users/sessions/tokens, titles/names/genres, teams/memberships/invitations, projects, chapters, releases/revisions/pages/variants/upload_sessions, jobs/outbox, library/progress/events, ratings/comments/reports/notifications/audit. Migrations are embedded, ordered, transactional where appropriate, with a version table and startup migration lock. Chapter labels are strings and order is an integer; no float numbering.

Public reader manifest: `{release_id,revision_id,title_id,title_name,chapter_label,team:{id,name},direction,previous_release_id,next_release_id,pages:[{id,position,width,height,variants:{eco:[{url,width,height,bytes,offset}],standard:[],high:[]}}]}`. Offset is normalized to original page. Original URLs are obtained only on explicit request; never expose source bucket credentials.

Authentication: email/password, Argon2id, random opaque session token hashed in database; HttpOnly cookie, Secure in production, SameSite=Lax. Unsafe requests validate Origin plus CSRF token for session-authenticated browser operations. Public registration/login/password-reset have rate limiting. Email transport uses configured SMTP and local development mail capture; verification/reset tokens are never returned by production APIs. User approval of executor covers implementing this already-specified subsystem; no account secrets are included in task prompts.

## Task 1: Runnable foundation and stable contracts

Files: root manifests and scripts, `apps/web` minimal SSR shell, `backend/go.mod`, `backend/cmd/api/main.go`, `backend/internal/{config,httpapi,store,storage}`, initial migrations, `compose.yaml`, Dockerfiles, `.env.example`, `.gitignore`, `.dockerignore`, README.

Produces: working web → Go health call; healthy PostgreSQL and MinIO; explicit config validation; startup migrations; Alpine application containers. Choose compatible pinned dependency versions and lock them. Preserve the Chrome 109 target: Tailwind 3.4 is acceptable if newer Tailwind requires a newer browser.

- [ ] Add config tests for missing production secrets and health handler tests.
- [ ] Observe the new behavior tests fail, implement foundation, pass `go test ./...` and frontend checks.
- [ ] Provide Compose health checks, dependency readiness, named volumes, non-root app users, graceful stop and a same-origin gateway.
- [ ] Provide a safe local setup command that creates ignored `.env` secrets without overwriting existing configuration.
- [ ] Verify `docker compose config`, build and startup of the foundation.
- [ ] Review and commit `chore: scaffold Alpine application stack`.

## Task 2: Identity and team authorization

Files: `backend/internal/identity`, `backend/internal/teams`, route integration, migrations, web auth/settings pages.

Consumes Task 1 config/store/router. Produces session middleware and `RequireUser`, role/project access checks used by later handlers. Routes: auth register/login/logout/me/csrf/verify-email/forgot-password/reset-password; team membership and invitation acceptance. Configurable bootstrap admin command reads password from stdin/environment and does not create known default credentials.

- [ ] Test duplicate registration, wrong password, expiry/revocation, token single-use, CSRF, origin rejection, team role isolation and revoked memberships.
- [ ] Implement real persistence, SMTP flow and Kazakh auth screens with return paths.
- [ ] Run tests and browser login/reset flow with local mail capture.
- [ ] Review and commit `feat: add accounts and scoped team access`.

## Task 3: Catalogue and translation projects

Files: catalog store/services/handlers/tests, title/genre/project migrations, web catalogue/title/team pages, original demo seed assets/data.

Produces public titles/search/filter/detail/chapters/team APIs and admin catalogue/project creation. Search keeps Kazakh characters and aliases. Public results contain only published content. Library status is separate from public caching.

- [ ] Test Kazakh names/aliases, empty results, pagination, chapter labels `1`, `1.1`, `1.10`, `2`, two teams for one chapter and hidden drafts.
- [ ] Implement main mobile navigation, dark/light/system theme, search, filters, title and team views.
- [ ] Seed original fictional sample titles and original demonstration pages; mark them demo.
- [ ] Run database integration tests and web checks.
- [ ] Review and commit `feat: add Kazakh catalogue and translation projects`.

## Task 4: Publishing and image processing

Files: publishing/media/jobs modules, worker entrypoint, migrations, studio and moderation pages, image tool Docker dependencies.

Produces scoped resumable upload sessions, draft editing, natural page order, server-side reorder/version checks, processed profiles, submit/review/publish/withdraw and original-page access. Source formats JPG/PNG/WebP; 25MiB per file, 500MiB/release, 300 source pages, 80MP/source. Profiles at 720/1080/1600 max width; slices up to 2048px. Never upscale.

- [ ] Test cross-team access, forged file metadata, truncated/invalid images, oversize/pixel limits, retry after worker crash, concurrent submit/publish and all-or-nothing revision replacement.
- [ ] Implement private uploads, bounded image conversion, PG leased jobs and transactional outbox.
- [ ] Implement mobile studio upload progress, failed retry, order controls and reader preview; moderation decisions require reason on rejection/withdrawal.
- [ ] Test two team releases of the same chapter and edits to a published release without breaking its current revision.
- [ ] Review and commit `feat: publish versioned chapters with processed images`.

## Task 5: Reader and low-bandwidth behavior

Files: manifest/media handlers, `apps/web/src/lib/reader/*`, reader route and focused tests.

Consumes the exact shared manifest. Produces strip/page modes, direction, quality controls, manual original view, bounded fetch/decode window, retries and retained page-relative position.

- [ ] Test quality switching without changing source position, no silent team switch, missing variant, max two image requests in eco, cancellation, retry limits and withdrawn response.
- [ ] Implement early first image, dimensions before load, nearby image window/placeholders and native zoom.
- [ ] Exercise a long original demo chapter, network failure, 360px navigation and both reading modes.
- [ ] Review and commit `feat: add adaptive mobile manga reader`.

## Task 6: Library and progress synchronization

Files: library/progress modules and migrations, frontend library/settings/progress storage.

Produces independent favorite flag and reading statuses, preferred project, local guest progress, versioned progress event API, last position and furthest completed chapter, explicit conflicts and import-on-login.

- [ ] Test event idempotency, out-of-order sequence, stale base version, two devices, rereading early chapters and account switching without data leakage.
- [ ] Implement local periodic saves, network batching and visible pending/conflict states.
- [ ] Implement library views and continue-reading cards.
- [ ] Review and commit `feat: synchronize library and reading progress`.

## Task 7: Offline chapters

Files: service worker, IndexedDB download manager, downloads view and browser tests.

Produces explicit single-chapter save operations, byte estimate, progress/error/ready states, offline restart reading and deletion controls. Cache public reader shell/media only. 50MiB automatic LRU; explicit saved soft budget min(200MiB,20% estimated remaining quota); handle quota failures regardless of estimates.

- [ ] Test partial download is never ready, missing cache entries, quota rejection, cancelled save, retained explicit chapters, service-worker update and offline restart.
- [ ] Implement persisted manifest and files, resume failed saves without refetching complete objects, capability fallback and account-private data isolation.
- [ ] Run real browser offline flows.
- [ ] Review and commit `feat: save chapters for offline reading`.

## Task 8: Community and notifications

Files: comments/ratings/reports/notifications modules and UI, migrations, outbox delivery.

Produces text comments and one reply level, spoiler, own edit/delete, moderation hide, one rating 1–10/account/title, report categories, in-app chapter and moderation notifications with dedupe.

- [ ] Test access checks, XSS output, reply depth, duplicate rating updates, reporting page context, duplicate outbox delivery and notification preference/project filtering.
- [ ] Implement Kazakh screens and empty/error states with bounded pagination.
- [ ] Review and commit `feat: add community moderation and notifications`.

## Task 9: Full integration and release verification

Files: e2e tests, CI workflow, README/runbook, targeted fixes returned to AGY.

- [ ] Start from a fresh generated `.env`, build all Alpine images, migrate and seed only original demo content.
- [ ] Run Go unit/integration tests, Svelte checks/build and browser tests for both guest and approved-team paths.
- [ ] Check 320/360/390/412/768/1280px, no horizontal overflow, keyboard/back behavior, quality changes and offline restart.
- [ ] Verify Docker restarts retain chapters/accounts and application containers report Alpine and non-root identity.
- [ ] Record first-load JS/media sizes and throttled reading results; explicitly distinguish browser emulation from physical Android validation.
- [ ] Run final review, delegate fixes, rerun affected checks, commit, push accepted milestones to private GitHub.
- [ ] Deliver local URL, startup instructions, commits, implemented scope and any verification limitations honestly.

## Execution ledger

Initial state: specifications accepted for execution by the user's request on 2026-10-17. User explicitly chose AGY Gemini 3.8 Flash High and Codex review, authorized frequent commits and Alpine Docker. No repeat execution approval is needed.

Foundation dependency choices, endpoint refinements and review outcomes will be recorded in `docs/PROGRESS.md`. The specification remains the coverage checklist; a feature is not complete until its task evidence is recorded.
