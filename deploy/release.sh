#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/lib.sh"
target_config "${1:?target required}"
desired=${2:?tested 40-character SHA required}
valid_sha "$desired" || { echo 'Release SHA must be 40 lowercase hex characters' >&2; exit 2; }
load_public_origin

exec 9>"$QMANGA_HOST_ROOT/state/deployment.lock"
flock -n 9 || { echo 'Another Qmanga deployment or rollback is running' >&2; exit 75; }
git -C "$QMANGA_REPO_ROOT" fetch --quiet origin main
git -C "$QMANGA_REPO_ROOT" cat-file -e "$desired^{commit}"
git -C "$QMANGA_REPO_ROOT" merge-base --is-ancestor "$desired" origin/main || {
  echo 'Refusing a release outside main history' >&2; exit 2;
}
git -C "$QMANGA_REPO_ROOT" checkout --detach --quiet "$desired"

if held=$(hold_sha) && test "$held" = "$desired"; then
  echo "Release $desired is held after a failed or rolled-back deployment; remove $QMANGA_HOST_ROOT/state/deployment-hold to retry" >&2
  exit 3
fi

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
  if git -C "$QMANGA_REPO_ROOT" merge-base --is-ancestor "$desired" "$old_sha"; then
    echo "Refusing to move $DEPLOY_TARGET backwards from $old_sha to $desired" >&2
    exit 2
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
# Staging derives the composition from changed sources since its active
# release. Production promotes the exact staging manifest, so the public
# composition is always the one that passed staging smoke.
old_available=true
if test -n "$old_sha" && ! git -C "$QMANGA_REPO_ROOT" cat-file -e "$old_sha^{commit}" 2>/dev/null; then
  old_available=false
fi
changed_paths=()
changed_text=''
migrations_changed=true
if test "$old_available" = true && test -n "$old_sha"; then
  mapfile -t changed_paths < <(git -C "$QMANGA_REPO_ROOT" diff --name-only "$old_sha" "$desired")
  changed_text=$(printf '%s\n' "${changed_paths[@]}")
  if grep -q '^backend/migrations/' <<<"$changed_text"; then
    migrations_changed=true
  else
    migrations_changed=false
  fi
fi
path_changed() { # regex
  test -n "$changed_text" && grep -qE "$1" <<<"$changed_text"
}
if test "$DEPLOY_TARGET" = production; then
  read -r tag_gateway tag_api tag_web < <(read_manifest_for staging "$desired")
else
  if test "$old_available" = true && test -n "$old_sha"; then
    read -r old_gateway old_api old_web < <(read_manifest "$old_sha")
  else
    old_gateway=$desired old_api=$desired old_web=$desired
  fi
  select_tag() { # component previous-tag regex
    if test -z "$changed_text" || path_changed "$3" || image_present "$1" "$desired"; then
      printf '%s\n' "$desired"
    else
      printf '%s\n' "$2"
    fi
  }
  tag_gateway=$(select_tag gateway "$old_gateway" '^infra/(nginx\.conf|gateway\.Dockerfile)$|^deploy/Caddyfile$')
  tag_api=$(select_tag api "$old_api" '^backend/|^infra/(api\.Dockerfile|imagemagick-policy\.xml)$')
  tag_web=$(select_tag web "$old_web" '^apps/web/|^infra/web\.Dockerfile$')
fi
valid_sha "$tag_gateway" && valid_sha "$tag_api" && valid_sha "$tag_web" || {
  echo 'Release component manifest is invalid' >&2; exit 2;
}

"$QMANGA_REPO_ROOT/deploy/bootstrap.sh" "$DEPLOY_TARGET"

if test "$migrations_changed" = true || ! backup_fresh; then
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
  trap - EXIT TERM INT HUP
  if test "$result" -ne 0; then
    record_hold "$desired"
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
trap 'exit 143' TERM
trap 'exit 130' INT HUP

release_tags --profile worker stop worker >/dev/null 2>&1 || true

# Pull only the immutable images that this release actually introduces.
pull_components=()
if test "$tag_gateway" = "$desired"; then pull_components+=(gateway); fi
if test "$tag_api" = "$desired"; then pull_components+=(api); fi
if test "$tag_web" = "$desired"; then pull_components+=(web); fi
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
trap - EXIT TERM INT HUP
echo "$DEPLOY_TARGET now serves $desired on $new_color (gateway $tag_gateway, api $tag_api, web $tag_web); previous app remains ready for rollback"
