# Project Name

> Built with [DevRail](https://devrail.dev) `v1` standards. See [STABILITY.md](STABILITY.md) for component status.

<!-- TODO: Replace with your project name and one-line description -->

A new project bootstrapped from the [DevRail GitHub template](https://github.com/devrail-dev/github-repo-template).

<!-- badges-start -->
<!-- TODO: Add CI status badge: ![Lint](https://github.com/OWNER/REPO/actions/workflows/lint.yml/badge.svg) -->
[![DevRail compliant](https://devrail.dev/images/badge.svg)](https://devrail.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
<!-- badges-end -->

## Quick Start

1. Click **"Use this template"** on [github.com/devrail-dev/github-repo-template](https://github.com/devrail-dev/github-repo-template) to create a new repository.
2. Edit `.devrail.yml` and uncomment the languages used in your project.
3. Run `make install-hooks` to set up pre-commit hooks.

## Usage

The Makefile is the universal execution interface. Every target produces consistent behavior whether invoked by a developer, CI pipeline, or AI agent.

| Target | Purpose |
|---|---|
| `make help` | Show available targets (default) |
| `make lint` | Run all linters for declared languages |
| `make format` | Run all formatters for declared languages |
| `make fix` | Auto-fix formatting issues in-place |
| `make test` | Run project test suite |
| `make security` | Run language-specific security scanners |
| `make scan` | Run universal scanning (trivy, gitleaks) |
| `make docs` | Generate documentation |
| `make check` | Run all of the above; report composite summary |
| `make install-hooks` | Install pre-commit and pre-push hooks |

All targets except `help` and `install-hooks` delegate to the dev-toolchain Docker container (`ghcr.io/devrail-dev/dev-toolchain:v1`).

## Configuration

### `.devrail.yml`

Every DevRail-managed repository includes a `.devrail.yml` file at the repo root. This file declares the project's languages and settings, and is read by the Makefile, CI pipelines, and AI agents.

```yaml
languages:
  - python
  - bash

fail_fast: false
log_format: json
```

Uncomment the languages used in your project and configure settings as needed.

### Branch Protection

To enforce CI checks before merging pull requests:

1. Go to **Settings > Branches > Branch protection rules**
2. Add a rule for the `main` branch
3. Enable **"Require status checks to pass before merging"**
4. Select all five status checks: `lint`, `format`, `security`, `test`, `docs`

### GitHub Template Repository

This repo is configured as a GitHub template. To enable this on your fork:

1. Go to **Settings > General**
2. Check **"Template repository"** under the repository name section
3. Users will then see a **"Use this template"** button on the repo page

## Contributing

See [DEVELOPMENT.md](DEVELOPMENT.md) for development standards, coding conventions, and contribution guidelines.

To add a new language ecosystem to DevRail, see the [Contributing to DevRail](https://github.com/devrail-dev/devrail-standards/blob/main/standards/contributing.md) guide.

This project follows [Conventional Commits](https://www.conventionalcommits.org/). All commits use the `type(scope): description` format.

## Retrofit Existing Project

To add DevRail standards to an existing GitHub repository:

### Step 1: Core Configuration

- [ ] Copy `.devrail.yml` and uncomment your project's languages
- [ ] Copy `.editorconfig`
- [ ] Merge `.gitignore` patterns into your existing .gitignore
- [ ] Copy `Makefile` (or merge targets if you have an existing Makefile)

### Step 2: Pre-Commit Hooks

- [ ] Copy `.pre-commit-config.yaml` and uncomment hooks for your languages
- [ ] Run `make install-hooks`

### Step 3: Agent Instruction Files

- [ ] Copy `DEVELOPMENT.md`, `CLAUDE.md`, `AGENTS.md`, `.cursorrules`
- [ ] Copy `.opencode/agents.yaml`

### Step 4: CI Workflows

- [ ] Copy `.github/workflows/` directory (lint.yml, format.yml, security.yml, test.yml, docs.yml)
- [ ] Configure branch protection: Settings > Branches > Require status checks

### Step 5: Project Documentation

- [ ] Copy `.github/PULL_REQUEST_TEMPLATE.md`
- [ ] Copy `.github/CODEOWNERS` and configure for your team
- [ ] Copy `CHANGELOG.md` if not already present

### Step 6: Verify

- [ ] Run `make check` and fix any issues
- [ ] Create a test commit to verify pre-commit hooks fire
- [ ] Create a test PR to verify CI workflows run

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
