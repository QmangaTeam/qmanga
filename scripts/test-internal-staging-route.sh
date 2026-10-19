#!/usr/bin/env bash
set -euo pipefail

repo=$(cd "$(dirname "$0")/.." && pwd)
root=$(mktemp -d)
trap 'rm -rf "$root"' EXIT
mkdir -p "$root/config" "$root/state/sites" "$root/bin"
: > "$root/config/staging.env"
: > "$root/config/production.env"
cat > "$root/bin/docker" <<'EOF'
#!/usr/bin/env sh
exit 0
EOF
chmod +x "$root/bin/docker"
export PATH="$root/bin:$PATH" QMANGA_HOST_ROOT="$root" QMANGA_REPO_ROOT="$repo"

printf 'qmanga-dev.maqsatto.dev {\n  respond 200\n}\n' > "$root/state/sites/staging.caddy"
for color in blue green; do
  "$repo/deploy/route.sh" staging "$color" normal
  test ! -e "$root/state/sites/staging.caddy"
done
"$repo/deploy/route.sh" staging blue readonly
test ! -e "$root/state/sites/staging.caddy"

"$repo/deploy/route.sh" production blue normal
grep -q 'qmanga.maqsatto.dev {' "$root/state/sites/production.caddy"
grep -q '127.0.0.1:18180' "$root/state/sites/production.caddy"

echo 'Internal staging route and production route checks passed'
