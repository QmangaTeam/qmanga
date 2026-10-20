#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/lib.sh"
target_config "${1:?target required}"
desired=${2:?tested 40-character SHA required}
valid_sha "$desired" || { echo 'Release SHA must be 40 lowercase hex characters' >&2; exit 2; }
load_public_origin

exec 9>"$QMANGA_HOST_ROOT/state/deployment.lock"
flock -n 9 || { echo 'Another Qmanga deployment or rollback is running' >&2; exit 1; }
git -C "$QMANGA_REPO_ROOT" fetch --quiet origin main
git -C "$QMANGA_REPO_ROOT" cat-file -e "$desired^{commit}"
git -C "$QMANGA_REPO_ROOT" merge-base --is-ancestor "$desired" origin/main || {
  echo 'Refusing a release outside main history' >&2; exit 2;
}
git -C "$QMANGA_REPO_ROOT" checkout --detach --quiet "$desired"

active_file="$QMANGA_HOST_ROOT/state/$DEPLOY_TARGET.active"
old_color= old_sha=
if test -f "$active_file"; then
  read -r old_color old_sha < "$active_file"
  valid_sha "$old_sha" && port_for_color "$old_color" >/dev/null || {
    echo 'Active release state is invalid' >&2; exit 2;
  }
  if test "$old_sha" = "$desired"; then
    local_health "$(port_for_color "$old_color")"
    if test "$DEPLOY_TARGET" = staging; then
      "$QMANGA_REPO_ROOT/deploy/route.sh" staging off
    fi
    target_health "$(port_for_color "$old_color")"
    echo "$DEPLOY_TARGET is already healthy on $desired"
    exit 0
  fi
fi

if test "$DEPLOY_TARGET" = production; then
  test -f "$QMANGA_HOST_ROOT/state/staging.active" &&
    grep -Eq "^[a-z]+ $desired$" "$QMANGA_HOST_ROOT/state/staging.active" || {
      echo 'Production SHA must currently serve staging' >&2; exit 2;
    }
  test -f "$QMANGA_HOST_ROOT/state/production.restore-verified" || {
    echo 'Restore has not been verified on staging; production release refused' >&2; exit 2;
  }
fi

# ---- selective component images -------------------------------------------
# Only components whose sources changed (or whose immutable image is already
# present for this SHA) are assigned the new tag. Everything else keeps the
# exact immutable tag that the currently active release runs, so a CSS-only
# change does not rebuild or restart unrelated API/gateway code.
changed_paths=()
if test -n "$old_sha"; then
  mapfile -t changed_paths < <(git -C "$QMANGA_REPO_ROOT" diff --name-only "$old_sha" "$desired")
fi
path_changed() {
  test -n "$old_sha" || return 0
  printf '%s\n' "${changed_paths[@]}" | grep -qE "$1"
}
select_tag() { # component
  local component=$1
  if test -z "$old_sha" || path_changed "$2" || image_present "$component" "$desired"; then
    printf '%s\n' "$desired"
  else
    printf '%s\n' "$old_sha"
  fi
}
tag_gateway=$(select_tag gateway '^infra/(nginx\.conf|gateway\.Dockerfile)$|^deploy/Caddyfile$')
tag_api=$(select_tag api '^backend/|^infra/(api\.Dockerfile|imagemagick-policy\.xml)$')
tag_web=$(select_tag web '^apps/web/|^infra/web\.Dockerfile$')

migrations_changed=false
if test -n "$old_sha" && printf '%s\n' "${changed_paths[@]}" | grep -q '^backend/migrations/'; then
  migrations_changed=true
fi

if test "$migrations_changed" = true || test -z "$old_sha"; then
  "$QMANGA_REPO_ROOT/deploy/backup.sh" "$DEPLOY_TARGET"
elif test "$DEPLOY_TARGET" = production && ! backup_fresh; then
  echo "No verified backup younger than ${QMANGA_MAX_BACKUP_AGE_H:-30}h; taking one before release" >&2
  "$QMANGA_REPO_ROOT/deploy/backup.sh" "$DEPLOY_TARGET"
else
  echo "Skipping full backup: no migrations and a verified ${DEPLOY_TARGET} backup is younger than ${QMANGA_MAX_BACKUP_AGE_H:-30}h"
fi

if test "$old_color" = blue; then new_color=green; else new_color=blue; fi
new_port=$(port_for_color "$new_color")
switched=false
new_worker=false
old_worker_stopped=false
release_tags() {
  IMAGE_TAG_GATEWAY=$tag_gateway IMAGE_TAG_API=$tag_api IMAGE_TAG_WEB=$tag_web \
    app_compose "$new_color" "$desired" "$@"
}
restore_previous() {
  local result=$?
  trap - EXIT
  if test "$result" -ne 0; then
    if "$new_worker"; then
      release_tags --profile worker stop worker || true
    fi
    if "$switched"; then
      if test -n "$old_sha"; then
        "$QMANGA_REPO_ROOT/deploy/route.sh" "$DEPLOY_TARGET" "$old_color" normal || true
      else
        "$QMANGA_REPO_ROOT/deploy/route.sh" "$DEPLOY_TARGET" off || true
      fi
    fi
    if "$old_worker_stopped"; then
      app_compose "$old_color" "$old_sha" --profile worker up -d --no-deps worker --wait || true
    fi
    echo "Release $desired failed; previous $DEPLOY_TARGET route retained/restored" >&2
  fi
  exit "$result"
}
trap restore_previous EXIT

"$QMANGA_REPO_ROOT/deploy/bootstrap.sh" "$DEPLOY_TARGET"
release_tags --profile worker stop worker >/dev/null 2>&1 || true

# Pull only the immutable images that this release actually introduces.
pull_components=()
test "$tag_gateway" = "$desired" && pull_components+=(gateway)
test "$tag_api" = "$desired" && pull_components+=(api)
test "$tag_web" = "$desired" && pull_components+=(web)
if test ${#pull_components[@]} -gt 0; then
  release_tags pull "${pull_components[@]}"
fi
release_tags up -d --wait --wait-timeout 180 gateway api web
local_health "$new_port"

"$QMANGA_REPO_ROOT/deploy/route.sh" "$DEPLOY_TARGET" "$new_color" normal
switched=true
healthy=false
for _ in {1..18}; do
  if target_health "$new_port"; then healthy=true; break; fi
  sleep 5
done
"$healthy" || { echo 'Target smoke failed after route switch' >&2; exit 1; }

if test -n "$old_sha"; then
  app_compose "$old_color" "$old_sha" --profile worker stop worker
  old_worker_stopped=true
fi
release_tags --profile worker up -d --no-deps worker --wait
new_worker=true
target_health "$new_port"

# Confirm the public page renders this deployment, not a cached predecessor.
homepage=$(served_homepage "$new_port")
grep -q "$desired" <<< "$homepage" || {
  echo "Homepage did not render release $desired" >&2; exit 1;
}

if test -n "$old_sha"; then
  printf '%s %s\n' "$old_color" "$old_sha" > "$QMANGA_HOST_ROOT/state/$DEPLOY_TARGET.previous"
fi
manifest=$(manifest_path "$desired")
mkdir -p "$(dirname "$manifest")"
manifest_tmp=$(mktemp "$(dirname "$manifest")/.manifest.XXXXXX")
printf 'gateway %s\napi %s\nweb %s\n' "$tag_gateway" "$tag_api" "$tag_web" > "$manifest_tmp"
mv "$manifest_tmp" "$manifest"
state_tmp=$(mktemp "$QMANGA_HOST_ROOT/state/.active.XXXXXX")
printf '%s %s\n' "$new_color" "$desired" > "$state_tmp"
mv "$state_tmp" "$active_file"
trap - EXIT
echo "$DEPLOY_TARGET now serves $desired on $new_color (gateway $tag_gateway, api $tag_api, web $tag_web); previous app remains ready for rollback"
