# Qmanga

Қазақ тіліндегі манга платформасы. Mobile-first SvelteKit + Bun + Tailwind, Go/Gin, PostgreSQL, private MinIO. All images and uploads pass through Go.

Статус и правила работы: [AGENTS.md](AGENTS.md). Деплой: [RUNBOOK](docs/deployment/RUNBOOK.md).

## Local start

For Oracle staging/production setup and safe releases, see the [deployment runbook](docs/deployment/RUNBOOK.md).

Use Docker Desktop/Engine with Compose v2. Windows PowerShell 7, from this repository:

```powershell
./scripts/setup.ps1
docker compose --parallel 1 up -d --build --wait --wait-timeout 180
```

Linux/macOS:

```sh
sh scripts/setup.sh
docker compose --parallel 1 up -d --build --wait --wait-timeout 180
```

Open http://localhost:8080. Setup preserves an existing env and generates credentials only for a new environment. **If old volumes exist but the env is missing, restore its original credentials; do not regenerate them or delete data.** Follow the single [local development/recovery guide](docs/LOCAL-DEVELOPMENT.md) for a separate sandbox, synthetic publication, browser checks and the staff access policy. Never commit .env. APP_PORT and APP_ORIGIN must match when changing the local port.

MinIO and PostgreSQL have no host ports. Only the same-origin gateway is exposed, on loopback by default. Application runtimes and MinIO run as non-root Alpine containers. Image manifests are pinned by digest; review pins periodically.

```sh
docker compose ps
docker compose logs --tail=50 api
docker compose down
```

Named volumes survive restarts and compose down. Do not remove volumes to upgrade or restart.

## Checks

```sh
go -C backend test ./...
docker compose --profile tools run --rm go-tests
docker compose build web
```

PowerShell can run the real SMTP/auth smoke flow with `./scripts/smoke-auth.ps1`. It creates disposable local test accounts with random credentials and requires localhost only. Mailpit at http://localhost:8025 captures development email; it does not deliver to external recipients.

Create an administrator by setting ADMIN_EMAIL and supplying ADMIN_PASSWORD through an environment variable or stdin to the `qmanga-admin` executable in the API container. It creates a verified admin, never a known default password. For deployment set SMTP_ADDRESS, SMTP_FROM and optional SMTP_USERNAME/SMTP_PASSWORD; production requires SMTP STARTTLS.

The first command skips PostgreSQL integration without TEST_DATABASE_URL. The tools service runs integration against isolated test schemas in the local database and checks concurrent migrations and cancellation. The web build runs Svelte checks and frontend dependency-boundary checks.

## Interface and local preview

The interface defaults to Kazakh. An explicit Russian choice is available in settings and reader settings; it does not translate manga content or reset the reading position. Dark and light themes use the approved orange/brown palette. Help and reporting guidance are at /help.

For the isolated component gallery run `docker compose --profile tools up -d web-ui` and open http://localhost:3001. It is development tooling, not a production route. The web image build runs application tests, Svelte SSR component tests, architecture/contrast checks, both Svelte checks and the production build.

## Architecture
See docs/ARCHITECTURE.md and docs/MEDIA-GATEWAY.md. Feature contexts contain domain, application ports/use cases, and adapters; Go and browser composition roots provide concrete dependencies. PostgreSQL migrations are embedded and serialized by an advisory lock.

The original MinIO upload sources remain private. Reader original quality is an explicitly requested, sanitized full-resolution derivative through Go. Already downloaded offline files cannot be revoked while the device is disconnected. IndexedDB/Cache Storage can be evicted by the browser; the saved list can recheck completeness.
