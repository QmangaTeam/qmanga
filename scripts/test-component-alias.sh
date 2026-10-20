#!/usr/bin/env bash
set -euo pipefail
script=$(cd "$(dirname "$0")" && pwd)/publish-component-alias.sh
repo=$(cd "$(dirname "$script")/.." && pwd)
# Both workflows must serialize the complete HEAD + alias publication job.
for workflow in release verify; do
 grep -Fq 'group: component-image-${{ github.repository }}-${{ matrix.component }}' "$repo/.github/workflows/$workflow.yml"
done
fixture=$(mktemp -d)
trap 'rm -rf "$fixture"' EXIT
cd "$fixture"
git init -q
git config user.email test@example.invalid
git config user.name Test
mkdir -p backend apps/web infra bin
printf api > backend/input
git add backend
git commit -qm api
api=$(git rev-parse HEAD)
printf web > apps/web/input
git add apps
git commit -qm web
web=$(git rev-parse HEAD)
printf gateway > infra/nginx.conf
git add infra
git commit -qm gateway
gateway=$(git rev-parse HEAD)
canonical=$api
printf docs > note
git add note
git commit -qm workflow-head
head=$(git rev-parse HEAD)
export EVENT_FILE="$fixture/events" ALIAS_MODE=missing
cat > bin/docker <<'DOCKER'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$EVENT_FILE"
if test "$1" = manifest; then
 case "$ALIAS_MODE" in
  existing) exit 0 ;;
  missing) echo 'no such manifest: fixture' >&2; exit 1 ;;
  denied) echo 'unauthorized: authentication required' >&2; exit 1 ;;
  timeout) echo 'TLS handshake timeout' >&2; exit 1 ;;
 esac
fi
DOCKER
chmod +x bin/docker
export PATH="$fixture/bin:$PATH"
for component in api web gateway; do
 canonical=${!component}
 : > "$EVENT_FILE"
 bash "$script" "$component" example/image "$head"
 grep -Fx "buildx imagetools create --tag example/image:$canonical example/image:$head" "$EVENT_FILE"
done
canonical=$api
for ALIAS_MODE in existing denied timeout; do
 export ALIAS_MODE
 : > "$EVENT_FILE"
 if test "$ALIAS_MODE" = existing; then
  bash "$script" api example/image "$head"
 else
  if bash "$script" api example/image "$head"; then echo 'Registry error accepted' >&2; exit 1; fi
 fi
 ! grep -q '^buildx ' "$EVENT_FILE"
done
: > "$EVENT_FILE"
bash "$script" api example/image "$canonical"
test ! -s "$EVENT_FILE"
for ALIAS_MODE in existing missing; do
 export ALIAS_MODE
 [[ $(bash "$script" --head-status example/image "$head") == "$ALIAS_MODE" ]]
done
for ALIAS_MODE in denied timeout; do
 export ALIAS_MODE
 if bash "$script" --head-status example/image "$head"; then echo 'Unsafe HEAD overwrite authorized' >&2; exit 1; fi
done
echo 'Component aliases: all mappings, missing publication, existing preservation, auth/network rejection and head identity pass.'
