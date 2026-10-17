# Architecture decision: Clean Architecture + DDD

Status: required by the owner on 2026-10-17 for both backend and frontend. This decision supersedes the flat package paths in the original implementation plan. Product scope is unchanged. AGY Gemini 3.8 Flash High implements; Codex reviews.

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
    httpserver/        # cross-cutting transport concerns only
  internal/<context>/
    domain/            # entities, value types, state transitions, domain errors
    application/       # use cases, commands/read models and required ports
    adapters/
      http/            # transport DTOs, validation, status/error mapping
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

The unaccepted WIP scaffold has flat `internal/config`, `store`, `storage`, `httpapi`. AGY must move infrastructure into platform packages and move transport/composition out of business logic before further feature integration. The pending identity task is re-scoped to domain/application/adapters before implementation. Existing product documents remain the functional acceptance checklist.
