# Qmanga deployment runbook

The public names are `qmanga-dev.maqsatto.dev` (staging) and `qmanga.maqsatto.dev` (production). Both may share one Oracle Ampere A1 host, but use separate Compose projects, PostgreSQL databases, MinIO volumes and credentials. This is **one-host fault tolerance during application updates**, not high availability if Oracle, the VM, its disk or the network fails.

## Before the first release

1. In the Oracle home region, check actual Always Free capacity and create an **ARM64** Ubuntu VM with at least 2 OCPU, 12 GB RAM and enough persistent disk for two sites, two warm application versions and two backups. Reserve a stable public IP if available. Keep a separate off-host backup destination; a second folder on this VM is insufficient.
2. Set OCI network rules and host firewall for inbound TCP 80/443 from the internet and TCP 22 only from the operator's allowed source. Do not expose PostgreSQL, MinIO, API or gateway ports. Install Docker Engine with Compose v2, Git, curl, sha256sum and `flock`. Create a dedicated deploy account and install its **public** SSH key. Docker group access effectively grants root-equivalent host control; restrict this account and key.
3. Create DNS A records for both names pointing to the VM public IP. Do not add AAAA records unless IPv6 is configured and reachable. Confirm both names resolve before expecting TLS issuance.
4. Under `/opt/qmanga`, create `repo`, `config`, `state/sites`, `backups`, `bin` and `hooks`. Clone the public repository into `repo`, owned by the deploy account. Copy `deploy/host.env.example` to `config/staging.env` and `config/production.env`, adjusting target and origin. Generate distinct high-entropy alphanumeric secrets for each environment. Set owner-only permissions (`chmod 600`) on both files. Configure real STARTTLS SMTP and a working sender address. Keep all secrets outside Git and GitHub logs.
5. Authenticate the deploy account to GHCR with a read-only `read:packages` token using `docker login ghcr.io`; this is needed if the packages are private. Store Docker credentials in the account's protected home. Keep the machine's SSH **private** key outside GitHub except the environment secret described below. Pin the VM's SSH host key fingerprint out of band.
6. In GitHub, create `staging` and `production` environments. Set repository/environment variables `QMANGA_DEPLOY_HOST`, `QMANGA_DEPLOY_USER` and repository variable `QMANGA_STAGING_ENABLED=true` only after the host is ready. Set secrets `QMANGA_DEPLOY_KEY` (dedicated private deployment key) and `QMANGA_SSH_KNOWN_HOSTS` (exact verified host-key line). Add required reviewers to `production` if available for this repository/account. Never paste these values in a chat or issue.
7. Configure an executable `/opt/qmanga/hooks/offsite-backup` accepting one local backup-directory argument. It must copy both archives and `SHA256SUMS` to storage outside the VM, verify bytes/checksums at the destination, and return nonzero on failure. The production release refuses to run without it. Arrange regular backups independently of releases too; deployment backups alone do not protect against failures between updates.

`deploy/bootstrap.sh staging` starts staging PostgreSQL and MinIO, creates private buckets and a restricted application account. `release.sh` invokes bootstrap itself. The host must be `aarch64`; the scripts refuse another architecture. `config/*.env` files are shell-sourced by the owner-controlled deployment scripts, so only the deploy account may edit them.

## Release flow

Push to `main` runs the existing full acceptance suite. Only a successful run publishes API, web and gateway Linux/arm64 images to GHCR, tagged with the exact commit SHA. If staging is enabled, the same workflow calls `deploy/release.sh staging SHA` through SSH with a pinned host key. The release keeps the active color serving while it pulls and starts the candidate, checks local API/homepage health, switches Caddy, and checks public HTTPS. Failure before switch leaves the old route; failure afterward restores it. The old application stack remains warm. One worker runs per environment. Database migrations occur when the candidate API starts, so **every new migration must remain compatible with the currently active application for the rollback window**.

Observe the GitHub `Verify Qmanga` run and verify `https://qmanga-dev.maqsatto.dev/api/v1/health`, the homepage, login, image reading, and staff publishing on staging. To promote, run `Promote Qmanga` with the exact 40-character SHA. The workflow requires a successful push CI run and the same SHA in the staging active-state file. The production environment can require a reviewer. Promotion creates a fresh backup, invokes the off-host hook, and uses the same health-gated switch at `https://qmanga.maqsatto.dev`.

Manual host commands, when investigating a failed automatic run:

```sh
bash /opt/qmanga/repo/deploy/release.sh staging <tested-40-character-sha>
bash /opt/qmanga/repo/deploy/rollback.sh staging
bash /opt/qmanga/repo/deploy/rollback.sh production
docker compose -f /opt/qmanga/repo/deploy/compose.front.yaml -p qmanga-front logs --tail=100 caddy
```

Rollback switches only application containers and traffic. It **does not undo database migrations or user writes**. If a migration is incompatible, restore requires a separate maintenance decision and may lose writes since the backup. Never run `docker compose down -v` on these projects.

## Recovery and monitoring

Each backup contains `postgres.dump`, `objects.tar`, `SHA256SUMS` and `COMPLETE`. It is made without pausing readers or writers: the DB snapshot is taken first, then objects are copied. This relies on Qmanga's current immutable S3 object keys and lack of physical deletion. **Change the backup procedure before implementing overwrite or physical object deletion.** An incomplete backup has no `COMPLETE` marker; a failed backup blocks promotion. Keep enough free disk for the largest backup and incoming images.

Before production opening, restore one staging backup into **new, isolated test volumes and a new database**, validate `sha256sum -c SHA256SUMS`, run `pg_restore` against the test database, import both object folders from `objects.tar` into test MinIO, compare bucket object counts and fetch sample source/derivative objects. Then start the tested application against that restored environment, verify login, catalogue, reader and publishing, and record the tested backup/SHA/time. Only after this real drill create `/opt/qmanga/state/production.restore-verified`. Repeat the drill regularly and whenever the backup format or storage layer changes. Never restore over live staging or production volumes for a drill.

Check both public `/api/v1/health` URLs, homepage and TLS certificates with external monitoring every few minutes. Alert on failures, disk usage, Docker restarts, backup-hook failures and certificate renewal. SSH access, CI, GHCR, DNS, TLS, SMTP and off-host backup are all operational dependencies. The first real release must be observed on the VM and in a browser; configuration validation alone does not establish uptime.
