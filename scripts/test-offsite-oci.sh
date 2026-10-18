#!/usr/bin/env bash
set -euo pipefail

repo=$(cd "$(dirname "$0")/.." && pwd)
root=$(mktemp -d)
trap 'rm -rf -- "$root"' EXIT
mkdir -p "$root/config" "$root/state" "$root/backups/production/20261018T000000Z"
: > "$root/config/production.env"
backup="$root/backups/production/20261018T000000Z"
printf 'database fixture\n' > "$backup/postgres.dump"
printf 'object fixture\n' > "$backup/objects.tar"
(cd "$backup" && sha256sum postgres.dump objects.tar > SHA256SUMS)

cat > "$root/fake-oci" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
test "$1 $2" = 'os object'
operation=$3
shift 3
name= file= force=false
while test "$#" -gt 0; do
  case "$1" in
    --name) name=$2; shift 2 ;;
    --file) file=$2; shift 2 ;;
    --force) force=true; shift ;;
    *) shift ;;
  esac
done
test -n "$name" && test -n "$file"
object="$FAKE_BUCKET_DIR/$name"
case "$operation" in
  put) mkdir -p "$(dirname "$object")"; cp "$file" "$object" ;;
  get) test "$force" = false || { echo 'Object get does not support --force' >&2; exit 2; }; cp "$object" "$file" ;;
  *) exit 2 ;;
esac
EOF
chmod +x "$root/fake-oci"

export QMANGA_REPO_ROOT="$repo" QMANGA_HOST_ROOT="$root"
export OCI_BIN="$root/fake-oci" FAKE_BUCKET_DIR="$root/bucket"
bash "$repo/deploy/offsite-oci.sh" "$backup"
test -s "$root/bucket/production/20261018T000000Z/COMPLETE"
test "$(cat "$root/bucket/production/20261018T000000Z/COMPLETE")" = \
  "$(sha256sum "$backup/SHA256SUMS" | awk '{print $1}')"

printf 'corrupt\n' > "$backup/postgres.dump"
if bash "$repo/deploy/offsite-oci.sh" "$backup" >/dev/null 2>&1; then
  echo 'Corrupt backup was accepted' >&2
  exit 1
fi
echo 'OCI backup hook passed upload, download and checksum checks'
