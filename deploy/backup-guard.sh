#!/usr/bin/env bash
# Sourced by backup.sh. All avatar deletion uses the same two-int session lock.
# The psql process is deliberately long-lived, separate from pg_dump.
backup_guard_start() {
  coproc QMANGA_BACKUP_LOCK { data_compose exec -T postgres psql -X -qAt -v ON_ERROR_STOP=1 -U qmanga -d qmanga; }
  backup_guard_pid=$QMANGA_BACKUP_LOCK_PID
  backup_guard_input=${QMANGA_BACKUP_LOCK[1]}
  backup_guard_output=${QMANGA_BACKUP_LOCK[0]}
  local acknowledgement
  printf '%s\n' "SELECT CASE WHEN pg_try_advisory_lock(482161,1) THEN 'QMANGA_BACKUP_GUARD_READY:' || pg_backend_pid()::text ELSE 'QMANGA_BACKUP_GUARD_BUSY' END;" >&"$backup_guard_input"
  if ! IFS= read -r -t 10 acknowledgement <&"$backup_guard_output" || [[ ! "$acknowledgement" =~ ^QMANGA_BACKUP_GUARD_READY:([0-9]+)$ ]]; then
    echo 'Backup deletion guard was not acquired; refusing database snapshot' >&2
    return 1
  fi
  backup_guard_backend=${BASH_REMATCH[1]}
  backup_guard_check
}
backup_guard_check() {
  local acknowledgement
  if ! kill -0 "$backup_guard_pid" 2>/dev/null; then
    echo 'Backup deletion guard process exited; backup is incomplete' >&2; return 1
  fi
  if ! printf '%s\n' "SELECT 'QMANGA_BACKUP_GUARD_CHECK:' || pg_backend_pid()::text || ':' || EXISTS(SELECT 1 FROM pg_locks WHERE locktype='advisory' AND classid=482161 AND objid=1 AND objsubid=2 AND mode='ExclusiveLock' AND granted AND pid=pg_backend_pid())::text;" >&"$backup_guard_input" ||
     ! IFS= read -r -t 10 acknowledgement <&"$backup_guard_output" ||
     [[ "$acknowledgement" != "QMANGA_BACKUP_GUARD_CHECK:$backup_guard_backend:true" ]]; then
    echo 'Backup deletion guard connection/lock was lost; backup is incomplete' >&2; return 1
  fi
}
backup_guard_stop() {
  if test -n "${backup_guard_pid:-}"; then
    if kill -0 "$backup_guard_pid" 2>/dev/null; then
      { printf '\\q\n' >&"$backup_guard_input"; } 2>/dev/null || true
    fi
    # Do not hang failure cleanup if the client/connection has already failed.
    for _ in {1..20}; do
      kill -0 "$backup_guard_pid" 2>/dev/null || break
      sleep 0.1
    done
    if kill -0 "$backup_guard_pid" 2>/dev/null; then kill "$backup_guard_pid" 2>/dev/null || true; fi
    wait "$backup_guard_pid" 2>/dev/null || true
  fi
}
