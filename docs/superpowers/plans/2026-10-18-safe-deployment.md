# Safe two-environment deployment implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship repeatable staging and production releases on one Oracle Free Tier VM without replacing a healthy site with an unverified build.

**Architecture:** GitHub tests and builds immutable arm64 images. Two isolated data projects feed blue/green application stacks. Caddy switches each hostname after candidate health checks; release scripts restore the previous route on failure. Backups and runbooks make recovery repeatable.

**Tech Stack:** GitHub Actions, Docker Compose, Caddy, Bash, Go/Gin, PostgreSQL, MinIO.

**Spec:** `docs/deployment/2026-10-18-deployment-spec.md`

## Global constraints

- Preserve local `compose.yaml` and current development data.
- Production is not made public until DNS, TLS, off-host backup and restore have been checked on the actual VM.
- Never write credentials to Git or GitHub logs. No password or private SSH key is requested through chat.
- Publish only commit-SHA-tagged Linux/arm64 images after full CI success.

## Review focus

- Candidate API or web fails health: active route remains unchanged.
- Post-switch smoke fails: old route and worker resume, no database rollback attempted.
- Supplied SHA lacks a green main CI run: production promotion fails.
- A direct client forges X-Forwarded-For: identity limiter uses the real peer.
- Backup misses either PostgreSQL or MinIO: release stops before applying migrations.

---

### Task 1: Trust only the internal gateway

**Files:** `backend/internal/platform/config/config.go`, `backend/internal/bootstrap/bootstrap.go`, `backend/internal/transport/http/router.go`, `backend/internal/transport/http/router_test.go`, `infra/nginx.conf`.

**Interfaces:** `TRUSTED_PROXY_HOST=gateway`; production startup resolves this service to exact IPs before Gin trusts forwarding headers. nginx passes only the X-Forwarded-For supplied by the trusted front proxy.

- [ ] Write and run a failing Go regression: forwarded client IP is ignored for a non-gateway peer and honored for the configured gateway.
- [ ] Implement exact gateway trust and fail startup on unresolved/invalid proxy configuration.
- [ ] Run focused and full Go tests.

### Task 2: Isolated host runtime and backups

**Files:** `deploy/compose.data.yaml`, `deploy/compose.app.yaml`, `deploy/compose.front.yaml`, `deploy/Caddyfile`, `deploy/backup.sh`, `deploy/bootstrap.sh`, `deploy/host.env.example`.

**Interfaces:** target `staging|production`, color `blue|green`, image `RELEASE_SHA`, loopback gateway ports and separate Compose projects/volumes. Backup outputs a checksummed directory including pg_dump and both bucket objects.

- [ ] Add static config checks demonstrating separation, loopback-only gateways and no data volume removal.
- [ ] Implement Compose files, Caddy configuration, host bootstrap and backup.
- [ ] Validate Compose and Caddy configurations; prove backup on a disposable project and restore on a separate disposable project.

### Task 3: Atomic release, rollback and CI packaging

**Files:** `deploy/release.sh`, `deploy/rollback.sh`, `deploy/test-release.sh`, `.github/workflows/verify.yml`, `.github/workflows/promote.yml`.

**Interfaces:** `release.sh staging|production SHA`; `rollback.sh staging|production`. State records active color/SHA. Workflow uses a successful main SHA, then sends it over pinned SSH to the host.

- [ ] Write failing dry-run/fake-command tests for failed candidate, failed switch and successful switch.
- [ ] Implement health-gated Caddy swap, one worker per target, lock, recovery and immutable image pulls.
- [ ] Add CI image build after current acceptance; staging deploy opt-in; manual verified production promotion.
- [ ] Run script, workflow and local Docker acceptance; review full diff once.

### Task 4: Operator runbook and real provisioning

**Files:** `docs/deployment/RUNBOOK.md`, `README.md`.

- [ ] Document exact Oracle/DNS/GitHub setup, release, monitoring, backups, restore and rollback.
- [ ] When account access arrives, verify actual tenancy limits, create VM/network/DNS/env/secrets, test backup restoration and deploy staging first.
- [ ] Release production only after staging and recovery gates succeed; record observed URLs and CI runs.
