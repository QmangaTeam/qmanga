#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/lib.sh"
target_config staging

# The VM only makes outbound requests. A release is eligible only when the
# trusted CI workflow advanced refs/deploy-ready: that ref moves exclusively
# after the immutable ARM64 images for the exact SHA were published. Git ref
# polling stays far below GitHub API rate limits and needs no private key.
exec 8>"$QMANGA_HOST_ROOT/state/staging-poll.lock"
flock -n 8 || exit 0

git -C "$QMANGA_REPO_ROOT" fetch --quiet --force \
  origin '+refs/deploy-ready:refs/remotes/origin/deploy-ready' || true
desired=$(git -C "$QMANGA_REPO_ROOT" rev-parse --verify --quiet refs/remotes/origin/deploy-ready || true)
if ! valid_sha "$desired"; then
  echo 'No release has been marked ready yet'
  exit 0
fi

failed_within() { # file seconds
  local file="$QMANGA_HOST_ROOT/state/${1:?file required}" failed_sha failed_at now
  test -f "$file" || return 1
  read -r failed_sha failed_at < "$file"
  test "$failed_sha" = "$desired" || return 1
  [[ "$failed_at" =~ ^[0-9]+$ ]] || return 1
  now=$(date +%s)
  test $((now - failed_at)) -lt "${2:?seconds required}" && return 0
  return 1
}
record_failure() { # file
  printf '%s %s\n' "$desired" "$(date +%s)" > "$QMANGA_HOST_ROOT/state/${1:?file required}"
}

staging_done=false
if grep -Eq "^[a-z]+ $desired$" "$QMANGA_HOST_ROOT/state/staging.active" 2>/dev/null; then
  staging_done=true
fi
production_done=false
if grep -Eq "^[a-z]+ $desired$" "$QMANGA_HOST_ROOT/state/production.active" 2>/dev/null; then
  production_done=true
fi
if "$staging_done" && "$production_done"; then
  exit 0
fi

# release.sh takes the deployment lock on its own descriptor; holding this
# poller descriptor across a child process would make that lock conflict.
exec 8>&-

if ! "$staging_done"; then
  if failed_within staging.last-failure 1800; then
    echo "Staging release $desired failed recently; not retrying yet"
    exit 0
  fi
  if ! bash "$QMANGA_REPO_ROOT/deploy/release.sh" staging "$desired"; then
    record_failure staging.last-failure
    exit 1
  fi
fi

# Staging smoke passed in the same release run: production is promoted
# automatically for this exact SHA. The production release re-checks the
# staging state, restore gate, backup freshness and health with auto-rollback.
if ! "$production_done"; then
  if failed_within production.last-failure 1800; then
    echo "Production release $desired failed recently; not retrying yet"
    exit 0
  fi
  if ! bash "$QMANGA_REPO_ROOT/deploy/release.sh" production "$desired"; then
    record_failure production.last-failure
    exit 1
  fi
fi
