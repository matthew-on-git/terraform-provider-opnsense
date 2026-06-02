#!/usr/bin/env bash
# Validate that every service package's exports.go registers a constructor for
# each resource/data-source implemented in that package, so nothing is silently
# left unwired. Idempotent and read-only.
#
# Note: intentionally no `set -e` — grep returning non-zero (a package with no
# data sources) is expected; the script tracks failures via `status`.
set -uo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

status=0
for dir in internal/service/*/; do
  pkg="$(basename "$dir")"
  exports="$dir/exports.go"
  [ -f "$exports" ] || { echo "MISSING exports.go: $pkg" >&2; status=1; continue; }

  # Resource constructors defined in the package vs. registered in Resources().
  defined_res="$(grep -rhoE 'func (new[A-Za-z0-9]+Resource)\(\) resource\.Resource' "$dir" 2>/dev/null \
    | grep -oE 'new[A-Za-z0-9]+Resource' | sort -u)"
  defined_ds="$(grep -rhoE 'func (new[A-Za-z0-9]+DataSource)\(\) datasource\.DataSource' "$dir" 2>/dev/null \
    | grep -oE 'new[A-Za-z0-9]+DataSource' | sort -u)"
  registered="$(grep -oE 'new[A-Za-z0-9]+(Resource|DataSource)' "$exports" | sort -u)"

  for c in $defined_res $defined_ds; do
    if ! grep -qxF "$c" <<<"$registered"; then
      echo "UNREGISTERED in $pkg/exports.go: $c" >&2
      status=1
    fi
  done
done

if [ "$status" -eq 0 ]; then
  echo "structure OK: all resource/data-source constructors are registered"
fi
exit "$status"
