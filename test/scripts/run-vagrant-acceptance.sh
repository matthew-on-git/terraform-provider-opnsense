#!/usr/bin/env bash
# Run provider acceptance tests against the local Vagrant OPNsense appliance.

set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
PROJECT_ROOT=$(cd -- "${SCRIPT_DIR}/../.." && pwd)
# shellcheck source=test/scripts/lib/log.sh
. "${SCRIPT_DIR}/lib/log.sh"

DEVRAIL_IMAGE=${DEVRAIL_IMAGE:-ghcr.io/devrail-dev/dev-toolchain:1.12.0}
TEST_REGEX=${TEST_REGEX:-TestAcc}
TEST_TIMEOUT=${TEST_TIMEOUT:-30m}
RESULTS_DIR=${RESULTS_DIR:-${PROJECT_ROOT}/test/acceptance-results}
RUN_ID=${RUN_ID:-$(date -u +%Y%m%dT%H%M%SZ)}
RUN_DIR=${RESULTS_DIR}/${RUN_ID}
PACKAGE_LIST=${OPNSENSE_ACCEPTANCE_PACKAGES:-./internal/service/...}

usage() {
  cat <<'USAGE'
Usage: test/scripts/run-vagrant-acceptance.sh [--package PKG ...]

Runs Terraform acceptance tests against the local Vagrant OPNsense VM through
the dev-toolchain container. Required environment:

  OPNSENSE_URI=https://localhost:10444
  OPNSENSE_API_KEY=...
  OPNSENSE_API_SECRET=...
  OPNSENSE_ALLOW_INSECURE=true

Optional environment:

  TEST_REGEX=TestAcc                       Go -run pattern
  TEST_TIMEOUT=30m                         Per-package Go test timeout
  DEVRAIL_IMAGE=ghcr.io/devrail-dev/dev-toolchain:1.12.0
  OPNSENSE_ACCEPTANCE_PACKAGES='./internal/service/firewall ./internal/service/haproxy'
  RESULTS_DIR=test/acceptance-results

Optional tests may self-skip unless their extra env vars are set, for example
OPNSENSE_ACME_ISSUE=1 or OPNSENSE_HAPROXY_CERT_REFID.
USAGE
}

packages_from_args=()
while [ "$#" -gt 0 ]; do
  case "$1" in
    --help|-h)
      usage
      exit 0
      ;;
    --package)
      [ "$#" -ge 2 ] || die "--package requires a package path"
      packages_from_args+=("$2")
      shift 2
      ;;
    *)
      die "unknown argument: $1"
      ;;
  esac
done

for required_env in OPNSENSE_URI OPNSENSE_API_KEY OPNSENSE_API_SECRET OPNSENSE_ALLOW_INSECURE; do
  [ -n "${!required_env:-}" ] || die "${required_env} must be set"
done

command -v docker >/dev/null 2>&1 || die "docker is required"
mkdir -p "${RUN_DIR}"

docker_env=(
  -e TF_ACC=1
  -e "OPNSENSE_URI=${OPNSENSE_URI}"
  -e "OPNSENSE_API_KEY=${OPNSENSE_API_KEY}"
  -e "OPNSENSE_API_SECRET=${OPNSENSE_API_SECRET}"
  -e "OPNSENSE_ALLOW_INSECURE=${OPNSENSE_ALLOW_INSECURE}"
)

for optional_env in OPNSENSE_ACME_ISSUE OPNSENSE_ACME_CERT_DOMAIN OPNSENSE_ACME_ACCOUNT_UUID OPNSENSE_ACME_VALIDATION_UUID OPNSENSE_HAPROXY_CERT_REFID; do
  if [ -n "${!optional_env:-}" ]; then
    docker_env+=( -e "${optional_env}=${!optional_env}" )
  fi
done

docker_base=(
  docker run --rm --network host
  -v "${PROJECT_ROOT}:/workspace"
  -w /workspace
  "${docker_env[@]}"
  "${DEVRAIL_IMAGE}"
)

if [ "${#packages_from_args[@]}" -gt 0 ]; then
  packages=("${packages_from_args[@]}")
else
  # shellcheck disable=SC2206
  package_patterns=(${PACKAGE_LIST})
  mapfile -t packages < <("${docker_base[@]}" go list "${package_patterns[@]}")
fi

[ "${#packages[@]}" -gt 0 ] || die "no packages resolved"

summary_file=${RUN_DIR}/summary.tsv
printf 'status\tpackage\tlog\n' > "${summary_file}"

log_info "running ${#packages[@]} package(s) with -run ${TEST_REGEX}"
log_info "results: ${RUN_DIR}"

failures=0
for pkg in "${packages[@]}"; do
  safe_name=${pkg//\//_}
  safe_name=${safe_name//./_}
  log_file=${RUN_DIR}/${safe_name}.log
  log_info "testing ${pkg}"
  if "${docker_base[@]}" go test -run "${TEST_REGEX}" -count=1 -p 1 -timeout "${TEST_TIMEOUT}" "${pkg}" >"${log_file}" 2>&1; then
    printf 'pass\t%s\t%s\n' "${pkg}" "${log_file}" >> "${summary_file}"
  else
    failures=$((failures + 1))
    printf 'fail\t%s\t%s\n' "${pkg}" "${log_file}" >> "${summary_file}"
    log_warn "failed ${pkg}; see ${log_file}"
  fi
done

if [ "${failures}" -gt 0 ]; then
  log_error "${failures} package(s) failed; summary: ${summary_file}"
  exit 1
fi

log_info "all package acceptance runs passed; summary: ${summary_file}"
