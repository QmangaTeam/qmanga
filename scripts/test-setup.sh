#!/bin/sh
set -eu
repository=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT
mkdir -p "$temporary/scripts" "$temporary/bin"
cp "$repository/scripts/setup.sh" "$temporary/scripts/setup.sh"
cat > "$temporary/bin/docker" <<'MOCK'
#!/bin/sh
if [ "${FAKE_DOCKER_FAIL:-0}" = 1 ]; then exit 1; fi
printf '%s' "${FAKE_VOLUMES:-}"
MOCK
chmod +x "$temporary/bin/docker"
export PATH="$temporary/bin:$PATH"
export COMPOSE_PROJECT_NAME=qmanga-setup-test
export FAKE_VOLUMES=qmanga-setup-test_postgres-data
if sh "$temporary/scripts/setup.sh"; then printf '%s\n' 'Setup accepted old volumes without credentials.' >&2; exit 1; fi
test ! -e "$temporary/.env"
export FAKE_VOLUMES=''
export FAKE_DOCKER_FAIL=1
if sh "$temporary/scripts/setup.sh"; then printf '%s\n' 'Setup ignored Docker failure.' >&2; exit 1; fi
test ! -e "$temporary/.env"
export FAKE_DOCKER_FAIL=0
sh "$temporary/scripts/setup.sh"
test -s "$temporary/.env"
cp "$temporary/.env" "$temporary/original.env"
export FAKE_DOCKER_FAIL=1
sh "$temporary/scripts/setup.sh"
cmp "$temporary/original.env" "$temporary/.env"
printf '%s\n' 'Shell setup recovery guards and credential preservation passed.'
