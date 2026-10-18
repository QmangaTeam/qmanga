#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/lib.sh"
target_config "${1:?target required}"
color=${2:?blue, green or off required}
mode=${3:-normal}
site="$QMANGA_HOST_ROOT/state/sites/$DEPLOY_TARGET.caddy"
old=$(mktemp "$QMANGA_HOST_ROOT/state/sites/.previous.XXXXXX")
draft=$(mktemp "$QMANGA_HOST_ROOT/state/sites/.draft.XXXXXX")
had_old=false
if test -f "$site"; then cp "$site" "$old"; had_old=true; fi
cleanup() { rm -f "$old" "$draft"; }
trap cleanup EXIT

if test "$color" = off; then
  rm -f "$site"
else
  port=$(port_for_color "$color")
  case "$mode" in normal|readonly) ;; *) echo 'Invalid route mode' >&2; exit 2 ;; esac
  {
    printf '%s {\n' "$DOMAIN"
    if test "$mode" = readonly; then
      printf '  @mutations method POST PUT PATCH DELETE\n'
      printf '  handle @mutations {\n    header Retry-After "120"\n    respond "Temporarily read-only for a consistent backup" 503\n  }\n'
    fi
    printf '  handle {\n    reverse_proxy 127.0.0.1:%s\n  }\n}\n' "$port"
  } > "$draft"
  mv "$draft" "$site"
fi

if test -n "$(front_compose ps --status running -q caddy 2>/dev/null)"; then
  if ! front_compose exec -T caddy caddy reload --config /etc/caddy/config/Caddyfile --adapter caddyfile; then
    if "$had_old"; then cp "$old" "$site"; else rm -f "$site"; fi
    front_compose exec -T caddy caddy reload --config /etc/caddy/config/Caddyfile --adapter caddyfile || true
    echo 'Caddy refused route; previous site restored' >&2
    exit 1
  fi
elif test "$color" != off; then
  if ! front_compose up -d --wait; then
    if "$had_old"; then cp "$old" "$site"; else rm -f "$site"; fi
    echo 'Caddy failed to start; previous site restored' >&2
    exit 1
  fi
fi
