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

- For releasable provider changes, update the top `CHANGELOG.md` section in the same PR with an explicit `## X.Y.Z (Unreleased)` heading. Append to an existing unreleased version; otherwise compare published tags and choose the next semantic version. Include migration requirements and the full unreleased scope, not only the latest PR.
- This heading is an executable release input, not optional housekeeping: successful push CI on `main` triggers `auto-release-tag.yml`, which creates the missing `vX.Y.Z` tag and dispatches `release.yml`. A missing heading skips release; an already-existing tag also skips it. A PR containing a new release heading will publish after merge and successful CI.
- Before opening a PR, inspect the release version/tag decision and explain any deliberate changelog exemption in the PR. Docs/planning-only changes may be exempt; do not silently omit provider behavior changes.
- After merge, distinguish CI success, auto-tag/dispatch success, GitHub release/assets and Terraform Registry availability. Verify each before claiming publication. Never add a manual tag while the automatic path is pending; follow the documented recovery procedure if it fails.
- See [Provider Release Automation](DEVELOPMENT.md#provider-release-automation) for the exact trigger and verification checklist.
