#!/usr/bin/env bash
set -euo pipefail

QMANGA_REPO_ROOT=${QMANGA_REPO_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}
QMANGA_HOST_ROOT=${QMANGA_HOST_ROOT:-/opt/qmanga}
export QMANGA_HOST_ROOT

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

app_compose() {
  local color=${1:?color required} sha=${2:?SHA required}
  shift 2
  valid_sha "$sha" || { echo 'Release SHA must be 40 lowercase hex characters' >&2; return 2; }
  RELEASE_SHA=$sha GATEWAY_PORT=$(port_for_color "$color") \
    docker compose --env-file "$CONFIG_FILE" -f "$QMANGA_REPO_ROOT/deploy/compose.app.yaml" \
      -p "qmanga-$DEPLOY_TARGET-$color" "$@"
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

load_public_origin() {
  APP_ORIGIN=$(sed -n 's/^APP_ORIGIN=//p' "$CONFIG_FILE" | tail -n 1)
  test "$APP_ORIGIN" = "https://$DOMAIN" || {
    echo "APP_ORIGIN must be https://$DOMAIN" >&2; return 2;
  }
  export APP_ORIGIN
}
