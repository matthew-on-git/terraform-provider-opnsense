#!/usr/bin/env bash
# Purpose: Reproducible FRR compatibility source/spec/dependency digest.
# Usage: test/scripts/story-31-0h-code-digest.sh
# Dependencies: shasum, find, sort
set -euo pipefail
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${repo_root}"
if [[ "${1:-}" == --help ]]; then
  printf '%s\n' 'Usage: test/scripts/story-31-0h-code-digest.sh'
  exit 0
fi
paths=(.gitleaksignore .devrail.yml .goreleaser.yml DEVELOPMENT.md README.md Makefile go.mod go.sum main.go
  _bmad-output/implementation-artifacts/spec-31-0h-frr2-compatibility-gate.md
  _bmad-output/implementation-artifacts/spec-31-0h-advanced-frr-parity.md
  _bmad-output/implementation-artifacts/31-0h-advanced-source-evidence.json
  _bmad-output/implementation-artifacts/epic-31-context.md
  _bmad-output/implementation-artifacts/31-0h-quagga-bgp-field-parity.md
  _bmad-output/implementation-artifacts/sprint-status.yaml
  _bmad-output/planning-artifacts/support-matrix.md
  _bmad-output/planning-artifacts/provider-ecosystem-coverage-audit.json
  _bmad-output/planning-artifacts/provider-ecosystem-coverage-audit.md
  test/README.md test/scripts/lib/log.sh test/scripts/run-vagrant-acceptance.sh
  test/scripts/run-story-31-0h-acceptance.sh test/scripts/frr_acceptance_supervisor.py test/scripts/test_frr_acceptance_supervisor.py test/scripts/story-31-0h-code-digest.sh test/scripts/test-story-31-0h-code-digest.sh)
inventory_dir="$(mktemp -d)"
trap 'rm -rf -- "${inventory_dir}"' EXIT
find internal pkg .github/workflows -type f -print0 >"${inventory_dir}/sources" || {
  printf '%s\n' 'source inventory failed' >&2
  exit 1
}
find templates examples docs -type f -path '*quagga*' -print0 >"${inventory_dir}/docs" || {
  printf '%s\n' 'documentation inventory failed' >&2
  exit 1
}
[[ -r "${inventory_dir}/sources" && -s "${inventory_dir}/sources" && -r "${inventory_dir}/docs" && -s "${inventory_dir}/docs" ]] || {
  printf '%s\n' 'missing or unreadable inventory' >&2
  exit 1
}
while IFS= read -r -d '' path; do paths+=("${path}"); done <"${inventory_dir}/sources"
while IFS= read -r -d '' path; do paths+=("${path}"); done <"${inventory_dir}/docs"
printf '%s\n' "${paths[@]}" | LC_ALL=C sort -u >"${inventory_dir}/sorted"
# Live evidence deliberately excluded: it records this digest without recursion.
while IFS= read -r path; do
  printf 'path %s\n' "${path}"
  shasum -a 256 "${path}"
done <"${inventory_dir}/sorted" | shasum -a 256 | awk '{print $1}'
