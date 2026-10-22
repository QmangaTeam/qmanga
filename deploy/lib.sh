#!/usr/bin/env bash
set -euo pipefail

QMANGA_REPO_ROOT=${QMANGA_REPO_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}
QMANGA_HOST_ROOT=${QMANGA_HOST_ROOT:-/opt/qmanga}
IMAGE_PREFIX=${IMAGE_PREFIX:-ghcr.io/qmangateam/qmanga}
export QMANGA_HOST_ROOT IMAGE_PREFIX

target_config() {
  DEPLOY_TARGET=${1:?staging or production required}
  case "$DEPLOY_TARGET" in
    staging) DOMAIN=qmanga-dev.maqsatto.dev; BLUE_PORT=18280; GREEN_PORT=18281 ;;
    production) DOMAIN=qmanga.maqsatto.dev; BLUE_PORT=18180; GREEN_PORT=18181 ;;
    *) echo 'Target must be staging or production' >&2; return 2 ;;
  esac
  CONFIG_FILE="$QMANGA_HOST_ROOT/config/$DEPLOY_TARGET.env"
  test -f "$CONFIG_FILE" || { echo "Missing $CONFIG_FILE" >&2; return 2; }
  mkdir -p "$QMANGA_HOST_ROOT/state/sites" "$QMANGA_HOST_ROOT/backups/$DEPLOY_TARGET"
  export DEPLOY_TARGET
}

port_for_color() {
  case "${1:?color required}" in
    blue) printf '%s\n' "$BLUE_PORT" ;;
    green) printf '%s\n' "$GREEN_PORT" ;;
    *) echo 'Color must be blue or green' >&2; return 2 ;;
  esac
}

valid_sha() { [[ ${1:-} =~ ^[0-9a-f]{40}$ ]]; }

data_compose() {
  docker compose --env-file "$CONFIG_FILE" -f "$QMANGA_REPO_ROOT/deploy/compose.data.yaml" \
    -p "qmanga-$DEPLOY_TARGET-data" "$@"
}

manifest_path() {
  printf '%s\n' "$QMANGA_HOST_ROOT/state/manifests/$DEPLOY_TARGET-${1:?sha required}"
}

manifest_path_for() {
  printf '%s\n' "$QMANGA_HOST_ROOT/state/manifests/${1:?target required}-${2:?sha required}"
}

# A release may reuse an unchanged component's immutable image from its
# predecessor. The manifest records the exact tag per component; tagged
# releases before manifests existed fall back to their release SHA.
read_manifest_file() {
  local file=${1:?manifest required} fallback=${2:?sha required} gateway api web
  if test -f "$file"; then
    gateway=$(awk '$1 == "gateway" {print $2; exit}' "$file")
    api=$(awk '$1 == "api" {print $2; exit}' "$file")
    web=$(awk '$1 == "web" {print $2; exit}' "$file")
    if valid_sha "$gateway" && valid_sha "$api" && valid_sha "$web"; then
      printf '%s %s %s\n' "$gateway" "$api" "$web"
      return
    fi
  fi
  printf '%s %s %s\n' "$fallback" "$fallback" "$fallback"
}

read_manifest() { # sha -> "gateway api web" tags
  read_manifest_file "$(manifest_path "${1:?sha required}")" "$1"
}

read_manifest_for() { # target sha -> "gateway api web" tags
  local target=${1:?target required} sha=${2:?sha required}
  read_manifest_file "$(manifest_path_for "$target" "$sha")" "$sha"
}

image_present() { # component sha
  docker image inspect "$IMAGE_PREFIX-${1:?component required}:${2:?sha required}" >/dev/null 2>&1
}

# A locally missing image may still be published. Never select a tag that does
# not exist: fall back to the previous working component instead.
image_available() { # component sha
  image_present "$1" "$2" ||
    timeout 10 docker manifest inspect "$IMAGE_PREFIX-$1:$2" >/dev/null 2>&1
}

app_compose() {
  local color=${1:?color required} sha=${2:?SHA required}
  shift 2
  valid_sha "$sha" || { echo 'Release SHA must be 40 lowercase hex characters' >&2; return 2; }
  if test -z "${IMAGE_TAG_GATEWAY:-}" || test -z "${IMAGE_TAG_API:-}" || test -z "${IMAGE_TAG_WEB:-}"; then
    read -r IMAGE_TAG_GATEWAY IMAGE_TAG_API IMAGE_TAG_WEB < <(read_manifest "$sha")
    export IMAGE_TAG_GATEWAY IMAGE_TAG_API IMAGE_TAG_WEB
  fi
  RELEASE_SHA=$sha GATEWAY_PORT=$(port_for_color "$color") \
    docker compose --env-file "$CONFIG_FILE" -f "$QMANGA_REPO_ROOT/deploy/compose.app.yaml" \
      -p "qmanga-$DEPLOY_TARGET-$color" "$@"
}

# Skip the full MinIO/PostgreSQL snapshot on the critical path only while the
# regularly scheduled verified backup is fresh. Migrations always force a new
# snapshot; missing or stale markers conservatively force one too.
backup_fresh() {
  local newest now age limit
  newest=$(find "$QMANGA_HOST_ROOT/backups/$DEPLOY_TARGET" -type f -name COMPLETE \
    -printf '%T@ %p\n' 2>/dev/null | sort -n | tail -n 1 | cut -d' ' -f2-)
  test -n "$newest" || return 1
  now=$(date +%s)
  age=$((now - $(stat -c %Y "$newest")))
  limit=$(( ${QMANGA_MAX_BACKUP_AGE_H:-30} * 3600 ))
  test "$age" -le "$limit"
}

# A failed or manually rolled-back release stays held until deploy-ready
# moves to a different SHA. Poller and manual releases respect the hold, so a
# bad SHA is not re-promoted every retry window.
hold_sha() {
  local file="$QMANGA_HOST_ROOT/state/deployment-hold" sha
  test -f "$file" || return 1
  read -r sha _ < "$file"
  printf '%s\n' "$sha"
}
record_hold() { # sha
  printf '%s %s\n' "${1:?sha required}" "$(date +%s)" > "$QMANGA_HOST_ROOT/state/deployment-hold"
}
locked_release() { # exit code from release.sh meaning "another run holds the lock"
  test "${1:-}" -eq 75
}

front_compose() {
  docker compose -f "$QMANGA_REPO_ROOT/deploy/compose.front.yaml" -p qmanga-front "$@"
}

local_health() {
  local port=${1:?port required}
  curl --fail --silent --show-error --max-time 10 -H "Host: $DOMAIN" \
    "http://127.0.0.1:$port/api/v1/health" >/dev/null
  curl --fail --silent --show-error --max-time 10 -H "Host: $DOMAIN" \
    "http://127.0.0.1:$port/" >/dev/null
}

public_health() {
  # OCI does not hairpin a VM's own public IP back to the VM. Connect to the
  # local Caddy listener while keeping the real hostname for TLS and routing.
  curl --fail --silent --show-error --max-time 15 \
    --resolve "$DOMAIN:443:127.0.0.1" "$APP_ORIGIN/api/v1/health" >/dev/null
  curl --fail --silent --show-error --max-time 15 \
    --resolve "$DOMAIN:443:127.0.0.1" "$APP_ORIGIN/" >/dev/null
}

target_health() {
  local port=${1:?port required}
  if test "$DEPLOY_TARGET" = staging; then
    local_health "$port"
  else
    public_health
  fi
}

served_homepage() {
  local port=${1:?port required}
  if test "$DEPLOY_TARGET" = staging; then
    curl --fail --silent --show-error --max-time 15 -H "Host: $DOMAIN" \
      "http://127.0.0.1:$port/"
  else
    curl --fail --silent --show-error --max-time 15 \
      --resolve "$DOMAIN:443:127.0.0.1" "$APP_ORIGIN/"
  fi
}

load_public_origin() {
  APP_ORIGIN=$(sed -n 's/^APP_ORIGIN=//p' "$CONFIG_FILE" | tail -n 1)
  test "$APP_ORIGIN" = "https://$DOMAIN" || {
    echo "APP_ORIGIN must be https://$DOMAIN" >&2; return 2;
  }
  export APP_ORIGIN
}
