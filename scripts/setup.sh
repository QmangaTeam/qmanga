#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
if [ -e .env ]; then printf '%s\n' '.env already exists; preserved.'; exit 0; fi
umask 077
secret() { od -An -N32 -tx1 /dev/urandom | tr -d ' \n'; }
set -C
{
    printf 'APP_ENV=development\nAPP_PORT=8080\nAPP_ORIGIN=http://localhost:8080\n'
    printf 'POSTGRES_PASSWORD=%s\n' "$(secret)"
    printf 'S3_ACCESS_KEY=qmanga%s\n' "$(secret | cut -c1-16)"
    printf 'S3_SECRET_KEY=%s\n' "$(secret)"
    printf 'SESSION_SECRET=%s\n' "$(secret)"
    printf 'CSRF_SECRET=%s\n' "$(secret)"
} > .env
printf '%s\n' 'Created local .env with random secrets.'

