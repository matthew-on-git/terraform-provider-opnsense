#!/usr/bin/env bash
# Purpose: Run guarded Trust lifecycle acceptance in the pinned toolchain.
# Usage: run-story-31-0f-acceptance.sh [--help]
# Dependencies: bash, docker, git, realpath
set -euo pipefail
if [[ "${1:-}" == --help ]]; then
  printf '%s\n' 'Usage: OPNSENSE_TRUST_RUNTIME=/private/runtime/runtime.json run-story-31-0f-acceptance.sh' 'Requires newly reserved disposable authority and private cleanup/restore oracle; never targets an ordinary configured endpoint.'
  exit 0
fi
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
# shellcheck source=test/scripts/lib/log.sh
. "${repo_root}/test/scripts/lib/log.sh"
: "${OPNSENSE_TRUST_RUNTIME:?new disposable runtime JSON required}"
[[ "${OPNSENSE_TRUST_RUNTIME}" = /* && -f "${OPNSENSE_TRUST_RUNTIME}" ]] || die 'runtime must be an absolute private file'
OPNSENSE_TRUST_RUNTIME="$(realpath "${OPNSENSE_TRUST_RUNTIME}")"
export OPNSENSE_TRUST_RUNTIME
runtime_root="$(cd "$(dirname "${OPNSENSE_TRUST_RUNTIME}")" && pwd -P)"
[[ "${runtime_root}" != / && "${runtime_root}" != "${repo_root}" &&
  "${runtime_root}" != "${repo_root}/"* && "${repo_root}" != "${runtime_root}/"* ]] || die 'runtime must not overlap the publication tree'
# Worktree .git files reference metadata outside /workspace. Preserve that exact
# absolute path, read-only, rather than copying or replacing source inventory.
git_common_dir="$(git -C "${repo_root}" rev-parse --path-format=absolute --git-common-dir)"
container_args=(-v "${git_common_dir}:${git_common_dir}:ro")
if [[ -n "${SSL_CERT_FILE:-}" ]]; then
  [[ "${SSL_CERT_FILE}" = /* && -f "${SSL_CERT_FILE}" ]] || die 'TLS CA must be an absolute private file'
  SSL_CERT_FILE="$(realpath "${SSL_CERT_FILE}")"
  [[ "${SSL_CERT_FILE}" == "${runtime_root}/"* ]] || die 'TLS CA must be inside private runtime root'
  export SSL_CERT_FILE
  container_args+=(-e SSL_CERT_FILE)
fi
docker run --rm "${container_args[@]}" -v "${repo_root}:/workspace:ro" -v "${runtime_root}:${runtime_root}" \
  -v opnsense-dhcp-go-mod:/root/go/pkg/mod -v opnsense-dhcp-go-build:/root/.cache/go-build \
  -w /workspace -e OPNSENSE_TRUST_RUNTIME -e OPNSENSE_URI -e OPNSENSE_API_KEY \
  -e OPNSENSE_API_SECRET -e OPNSENSE_ALLOW_INSECURE \
  ghcr.io/devrail-dev/dev-toolchain:1.12.24 \
  python3 test/scripts/trust_lifecycle_acceptance.py
