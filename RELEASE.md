# Release Preflight

Use this checklist before tagging a provider release.

## Required Checks

1. Run `make check` and confirm lint, format, test, security, scan, and docs all pass.
2. Confirm `CHANGELOG.md` lists all new resources and data sources for the release.
3. Confirm `terraform-registry-manifest.json` declares protocol version `6.0`.
4. Confirm `.goreleaser.yml` includes the Registry manifest in checksum and release artifacts.
5. Confirm GitHub secrets `GPG_PRIVATE_KEY` and `PASSPHRASE` are configured.
6. Confirm the GPG public key is registered with the Terraform Registry publisher account.
7. Run a GoReleaser snapshot/dry-run before the first public tag.

## Release Trigger

Push a semantic version tag matching `v*`, for example `v0.1.0`. The release workflow runs GoReleaser, builds provider ZIPs, creates SHA256 checksums, signs the checksum file, and publishes a GitHub Release for Terraform Registry discovery.

## Historical v0.1.0 Positioning

The first release advertised this historical v0.1.0 baseline:

- 90 supported resources.
- 34 supported data sources.
- Data-source parity as Coming.
- Interface assignment/IP config/PPPoE, gateway groups, and system general settings as Upstream-blocked; tunables/sysctl is supported through `opnsense_system_tunable`.

For every subsequent release, refresh the current release positioning from
`docs/`, `internal/service/*/exports.go`, and
`_bmad-output/planning-artifacts/support-matrix.md` before tagging.
