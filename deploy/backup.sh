#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/lib.sh"
target_config "${1:?target required}"
load_public_origin
stamp=$(date -u +%Y%m%dT%H%M%SZ)
backup="$QMANGA_HOST_ROOT/backups/$DEPLOY_TARGET/$stamp"
mkdir -m 700 "$backup"
# The mirror and tar briefly need two extra copies of object bytes. Refuse a
# backup before risking a full filesystem shared with live data.
object_kb=$(docker run --rm --network none \
  --mount "type=volume,source=qmanga-$DEPLOY_TARGET-minio,target=/snapshot,readonly" \
  alpine@sha256:ce64758a109eb420d874a118f87920e625e12d3634e03b4a5573fd9f6e5d3507 \
  du -sk /snapshot | awk '{print $1}')
database_bytes=$(data_compose exec -T postgres psql -At -U qmanga -d qmanga -c "SELECT pg_database_size('qmanga')")
[[ "$object_kb" =~ ^[0-9]+$ && "$database_bytes" =~ ^[0-9]+$ ]]
available_kb=$(df -Pk "$backup" | awk 'NR==2 {print $4}')
required_kb=$((object_kb * 2 + (database_bytes / 1024) * 2 + 1048576))
test "$available_kb" -ge "$required_kb" || {
  echo "Backup needs $required_kb KiB free, only $available_kb KiB available" >&2; exit 2;
}
# Database snapshot precedes object copy. This remains online-safe only while
# S3 keys are immutable and objects are never physically deleted.
data_compose exec -T postgres pg_dump -U qmanga -d qmanga -Fc --no-owner > "$backup/postgres.dump"
test -s "$backup/postgres.dump"
set -a
source "$CONFIG_FILE"
set +a
export MC_HOST_qmanga="http://$MINIO_ROOT_USER:$MINIO_ROOT_PASSWORD@minio:9000"
mc="$QMANGA_HOST_ROOT/bin/mc"
test -x "$mc"
for bucket in qmanga-media qmanga-source; do
  mkdir -p "$backup/$bucket"
  docker run --rm --user "$(id -u):$(id -g)" --network "qmanga-$DEPLOY_TARGET-data" \
    --mount "type=bind,source=$mc,target=/usr/local/bin/mc,readonly" \
    --mount "type=bind,source=$backup,target=/backup" \
    --env MC_HOST_qmanga \
    alpine@sha256:ce64758a109eb420d874a118f87920e625e12d3634e03b4a5573fd9f6e5d3507 \
    /usr/local/bin/mc --config-dir /tmp/mc mirror "qmanga/$bucket" "/backup/$bucket"
done
docker run --rm --user "$(id -u):$(id -g)" --network none \
  --mount "type=bind,source=$backup,target=/backup" \
  alpine@sha256:ce64758a109eb420d874a118f87920e625e12d3634e03b4a5573fd9f6e5d3507 \
  tar -C /backup -cf /backup/objects.tar qmanga-media qmanga-source
rm -r "$backup/qmanga-media" "$backup/qmanga-source"
test -s "$backup/objects.tar"
(cd "$backup" && sha256sum postgres.dump objects.tar > SHA256SUMS && sha256sum -c SHA256SUMS)
# Production must have a separately configured off-host copy, not just another
# directory on the same VM. The hook receives only the backup directory path.
if test "$DEPLOY_TARGET" = production; then
  hook="$QMANGA_HOST_ROOT/hooks/offsite-backup"
  test -x "$hook" || { echo "Missing executable off-host backup hook: $hook" >&2; exit 2; }
  "$hook" "$backup"
fi
printf '%s\n' "$DEPLOY_TARGET $stamp" > "$backup/COMPLETE"
echo "Verified $DEPLOY_TARGET backup: $backup"
