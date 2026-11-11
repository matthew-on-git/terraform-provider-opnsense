# Story 31.0i verification checkpoint

Status: implementation and historical bounded live read-only acceptance verified;
three independent reviews collected and triaged, repaired changes independently
reviewed. Final local gates pass; PR and exact-head hosted checks are separate publication gates.
No release claimed.

## Source binding

Baseline commit: `b090de1cf22e467c32e7711d2b09a6593f5d2bf1`.
The tested candidate was uncommitted. SHA256 of the sorted `sha256  path` inventory
of all tracked/nonignored Go source files (one trailing newline per file):
`6e25f18867fe194d7a06e470793159c71ea0c5460167f2e7b9ae36460ad600dd`.
The private runner retains the full inventory and separate failed/fresh receipts.
No Go source changed between the frozen binary build and this digest capture.

## Executed matrix coverage

| Matrix row | Executed passing controls |
| --- | --- |
| Diagnostic read | `TestLiveInterfaceFullAllowlist`, `TestLiveInterfaceSourceWireNames`, `TestLiveInterfaceMalformed`, `TestLiveInterfaceCoreStateAndNoChange` |
| Empty inventory | `TestLiveInterfaceMalformed`, `TestLiveInterfaceCoreExactSelectionAndEmptyInventory` |
| Overview | `TestLiveInterfaceFullAllowlist`, `TestLiveInterfaceSourceWireNames`, `TestLiveInterfaceMalformed`, `TestLiveInterfaceDuplicateSetMembers` |
| Determinism | `TestLiveInterfaceNullAndOrdering`, `TestLiveInterfaceCoreStateAndNoChange` |
| Failures | `TestLiveInterfaceHTTPAndBoundedRead`, `TestLiveInterfaceNoRedirectCredentialForwarding`, `TestLiveInterfaceDuplicateDiagnosticKey`, exact-selection Core errors |

Focused race tests ran the controls above and shared client/audit regressions.
The opt-in live harness was separately executed (not counted from its default
skip). Actual Core and an independent complete-field oracle passed for all four
sources after apply and refresh, with a no-change plan and zero upstream writes.
See the owning story for tested versions and bounded proof limitations.

## Independent review and repaired-source evidence — 2026-11-10 UTC

All three independent review layers completed against the identical frozen patch
SHA256 `d26983cd34c1d0ec419fab53479ef9e79ed207adcbbcf0594ab6c86f2032fc65`.
Blind review returned eleven findings; edge-case review found no unhandled path;
verification-gap review returned two regression gaps and a harness limitation.
Every finding has a separate verdict and disposition in the spec triage log.
The original unavailable implementer was not impersonated: parent applied the
trivial patches using the rendered step04 fallback. Independent bounded repair
review confirmed the fixes and unchanged allowlist; one documentation nit was
corrected. The standing pagination guidance is deferred under step04's explicit
agent-context-file rule. No unrelated parent story is marked complete.

Final Go source-inventory SHA256 (same inventory procedure as historical binding):
`26ad5c0251c9357755ad4f3ab83df80818d8b4f1e619e6e8e7f86173fe22eefd`.
Focused race/Core/client/audit tests passed. Final docs generation ran twice and
produced identical complete diffs. The repaired source was exercised against
privately retained historical appliance responses through actual Terraform Core
apply, refresh, detailed-exit-code no-change plan and the independent all-field
oracle. All four sources matched every allowlisted field. This was **recorded
appliance-observation replay**, with zero new upstream requests; it is not new
live-appliance proof and does not replace or broaden the historical live result.
The private replay listener was retired. Raw bodies, state and inventory remain
outside the public repository and PR.

## Scanner diagnostic and final delivery gates

The first final local `make check` passed lint, format, test, security and docs,
but scan identified a historical false positive on `ipsec_pending_activation_v1`
at commit `3a5a84b4f9a56604ef51dd2981b781d150a794bd`,
`internal/service/ipsec/advanced.go`, line 316 (`generic-api-key`). Source inspection
confirmed this is a deterministic Terraform private-state activation-marker name,
not credential material. The failure and diagnostic were retained separately.

The repository's `.gitleaksignore` specifically instructs recording false-positive
fingerprints and already records the same class of Kea/Unbound private-state names.
The repair adds only this exact historical commit/file/rule/line fingerprint,
without broad rule suppression, scanner disabling, or concealing a real secret.
The source tree's actual secret protection and all normal scanner invocations
remain enabled. Final local gates and eight exact-head hosted checks must pass
before PR readiness is reported. No merge or release authority is inferred.

## Final local checkpoint

Final pinned `make check` passed all six categories in 71.667 seconds with zero
skipped categories. This follows the retained failed first scan; the passing
result does not erase that diagnostic. Independent review verified the exact
non-credential marker classification. Focused race/Core/client/audit tests,
deterministic generation twice, whitespace checks and full public-diff privacy
inspection passed. Review handoffs and all raw appliance/state evidence remain
private and are not staged. The spec is implementation-complete and the story
is in maintainer `review`; unrelated parent stories remain unchanged.

The actual commit and eight hosted check outcomes are recorded in the linked
PR/maintainer handoff after publication, not inferred from this local checkpoint.
No automatic merge, release or new appliance proof is claimed.
