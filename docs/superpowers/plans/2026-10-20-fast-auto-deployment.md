# Fast Auto Deployment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** After `git push origin main`, ordinary changes reach production (qmanga.maqsatto.dev serving the pushed SHA) auto – without manual SSH and without full test suite gating – in ≤ 300 s (median target 120 s).

**Architecture:** Reuse the existing stack: `verify.yml` runs full acceptance/publishes arm64 images (stays as Full Verification, non-blocking on push), a new `release.yml` publishes only changed component images (fast path) and advances a git ref `refs/deploy-ready`. The VM outbound poller (`deploy/poll-staging.sh`, timer 20–30 s) detects the ref and runs through the existing `release.sh` blue-green staging→production path with per-component image manifests and backup-on-freshness policy.

**Tech Stack:** GitHub Actions, Docker Buildx with GHA cache, GHCR arm64 images, compose blue/green (existing), nginx/gateway, systemd timer (existing).

**Spec:** owner request (chat, 2026-10-20): fast path ≤300 s auto; frontend checks fully removed from deploy critical path (owner explicit instruction); Full Verification separate non-blocking; selective builds relative to last released state; auto-risk split: migrations/deploy infra → slow path gated by full verify; no mutable `latest`; race-safe `deploy-ready` monotonic advance; keep blue-green, deployment lock, restore gate, backup guard, rollback; no schema rollback; measurements push→public production.

## Global Constraints

- Only outbound connections from the VM; open no new public SSH ports or firewall rules.
- Never disable: deployment lock, restore gate (`production.restore-verified`), backup guard, CSRF/auth, SSH host-key validation, PostgreSQL/MinIO protection.
- Immutable `:40-hex-SHA` tags only; no mutable `latest` as source of truth; no incompatible mixing of API/images versions (unchanged component reuses its predecessor's tag deterministically).
- Regular production backup timer (~02:30 UTC) untouched; migrations ⇒ mandatory fresh backup; non-migration release requires a fresh verified backup (≤ MAX_BACKUP_AGE_H=30) else full backup runs (or release aborts).
- Automatic rollback keeps previous color/SHA on failure; no automatic DB schema rollback; force-push/outside-main-history SHA denied (existing ancestry check).
- Do not store secrets in repo or reports; `GITHUB_TOKEN` only.
- Full suite (`local-acceptance`) must not re-run unnecessarily for the same SHA.
- The final event is public production serving the released SHA (homepage renders `BUILD_VERSION` marker) — not merely a successful image push.

## Review Focus

(ref `deploy-ready` advance modes: only descendant of its current value and only from CI workflow — a cancelled workflow must not block later SHAs, mid-build cancel must leave no half manifest. Race of two close pushes — second push wins/cancels first, and no scenario may stage a mixed composition of stale api + new web unintentionally. Freshness policy must refuse release gracefully if stale backup and no way to make one (lines 3–5 styled).)
(Failure-mode tests pinned in tasks 4 and 6 below.)

- A concurrent second push during a deployment must never deploy a new SHA mixed with the first's component tags unless explicitly unchanged components of the *same* release → policed by per-release manifest computed server-side from repo history (Task 2/3) and `flock`.
- A cancelled first build with a failing second… actually cancelled first run must not leave a partial `deploy-ready` advancement: ref is updated only after all push steps complete, in one final guarded step (Task 5).
- Same-SHA re-run must be a no-op (existing early-exit, Task 2).
- Non-CI attempt to advance `refs/deploy-ready` is excluded: only the workflow pushes it, local git pushes by other identities can't authenticate; else guard `scripts/ci` checks sha descendant-of (Task 5 verification).
- Freshness check must fail conservatively if `COMPLETE` marker missing (Task 4).

## Baseline (measured 2026-10-20, runs 39335146020 / 39333412292)

- local-acceptance (2 shards): 300–390 s (cold start+build ~110–120 s; Go+postgres+codecs ~55–70 s; smoke ~15–30 s; browser ~55–80 s)
- publish-images: gateway 18–21 s, web 51–56 s, api 64–65 s (no cache)
- Staging poller: `OnUnitInactiveSec=5min` → 0–5 min wait
- Production: manual `release.sh production`, full backup on every release
- Push → production today: ~10–15 min of waiting + manual step. Prototype cuts the (no-migration) acceptance+wait layer.

---

### Task 1: web.Dockerfile — drop frontend checks from image build

**Files:**
- Modify: `infra/web.Dockerfile:7`

**Interfaces:** Produces `ghcr.io/maqsatto/qmanga-web:<sha>` containing only `build/` output (`bun run build`).

- [ ] Step 1: Replace `RUN bun run check && bun run check:architecture && bun run check:ui-contrast && bun test src/lib/features src/lib/shared src/lib/server && bun run test:ui && bun run check:ui-preview && bun run build` with `RUN bun run build`
- [ ] Step 2: Verify local docker build still succeeds and tests still run via verify.yml (acceptance unchanged there): `docker compose --env-file deploy/host.env.example -f deploy/compose.app.yaml config -q` unaffected; run web-only build `docker build -f infra/web.Dockerfile .` if docker available locally
- [ ] Step 3: Commit

### Task 2: Per-component image tags (release manifest) in compose

**Files:**
- Modify: `deploy/compose.app.yaml`

**Interfaces:**
- Produces env vars consumed by `release.sh`: `IMAGE_TAG_GATEWAY`, `IMAGE_TAG_API`, `IMAGE_TAG_WEB` (each replace `RELEASE_SHA` in image:- sanitizes: missing tag ⇒ compose error), plus `BUILD_VERSION: ${IMAGE_TAG_WEB}` on web (truthful served version).
- Consumes: nothing else changes.

- [ ] Step 1: For each of gateway/api/web/worker replace `image:` tag `${RELEASE_SHA:?…}` with `${IMAGE_TAG_GATEWAY:?…}` / `${IMAGE_TAG_API:?…}` / `${IMAGE_TAG_WEB:?…}` (worker shares api image); web env `BUILD_VERSION: ${IMAGE_TAG_WEB}`
- [ ] Step 2: `docker compose --env-file deploy/host.env.example -f deploy/compose.app.yaml config --quiet` with env vars set (release.sh-level integration verified in Task 3)
- [ ] Step 3: Commit

### Task 3: release.sh — component manifest, backup freshness, no-op and web-only jobs

**Files:**
- Modify: `deploy/release.sh`

**Interfaces:**
- Consumes: compose env vars from Task 2 (`IMAGE_TAG_*`), existing state files (`staging.active`, `production.active`, `*.previous`, backups `COMPLETE` markers).
- Produces: for each component c ∈ {gateway api web} an `IMAGE_TAG_<c>` selection rule:
  - tag = `desired` if `git diff --name-only "$old_sha" "$desired"` touches that component's paths (see map below) or if the image `$PREFIX-$c:$desired` already exists locally; else tag = `old_sha` (deterministic minimal re-use of the running predecessor image), pulled no more than needed.
  - Component path map: **web** → `apps/web/`; **gateway** → `deploy/Caddyfile`, `infra/nginx.conf`, `infra/gateway.Dockerfile`; **api** → `backend/` excluding `backend/migrations/` but treating `backend/migrations/` separately (slow/backup trigger below).
  - Empty `old_sha` → all three use `desired` (first deploy).
- Backup policy (production target): run full `backup.sh` only if (a) `backend/migrations/` files changed old→new, or (b) `newest COMPLETE marker under /opt/qmanga/backups/production is older than 30h or absent` (MAX_BACKUP_AGE_H env, default 30); else skip full backup. Staging target: full backup only on migrations (existing cadence acceptable; never disabled). Migrations/destructive ⇒ always full.
- web-only release: still deploys a new color with only the web tag pointing at `desired`.

- [ ] Step 1: Add helper `changed_components old new` emitting `gateway|api|web` names per map; empty `old` ⇒ emit all
- [ ] Step 2: Replace the unconditional `backup.sh` line with the policy above (impl stays inline; no new deps)
- [ ] Step 3: Replace `app_compose … pull/up` invocations with per-component env (`IMAGE_TAG_GATEWAY=… IMAGE_TAG_API=… IMAGE_TAG_WEB=… app_compose "$new_color" "$desired" …`), and pull only components whose tag changed (`--pull missing` semantics already minimize traffic; verified in compose `pull_policy: missing`)
- [ ] Step 4: Worker handling unchanged (worker image tag = `IMAGE_TAG_API`)
- [ ] Step 5: Verification: `bash -n deploy/release.sh`; locally exercise `$XDG_RUNTIME_DIR`-free dry run? — real exercise is Task 10 (prototype); no unit tests exist for release.sh (infra shell)
- [ ] Step 6: Commit

### Task 4: Backup freshness helper + conservative failure

**Files:**
- Modify: `deploy/lib.sh` (add `fresh_backup_max_seconds` reader + `newest_backup_age_seconds`); `deploy/release.sh` (consume)

**Interfaces:**
- Produces: `backup_fresh_ok target` returning 0 iff newest `COMPLETE` marker under backups/<target> is ≤ `${QMANGA_MAX_BACKUP_AGE_H:-30}` hours; conservative FAILURE (nonzero) when the marker is missing.
- Consumed by: Task 3's policy.

- [ ] Step 1: Add to `lib.sh`: `newest_backup_age_seconds() { local d file; d="$QMANGA_HOST_ROOT/backups/$DEPLOY_TARGET"; file=$(find "$d" -name COMPLETE -readable -type f 2>/dev/null | sort | tail -n 1); if test -z "$file"; then return 3; fi; echo $(( $(date +%s) - $(stat -mt +%s "$file" 2>/dev/null || stat -c %Y "$file" 2>/dev/null || date +%s) )); }`
- [ ] Step 2: In `release.sh` policy (Task 3 Step 2): `backup_fresh_ok_staging_pick(){ if newest_backup_age_seconds | awk '{exit $1>30*3600?1:0}' || age-missing; then full backup; fi }` — concretized in the script, no new behavior
- [ ] Step 3: Verification: `bash -n deploy/lib.sh deploy/release.sh`
- [ ] Step 4: Commit (may merge with Task 3 in one commit if cleaner)

### Task 5: fast-release.yml — fast path workflow

**Files:**
- Create: `.github/workflows/release.yml`
- Create: `scripts/advance-deploy-ready.sh`

**Interfaces:**
- Trigger: `push` to `main`; `concurrency.group: release-main`, `cancel-in-progress: true` (latest push wins).
- Classification vs `refs/deploy-ready` (fallback `HEAD~1`…in practice exists): `github/git` `dispatch`. Inputs none.
- Fast class: changed paths only under `apps/web/**` and/or `backend/**` with no files in `backend/migrations/**`, `deploy/**`, `infra/**`, `.github/**`, `bun.lock` (non-frontend) and not **both** app trees (web+backend is still fast — both are app code; migrations/deploy/infra are slow); slow class: any change under `backend/migrations/**`, `deploy/**`, `infra/**`, `.github/**` ⇒ wait for Verify Qmanga (verify.yml) success for the SHA (poll `gh api …/runs?head_sha=…` every 30 s up to 45 min) then advance; if missing ⇒ exit 1.
- Fast class: build each changed component with `docker/build-push-action@v6`, `runners: ubuntu-24.04-arm`, `cache-from: type=gha`, `cache-to type=gha,mode=max`, tag `ghcr.io/maqsatto/qmanga-<component>:${{ github.sha }}`, push.
- Final step (both classes): `scripts/advance-deploy-ready.sh $SHA`:
  1. `git ls-remote origin refs/deploy-ready` → current (may be empty)
  2. if not empty: `git merge-base --is-ancestor current $SHA` (with both refs fetched); if already ancestor and `git rev-parse $SHA != current`… specification: only advance when `desired` is a descendant of current and differ; if current is a descendant of `desired` (race loser) → skip
  3. `git push --force-with-lease=refs/deploy-ready:<current> origin $SHA:refs/deploy-ready` (empty current ⇒ plain push)
- permissions: `contents: write`; `persist-credentials: true` for the push step only.

- [ ] Step 1: Write `scripts/advance-deploy-ready.sh` exactly to the guard above; `bash -n`
- [ ] Step 2: Write `.github/workflows/release.yml` classification (`git diff --name-only refs/heads/deploy-ready...HEAD` fetched defensively; slow class ⇒ wait-for-verify job built on gh CLI polling with `GITHUB_TOKEN` `actions: read`)
- [ ] Step 3: Local: `bash -n`; `docker compose … config -q` unaffected
- [ ] Step 4: Commit; monitor first real run (self-satisfying since push includes `.github/**` ⇒ slow class ⇒ waits for the very verify run on this push, this validates both halves)
- [ ] Step 5: Commit any latency fix observed (site live in Task 10)

### Task 6: pollstaging — deploy-ready-based poller

**Files:**
- Modify: `deploy/poll-staging.sh`

**Interfaces:**
- Consumes: `refs/deploy-ready`, `release.sh` (Tasks 2–3), `lib.sh` `valid_sha`.
- Produces: automatic oneshot staging→chain: if `staging.active` ≠ desired → `release.sh staging $desired`; then if `production.active` ≠ same desired → `release.sh production $desired` (release.sh itself enforces staging-first and lock; production auto-promotion is the confirmed owner decision).
- No GitHub REST API calls anymore → no rate limits.

- [ ] Step 1: Rewrite poll-staging.sh: resolve `git ls-remote origin refs/deploy-ready | awk '{print $1}'` (origin fetch of exact ref via `git fetch origin refs/deploy-ready:refs/remote/deploy-ready`, then `rev-parse`); early-exit if active==desired
- [ ] Step 2: staging check+deploy (release.sh), then production check+deploy in same cycle (both no-op on same SHA)
- [ ] Step 3: `bash -n`; retention of the previous 5-min behavior only as fallback comment in docs, not code
- [ ] Step 4: Commit

### Task 7: Timer 20–30 s

**Files:**
- Modify: `deploy/qmanga-staging-poll.timer` (`OnUnitInactiveSec=20s` through `OnBootSec=2min`)
- One-time ops step (documented in docs): server-side `cp deploy/qmanga-staging-poll.timer … && systemctl daemon-reload && systemctl restart qmanga-staging-poll.timer`

- [ ] Step 1: Update file header Comment (was "every five minutes" → 20 s poll)
- [ ] Step 2: Commit (server-side application is an owner-visible SSH step; guarded but allowed since already agreed)
- [ ] Step 3: Ops execution recorded in [OPERATIONS.md](../../docs/deployment/OPERATIONS.md)

### Task 8: verify.yml — Full Verification continuing job (non-blocking, no dead deploy)

**Files:**
- Modify: `.github/workflows/verify.yml`

**Interfaces:**
- orchestrator: `concurrency` group per-ref stays; `cancel-in-progress: true` only for `github.event_name == 'push'` context (matrices are all-or… cannot use `if` on group… check YAML alternative: use one group `verify-${{ github.ref }}` with `cancel-in-progress: ${{ github.event_name == 'push' }}`… valid as workflow level expression; keep `false` as default)
- Remove the dead SSH `deploy-staging` job (job `deploy-staging` in current file, lines 146–171) — the poller replaces it; this closes the "two deploy paths" rule for `vars.QMANGA_STAGING_ENABLED`.
- publish-images / local-acceptance untouched (full verification, arm64, restore of coverage for migrations/infra path).

- [ ] Step 1: preference edit per above
- [ ] Step 2: Commit
- [ ] Step 3: Note: publish-images matrix keeps publishing all three components per SHA (redundant fast-class coverage but harmless; only slow-SHA builds depend on it)

### Task 9: Docs

**Files:**
- Modify: `docs/deployment/OPERATIONS.md` (CI → staging → production section: new fast path, auto-promotion, backup freshness policy, deploy-ready ref; poller no longer 5-min REST)
- Modify: `docs/deployment/RUNBOOK.md` (deployment gate sections)
- Modify: root `AGENTS.md` release contract line ("после …30SHA" reference immutability; auto-promote mention) — only if trivially placeable
- Update the backlog card via CLI not manual markdown

- [ ] Step 1: Draft; push-referenced SHA values where known
- [ ] Step 2: Commit

### Task 10: Prototype measurement + 3 acceptance releases (verification of the whole path)

**Files:**
- No doc; measurements into cards/logs (ignored `tmp/` allowed), report in final answer + OPERATIONS.md

- [ ] Step 1: SSH deploy of new repo code + timer reload (one-time)
- [ ] Step 2: Release #1 — docs-only (words) → expect no deploy of nothing (ref unchanged, production unchanged) — verify nothing actually advanced on VM (`production.active` SHA unchanged)
- [ ] Step 3: Release #2 — CSS-only under apps/web → verify web-only composition: `IMAGE_TAG_API` = prior SHA, `IMAGE_TAG_WEB` = current SHA; measure durations of: workflow run, ref advance, poller staging deploy, production switch; public health/home page
- [ ] Step 4: Release #3 — Go-only backend no-migration change → same measure; verify staging+production smoke
- [ ] Step 5: Failure drill — introduce deliberately failing new release (a bad health check e.g. `curl` on a bad URL passed through, or new web page broken) ⇒ verify automatic rollback retains previous `production.active` and public health
- [ ] Step 6: Double-push — two commits pushed B then A within seconds… pushed A then B close together ⇒ B's run cancels A; verify network shows only B's SHA in production & AA rollback marker intact
- [ ] Step 7: Update timing table in docs/cards

## Execution notes

- Worktree not needed: single pipeline workstream on main, owner already coordinates.
- Docker on this laptop may require sudo; docker desktop compose config check okay without sudo.
- Tests availability: `bash -n` only for deploy scripts (no test framework for these) — release/CI measurements are the tests.
- Task boundaries same-session; subagent review after tasks 1–8 ready.

## Self-review notes

- Spec coverage: fast path (T5/T6/T7), selective builds (T3), manifest (T2), web.Dockerfile (T1), backup freshness (T3/T4), auto-promotion (T6), rollback (existing release.sh mechanism, T10 step 5), docs (T9), no-latest rule honored, deploy-ready race handled (T5).
- No failing tests for new script except live prototype tasks (T10) — limitation accepted by owner (deploy infra; no harness).
- Proportional: plan is descriptive rather than exhaustive — mirrors spec's constraints verbatim.
