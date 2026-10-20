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

# A failed or rolled-back SHA stays held until deploy-ready moves elsewhere.
if held=$(hold_sha) 2>/dev/null && test "$held" = "$desired"; then
  echo "Release $desired is held after a failed or rolled-back deployment"
  exit 0
fi

active_sha_of() { # state file
  awk '{print $2; exit}' "$1" 2>/dev/null || true
}
newer_already_active() { # target state file
  local active
  active=$(active_sha_of "$1")
  valid_sha "$active" && {
    test "$active" = "$desired" ||
      git -C "$QMANGA_REPO_ROOT" merge-base --is-ancestor "$desired" "$active"
  }
}

staging_done=false
if newer_already_active "$QMANGA_HOST_ROOT/state/staging.active"; then
  staging_done=true
fi
production_done=false
if newer_already_active "$QMANGA_HOST_ROOT/state/production.active"; then
  production_done=true
fi
if "$staging_done" && "$production_done"; then
  exit 0
fi

# release.sh takes the deployment lock on its own descriptor; holding this
# poller descriptor across a child process would make that lock conflict.
exec 8>&-

run_release() { # target
  local code
  if bash "$QMANGA_REPO_ROOT/deploy/release.sh" "$1" "$desired"; then
    return 0
  else
    code=$?
  fi
  if locked_release "$code"; then
    echo "Deployment lock is busy; retrying on the next poll"
    return 0
  fi
  record_hold "$desired"
  echo "Release $desired failed for $1; deployment held until a newer release is ready" >&2
  return 1
}

if test "$staging_done" = false; then
  run_release staging || exit 1
  # release.sh checks out the released commit, which may be a newer poller.
  # Re-exec so the remaining phase always runs the code from that release.
  exec bash "$QMANGA_REPO_ROOT/deploy/poll-staging.sh"
fi

# Staging smoke passed in the same release run: production is promoted
# automatically for this exact SHA. The production release re-checks the
# staging state, restore gate, backup freshness and health with auto-rollback.
if test "$production_done" = false; then
  run_release production || exit 1
  exec bash "$QMANGA_REPO_ROOT/deploy/poll-staging.sh"
fi
