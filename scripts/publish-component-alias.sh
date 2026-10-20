#!/usr/bin/env bash
set -euo pipefail
# Return 0 for present, 1 only for a registry-confirmed missing manifest.
# Every other failure exits rather than authorizing an overwrite.
manifest_present() {
  local error
  error=$(mktemp)
  if docker manifest inspect "$1" >/dev/null 2>"$error"; then
    rm -f "$error"
    return 0
  fi
  if grep -qiE 'manifest unknown|no such manifest' "$error"; then
    rm -f "$error"
    return 1
  fi
  rm -f "$error"
  echo 'Cannot establish image existence; refusing to overwrite' >&2
  exit 2
}
if test "${1:-}" = --head-status; then
  image=${2:?image repository required}
  sha=${3:?head SHA required}
  [[ "$sha" =~ ^[0-9a-f]{40}$ ]] || exit 2
  if manifest_present "$image:$sha"; then echo existing; else echo missing; fi
  exit 0
fi
component=${1:?component required}
image=${2:?image repository required}
desired=${3:?published head SHA required}
[[ "$desired" =~ ^[0-9a-f]{40}$ ]] || exit 2
case "$component" in
  api) paths=(backend infra/api.Dockerfile infra/imagemagick-policy.xml) ;;
  web) paths=(apps/web infra/web.Dockerfile) ;;
  gateway) paths=(infra/nginx.conf infra/gateway.Dockerfile) ;;
  *) echo 'Unknown component' >&2; exit 2 ;;
esac
canonical=$(git log -1 --format=%H "$desired" -- "${paths[@]}")
[[ "$canonical" =~ ^[0-9a-f]{40}$ ]] || { echo 'Cannot resolve component inputs' >&2; exit 2; }
if test "$canonical" = "$desired"; then exit 0; fi
if manifest_present "$image:$canonical"; then
  echo "Keeping existing immutable $component:$canonical"
  exit 0
fi
docker buildx imagetools create --tag "$image:$canonical" "$image:$desired"
echo "Published component alias $component:$canonical from $desired"
