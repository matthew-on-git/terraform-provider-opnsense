# Agent Instructions

This project follows [DevRail](https://devrail.dev) development standards.
See DEVELOPMENT.md for the complete reference.

## Critical Rules

1. **Run `make check` before completing any story or task.** Never mark work done without passing checks. This is the single gate for all linting, formatting, security, and test validation.
2. **Use conventional commits.** Every commit message follows the `type(scope): description` format. No exceptions.
3. **Never install tools outside the container.** All linters, formatters, scanners, and test runners live inside `ghcr.io/devrail-dev/dev-toolchain:v1`. The Makefile delegates to Docker. Do not install tools on the host.
4. **Respect `.editorconfig`.** Never override formatting rules (indent style, line endings, trailing whitespace) without explicit instruction.
5. **Write idempotent scripts.** Every script must be safe to re-run. Check before acting: `command -v tool || install_tool`, `mkdir -p`, guard file writes with existence checks.
6. **Use the shared logging library.** No raw `echo` for status messages. Use `log_info`, `log_warn`, `log_error`, `log_debug`, and `die` from `lib/log.sh`.
7. **Never suppress failing checks.** When a lint, format, security, or test check fails, fix the underlying issue. Never comment out code, add suppression annotations, disable rules, or mark CI jobs as allowed-to-fail to bypass a failing check.
8. **Update documentation when changing behavior.** When a change affects public interfaces, configuration, CLI usage, or setup steps, update the relevant documentation (README, DEVELOPMENT.md, inline docs) in the same commit or PR. Do not leave documentation out of sync with code.

## Quick Reference

- Run `make check` to validate all standards
- Run `make help` to see available targets
- All tools run inside the dev-toolchain container

## Release Readiness

- Use conventional commit messages (including the final squash/merge message). `fix` bumps patch, `feat` bumps minor; `!` / `BREAKING CHANGE:` bumps minor before 1.0 and major thereafter. Include user-visible behavior and migration requirements in the PR/commit body.
- Contributors do **not** select versions, edit the release manifest, or add versioned Unreleased headings. Release Please accumulates merged changes into one maintainer-reviewed release PR, owning `CHANGELOG.md` and `.release-please-manifest.json`.
- Review the generated release PR for full scope and migration notes before merging. After its merge passes main CI, automation creates a tag and draft, then explicitly dispatches signed GoReleaser publication. Ordinary feature/fix merges only update the release PR.
- Run `make check` and `npm ci --prefix test/release && node --test test/release/release.test.cjs` inside the dev-toolchain container before changing release automation.
- Distinguish main CI, tag/draft creation, signed asset publication and Registry availability. Never race automation with manual tags; recover at the existing verified tag.
- See [Provider Release Automation](DEVELOPMENT.md#provider-release-automation) for bootstrap, token permissions and recovery.

## Public Project / Private Consumer Boundary

This is a public, reusable provider project. Keep operator-specific inventory,
site names, internal/public service hostnames tied to an operator, real network
allocations, credentials, local workspace paths, deployment/adoption plans and
operational notes out of this repository, its issues/PR bodies and CI artifacts.
Store those in the operator's private documentation or infrastructure repository.
Public product roadmaps and acceptance specifications must remain consumer-neutral.
Use RFC 5737/RFC 3849 documentation addresses and reserved example domains in
fixtures and examples; do not copy an operator's topology into a public test.
Record only sanitized test versions, behavior and results publicly. Before
publication, inspect the full diff (including generated files and planning
artifacts) for identifying information; a secret scanner alone does not enforce
this boundary. Removing HEAD content does not remove Git history: document any
historical exposure privately and do not rewrite shared history without a
maintainer-approved remediation plan.
