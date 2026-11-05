# Advanced IPsec disposable acceptance (unreleased)

## Evidence boundary

The 21 audited advanced controls are implemented on existing connection/child
owners. HTTP-backed framework protocol tests cover schema/lookup parity, actual
planning/apply, import, reads, omission/clears, drift, cancellation, malformed
responses and confirmed-save failure recovery. The separate bounded offline
`TestIPsecCoreImportedOmissionAndCSVNoChange` uses the actual local Terraform CLI/Core
against a loopback fixture to construct real no-change plans for configured CSV
presentation and imported omitted Optional+Computed fields. Neither category is
live proof. The final guarded **26.7.8_2** run now passes at `c03b2e0`, including
real service activation, Terraform deletion, namespace zero, exact restoration,
final digests and independently verified retirement. The sanitized
[receipt](../_bmad-output/implementation-artifacts/31-0g-live-proof.json) identifies
the exact tested source. This is the bounded 21-control increment, not negotiated
IKE/packet proof or completion of inherited legacy limitations. Until merge and
automatic publication, this remains a pending-provider capability.

The oracle is released core **26.7.8**, commit
[`96978001a8343669154e4492172ce0a07dac7e4b`](https://github.com/opnsense/core/tree/96978001a8343669154e4492172ce0a07dac7e4b).
The [model](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/models/OPNsense/IPsec/Swanctl.xml)
and [controller](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/controllers/OPNsense/IPsec/Api/ConnectionsController.php)
define `/api/ipsec/connections/*`; the older swanctl-controller URL is not used.

## Operator preparation — private receipt, never production/shared targets

1. Newly reserve a dedicated disposable 26.7.8 guest, with no production
   connectivity, an empty IPsec connection/child inventory, and an active lease.
   Keep its infrastructure reservation/ownership record privately.
2. Independently establish that guest's identity through a read-only core or
   diagnostics API endpoint. Hash its exact stable response bytes with SHA256;
   choose a stable machine-specific endpoint, not firmware status alone. Do not
   fingerprint whichever appliance an inherited credential happens to reach.
3. Write a private JSON receipt matching the shape below. All attestations are
   mandatory; the lease must be active, created within 24 hours and expire no
   more than two hours after preflight, with at least **60 minutes remaining**
   (45 minutes execution plus 15 minutes reserved cleanup). Reserve about 90
   minutes to avoid racing the minimum at startup. The test checks the exact target origin,
   independent identity hash, released version and empty inventory before writes.
   Infrastructure reservation truth is the operator's responsibility; the test
   cannot query an infrastructure reservation service or validate a signature.
4. Supply only that guest's ephemeral API credentials and explicitly trusted TLS
   certificate. The bounded runner rejects insecure TLS. There is no discovery,
   credential or target fallback. Keep the root, receipt and recovery journal
   private (0700 directories and 0600 data).

```json
{
  "target_uri": "https://firewall.example.test",
  "guest_id": "new-disposable-guest",
  "reservation_id": "private-reservation-receipt",
  "reserved_by": "acceptance-operator",
  "purpose": "ipsec-disposable-acceptance",
  "created_at": "2026-11-03T12:00:00Z",
  "expires_at": "2026-11-03T13:30:00Z",
  "isolated": true,
  "newly_created": true,
  "disposable": true,
  "no_production_connectivity": true,
  "allow_service_reconfigure": true,
  "identity_endpoint": "/api/diagnostics/system/systemInformation",
  "identity_sha256": "<independently-established-64-hex-response-hash>"
}
```

The endpoint above is a shape example, not a verified identity oracle: choose
an endpoint the reserved guest actually exposes with stable machine-specific
content. Never commit the completed receipt, hostnames, private inventory,
credentials or unsanitized test output.

## Dedicated explicit invocation

The general `TF_ACC=1` acceptance workflow **skips this dedicated test** when
`OPNSENSE_IPSEC_DISPOSABLE_ACK` is absent. Supplying the acknowledgment opts in:
a wrong acknowledgment, missing/invalid receipt or failed isolation proof is an
error, never a fallback or silent skip. Do not add the acknowledgment to general
acceptance environments.

## Executable guarded path

Run this host-side launcher; it starts the pinned DevRail container (all toolchains stay in that container):

```sh
OPNSENSE_IPSEC_GUARDED_RUNTIME=/private/runtime.json \
OPNSENSE_URI=https://firewall.example.test \
OPNSENSE_API_KEY="$DISPOSABLE_API_KEY" \
OPNSENSE_API_SECRET="$DISPOSABLE_API_SECRET" \
SSL_CERT_FILE=/private/disposable-ca.pem \
bash test/scripts/run-story-31-0g-acceptance.sh
```

The runtime JSON binds owner, guest UUID, endpoint, TLS pin, immutable guard/oracle
executables and their SHA256 hashes, the reviewed source digest, a private
`isolation_receipt` and a durable `journal` path. Its helpers must independently
attest live infrastructure isolation, authentic installed **26.7.8_2**, and exact
reserved identity; they must not synthesize firmware or API responses. The
wrapper verifies source/runtime/receipt digests and requires a non-skipped Go
PASS, independent namespace-zero observations and matching final configuration,
Swanctl-settings, routes and runtime hashes. See the Python runner's documented
helper contract. A raw `go test` pass alone is not the complete guarded receipt.

The default outer command budget is 40 minutes (maximum 40), with a shorter Go
budget; normal writes stop 15 minutes before hard lease expiry. On interruption
or timeout the external runner terminates the owned Go/Terraform process group,
proves no live group members remain, then independently attempts recorded-owned
cleanup, zero verification and exact restoration. It refuses restoration while
owned processes might still mutate the appliance. Fallback recovery cannot turn
a failed Go run into acceptance. Missing/expired/mismatched identity, reservation,
isolation, firmware or nonempty initial inventory fails closed. Ordinary offline
`go test` reports this dedicated test as **skipped**, not live proof.

The test creates disabled connection/child objects with all advanced controls,
uses UUID lookups, imports both, checks a no-change plan, injects owned connection
and child drift, observes/repairs it, applies allowed blank/false/zero clears,
includes an update omitting all advanced controls after import and compares
independent raw before/after saved-field snapshots, then deletes the child
independently before deleting its parent. Terraform's
cleanup owns only the UUIDs created by this run. Raw missing-object GET evidence
is required for **both** deletions; transport/read failure is never accepted as
absence. A guarded underlying retry transport rechecks lease time and the independently
pinned identity before every mutation attempt, including provider activation and
direct drift injection. Normal mutations stop with 15 minutes left for cleanup.
Only recorded owned-UUID deletion and its pending reconfigure remain allowed for
cleanup after that cutoff; identity must still match, and **all** writes stop at
hard lease expiry, including after slow identity verification or journal sync.
The journal atomically records each add dispatch before the HTTP call and binds
confirmed UUIDs to this reservation. Lost, truncated or otherwise unconfirmed add
outcomes remain unresolved and block further writes and clean reconciliation;
inventory searches never confer deletion authority. An independent `t.Cleanup`
fallback uses the runner's guarded client and a fresh two-minute context even if
Core refresh-before-destroy failed or was canceled. It snapshots only UUIDs
recorded from this run's affirmative adds, deletes children before parents,
activates confirmed owned cleanup, and proves each exact absence. Transport,
identity or absence uncertainty retains ownership and blocks parent cascading;
the fallback never clears the original failing Terraform outcome. Reconfigure affects the
whole IPsec service, hence the isolation gate.
No auth, certificate, key or PSK objects are created. Multi-proposal/address/pool
selection and unrelated policy preservation are additionally covered offline.

## Required review evidence

Retain sanitized test version, exact source revision, named test results,
identity/reservation attestation outcome, each create/import/no-change/drift/
clear/delete observation and independent child/parent absence, plus cleanup and
reservation release. Keep private raw receipts outside this public repository.
If cleanup fails, retain ownership/reservation and reconcile the recorded UUIDs;
do not delete arbitrary objects or release the guest as clean. A complete guarded live PASS is required before changing lifecycle readiness.
Merge authority is separate from the runner and its evidence; follow the current
maintainer authorization and exact-head review/CI requirements.

## Sequential merge review repairs (2026-11-04)

Confirmed UUID create failures now emit warnings instead of error-tainting the
saved resource in Core; unconfirmed creates still fail. Four actual Core controls
fail on the old code and pass after repair: two applies and a no-change plan keep
one add, zero saves and zero interim deletes, followed by one final cleanup
delete. A discarded speculative refresh may repeat read-time activation recovery;
it never replays an object write. Returned, consumed private markers are tested
separately for zero reactivation. Deletes require source-backed affirmative
results, and malformed selection members/domain values fail without erasing
prior state. True/false aggressive and SHA256-96 lookup/drift controls are executed
offline, including hard-coded-false negative controls.

The revised guarded live run passed completely; offline results remain distinct
from that appliance evidence. The current integration targets **v0.15.0**, following published Trust
v0.14.0 and retry-settings v0.13.0.

## Final guarded receipt (2026-11-04)

The fresh full retry passed the non-skipped real Terraform matrix: 11 configured
steps, both persistent imports, all 21 lookup assertions, omission/no-change,
both drift/repair cases, true-to-false and blank/zero clears, independent child
deletion and final parent deletion. Both resources remained disabled; no
negotiated IKE or packet-flow claim is made. Independent observations confirm
zero owned/unresolved objects immediately and finally. Original raw config,
Swanctl settings, routes and generated runtime hashes match; final source,
runtime and receipt digest gates passed and wrapper exit was zero. The exact
owned disposable, disk, marker, acceptance container, oracle, tunnel and
listeners were retired and independently absent.

The earlier full attempt remains **failed**, despite its Go pass and zero
objects. Its restoration comparison rejected only the model-save `persisted_at`
attribute. Pinned `BaseModel::toXML` stamps this cache/save metadata on every
versioned-model serialization; installed bytes match the pinned source. Recovery
permits comparison differences only in the unique owned Swanctl timestamp and
direct revision metadata, rejects policy/global/other-model or runtime/routes
drift, and restores the original raw config under the original inode lock.
**Final hashes still include every original byte**, including timestamps; they
are not normalized. Twenty independent counterexample controls pass. The
original failed inputs/receipts remain private, and only the subsequent full
fresh successful runner is acceptance evidence.
