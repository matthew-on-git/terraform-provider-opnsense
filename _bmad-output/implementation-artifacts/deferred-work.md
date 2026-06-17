## Deferred from: code review of 26-3-full-appliance-migration-guide.md (2026-06-02)

- Infeasible/no-API stories are marked backlog and may be reselected (`_bmad-output/implementation-artifacts/sprint-status.yaml:138`). The sprint tracker comments still say cancelled/infeasible, but status normalization made them backlog. Revisit the tracking model for infeasible work separately from this migration-guide story.
- README dev-toolchain tag differs from Makefile pin (`README.md:53`). README says `ghcr.io/devrail-dev/dev-toolchain:v1`, while `Makefile` currently pins `ghcr.io/devrail-dev/dev-toolchain:1.12.0`. Resolve documentation/toolchain-version policy outside this story.

## Deferred from: code review of 5-1-interface-resource.md (2026-06-14)

- ACME issuance marked done despite contradictory sprint-status note (`_bmad-output/implementation-artifacts/sprint-status.yaml:280`). The Epic 29 status line says `29-4-acme-certificate-issuance-and-refid: done` while its comment describes `certificate_resource.go` as plain CRUD with no `/sign`, poll, or refid. This appears unrelated to Story 5.1 and should be reviewed with the Epic 29 changes.

## Deferred from: code review of 5-6-gateway-group-resource.md (2026-06-14)

- Stale generated header timestamp in sprint status (`_bmad-output/implementation-artifacts/sprint-status.yaml:2`). The top comment still says 2026-06-09 while the actual YAML field is current; this is pre-existing tracker metadata drift outside Story 5.6.
- Generated provider index import guidance is stale versus migration guide (`templates/index.md.tmpl:113`). The HAProxy import ordering guidance predates Epic 29 mapfile/action and ACME-before-frontend guidance; this is unrelated to gateway-group revalidation.
- Blocked interface story marked done while remaining upstream-blocked (`_bmad-output/implementation-artifacts/sprint-status.yaml:84`). Story 5.1 is a completed revalidation record, but the underlying domain remains blocked; this is outside Story 5.6.

## Deferred from: code review of 28-4-implement-system-tunables-resource.md (2026-06-17)

- Mutation succeeds but failed `reconfigure` can orphan or desync Terraform state (`pkg/opnsense/crud.go:58`). Shared CRUD persists add/update/delete first and then calls reconfigure, so a post-mutation reconfigure failure can leave remote state changed while Terraform returns an error. This is pre-existing shared behavior across item resources and should be addressed as an API-client lifecycle design issue rather than only in `opnsense_system_tunable`.
