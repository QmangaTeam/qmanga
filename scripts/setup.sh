#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
destination=${QMANGA_ENV_FILE:-.env}
project=${COMPOSE_PROJECT_NAME:-qmanga}
if [ -e "$destination" ]; then printf '%s\n' 'Environment file already exists; preserved.'; exit 0; fi
case "$project" in ''|[!a-z0-9]*|*[!a-z0-9_-]*) printf '%s\n' 'Invalid Compose project name.' >&2; exit 1;; esac
volumes=$(docker volume ls --filter "label=com.docker.compose.project=$project" --format '{{.Name}}') || { printf '%s\n' 'Cannot inspect Docker volumes. Start Docker and retry; no credentials were generated.' >&2; exit 1; }
if [ -n "$volumes" ]; then printf '%s\n' 'Existing volumes but no env. Restore the original env or choose a NEW COMPOSE_PROJECT_NAME and QMANGA_ENV_FILE. Do not delete data.' >&2; exit 1; fi
umask 077
secret() { od -An -N32 -tx1 /dev/urandom | tr -d ' \n'; }
set -C
{
    printf 'APP_ENV=development\nCOMPOSE_PROJECT_NAME=%s\nAPP_PORT=8080\nAPP_ORIGIN=http://localhost:8080\nMAILPIT_PORT=8025\n' "$project"
    printf 'POSTGRES_PASSWORD=%s\n' "$(secret)"
    printf 'S3_ACCESS_KEY=qmanga%s\n' "$(secret | cut -c1-16)"
    printf 'S3_SECRET_KEY=%s\n' "$(secret)"
    printf 'SESSION_SECRET=%s\n' "$(secret)"
    printf 'CSRF_SECRET=%s\n' "$(secret)"
} > "$destination"
printf '%s\n' 'Created local .env with random secrets.'

