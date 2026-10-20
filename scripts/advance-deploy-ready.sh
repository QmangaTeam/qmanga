#!/usr/bin/env bash
# Move refs/deploy-ready forward to a published release SHA, never backwards.
# Only a trusted workflow with write access can call this; the ref is the
# single hand-off between CI and the outbound VM poller.
set -euo pipefail
desired=${1:?published release SHA required}
[[ "$desired" =~ ^[0-9a-f]{40}$ ]] || { echo 'Release SHA must be 40 lowercase hex characters' >&2; exit 2; }

for attempt in 1 2 3; do
  git fetch --quiet --no-tags origin '+refs/deploy-ready:refs/remotes/origin/deploy-ready' || true
  current=$(git rev-parse --verify --quiet refs/remotes/origin/deploy-ready || true)
  if test "$current" = "$desired"; then
    echo "deploy-ready already points at $desired"
    exit 0
  fi
  if test -n "$current"; then
    if git merge-base --is-ancestor "$desired" "$current"; then
      echo "A newer release ($current) already won; not moving deploy-ready back to $desired" >&2
      exit 0
    fi
    git merge-base --is-ancestor "$current" "$desired" || {
      echo "Release $desired is not a descendant of the released $current" >&2; exit 2;
    }
    if git push --quiet origin "$desired:refs/deploy-ready" --force-with-lease=refs/deploy-ready:"$current"; then
      echo "deploy-ready -> $desired"
      exit 0
    fi
  else
    if git push --quiet origin "$desired:refs/deploy-ready"; then
      echo "deploy-ready -> $desired"
      exit 0
    fi
  fi
  echo "deploy-ready moved concurrently; retrying ($attempt/3)" >&2
  sleep $((attempt * 2))
done
echo 'Could not advance deploy-ready; another release keeps winning' >&2
exit 1
