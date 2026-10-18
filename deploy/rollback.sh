#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/lib.sh"
target_config "${1:?target required}"
load_public_origin
exec 9>"$QMANGA_HOST_ROOT/state/deployment.lock"
flock -n 9 || { echo 'Another deployment or rollback is running' >&2; exit 1; }
active_file="$QMANGA_HOST_ROOT/state/$DEPLOY_TARGET.active"
previous_file="$QMANGA_HOST_ROOT/state/$DEPLOY_TARGET.previous"
test -f "$active_file" && test -f "$previous_file" || {
  echo 'No verified previous release is recorded' >&2; exit 2;
}
read -r active_color active_sha < "$active_file"
read -r previous_color previous_sha < "$previous_file"
valid_sha "$active_sha" && valid_sha "$previous_sha"
test "$active_color" != "$previous_color"
previous_port=$(port_for_color "$previous_color")
port_for_color "$active_color" >/dev/null

# The old stack remains running after a successful release. Do not redirect
# readers to it unless both its API and homepage currently answer.
local_health "$previous_port"
"$QMANGA_REPO_ROOT/deploy/route.sh" "$DEPLOY_TARGET" "$previous_color" normal
active_stopped=false
previous_started=false
restore_active() {
  local result=$?
  trap - EXIT
  if test "$result" -ne 0; then
    "$QMANGA_REPO_ROOT/deploy/route.sh" "$DEPLOY_TARGET" "$active_color" normal || true
    if "$previous_started"; then
      app_compose "$previous_color" "$previous_sha" --profile worker stop worker || true
    fi
    if "$active_stopped"; then
      app_compose "$active_color" "$active_sha" --profile worker up -d --no-deps worker --wait || true
    fi
    echo 'Rollback failed; active route and worker restored' >&2
  fi
  exit "$result"
}
trap restore_active EXIT
public_health
app_compose "$active_color" "$active_sha" --profile worker stop worker
active_stopped=true
app_compose "$previous_color" "$previous_sha" --profile worker up -d --no-deps worker --wait
previous_started=true
public_health
printf '%s %s\n' "$active_color" "$active_sha" > "$previous_file"
state_tmp=$(mktemp "$QMANGA_HOST_ROOT/state/.active.XXXXXX")
printf '%s %s\n' "$previous_color" "$previous_sha" > "$state_tmp"
mv "$state_tmp" "$active_file"
trap - EXIT
echo "$DEPLOY_TARGET rolled back to $previous_sha; database migrations were not reversed"
