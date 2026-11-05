#!/usr/bin/env bash
# Purpose: Non-skippable isolated, digest-bound FRR compatibility proof.
# Usage: source private acceptance.env; test/scripts/run-story-31-0h-acceptance.sh
# Dependencies: docker, python3, openssl, shasum
set -euo pipefail
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${repo_root}"
# shellcheck source=test/scripts/lib/log.sh
. test/scripts/lib/log.sh
if [[ "${1:-}" == --help ]]; then
  printf '%s\n' 'Usage: source private acceptance.env; test/scripts/run-story-31-0h-acceptance.sh'
  exit 0
fi
scan_secret_artifacts() {
  local artifact_dir="$1" secret_name scan_status
  for secret_name in OPNSENSE_API_KEY OPNSENSE_API_SECRET OPNSENSE_FRR_ORACLE_TOKEN; do
    if grep -rFq -- "${!secret_name}" "${artifact_dir}"; then
      die "secret scan failed for ${secret_name}; raw artifacts withheld"
    else scan_status=$?; fi
    [[ ${scan_status} -eq 1 ]] || die "secret scan operational error; raw artifacts withheld"
  done
}
# A fixed host-side operation is the only guest mutation outside provider/API
# lifecycle: regenerate the disabled template after exact API rollback. SSH keys
# remain on the host and guest-probe independently verifies VM identity and TLS.
disabled_sync_worker() {
  local handshake_dir="$1"
  while [[ ! -f "${handshake_dir}/disabled-config-sync.stop" ]]; do
    if [[ -f "${handshake_dir}/disabled-config-sync.request" && ! -f "${handshake_dir}/disabled-config-sync.done" ]]; then
      if [[ "$(cat "${handshake_dir}/disabled-config-sync.request")" == "${OPNSENSE_DISPOSABLE_VM_UUID}" ]] &&
        verify_disabled_settings >"${handshake_dir}/disabled-config-precheck.log" 2>&1 &&
        "${OPNSENSE_FRR_GUEST_PROBE}" '/usr/local/sbin/pluginctl -c frr_sync' >"${handshake_dir}/disabled-config-sync.log" 2>&1; then
        printf ok >"${handshake_dir}/disabled-config-sync.done"
      else
        printf failed >"${handshake_dir}/disabled-config-sync.done"
      fi
    fi
    sleep 0.2
  done
}
verify_disabled_settings() {
  python3 - <<'SYNC'
import base64, json, os, ssl, urllib.request
uri = os.environ["OPNSENSE_URI"].replace("host.docker.internal", "127.0.0.1")
headers = {"Authorization": "Basic " + base64.b64encode((os.environ["OPNSENSE_API_KEY"] + ":" + os.environ["OPNSENSE_API_SECRET"]).encode()).decode()}
for kind in ("general", "bgp"):
    request = urllib.request.Request(uri + "/api/quagga/" + kind + "/get", headers=headers)
    with urllib.request.urlopen(request, context=ssl._create_unverified_context(), timeout=15) as response:
        if json.load(response)[kind]["enabled"] != "0":
            raise SystemExit("disabled template sync refused enabled settings")
SYNC
}
required_env=(OPNSENSE_FRR_EXPECTED_CORE OPNSENSE_FRR_EXPECTED_PLUGIN OPNSENSE_FRR_EXPECTED_DAEMON OPNSENSE_FRR_GUEST_PROBE OPNSENSE_FRR_ORACLE_URI OPNSENSE_FRR_ORACLE_TOKEN OPNSENSE_URI OPNSENSE_API_KEY OPNSENSE_API_SECRET OPNSENSE_ALLOW_INSECURE OPNSENSE_DISPOSABLE_VM_NAME OPNSENSE_DISPOSABLE_VM_UUID OPNSENSE_FRR_IDENTITY_PROBE OPNSENSE_FRR_EXPECTED_DIGEST)
for name in "${required_env[@]}"; do
  [[ -n "${!name:-}" ]] || die "missing required environment variable: ${name}"
done
case "${OPNSENSE_URI}" in
https://127.0.0.1:18444 | https://host.docker.internal:18444) ;;
*) die "FRR acceptance only permits the isolated loopback forwarding endpoint" ;;
esac
[[ "${OPNSENSE_FRR_GUEST_PROBE}" = /* && -x "${OPNSENSE_FRR_GUEST_PROBE}" ]] || die "guest probe must be an absolute executable path"
probe="${OPNSENSE_FRR_IDENTITY_PROBE}"
[[ "${probe}" = /* && -x "${probe}" ]] || die "identity probe must be an absolute executable path"
identity_probe_before="$(shasum -a 256 "${probe}" | awk '{print $1}')"
guest_probe_before="$(shasum -a 256 "${OPNSENSE_FRR_GUEST_PROBE}" | awk '{print $1}')"
identity="$("${probe}")"
observed_name="$(sed -n 's/^name=//p' <<<"${identity}")"
observed_uuid="$(sed -n 's/^uuid=//p' <<<"${identity}")"
[[ -n "${observed_name}" && "${observed_name}" = "${OPNSENSE_DISPOSABLE_VM_NAME}" && "${observed_uuid}" =~ ^[0-9a-f-]{36}$ && "${observed_uuid}" = "${OPNSENSE_DISPOSABLE_VM_UUID}" ]] || die "disposable VM identity mismatch"
# Independently bind the configured API endpoint to the attested VM before any
# credentialed traffic. The probe gets this certificate through VM-owned QEMU
# forwarding over trusted SSH, not from OPNSENSE_URI.
attested_tls="$(sed -n 's/^tls_sha256=//p' <<<"${identity}")"
[[ "${attested_tls}" =~ ^[0-9a-f]{64}$ ]] || die "missing or malformed attested TLS fingerprint"
endpoint_tls="$(
  python3 - "${OPNSENSE_URI}" <<'TLS'
import hashlib, socket, ssl, sys, urllib.parse
uri = urllib.parse.urlsplit(sys.argv[1])
if uri.scheme != "https" or not uri.hostname or uri.username or uri.password:
    raise SystemExit("API URI must be HTTPS without credentials")
# Docker Desktop resolves this special hostname only in containers. The host
# leg uses the root-owned localhost tunnel; the Go precheck separately attests
# the exact container endpoint before constructing an authenticated client.
hostname = "127.0.0.1" if uri.hostname == "host.docker.internal" else uri.hostname
context = ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT)
context.check_hostname = False
context.verify_mode = ssl.CERT_NONE
with socket.create_connection((hostname, uri.port or 443), timeout=15) as sock:
    with context.wrap_socket(sock, server_hostname=uri.hostname) as tls:
        print(hashlib.sha256(tls.getpeercert(binary_form=True)).hexdigest())
TLS
)"
[[ "${endpoint_tls}" = "${attested_tls}" ]] || die "API endpoint TLS fingerprint mismatch"
# FRR dual-AF over IPv4 requires usable IPv6 next hops on this isolated WAN.
# Observe only: lab provisioning belongs to the parent, not to this runner.
wan_ipv6_observation="$("${OPNSENSE_FRR_GUEST_PROBE}" '/sbin/ifconfig vtnet0')"
python3 -c '
import ipaddress, sys
raw = sys.stdin.read()
if "IFDISABLED" in raw or "tentative" in raw or "duplicated" in raw:
    raise SystemExit("isolated WAN IPv6 is disabled or not ready")
addresses = [line.split()[1].split("%")[0] for line in raw.splitlines() if line.strip().startswith("inet6 ")]
if "2001:db8:2::15" not in addresses or not any(ipaddress.ip_address(address).is_link_local for address in addresses):
    raise SystemExit("isolated WAN requires provisioned global and link-local IPv6 next hops")
' <<<"${wan_ipv6_observation}"
export OPNSENSE_FRR_IPV6_READY_ATTESTATION="${observed_uuid}:vtnet0:2001:db8:2::15:linklocal:ready"
digest_before="$(test/scripts/story-31-0h-code-digest.sh)"
[[ "${digest_before}" = "${OPNSENSE_FRR_EXPECTED_DIGEST}" ]] || die "candidate digest mismatch"
# The disposable forwarding is exposed to Docker Desktop through the host
# alias. The exact alias endpoint is checked again, unauthenticated, in Go.
if [[ "${OPNSENSE_URI}" = https://127.0.0.1:18444 ]]; then
  export OPNSENSE_URI=https://host.docker.internal:18444
fi
export OPNSENSE_FRR_TLS_SHA256="${attested_tls}"
export OPNSENSE_FRR_COMPATIBILITY=1
export OPNSENSE_FRR_IDENTITY_ATTESTATION="${observed_name}:${observed_uuid}"
umask 077
mkdir -p test/acceptance-results
artifact_dir="$(mktemp -d "${repo_root}/test/acceptance-results/story31h.XXXXXX")"
printf 'identity_probe_sha256=%s\nguest_probe_sha256=%s\n' "${identity_probe_before}" "${guest_probe_before}" >"${artifact_dir}/external-probes-before.sha256"
container_artifacts="/workspace/${artifact_dir#"${repo_root}/"}"
printf '%s\n' "${digest_before}" >"${artifact_dir}/candidate-digest.sha256"
printf '%s\n' "${wan_ipv6_observation}" >"${artifact_dir}/wan-ipv6-readiness.txt"
mkdir -p "${artifact_dir}/.tmp" "${artifact_dir}/legacy-source" "${artifact_dir}/bin/legacy" "${artifact_dir}/bin/current"
git archive 154de46ec53a091a27edd27b69172247b214f3dd | tar -x -C "${artifact_dir}/legacy-source"
# Retain private raw artifacts on failure for reconciliation; the parent owns
# credential retirement. Nothing is printed until the full secret scan passes.
disabled_sync_worker "${artifact_dir}" &
sync_worker_pid=$!
trap 'kill "${sync_worker_pid}" 2>/dev/null || true' EXIT
set +e
docker run --rm --network host -v "${repo_root}:/workspace" \
  -v opnsense-go-mod:/root/go/pkg/mod -v opnsense-go-build:/root/.cache/go-build -w /workspace \
  -e TF_ACC=1 -e OPNSENSE_URI -e OPNSENSE_API_KEY -e OPNSENSE_API_SECRET -e OPNSENSE_ALLOW_INSECURE \
  -e OPNSENSE_DISPOSABLE_VM_NAME -e OPNSENSE_DISPOSABLE_VM_UUID \
  -e OPNSENSE_FRR_COMPATIBILITY -e OPNSENSE_FRR_IDENTITY_ATTESTATION -e OPNSENSE_FRR_TLS_SHA256 -e OPNSENSE_FRR_IPV6_READY_ATTESTATION \
  -e OPNSENSE_FRR_ORACLE_URI -e OPNSENSE_FRR_ORACLE_TOKEN \
  -e OPNSENSE_FRR_EXPECTED_CORE -e OPNSENSE_FRR_EXPECTED_PLUGIN -e OPNSENSE_FRR_EXPECTED_DAEMON \
  -e "OPNSENSE_FRR_ARTIFACT_DIR=${container_artifacts}" -e "TMPDIR=${container_artifacts}/.tmp" \
  ghcr.io/devrail-dev/dev-toolchain:1.12.24 bash -c \
  'go build -o "${OPNSENSE_FRR_ARTIFACT_DIR}/bin/current/terraform-provider-opnsense" . && (cd "${OPNSENSE_FRR_ARTIFACT_DIR}/legacy-source" && go build -o "${OPNSENSE_FRR_ARTIFACT_DIR}/bin/legacy/terraform-provider-opnsense" .) && go test -count=1 -v -timeout 60m -run "^TestAccFRRCompatibility$" ./internal/service/quagga' >"${artifact_dir}/runner.log" 2>&1
status=$?
set -e
printf stop >"${artifact_dir}/disabled-config-sync.stop"
wait "${sync_worker_pid}"
trap - EXIT
identity_probe_after="$(shasum -a 256 "${probe}" | awk '{print $1}')"
guest_probe_after="$(shasum -a 256 "${OPNSENSE_FRR_GUEST_PROBE}" | awk '{print $1}')"
printf 'identity_probe_sha256=%s\nguest_probe_sha256=%s\n' "${identity_probe_after}" "${guest_probe_after}" >"${artifact_dir}/external-probes-after.sha256"
scan_secret_artifacts "${artifact_dir}"
[[ "${identity_probe_before}" = "${identity_probe_after}" && "${guest_probe_before}" = "${guest_probe_after}" ]] || die "external proof probe changed during acceptance"
cat "${artifact_dir}/runner.log"
[[ ${status} -eq 0 ]] || die "FRR acceptance failed; private artifacts retained at ${artifact_dir}"
grep -Eq -- '--- PASS: TestAccFRRCompatibility( |$)' "${artifact_dir}/runner.log" || die "missing non-skipped live PASS"
for proof in 'FRR_ADVANCED public_fields=19 password_lookup=false' 'FRR_RESTORE settings_equal=true config_equal=true pf_equal=true rendered_pf_equal=true routes_equal=true' 'positive_prefix=198.18.31.0/24 default_export=false'; do
  grep -Fq "${proof}" "${artifact_dir}/runner.log" || die "missing ${proof}"
done
for kind in neighbor peergroup routemap prefixlist redistribution aspath communitylist; do
  for phase in TERRAFORM_DESTROY BEFORE_ROLLBACK FINAL_CLEANUP; do
    grep -Fq "FRR_${phase} namespace=${kind} total=0" "${artifact_dir}/runner.log" || die "missing zero-leftover proof"
  done
done
digest_after="$(test/scripts/story-31-0h-code-digest.sh)"
[[ "${digest_after}" = "${digest_before}" ]] || die "candidate changed during acceptance"
log_info "FRR_COMPATIBILITY_EVIDENCE digest=${digest_after} core=${OPNSENSE_FRR_EXPECTED_CORE} plugin=${OPNSENSE_FRR_EXPECTED_PLUGIN} daemon=${OPNSENSE_FRR_EXPECTED_DAEMON} restore_equal=true terraform_deleted=0 final_leftovers=0 secret_scan=pass legacy_live=unverified firewall_activation=explicit_lab_not_provider"
