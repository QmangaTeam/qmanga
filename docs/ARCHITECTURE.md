# Architecture decision: Clean Architecture + DDD

Status: required by the owner on 2026-10-17 for both backend and frontend. This decision supersedes the flat package paths in the original implementation plan. Product scope is unchanged. Codex implements and verifies directly following the owner's updated instruction.

## Principles

- Start as a modular monolith, not microservices. One Go API, one Go worker, one PostgreSQL database and private MinIO are sufficient.
- Organize by business capability, then layer. Do not create a global service/repository folder containing every feature.
- Dependencies point inward: adapters -> application -> domain. Domain code knows nothing about HTTP, SQL, MinIO, Svelte, browser APIs or deployment configuration.
- DDD means explicit language, ownership and business invariants, not one class/interface for every field. Use aggregates only where consistency boundaries matter. No generic CRUD repository, CQRS framework, event sourcing or dependency injection container.
- Compose concrete adapters only at entry points. Tests inject ports. Avoid service locators, global clients and environment reads in business code.

## Bounded contexts

| Context | Owns |
| --- | --- |
| Identity | Account, session, verification and reset lifecycle |
| Teams | Approved team, memberships, roles and invitations |
| Catalog | Title, alternate names, genres, chapter order and translation projects |
| Publishing | Release, revision, page assets, uploads, processing and publication decisions |
| Reading | Library entry, preferred translation and versioned reading progress |
| Community | Ratings, comments, spoilers, reports and moderation decisions |
| Notifications | Subscriptions, deduplicated in-app delivery and read state |
| Metrics | Private daily operational event aggregates and scoped dashboard read models |

A title, chapter, team release and immutable revision are different concepts. Do not collapse them into one manga table/model. A frontend view may combine read models from several contexts without merging their business ownership.

## Go layout

```text
backend/
  cmd/api/             # executable composition entry
  cmd/worker/
  cmd/admin/
  internal/bootstrap/  # builds concrete adapters and registers routes
  internal/platform/
    config/
    postgres/          # pool and migration runner, not business queries
    s3/                # private streaming storage adapter
    readiness/         # technical dependency checks
  internal/transport/http/
                       # Gin router, middleware, handlers and transport DTOs
  internal/<context>/
    domain/            # entities, value types, state transitions, domain errors
    application/       # use cases, commands/read models and required ports
    adapters/
      postgres/        # repository implementations and transactional SQL
      smtp/            # identity mail delivery, when applicable
  migrations/
```

- Application ports are small behavior-specific interfaces defined by their consumers. No `pgxpool.Pool`, `http.Request`, `minio.Client` or Svelte/JSON transport DTO crosses into the domain/application API.
- Application orchestrates authorization, repositories, transactions, clocks, ID generation, password hashing, email and storage via ports. Domain methods enforce valid transitions and value invariants.
- Repositories own SQL and translate persisted representations. Atomic operations such as consume-reset-token, publish-revision, reserve-upload-quota and save-progress-event have explicit transactional contracts; do not imitate a transaction with separate unrelated repository calls.
- Contexts exchange IDs, explicit application contracts and outbox events, not each other's database clients or mutable domain entities. A query adapter may join tables for a documented read model, but may not mutate another context's owned state.
- The composition root wires HTTP adapters and a Go media gateway. All media and uploads follow MEDIA-GATEWAY.md; MinIO never becomes a frontend adapter.
- Health/readiness is a technical application service with dependency-check ports, not a fabricated DDD aggregate. Missing dependencies are unhealthy.

## Frontend layout

```text
apps/web/src/
  routes/                      # thin SvelteKit entry points
  lib/
    features/<context>/
      domain/                  # pure TypeScript values and rules
      application/             # use cases and gateway/storage ports
      infrastructure/          # HTTP DTO mapping, IndexedDB/cache adapters
      presentation/            # Svelte components, view state and bindings
    shared/domain/             # only genuinely shared value types
    shared/presentation/       # reusable UI primitives
    server/composition.ts      # SSR-only configuration and concrete wiring
    composition/browser.ts     # browser-only concrete wiring
```

- Svelte components render state and invoke use cases; they do not contain publication rules, progress conflict resolution, HTTP payload mapping or IndexedDB operations.
- Domain/application modules have no imports from Svelte, SvelteKit, DOM APIs, `fetch`, `localStorage`, `indexedDB`, Cache Storage, Node or Bun. Pass platform capabilities through ports.
- Frontend domain models describe reader behavior, library choices, offline lifecycle and progress conflicts. Do not duplicate backend authorization as if it were authoritative. Go always verifies permissions and invariants again.
- Infrastructure validates/maps untrusted API DTOs to domain/read models. Keep backend `lower_snake_case` JSON out of UI components through explicit mapping.
- SSR and browser composition roots are separate. Importing a feature on the server must not touch `window` or browser storage. Route load/actions may import the server composition root, never a database or MinIO client.
- Reader image scheduling, quality decisions and offline completeness are testable pure policies/application services. Actual network, image decode, visibility and storage access are adapters.
- Mobile-first presentation remains required; architecture must not add a heavy state framework or unnecessary JavaScript to the reader.

## Testing and review gates

- Domain unit tests need neither Docker nor HTTP mocks. Test publication transitions, chapter labels, quality/position policies and progress conflict rules in their own contexts.
- Application tests use port doubles and cover failures and authorization decisions; adapter integration tests use real PostgreSQL/MinIO/SMTP when relevant.
- HTTP and browser tests verify complete paths, including uploads and media requests passing through Go.
- Add repeatable import-boundary checks for Go and TypeScript. At minimum forbid domain/application dependencies on outward layers/frameworks and cross-context private adapter imports. Static checks supplement code review; folder names alone do not prove architecture.
- Test composition independently: fresh Alpine Compose build, actual readiness, graceful stop and persistence. Do not report a layer complete because only isolated mocks pass.

## Immediate migration

The initial flat scaffold has been replaced by platform adapters, a Gin HTTP transport and explicit composition. Publication state and media limits live in a pure Go domain; reader quality, position and offline completeness policies live in pure TypeScript. Application scenarios consume ports; concrete persistence/network adapters are wired at the composition roots. Existing product documents remain the functional acceptance checklist.

## Catalogue age classification contract

`age_rating` is an integer: `0` means unclassified, and `12`, `16`, `18` are explicit editorial classifications. Existing titles default to `0`; the migration never assumes `12+`. Public summaries and details return the value. `GET /api/v1/titles?age_rating=18` filters the exact classification; `0` selects unclassified titles, and omission/empty value selects all classifications. Invalid values are rejected.

Only an administrator may change title metadata. On an update, omitted `age_rating` preserves the stored classification for older clients; explicit `0` clears it. This additive field does not introduce a reader access gate, a new genre, or automatic content classification.

## Reading completions and operational metrics

Metrics owns additive UTC daily event counters (`metrics_daily`) and a first-publication receipt per release. Collection starts at migration 000027; unknown historical confirmation/publication dates are not invented. Database triggers record confirmed-account transitions, first publication of a translation release, first rating, new/restored library entries and newly enabled title notification opt-ins in the same transaction as their source operation. Rating edits do not create new ratings. Removal/unsubscribe does not subtract an earlier event: these are additions over a period, not current balances. Republishing/revising an already published translation does not create another first-publication event.

Reading owns `reading_activity` (one account/release/UTC-day row, completion is monotonic) and `reading_completions` (one account/title/logical-chapter receipt). Only accepted, server-validated progress against the current published revision/page records activity; stale/conflicting/replayed receipts do not. Completion requires the last page and fraction >= .95. Different translations/revisions of the same logical chapter cannot advance the library suggestion threshold. After three distinct completed chapters an authenticated reader outside the library may receive a nonblocking suggestion; dismissing it suppresses it for that account/title in the browser tab session. New library preferences are inferred from actual accepted reading only when multiple projects compete; otherwise blank. Subsequent explicit preference edits remain authoritative.

**There is no public views counter.** Image/fragment requests, preloads and offline downloads never count as views. Authenticated progress and last-page completion are self-reported engagement signals, not proof of human reading; verified bots/accounts could submit them. Dashboard labels therefore say active accounts and completed readings, never human views. Any future public views count requires a separate approved qualification/rate-limit/bot-resistance policy; these operational counters must not be relabeled as views.

Daily active readers are distinct accounts with accepted progress; period readers are distinct over the entire selected window (not a sum of daily counts). Completed readings are deduplicated account/translation/day. Anonymous reading is deliberately excluded. No IP, email, user agent, fingerprint or page-image access history is collected by metrics. Personal activity rows expire after 90 UTC days during the first subsequent activity write, at most once per UTC day; aggregates use the same rolling horizon. Completion receipts remain reading state until the account/title/chapter is deleted. Existing progress idempotency receipts have their own synchronization lifecycle and are not exposed as metrics.

`GET /api/v1/admin/metrics?days=7|30|90` requires a verified unblocked administrator. `GET /api/v1/studio/metrics` derives the union of the actor's currently approved team memberships on every request, without a caller-controlled scope or reader-identifying response. Translator reading/publications cover their teams' translations; library/subscription events are attributed only to an explicitly preferred team, never all teams on a title. There is currently no separate team-follow feature: subscriptions mean title notification opt-ins. Membership removal immediately removes access. Every report is private/no-store, uses a repeatable-read snapshot, an application five-second deadline, indexed 90-day maximum ranges and exactly 7/30/90 returned days. Daily counters avoid rescanning account/rating/library histories; reading indexes bound the remaining distinct-account aggregation. Charts have an equivalent visible table and never depend on hover. Zero periods and the collection start date are explicit.
