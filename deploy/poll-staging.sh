#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/lib.sh"
target_config staging

# The VM only makes outbound requests. No CI private key or public SSH rule is
# needed. A release is eligible only when the exact main commit passed the full
# workflow, including all three ARM64 image publication jobs.
exec 8>"$QMANGA_HOST_ROOT/state/staging-poll.lock"
flock -n 8 || exit 0

desired=$(git -C "$QMANGA_REPO_ROOT" ls-remote origin refs/heads/main | awk '{print $1}')
valid_sha "$desired" || { echo 'Cannot resolve the current main commit' >&2; exit 1; }
active_file="$QMANGA_HOST_ROOT/state/staging.active"
if test -f "$active_file" && grep -Eq "^[a-z]+ $desired$" "$active_file"; then
  exit 0
fi

api=https://api.github.com/repos/maqsatto/qmanga/actions
request() {
  curl --fail --silent --show-error --max-time 20 \
    -H 'Accept: application/vnd.github+json' \
    -H 'X-GitHub-Api-Version: 2022-11-28' \
    -H 'User-Agent: qmanga-deploy-poller' "$1"
}
runs=$(request "$api/workflows/verify.yml/runs?head_sha=$desired&per_page=20")
mapfile -t successful_runs < <(jq -r --arg sha "$desired" \
  '.workflow_runs[] | select(.head_sha == $sha and .status == "completed" and .conclusion == "success") | .id' <<< "$runs")

for run in "${successful_runs[@]}"; do
  jobs=$(request "$api/runs/$run/jobs?per_page=100")
  if jq -e '[.jobs[] | select(.name | startswith("publish-images ("))] as $published |
    ($published | length) == 3 and all($published[]; .conclusion == "success")' \
    <<< "$jobs" >/dev/null; then
    exec bash "$QMANGA_REPO_ROOT/deploy/release.sh" staging "$desired"
  fi
done

echo "Staging is waiting for successful acceptance and three published images for $desired"
