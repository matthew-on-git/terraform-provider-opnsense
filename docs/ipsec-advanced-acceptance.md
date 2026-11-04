# Advanced IPsec disposable acceptance (unreleased)

## Evidence boundary

The 21 audited advanced controls are implemented on existing connection/child
owners. HTTP-backed framework protocol tests cover schema/lookup parity, actual
planning/apply, import, reads, omission/clears, drift, cancellation, malformed
responses and confirmed-save failure recovery. The separate bounded offline
`TestIPsecCoreImportedOmissionAndCSVNoChange` uses the actual local Terraform CLI/Core
against a loopback fixture to construct real no-change plans for configured CSV
presentation and imported omitted Optional+Computed fields. Neither category is
live proof. **No live acceptance has run and
zero appliance writes were made for this story.** Parent story 31.0g remains
in progress and its PR must remain draft until the lifecycle gate is satisfied.
These pending capabilities are not a released-provider claim.

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
4. Supply only that guest's API credentials explicitly. There is no discovery,
   credential or target fallback. `OPNSENSE_ALLOW_INSECURE=true` is optional only
   for an attested disposable guest with a self-signed certificate.

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
  "identity_endpoint": "/api/diagnostics/system/system_information",
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

Run inside the pinned DevRail container (all toolchains stay in the container):

```sh
TF_ACC=1 \
OPNSENSE_IPSEC_DISPOSABLE_ACK=I_RESERVED_A_NEW_ISOLATED_GUEST \
OPNSENSE_IPSEC_ISOLATION_RECEIPT=/private/isolation-receipt.json \
OPNSENSE_URI=https://firewall.example.test \
OPNSENSE_API_KEY="$DISPOSABLE_API_KEY" \
OPNSENSE_API_SECRET="$DISPOSABLE_API_SECRET" \
go test -count=1 -v ./internal/service/ipsec -run '^TestAccIPsecAdvancedDisposable$' -timeout 45m
```

Mount the private receipt read-only into the container and pass the explicitly
set environment variables; do not put credential values in a command history.
This is an executable Go acceptance test, not permission to use a shared target.
Missing/expired/mismatched identity, reservation, isolation, firmware or nonempty
inventory fails before mutation. Ordinary offline `go test` reports this test as
**skipped**, not live proof.

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
cleanup after that cutoff; identity must still match. An independent `t.Cleanup`
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
do not delete arbitrary objects or release the guest as clean. A live PASS is
required before changing the parent story/audit to complete. Do not auto-merge.
