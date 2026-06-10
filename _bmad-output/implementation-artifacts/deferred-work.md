## Deferred from: code review of 26-3-full-appliance-migration-guide.md (2026-06-02)

- Infeasible/no-API stories are marked backlog and may be reselected (`_bmad-output/implementation-artifacts/sprint-status.yaml:138`). The sprint tracker comments still say cancelled/infeasible, but status normalization made them backlog. Revisit the tracking model for infeasible work separately from this migration-guide story.
- README dev-toolchain tag differs from Makefile pin (`README.md:53`). README says `ghcr.io/devrail-dev/dev-toolchain:v1`, while `Makefile` currently pins `ghcr.io/devrail-dev/dev-toolchain:1.12.0`. Resolve documentation/toolchain-version policy outside this story.
