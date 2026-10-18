#!/usr/bin/env bash
set -euo pipefail
source "${QMANGA_REPO_ROOT:-/opt/qmanga/repo}/deploy/lib.sh"
target_config production

backup=$(realpath -e -- "${1:?production backup directory required}")
case "$backup" in
  "$QMANGA_HOST_ROOT/backups/production/"*) ;;
  *) echo 'Refusing to upload a path outside production backups' >&2; exit 2 ;;
esac
test -f "$backup/postgres.dump" && test -f "$backup/objects.tar" && test -f "$backup/SHA256SUMS"
(cd "$backup" && sha256sum -c SHA256SUMS)

oci=${OCI_BIN:-/opt/qmanga/bin/oci}
test -x "$oci" || { echo "OCI CLI is missing at $oci" >&2; exit 2; }
namespace=${OCI_BACKUP_NAMESPACE:-frf73cqcjqzh}
bucket=${OCI_BACKUP_BUCKET:-qmanga-backups}
region=${OCI_BACKUP_REGION:-eu-frankfurt-1}
prefix="production/$(basename "$backup")"
verification=$(mktemp -d "$QMANGA_HOST_ROOT/state/.oci-backup-verify.XXXXXX")
trap 'rm -rf -- "$verification"' EXIT

for file in postgres.dump objects.tar SHA256SUMS; do
  "$oci" os object put --auth instance_principal --region "$region" \
    --namespace "$namespace" --bucket-name "$bucket" \
    --name "$prefix/$file" --file "$backup/$file" --force --no-multipart >/dev/null
  "$oci" os object get --auth instance_principal --region "$region" \
    --namespace "$namespace" --bucket-name "$bucket" \
    --name "$prefix/$file" --file "$verification/$file" >/dev/null
  cmp "$backup/$file" "$verification/$file"
done
(cd "$verification" && sha256sum -c SHA256SUMS)
sha256sum "$verification/SHA256SUMS" | awk '{print $1}' > "$verification/COMPLETE"
"$oci" os object put --auth instance_principal --region "$region" \
  --namespace "$namespace" --bucket-name "$bucket" \
  --name "$prefix/COMPLETE" --file "$verification/COMPLETE" --force >/dev/null
echo "Verified off-host production backup: $bucket/$prefix"
