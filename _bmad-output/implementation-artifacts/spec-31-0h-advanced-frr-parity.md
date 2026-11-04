---
title: 'Story31.0h: advanced FRR neighbor and policy parity'
type: 'feature'
created: '2026-11-04'
status: 'done'
baseline_revision: '4bd107fc769586d5904919c8e69562897a87e84d'
review_loop_iteration: 0
followup_review_recommended: true
context: ['AGENTS.md', '_bmad-output/implementation-artifacts/epic-31-context.md']
warnings: [oversized]
deferred: []
---

<intent-contract>

## Intent

**Problem:** FRR2.0 compatibility is released, but the existing neighbor owner omits advanced BGP fields and peer groups omit listen ranges. Source-valid mode-only/inherited neighbors cannot be described by the required numeric remote AS. Policy validation and advanced lifecycle proof remain incomplete.

**Approach:** Finish remaining buildable neighbor/peer-group and route-map parity on the existing owners, preserving installed-plugin gating, grouped references, per-family semantics, recovery and old identities. Add safe read-only public fields to existing lookups, never passwords. Prove protocol/Core and disposable lifecycle boundaries separately, retaining explicitly unexecuted legacy/OpenTofu rows.

## Boundaries & Constraints

**Always:** Use pinned released plugins26.7.7 (484ea085f28fdf4f701a9e503ed33e117c1425cd, FRR2.0 model1.1.5), 26.7.6 (45b79b80aff277fc9edc97ce18bcd91177c40554,1.55 model1.1.2), and existing2.0_1 pin5ac2959c8514b4f60e94d65439be5f629f10521f for contract differences; inspect BGP.xml, template and API controller before behavior changes. Implement all20 omitted neighbor fields: remote_as_mode, local_as (wire localas), password, weight, local_ip (localip), link_local_interface (linklocalinterface), next_hop_self_all (nexthopselfall), multihop, route_reflector_client (rrclient), soft_reconfiguration_inbound, bfd, bfd_strict, connect_timer (connecttimer), default_originate (defaultoriginate), as_override (asoverride), remove_private_as (removeprivateas), allow_as_in (allowas_in), disable_connected_check, attribute_unchanged (attributeunchanged), capabilities. Match source wire names exactly. Add peer-group listen_ranges (listenranges) with strict CIDR semantics. New public properties Optional+Computed without static defaults preserve imports/omission; explicit false, zero where permitted and empty string/set are real configured intent. Null optional numeric is unset, not invented0. Relax neighbor remote_as to Optional+Computed while preserving existing numeric configs and type; mode takes precedence per template. Use source-valid relationships and validators (localas1..4294967295, weight0..65535, connecttimer1..65000; verify others). Authentication password is Optional Sensitive CONFIG-OWNED input only: omit null from requests, explicit empty clears, preserve prior configured input on refresh, never adopt remote password, never expose through lookup or diagnostics. Local IP must satisfy source password coupling without echoing secret. No password API response field or raw response logging. Document that Sensitive state is not encryption and remote password drift cannot be proven. Preserve global serialization, private pending activation, strict read/save/reconfigure outcomes and version/model discovery before writes. Reject configured unknown/unsupported values before writes. Keep prior schema-v0 migration working with new fields.

**Never:** Duplicate released compatibility; create another owner/registration; silently drop configured values; turn scalar attributeunchanged into a set; use XML enum element labels instead of option value attributes; treat policy UUID as route-map sequence; change unrelated story status; claim pending work released; print credentials, private inventory or local paths publicly; perform production writes or merges. Live evidence requires newly reserved positively attested isolated guest and exact restore/cleanup. Missing old firmware proof is explicit, not mocked appliance proof. Keep parent31.0h open if any required acceptance proof remains.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Advanced adoption | Imported neighbor with all public fields, omitted configuration | Authoritative state and actual Core no-change plan; unrelated update preserves options | Missing/malformed owned API values fail without replacing prior state |
| Explicit changes | false, weight0, clears, mode-only or inherited AS, listen-range set | Exact supported wire values; stable identity; update/drift/read converge | Unknown configured values, invalid ranges/enums, malformed CIDRs fail before mutation |
| Authentication | omitted/password/explicit empty, local IP coupling | Omitted secret never adopted or cleared; configured input preserved privately in sensitive resource state; lookup has no password | No secret in diagnostic; invalid coupling stops before write |
| Policies and families | IPv6-only/dual family with grouped refs and sequence changes | Group membership and stable UUIDs retained; sequence numeric not UUID; AF placement preserved | Unsupported model/plugin and stale relation fail closed |
| Activation failure | saved advanced update then failed read/reconfigure/cancellation | Confirmed state/identity and pending recovery survive; no repeated save | Strict failure remains visible and retry only activates |

</intent-contract>

## Code Map

- `internal/service/quagga/bgp_neighbor_{schema,model,resource,data_source}.go`: existing schemaVersion1 neighbor, generic guarded CRUD; currently17 fields; separate public lookup model or explicit omission-safe mapping is required for config-owned password because datasource schema lacks it.
- `compatibility{,_family,_decode,_presence,_lifecycle,_recovery,_mutation,_policies}.go`: inspect actual filenames; reflection-based scalar presence/validation and payload assembly, unknown/omission resolution, group canonicalization, pending activation. Extend these contracts rather than bypassing them. New fields must not make old version0 state upgrade lose values or become unreadable. Existing discovery accepts1.46–1.55; preserve shipped base support, but advanced configured fields require exact supported capability or authoritative field-presence proof BEFORE mutation (not failed readback after save). For unconfigured unsupported advanced fields retain null; never infer full1.55 schema on older packages. Extend exactBGPPayload selector handling for all new selected fields/capabilities. Transport-wide vs AF placement follows pinned templates.
- `bgp_peergroup_{schema,model,data_source}.go`: sole model field gap listenranges; existing peer remote-AS mode uses numeric0 sentinel, preserve compatibility. Add source-backed name/mode/numeric validators without arbitrary stricter policy.
- `route_map_{schema,model,resource}.go`: AS-path/community/prefix relations already implemented, API id is sequence; add range1..65535 and source regex `^[a-zA-Z0-9._-]{1,64}$` where applicable. No duplicate policy linkage implementation.
- `compatibility_protocol_test.go`, `compatibility_review_test.go`, `frr_*test.go`: reuse real protocol server, HTTP fixtures, grouped membership oracles and recovery tests; add full actual Core integration with fixture and semantic negative controls.
- Source models `net/frr/src/opnsense/mvc/app/models/OPNsense/Quagga/BGP.xml`, matching templates and `Api/BgpController.php` at pins above. Enums use XML option values: removeprivateas `''`, `remove-private-AS`, `remove-private-AS all`, `remove-private-AS replace-AS`, `remove-private-AS all replace-AS`; allowas_in `''`,`1`..`10`,`origin`; attributeunchanged `''`,`as-path`,`next-hop`,`med`; capabilities set extended-nexthop,dynamic,fqdn,link-local. Verify source before using.
- `templates/{resources,data-sources}/quagga_bgp_neighbor.md.tmpl`, peer-group/route-map templates and generated docs/examples: document omission, source semantics, authentication sensitive state and existing registration identity. `CHANGELOG.md` next0.13.0 Unreleased from publishedv0.12.0; pending sibling PRs not dependencies, version refresh required before merge.
- Owning `31-0h-quagga-bgp-field-parity.md`, ecosystem audit JSON/report and focused audit tests: update bounded remaining scope/evidence without falsely promoting full lifecycle or unrelated audit rows.
- Existing `frr_compatibility_acceptance_test.go` has historical hardcoded VM identity/version. Do not run against new target unchanged: add configurable identity/version/oracle bindings with fail-closed checks, no operator identifiers in added public lines. Extend advanced matrix and real source digest evidence. Parent owns private provisioning/oracle/restore and safe runtime configuration, not public repository.

## Tasks & Acceptance

**Execution:**
- Neighbor/peer-group models, schemas, payload/read/lookup boundaries -- implement the complete remaining field surface and correct enums/null semantics.
- Shared owner-local compatibility and policy code -- maintain strict discovery, omission and recovery; validate route-map order/name and peer fields from source.
- Focused protocol/Core/negative-control tests -- cover each matrix row with executed assertions, legacy schema upgrade, no secret lookup, no writes for invalid intent, actual plan/apply/import/refresh and no-change outcomes.
- Templates/docs/examples/changelog/audit/owning story -- generate deterministic outputs and label released vs candidate vs appliance proof honestly.
- Guarded acceptance path -- allow fresh isolated target bindings, advanced read/import/update/drift/AF/generated-policy observations, and restoration/zero-leftover checks; leave draft if any live row is unavailable.

**Acceptance Criteria:**
- Given source-supported legacy and2.0 models, when the actual provider protocol/Core operates all new fields, then import/adoption/update/drift and omission/clear converge at existing addresses with exact enums and no identity churn.
- Given grouped route maps and selected families, when sequence/references change, then stable UUID relationships and source-generated ordering/AF semantics remain correct.
- Given authentication input and a read-only lookup, when read/update/import happen, then no appliance secret is synthesized into state or cleared by omission, and lookup cannot expose it.
- Given a newly attested disposable runtime, when advanced lifecycle executes, then real core/plugin/digest, effective generated settings, activation diagnostics, exact restoration and zero owned leftovers are evidenced; any missing legacy/OpenTofu/route-collector row stays explicitly unverified and parent story open.

## Spec Change Log

## Review Triage Log

### 2026-11-04 — Review pass1
- Execution: all four independent layers completed; the native fourth simultaneous launch hit the thread limit and intent ran after a slot freed. No layer skipped.
- verdicts:17 findings — high2, medium11, low1, false3, maybe-false0.
- findings:
  - `[high]` `[patch]` Blind1 peer-field absence exemption — generic neighbor-name exemptions weakened old peer reads; scoped exemptions by concrete response type and tested missing required peer fields.
  - `[medium]` `[patch]` Blind2 malformed capability domains — decodable invalid numbers/enums counted as proof; template proof now validates source domains with negative controls.
  - `[medium]` `[patch]` Blind3 interface metadata — pinned versions could admit malformed options; discovery now validates selected/value metadata before accepting interface keys.
  - `[low]` `[reject]` Blind4 earlier schema validation — actual runtime validators reject invalid values before any mutation with actionable diagnostics. Additional custom schema-validator branches offer earlier timing only; not worth duplicating host/CIDR logic. Per-element safety is separately repaired under Edge1.
  - `[medium]` `[patch]` Blind5 numeric clearing — real Optional+Computed adoption limitation, shared with existing timers. Documented no numeric clear sentinel and explicit parent-story interface-design follow-up; null is never represented as a remote clear. No new public clearing API invented during review.
  - `[medium]` `[patch]` Blind6 password clear Core coverage — actual Core now proves one empty-password save and subsequent no-change/no-save.
  - `[medium]` `[patch]` Blind7 advanced activation recovery — saved update failure tests now carry configured authentication, zero weight, scalar and set intent; assert stable identity/private recovery and no replay.
  - `[medium]` `[patch]` Blind8 realistic old model — added genuinely absent1.46 schema/base lifecycle and advanced-intent zero-write rejection, rather than assuming modern fields.
  - `[medium]` `[patch]` Blind9 generated flags — all four configured capabilities and disable-connected-check are now required in the live oracle.
  - `[medium]` `[patch]` Blind10 group and reflector binding — live assertions bind each range to its intended peer-group and require reflector exclusively within IPv6 AF.
  - `[medium]` `[patch]` Gap1 adoption regex — literal backslash-n removed nothing; corrected expression, asserted actual assignment removal, and checked non-null zero/false/scalar/set no-change Core state.
  - `[high]` `[patch]` Edge1 CSV set boundary — one configured element could expand to multiple remote members; model resolution now validates/rejects comma-containing members before serialization, with protocol zero-write tests for both sets.
  - `[medium]` `[patch]` Edge2 version prefixes — exact parsed version tokens now replace substring matching and logs use observed verified values; mismatch controls executed.
  - `[false]` `[reject]` Intent1 API tests are not executed cross-version lifecycle — source evidence explicitly labels read-only comparison and no live lifecycle claim was made. Newly reserved actual matrix evidence is a separate required publication gate; old firmware proof remains unverified.
  - `[medium]` `[patch]` Intent2 order oracle — sequence presence alone was insufficient; live matrix now reverses relative order, requires generated relative positions/same UUIDs/group binding and active-session current-prefix withdrawal, then restores original positive gate.
  - `[false]` `[reject]` Intent3 comprehensive operational AF matrix — this diff makes no blanket operational/legacy/OpenTofu claim; owning story remains open and the additional IPv6 generated assertions are explicitly bounded. Existing positive collector and no-default gates remain required.
  - `[false]` `[reject]` Intent4 more route-map controls — pinned BGP model fields are already mapped by existing match/match2/match3/set owners; remaining substantive omissions were neighbor fields and listen ranges, not a missing parallel route-map surface.

## Design Notes

Additive config-owned password needs a separate datasource state type or equivalent mapping: the framework rejects a resource model containing fields absent in datasource schema. Reuse existing tested omission/recovery machinery. BFD strict/next-hop-self-all dependencies and password/local-IP coupling must follow released template/model; never silently report an ineffective advanced flag as active.

## Verification

- All Go/test/lint/generation commands in `ghcr.io/devrail-dev/dev-toolchain:1.12.24` only, caches permitted.
- Focused `go test -count=1 -race -v ./internal/service/quagga/... ./internal/coverageaudit/...`: every matrix test actually runs, none skipped except explicitly separate live suite; repair failures.
- Actual Terraform Core fixture tests and meaningful counterfactuals verify no-change, omission, drift, authentication boundary and ordering; parent independently reruns.
- Stage intended outputs, container `go generate ./...`, repeat and require zero unstaged output differences.
- `make check`: all6gates pass; `git diff --check`; full added-line privacy audit.
- Live advanced acceptance only after parent attests new runtime and freezes source; exact restore/cleanup required. Public proof uses versions/digests/counts only, no private paths/identities.

## Auto Run Result

Implementation run delivered the remaining20neighbor fields, password-free public lookup, config-owned Sensitive authentication, strict peer listen ranges and route-map validation on existing identities. Files: neighbor/peer schema-model-lookup and owner-local compatibility code implement fields, capability checks, read/omission/recovery; protocol/Core tests establish lifecycle and invalid-input boundaries; templates/generated docs/examples/changelog/audit capture supported semantics and residual limits; guarded runner/live oracle bind fresh identity/version/source and test advanced generated/collector behavior.

All17 review findings have individual verdicts above:13patch findings in13root-cause corrections (2high,11medium), one cosmetic timing improvement rejected and3false claims refuted. No pre-existing finding deferred in this run. Follow-up review recommended:true because capability/read omission, authentication recovery and per-family policy observation interact across multiple repaired paths; fresh live proof does not replace independent older-plugin/OpenTofu review.

Parent initial verification: container focused race suites and actual Core matrix passed; deterministic generation produced zero unstaged diff; make check passed6/6, none skipped. Counterfactual read-only overlay dropping numeric zero failed the expected wire assertion without changing candidate source. After review repairs, parent must repeat focused/full checks and the frozen isolated advanced matrix before publication. Live evidence records the exact candidate digest separately in31-0h-advanced-live-evidence.md (deliberately excluded from recursive source hashing); absence of a passing evidence artifact is not successful appliance proof.

Only this implementation/review run may be marked done. Parent31.0h remains open for numeric clear-interface design, legacy appliance and OpenTofu proof and any unexecuted required live rows. No merge/release/production adoption is asserted.

### Parent post-review verification

Focused container race suites passed (quagga7.189s, actual Core3.68s); all new read/capability/CSV/old-schema/clear/recovery matrix tests ran and passed. Full make check passed6/6, none skipped (84464ms). Parent deterministic generation and whitespace/privacy audit passed. Bounded independent source recheck confirmed original read/capability/interface and live-binding findings resolved; no new defect in that scope. Live acceptance is recorded separately, never inferred from these offline results.

### Actual acceptance repair

The first isolated matrix reached advanced creation but failed its public-lookup assertion because Terraform's synthetic flatmap `%` property counts the resource-only password. This is retained as a failed run, not appliance PASS. Terraform deletion independently reported zero in all7namespaces; exact settings/generated FRR/effective and rendered PF/routes restoration passed. The smallest test-only repair ignores only top-level synthetic `%` and password while comparing every actual public key in both directions, including explicit empty values and set counts/members. Focused negative controls prove real missing/extra/changed fields still fail. The whole frozen matrix must rerun after this repair.

### Further actual acceptance observations

Both generic/advanced public flatmap sites now share a bidirectional comparator that excludes only synthetic top-level `%`, resource-owned password and independently canonicalized relation prefixes; negative controls retain every real public/empty/set/password boundary. Third actual attempt reached Step12 after create/no-change/nine imports and advanced generated/effective/collector checks. It failed an unsupported textual sorting assertion: released sortDictList renders string IDs30before5, while the real active collector withdrew the approved prefix under numeric deny5 precedence. Historical review row Intent2 above records the initial repair, not a passing claim. The smallest source-backed correction now requires exact declared numeric values and mandatory active-session current-prefix withdrawal, retains UUID/group/default/restore gates and has a lexical-rendering regression. Full corrected17-step proof remains pending; all failed attempts and their exact restoration/zero-leftover evidence are retained in the separate advanced live evidence checkpoint.
