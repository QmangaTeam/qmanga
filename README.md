# Qmanga

Қазақ тіліндегі манга платформасы. Mobile-first SvelteKit + Bun + Tailwind, Go/Gin, PostgreSQL, private MinIO. All images and uploads pass through Go.

## Current development state

The Alpine stack, Kazakh mobile shell, accounts and approved-team workflows run. Gin handles HTTP; accounts support email verification, session cookies, CSRF and password recovery. Teams have owner/editor/member access, invitations and admin approval. Catalogue, publishing, reader, offline and community remain in development; this is not yet the full MVP.

## Local start

Windows PowerShell, from this repository:

```powershell
./scripts/setup.ps1
docker compose up -d --build --wait
```

Linux/macOS:

```sh
sh scripts/setup.sh
docker compose up -d --build --wait
```

Open http://localhost:8080. The setup scripts preserve existing configuration and generate random local credentials. Never commit .env. APP_PORT and APP_ORIGIN must match when changing the local port.

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

## Architecture

See docs/ARCHITECTURE.md and docs/MEDIA-GATEWAY.md. Feature contexts contain domain, application ports/use cases, and adapters; Go and browser composition roots provide concrete dependencies. PostgreSQL migrations are embedded and serialized by an advisory lock.

The original MinIO upload sources remain private. Reader original quality will be a sanitized full-resolution published derivative through the Go gateway. Already downloaded offline files cannot be revoked while the device is disconnected.

## Product specification

QMANGA-PLAN.md, QMANGA-READER.md and QMANGA-ROADMAP.md define the remaining MVP acceptance criteria. docs/PROGRESS.md records tested milestones.
