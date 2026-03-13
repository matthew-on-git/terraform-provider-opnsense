# DevRail Development Standards

This document is the single canonical source of truth for all DevRail development standards. Every agent instruction file (CLAUDE.md, AGENTS.md, .cursorrules, .opencode/) points here. Every template repo ships a copy of this document.

Sections are wrapped in HTML comment markers (`<!-- devrail:section-name -->` / `<!-- /devrail:section-name -->`) so that tooling can extract individual sections programmatically. Markers are flat (never nested) and invisible when rendered.

---

<!-- devrail:critical-rules -->

## Critical Rules

These eight rules are non-negotiable. Every developer and every AI agent must follow them without exception.

1. **Run `make check` before completing any story or task.** Never mark work done without passing checks. This is the single gate for all linting, formatting, security, and test validation.

2. **Use conventional commits.** Every commit message follows the `type(scope): description` format. No exceptions. See the [Conventional Commits](#conventional-commits) section for types and scopes.

3. **Never install tools outside the container.** All linters, formatters, scanners, and test runners live inside `ghcr.io/devrail-dev/dev-toolchain:v1`. The Makefile delegates to Docker. Do not install tools on the host.

4. **Respect `.editorconfig`.** Never override formatting rules (indent style, line endings, trailing whitespace) without explicit instruction. The `.editorconfig` file in each repo is authoritative.

5. **Write idempotent scripts.** Every script must be safe to re-run. Check before acting: `command -v tool || install_tool`, `mkdir -p`, guard file writes with existence checks.

6. **Use the shared logging library.** No raw `echo` for status messages. Use `log_info`, `log_warn`, `log_error`, `log_debug`, and `die` from `lib/log.sh`.

7. **Never suppress failing checks.** When a lint, format, security, or test check fails, fix the underlying issue. Never comment out code, add suppression annotations (`# noqa`, `# nosec`, `#tfsec:ignore`, `// nolint`), disable rules, or mark CI jobs as allowed-to-fail to bypass a failing check. If a finding is a confirmed false positive, document the justification inline alongside the tool's designated suppression mechanism.

8. **Update documentation when changing behavior.** When a change affects public interfaces, configuration, CLI usage, or setup steps, update the relevant documentation (README, DEVELOPMENT.md, inline docs) in the same commit or PR. Do not leave documentation out of sync with code.

<!-- /devrail:critical-rules -->

<!-- devrail:makefile-contract -->

## Makefile Contract

The Makefile is the universal execution interface. Developers, CI pipelines, and AI agents all interact with the project through `make` targets. Behavior is identical regardless of invocation context.

### Two-Layer Delegation Pattern

- **Layer 1 (public targets):** Run on the host. Delegate to the dev-toolchain Docker container.
- **Layer 2 (internal targets):** Run inside the container. Execute actual tool commands.

Public targets use `lower-kebab-case`. Internal targets use a `_` prefix (e.g., `_lint`, `_format`).

### Target Contract

Every DevRail-managed repo exposes these public targets:

| Target | Purpose |
|---|---|
| `make help` | List all targets with descriptions (default target) |
| `make lint` | Run all linters for declared languages |
| `make format` | Run all formatters for declared languages |
| `make fix` | Auto-fix formatting issues in-place |
| `make test` | Run all test suites for declared languages |
| `make security` | Run security scanners (bandit, tfsec, checkov, etc.) |
| `make scan` | Run universal scanners (trivy, gitleaks) |
| `make docs` | Generate documentation (terraform-docs, tool version report) |
| `make changelog` | Generate CHANGELOG.md from conventional commits (git-cliff) |
| `make check` | Run all of the above in sequence |
| `make install-hooks` | Install pre-commit and pre-push hooks |
| `make init` | Scaffold config files for declared languages |

### Naming Rules

- **No abbreviations** -- `security` not `sec`, `format` not `fmt`
- Every public target has a `## description` comment for `make help` auto-generation
- Variables use `UPPER_SNAKE_CASE` with `?=` for overridability (e.g., `DEVRAIL_IMAGE ?= ghcr.io/devrail-dev/dev-toolchain:v1`)

### File Structure

Makefiles follow this order:

1. Variables
2. `.PHONY` declarations
3. Public targets (with `## description` comments)
4. Internal `_`-prefixed targets

### Error Handling

- **Default:** Run-all-report-all. Every check runs to completion and all issues are reported.
- **Fail-fast:** Available via `DEVRAIL_FAIL_FAST=1` env var or `fail_fast: true` in `.devrail.yml`. Stops at first failure.

<!-- /devrail:makefile-contract -->

<!-- devrail:shell-conventions -->

## Shell Script Conventions

### Mandatory Header

Every shell script begins with:

```bash
#!/usr/bin/env bash
set -euo pipefail
```

No exceptions.

### Idempotency

Scripts must be safe to re-run. Use these patterns:

- `command -v tool || install_tool` -- check before installing
- `mkdir -p` -- never fail on existing directory
- Guard file writes with existence checks
- No interactive prompts -- scripts run in containers and CI

### Naming Conventions

| Element | Convention | Example |
|---|---|---|
| Environment variables / constants | `UPPER_SNAKE_CASE` with `readonly` | `readonly DEVRAIL_VERSION="1.0.0"` |
| Local variables | `lower_snake_case` | `local output_dir` |
| Functions | `lower_snake_case`, prefixed by purpose | `install_python`, `check_deps`, `log_info` |

### Argument Parsing

- Use `getopts` for argument parsing
- Every script supports `--help`
- `--help` auto-extracts usage from the structured header comment

### Self-Documenting Scripts

Every script has a structured header comment:

```bash
#!/usr/bin/env bash
# Purpose: One-line description of what this script does
# Usage: script-name [options] <args>
# Dependencies: docker, make
set -euo pipefail
```

### Shared Library (`lib/`)

Scripts source shared libraries from `lib/`:

- **`lib/log.sh`** -- logging functions and verbosity control
- **`lib/platform.sh`** -- platform detection helpers (`on_mac`, `on_linux`, `on_arm64`)

### Validation & Helpers

- `is_empty`, `is_not_empty`, `is_set` -- consistent variable checking
- `require_cmd "docker" "Install Docker to continue"` -- dependency guards at script start

### Cleanup & Safety

- Register trap handlers at script start: `trap cleanup EXIT`
- Create temp files with `mktemp` only, cleaned up by the trap
- Never use interactive prompts

### Linting

All shell scripts must pass `shellcheck`. This is enforced by the `make lint` target.

### Python CLIs

When a CLI tool requires more than basic argument parsing, use Python with Click instead of complex shell scripts.

<!-- /devrail:shell-conventions -->

<!-- devrail:logging -->

## Output & Logging

### Log Functions

All scripts use these shared log functions from `lib/log.sh`:

| Function | Purpose |
|---|---|
| `log_info "message"` | Informational status messages |
| `log_warn "message"` | Warning conditions |
| `log_error "message"` | Error conditions |
| `log_debug "message"` | Debug-level detail (only shown when `DEVRAIL_DEBUG=1`) |
| `die "message"` | Log error and `exit 1` in one call |

### Output Format

**JSON (default):**

```json
{"level":"info","msg":"Running lint checks","script":"lint.sh","ts":"2026-02-19T10:00:00Z"}
```

**Human-readable** (via `DEVRAIL_LOG_FORMAT=human`):

```
[INFO]  Running lint checks
```

Error entries include `exit_code` when applicable.

### Verbosity Levels

| Level | Env Var | Behavior |
|---|---|---|
| Quiet | `DEVRAIL_QUIET=1` | Suppress all output except errors |
| Normal | (default) | Standard informational output |
| Debug | `DEVRAIL_DEBUG=1` | Include debug-level messages |

### Stream Discipline

- All log output goes to **stderr**
- **stdout** is reserved for tool output (data, JSON results, etc.)
- No raw `echo` for status messages -- use the log functions
- No inline ANSI colors -- the log library handles formatting

### Makefile Target Output

Each target produces a JSON summary:

```json
{"target":"lint","status":"pass","duration_ms":1234,"errors":[]}
```

`make check` produces a final summary of all targets with pass/fail status and total duration.

In human mode, output is a simple table with status indicators.

### Exit Codes

| Code | Meaning |
|---|---|
| `0` | Pass |
| `1` | Failure (lint errors, test failures, security findings) |
| `2` | Misconfiguration (missing `.devrail.yml`, unknown language, container pull failure) |

### CI Output

- CI job names match Makefile target names: `lint`, `format`, `security`, `test`, `docs`
- Each job writes JSON output to an artifact file
- Exit codes are propagated -- no swallowed failures

### Pre-commit Output

- Human format by default (respects framework conventions)
- CI can override to JSON when needed

<!-- /devrail:logging -->

<!-- devrail:commits -->

## Conventional Commits

All commits in DevRail-managed repositories follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format

```
type(scope): description
```

### Types

| Type | When to Use |
|---|---|
| `feat` | A new feature or capability |
| `fix` | A bug fix |
| `docs` | Documentation-only changes |
| `chore` | Maintenance tasks (dependencies, config) |
| `ci` | CI/CD pipeline changes |
| `refactor` | Code restructuring without behavior change |
| `test` | Adding or updating tests |

### Scopes

| Scope | Applies To |
|---|---|
| `python` | Python tooling, configs, or standards |
| `bash` | Bash tooling, configs, or standards |
| `terraform` | Terraform tooling, configs, or standards |
| `ansible` | Ansible tooling, configs, or standards |
| `ruby` | Ruby tooling, configs, or standards |
| `go` | Go tooling, configs, or standards |
| `javascript` | JavaScript/TypeScript tooling, configs, or standards |
| `container` | Dev-toolchain container image |
| `ci` | CI/CD pipeline configuration |
| `makefile` | Makefile targets and patterns |
| `changelog` | Changelog generation and git-cliff config |
| `standards` | DevRail standards documentation |

### Code Comments

- Comments explain *why*, not *what* -- do not over-comment obvious code
- No commented-out code -- delete it; git has history
- TODO format: `# TODO(devrail#123): description` -- always linked to an issue

### Changelog

- Auto-generated from conventional commits
- `CHANGELOG.md` per repo, following [Keep a Changelog](https://keepachangelog.com/) format

<!-- /devrail:commits -->

<!-- devrail:coding-practices -->

## Coding Practices

General software engineering standards that apply across all languages. For the full reference, see [`standards/coding-practices.md`](standards/coding-practices.md).

- **DRY, KISS, YAGNI** -- don't repeat yourself, keep it simple, build only what is needed now
- **Single responsibility** -- each function, class, or module does one thing
- **Fail fast** -- validate inputs at boundaries, return or raise immediately on invalid state
- **No swallowed exceptions** -- every error branch handles the error meaningfully or propagates it
- **Test behavior, not implementation** -- assert on outputs and side effects, follow the test pyramid (unit > integration > e2e)
- **New code must include tests** -- PRs that add logic without tests are incomplete
- **~50 line function guideline** -- split long functions into focused helpers
- **Pin dependency versions** -- commit lock files, update regularly, respond to security advisories promptly

<!-- /devrail:coding-practices -->

<!-- devrail:git-workflow -->

## Git Workflow

Git discipline and collaboration standards. For the full reference, see [`standards/git-workflow.md`](standards/git-workflow.md).

- **Never push directly to `main`** -- all changes reach the default branch through a pull/merge request
- **Branch naming** -- `type/short-description` (e.g., `feat/add-auth`, `fix/login-error`)
- **Minimum 1 approval required** before merging, no self-merge
- **Atomic commits** -- one logical change per commit, conventional commit format
- **No `--force-push` to shared branches** -- only force push your own feature branches
- **Squash-merge feature branches** for clean, linear history on `main`
- **No secrets in commits** -- enforced by gitleaks pre-commit hook and `make scan`
- **Branch protection on `main`** -- require PR, approvals, and CI pass

<!-- /devrail:git-workflow -->

<!-- devrail:release-versioning -->

## Release & Versioning

Release management and versioning discipline. For the full reference, see [`standards/release-versioning.md`](standards/release-versioning.md).

- **Semantic versioning** -- `MAJOR.MINOR.PATCH` with strict adherence after `v1.0.0`
- **Annotated tags only** -- `vX.Y.Z` format, tagged from `main`, never moved or deleted after push
- **Release process** -- review changelog, tag, push, create platform release with artifacts
- **Hotfixes** -- branch from tag, fix, merge to `main`, tag new patch release
- **Pre-release versions** -- `v1.0.0-rc.1`, `v1.0.0-beta.1` conventions for unstable releases
- **Libraries vs services** -- libraries follow semver strictly; services may use date-based versioning
- **Changelog** -- auto-generated from conventional commits, [Keep a Changelog](https://keepachangelog.com/) format

<!-- /devrail:release-versioning -->

<!-- devrail:ci-cd-pipelines -->

## CI/CD Pipelines

Continuous integration and deployment standards. For the full reference, see [`standards/ci-cd-pipelines.md`](standards/ci-cd-pipelines.md).

- **Standard stages** -- `lint → format → test → security → scan → build → deploy` (in order)
- **Stage contract** -- each CI stage calls a `make` target; identical behavior locally and in CI
- **Required jobs** -- lint, format, test, security, scan must pass before merge
- **Deployment gates** -- auto-deploy to staging on merge to `main`; manual approval for production
- **Pipeline types** -- library (test+publish), service (test+build+deploy), infrastructure (plan+apply)
- **Artifact management** -- release tags are immutable, pin toolchain versions, commit lock files
- **Performance** -- cache dependencies, parallelize independent stages, target < 10 min for PR checks

<!-- /devrail:ci-cd-pipelines -->

<!-- devrail:container-standards -->

## Container Standards

Container image build and runtime standards. For the full reference, see [`standards/container-standards.md`](standards/container-standards.md).

- **Pin base images** -- use specific tags or digests, never `latest`
- **Multi-stage builds** -- separate build dependencies from the runtime image
- **Layer ordering** -- least-changing layers first to maximize cache reuse
- **Non-root user** -- never run containers as root in production
- **No secrets in images** -- inject at runtime via env vars or mounted volumes
- **Image tagging** -- `vX.Y.Z` for releases, `sha-<short>` for CI builds, never overwrite release tags
- **Health checks** -- every service container exposes `/healthz` and `/readyz` endpoints
- **`.dockerignore` required** -- exclude `.git`, tests, docs, and build artifacts from the context

<!-- /devrail:container-standards -->

<!-- devrail:secrets-management -->

## Secrets Management

Standards for handling secrets and sensitive configuration. For the full reference, see [`standards/secrets-management.md`](standards/secrets-management.md).

- **Classify correctly** -- secrets vs sensitive config vs environment config vs application config
- **Never in source control** -- no API keys, passwords, or tokens in committed files (enforced by gitleaks)
- **Platform secrets** -- use GitHub/GitLab secrets or a dedicated manager (Vault, AWS SM, GCP SM)
- **`.env` gitignored, `.env.example` committed** -- document required variables with placeholders
- **`UPPER_SNAKE_CASE` naming** -- prefix by service or context to avoid collisions
- **Rotate on schedule** -- 90-day minimum for keys and credentials; immediately on exposure
- **Least privilege** -- no shared credentials, service accounts over personal, audit access

<!-- /devrail:secrets-management -->

<!-- devrail:api-cli-design -->

## API & CLI Design

Standards for designing APIs and CLIs. For the full reference, see [`standards/api-cli-design.md`](standards/api-cli-design.md).

- **Version APIs from day one** -- URL path (`/v1/`) preferred; never break clients without a version bump
- **JSON by default** -- consistent envelope, ISO 8601 timestamps, request IDs in every response
- **Structured errors** -- machine-readable `code`, human-readable `message`, detailed `fields`; correct HTTP status codes
- **CLI conventions** -- `--help` on every command, exit codes 0/1/2, JSON output for machines
- **Backward compatibility** -- additive changes are safe; removals require deprecation + version bump
- **OpenAPI for APIs** -- spec is the source of truth, kept in sync with code
- **Pagination and rate limiting** -- standard patterns for collection endpoints

<!-- /devrail:api-cli-design -->

<!-- devrail:monitoring-observability -->

## Monitoring & Observability

Runtime monitoring and observability standards. For the full reference, see [`standards/monitoring-observability.md`](standards/monitoring-observability.md).

- **Health endpoints** -- `/healthz` (liveness) and `/readyz` (readiness) for every service
- **Structured logging** -- JSON format, correlation IDs, log levels (`debug`, `info`, `warn`, `error`)
- **RED metrics** -- Rate, Errors, Duration for every service; Prometheus-style exposition
- **Alerting** -- alert on symptoms not causes, every alert links to a runbook, minimize noise
- **Dashboards** -- one per service minimum, golden signals visible at a glance
- **Never log PII** -- no secrets, tokens, emails, or government IDs in logs; redact if unavoidable

<!-- /devrail:monitoring-observability -->

<!-- devrail:incident-response -->

## Incident Response

Incident detection, response, and learning standards. For the full reference, see [`standards/incident-response.md`](standards/incident-response.md).

- **Severity levels** -- SEV1 (15 min response) through SEV4 (1 business day)
- **Incident workflow** -- detect → triage → mitigate → resolve → post-mortem
- **Communication** -- status page updates, stakeholder notification cadence per severity
- **Post-mortems** -- required for SEV1-SEV2, blameless, concrete action items with owners and due dates
- **Runbooks** -- required for every production service, stored alongside code, reviewed quarterly
- **On-call** -- defined rotation, clean handoffs, escalation path documented

<!-- /devrail:incident-response -->

<!-- devrail:data-handling -->

## Data Handling

Data classification, privacy, and compliance standards. For the full reference, see [`standards/data-handling.md`](standards/data-handling.md).

- **Data classification** -- public, internal, confidential, restricted; classify at collection time
- **PII handling** -- identify, minimize collection, encrypt at rest and in transit, document what is collected
- **Retention** -- define periods per data type, automate deletion, support right-to-deletion requests
- **Backups** -- regular, tested restores, encrypted, offsite copy, automated
- **Encryption** -- TLS 1.2+ in transit, AES-256 at rest, keys managed via secrets manager
- **Compliance awareness** -- GDPR, CCPA, HIPAA, PCI DSS as applicable; breach notification process documented
- **Never log PII** -- redact or mask if logging is unavoidable; route to restricted log stream

<!-- /devrail:data-handling -->

## Configuration

### `.devrail.yml`

Every DevRail-managed repo has a `.devrail.yml` at its root. This file declares languages, settings, and project metadata. It is read by the Makefile, CI pipelines, and AI agents.

For the complete schema specification, see [`standards/devrail-yml-schema.md`](standards/devrail-yml-schema.md).

### `.editorconfig`

Every repo includes an `.editorconfig` file that defines formatting rules (indent style, indent size, line endings, trailing whitespace). All editors and agents must respect these settings.

## Language Standards

<!-- devrail:python -->

### Python

| Concern | Tool | Notes |
|---|---|---|
| Linter | **ruff** | Fast, replaces flake8/isort/pyupgrade |
| Formatter | **ruff format** | Consistent with linter config |
| Security | **bandit**, **semgrep** | Static analysis for security issues |
| Tests | **pytest** | Standard test runner |
| Type Check | **mypy** | Static type checking |

**Key rules:**

- `ruff` handles both linting and formatting -- no separate isort or black
- `bandit` and `semgrep` run as part of `make security`
- `pytest` runs as part of `make test`
- `mypy` runs as part of `make lint`
- All tools are pre-installed in the dev-toolchain container

<!-- /devrail:python -->

<!-- devrail:bash -->

### Bash

| Concern | Tool | Notes |
|---|---|---|
| Linter | **shellcheck** | Static analysis for shell scripts |
| Formatter | **shfmt** | Consistent formatting |
| Tests | **bats** | Bash Automated Testing System |

**Key rules:**

- All scripts must pass `shellcheck` with zero warnings
- `shfmt` enforces consistent indentation and style
- `bats` test files live in the `tests/` directory
- See [Shell Script Conventions](#shell-script-conventions) for coding standards

<!-- /devrail:bash -->

<!-- devrail:terraform -->

### Terraform

| Concern | Tool | Notes |
|---|---|---|
| Linter | **tflint** | Terraform-specific linting rules |
| Formatter | **terraform fmt** | Canonical formatting |
| Security | **tfsec**, **checkov** | Infrastructure security scanning |
| Tests | **terratest** | Go-based infrastructure testing |
| Docs | **terraform-docs** | Auto-generate module documentation |

**Key rules:**

- `terraform fmt` is enforced -- no manual formatting
- Both `tfsec` and `checkov` run to catch complementary issues
- `terraform-docs` auto-generates `README.md` content for modules
- `terratest` tests live in the `tests/` directory

<!-- /devrail:terraform -->

<!-- devrail:ansible -->

### Ansible

| Concern | Tool | Notes |
|---|---|---|
| Linter | **ansible-lint** | Playbook and role linting |
| Tests | **molecule** | Role testing framework |

**Key rules:**

- `ansible-lint` enforces best practices for playbooks and roles
- `molecule` scenarios live alongside roles
- No formatter is enforced -- YAML formatting is handled by `.editorconfig`

<!-- /devrail:ansible -->

<!-- devrail:universal -->

### Universal Security Tools

These tools run for every project regardless of language.

| Tool | Purpose |
|---|---|
| **trivy** | Container image and filesystem vulnerability scanning |
| **gitleaks** | Secret detection in git history and staged changes |

**Key rules:**

- `trivy` runs as part of `make scan`
- `gitleaks` runs as part of `make scan` and as a pre-commit hook
- Both tools produce JSON output in CI
- Findings at any severity level cause a non-zero exit code

<!-- /devrail:universal -->

## Agent Enforcement

All AI agents (Claude, Cursor, OpenCode, Windsurf, and others) operating on DevRail-managed repos must follow these guidelines:

1. **Read this document first.** Agent shim files (CLAUDE.md, AGENTS.md, .cursorrules, .opencode/) point here. This is the canonical source.

2. **Follow the Critical Rules.** The eight rules in the [Critical Rules](#critical-rules) section are inlined in every agent shim file. There is no excuse for missing them.

3. **Run `make check` and fix all issues before declaring work complete.** Do not rely on CI to catch problems.

4. **Use conventional commits.** Every commit follows the format defined in [Conventional Commits](#conventional-commits).

5. **Do not modify tool configurations without explicit instruction.** Ruff configs, shellcheck directives, tflint rules -- these are set by the standards. Do not change them to make code pass.

6. **Consult per-language sections for tool-specific guidance.** Each language section above defines which tools run and how.

Agent shim files use a hybrid strategy: critical rules are inlined directly in the shim, and the shim contains a pointer to this DEVELOPMENT.md for full standards. This ensures critical behaviors are present regardless of whether the agent follows cross-references.
