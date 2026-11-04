#!/usr/bin/env bash
# Purpose: Run guarded Trust lifecycle acceptance in the pinned toolchain.
# Usage: run-story-31-0f-acceptance.sh [--help]
# Dependencies: bash, docker
set -euo pipefail
if [[ "${1:-}" == --help ]]; then
  printf '%s\n' 'Usage: OPNSENSE_TRUST_RUNTIME=/private/runtime/runtime.json run-story-31-0f-acceptance.sh' 'Requires newly reserved disposable authority and private cleanup/restore oracle; never targets an ordinary configured endpoint.'
  exit 0
fi
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
# shellcheck source=test/scripts/lib/log.sh
. "${repo_root}/test/scripts/lib/log.sh"
: "${OPNSENSE_TRUST_RUNTIME:?new disposable runtime JSON required}"
[[ "${OPNSENSE_TRUST_RUNTIME}" = /* && -f "${OPNSENSE_TRUST_RUNTIME}" ]] || die 'runtime must be an absolute private file'
runtime_root="$(cd "$(dirname "${OPNSENSE_TRUST_RUNTIME}")" && pwd)"
[[ "${runtime_root}" != / && "${runtime_root}" != "${repo_root}"* ]] || die 'runtime must be outside the publication tree'
docker run --rm -v "${repo_root}:/workspace" -v "${runtime_root}:${runtime_root}" \
  -v opnsense-dhcp-go-mod:/root/go/pkg/mod -v opnsense-dhcp-go-build:/root/.cache/go-build \
  -w /workspace -e OPNSENSE_TRUST_RUNTIME -e OPNSENSE_URI -e OPNSENSE_API_KEY \
  -e OPNSENSE_API_SECRET -e OPNSENSE_ALLOW_INSECURE \
  ghcr.io/devrail-dev/dev-toolchain:1.12.24 \
  python3 test/scripts/trust_lifecycle_acceptance.py
