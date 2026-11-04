## Deferred from: code review of 26-3-full-appliance-migration-guide.md (2026-06-02)

- Infeasible/no-API stories are marked backlog and may be reselected (`_bmad-output/implementation-artifacts/sprint-status.yaml:138`). The sprint tracker comments still say cancelled/infeasible, but status normalization made them backlog. Revisit the tracking model for infeasible work separately from this migration-guide story.

## Deferred from: code review of 5-6-gateway-group-resource.md (2026-06-14)

- Blocked interface story marked done while remaining upstream-blocked (`_bmad-output/implementation-artifacts/sprint-status.yaml:84`). Story 5.1 is a completed revalidation record, but the underlying domain remains blocked; this is outside Story 5.6.

## Deferred from: code review of 28-4-implement-system-tunables-resource.md (2026-06-17)

- Post-mutation reconfigure failures now return a typed `MutationReconfigureError`, and `opnsense_system_tunable` create does a best-effort state preservation when the add succeeded but reconfigure failed. Remaining work: decide whether other create paths should also preserve state after `MutationReconfigureError`, or whether the typed shared error is sufficient for non-safety-sensitive resources.

## Deferred from: v0.3.0 release follow-up (2026-06-18)

- GitHub Actions emitted Node.js 20 deprecation warnings for upstream actions during the successful `v0.3.0` release workflow. The repository already uses current major action pins (`actions/checkout@v4`, `actions/setup-go@v5`, `crazy-max/ghaction-import-gpg@v6`, and `goreleaser/goreleaser-action@v6`), so treat this as ecosystem maintenance unless a future workflow starts failing.
