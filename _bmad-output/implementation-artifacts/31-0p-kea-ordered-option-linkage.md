# Story 31.0p: Kea Ordered Generic-Option Linkage

## Status

blocked — released-source gate complete on provider `main`
`b901ddd3543567075bc282573f0c62905bb28df2` and OPNsense core tag 26.7.8.
No authoritative ordered parent read exists. Parent links are not
implemented. This is a resumable gate, not feature support.

Evidence:
`../planning-artifacts/kea-ordered-option-linkage-api-research.md`.
Unsubmitted proposal:
`../planning-artifacts/kea-ordered-option-linkage-upstream-proposal.md`.

## Scope and API research gate

Tagged core 26.7.8 has a multi-valued `option` relation on subnet and
reservation models. Stored CSV order drives order-sensitive flex-option
generation: the last matching same-code action wins, and reservation
actions follow subnet actions. Current provider parent owners expose
neither relation. Parent GET choices are description-sorted rather than
stored order. Do not model this as an unordered set or infer order from
JSON map keys.

1. Authoritative ordered read: absent on released 26.7.8. `get_subnet`
   and `get_reservation` return label-sorted selected maps because the
   relations do not set `Sorted`. Grid search CSV and raw XML are not
   substitutes. Details and file hashes are in the research record.
2. Do not extend `opnsense_kea_dhcpv4_subnet` or the reservation owner
   until a later released tag documents an order-preserving parent read
   and that read is proven on that firmware. Then add typed ordered
   option UUID references on those existing owners. Preserve omitted
   fields and existing `option_data`/PXE ownership. Do not create a
   competing parent owner.
3. After that contract exists, prove order, reorder, import, no-change,
   drift, clear, and dependency deletion through real Core plus
   independent installed-model observations on disabled disposable
   services. Protect already configured links across unknowns and
   restarts.
4. Update release notes, docs, support/audit, and complete
   full/review/hosted gates only when implementation is actually made.
   This research change is planning/docs only and must not add an
   unreleased changelog heading.

## Acceptance criteria

- Given two same-code options, when linked or reordered in supported
  HCL, then the authoritative order and generated precedence match,
  parent UUIDs stay stable, and import plans no change. Not met.
  Implementation is blocked until the read contract exists and is
  proven.
- Given unavailable ordered reads, when planning support, then
  ownership remains explicitly research-gated rather than falsely
  supported through sorted maps, search CSV, or XML. Met by the
  26.7.8 source gate and the upstream-blocked classification.
- Given omission, unknown intent, or failed write/activation, when
  recovering, then existing relations and unrelated parent settings are
  not lost. Preserved by not writing links in this story.

Active packet delivery, OpenTofu lifecycle, and private adoption remain
separate validation gates. DHCPv6 has the same unsorted relation shape
and is not implemented here.
