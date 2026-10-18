#!/usr/bin/env bash
set -euo pipefail
# Exercise the exact production helper against a disposable Compose PostgreSQL.
: "${QMANGA_TEST_ENV_FILE:?isolated env file required}"
: "${QMANGA_TEST_PROJECT:?isolated project required}"
[[ "$QMANGA_TEST_PROJECT" =~ ^qmanga-(ci|stage[0-9]+-[a-z0-9-]+)$ ]] || { echo 'Disposable project required' >&2; exit 2; }
source "$(dirname "$0")/../deploy/backup-guard.sh"
data_compose() { docker compose --env-file "$QMANGA_TEST_ENV_FILE" -p "$QMANGA_TEST_PROJECT" "$@"; }
query() { data_compose exec -T postgres psql -X -qAt -v ON_ERROR_STOP=1 -U qmanga -d qmanga -c "$1"; }
trap backup_guard_stop EXIT
backup_guard_start
backup_guard_check
[[ "$(query 'SELECT pg_try_advisory_lock(482161,1)')" = f ]]
# A second session cannot acquire the collector/backup guard while held.
backend=$backup_guard_backend
[[ "$(query "SELECT pg_terminate_backend($backend)")" = t ]]
if backup_guard_check; then echo 'Lost database guard accepted' >&2; exit 1; fi
backup_guard_stop
# Termination releases the session lock; acquisition/release works again.
backup_guard_start
backup_guard_check
backup_guard_stop
backup_guard_pid=''
[[ "$(query 'SELECT pg_try_advisory_lock(482161,1)')" = t ]]
echo 'Backup guard: acquisition, collector exclusion, lost connection rejection and release pass.'
