# DevRail Reference Makefile — Two-Layer Delegation Pattern
#
# This Makefile implements the DevRail contract. Public targets run on the
# host and delegate to the dev-toolchain Docker container. Internal targets
# (prefixed with _) run inside the container where all tools are installed.
#
# Usage:
#   make              Show available targets (help)
#   make check        Run all checks (lint, format, test, security, scan, docs)
#   make lint         Run all linters
#   DEVRAIL_FAIL_FAST=1 make check   Stop on first failure
#
# Configuration is read from .devrail.yml at project root.

# ---------------------------------------------------------------------------
# Variables (overridable via environment)
# ---------------------------------------------------------------------------
DEVRAIL_IMAGE     ?= ghcr.io/devrail-dev/dev-toolchain:v1
DEVRAIL_FAIL_FAST ?= 0
DEVRAIL_LOG_FORMAT ?= json

DOCKER_RUN := docker run --rm \
	-v "$$(pwd):/workspace" \
	-w /workspace \
	-e DEVRAIL_FAIL_FAST=$(DEVRAIL_FAIL_FAST) \
	-e DEVRAIL_LOG_FORMAT=$(DEVRAIL_LOG_FORMAT) \
	$(DEVRAIL_IMAGE)

.DEFAULT_GOAL := help

# ---------------------------------------------------------------------------
# .devrail.yml language detection (runs inside container where yq is available)
# ---------------------------------------------------------------------------
DEVRAIL_CONFIG := .devrail.yml
LANGUAGES      := $(shell yq '.languages[]' $(DEVRAIL_CONFIG) 2>/dev/null)
HAS_PYTHON     := $(filter python,$(LANGUAGES))
HAS_BASH       := $(filter bash,$(LANGUAGES))
HAS_TERRAFORM  := $(filter terraform,$(LANGUAGES))
HAS_ANSIBLE    := $(filter ansible,$(LANGUAGES))
HAS_RUBY       := $(filter ruby,$(LANGUAGES))
HAS_GO         := $(filter go,$(LANGUAGES))
HAS_JAVASCRIPT := $(filter javascript,$(LANGUAGES))
HAS_RUST       := $(filter rust,$(LANGUAGES))

# ---------------------------------------------------------------------------
# .PHONY declarations
# ---------------------------------------------------------------------------
.PHONY: help lint format fix test security scan docs changelog check install-hooks init
.PHONY: _lint _format _fix _test _security _scan _docs _changelog _check _check-config _init

# ===========================================================================
# Public targets (run on host, delegate to Docker container)
# ===========================================================================

help: ## Show this help
	@echo "DevRail — developer infrastructure platform"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

changelog: ## Generate CHANGELOG.md from conventional commits
	$(DOCKER_RUN) make _changelog

check: ## Run all checks (lint, format, test, security, scan, docs)
	$(DOCKER_RUN) make _check

docs: ## Generate documentation
	$(DOCKER_RUN) make _docs

fix: ## Auto-fix formatting issues in-place
	$(DOCKER_RUN) make _fix

format: ## Run all formatters
	$(DOCKER_RUN) make _format

init: ## Scaffold config files for declared languages
	$(DOCKER_RUN) make _init

install-hooks: ## Install pre-commit hooks
	@if ! command -v python3 >/dev/null 2>&1; then \
		echo "Error: Python 3 is required to install pre-commit. Install Python 3 and try again."; \
		exit 2; \
	fi
	@if ! git rev-parse --git-dir >/dev/null 2>&1; then \
		echo "Error: Not in a git repository. Run 'git init' first."; \
		exit 2; \
	fi
	@if ! command -v pre-commit >/dev/null 2>&1; then \
		echo "Installing pre-commit..."; \
		if command -v pipx >/dev/null 2>&1; then \
			pipx install pre-commit; \
		else \
			pip install --user pre-commit; \
		fi; \
	fi
	@pre-commit install
	@pre-commit install --hook-type commit-msg
	@pre-commit install --hook-type pre-push
	@echo "Pre-commit hooks installed successfully. Hooks will run on commit and push."

lint: ## Run all linters
	$(DOCKER_RUN) make _lint

scan: ## Run universal scanners (trivy, gitleaks)
	$(DOCKER_RUN) make _scan

security: ## Run language-specific security scanners
	$(DOCKER_RUN) make _security

test: ## Run all tests
	$(DOCKER_RUN) make _test

# ===========================================================================
# Internal targets (run inside container — do NOT invoke directly)
#
# These targets are invoked by the public targets above via Docker.
# They read .devrail.yml to determine which language-specific tools to run.
# All internal targets follow the run-all-report-all pattern by default,
# switching to fail-fast when DEVRAIL_FAIL_FAST=1 is set.
#
# Exit codes:
#   0 — pass (all tools succeeded or skipped)
#   1 — failure (one or more tools reported issues)
#   2 — misconfiguration (missing .devrail.yml, missing tools, etc.)
#
# Each internal target emits a JSON summary line to stdout:
#   {"target":"<name>","status":"pass|fail|skip","duration_ms":<N>}
# ===========================================================================

_check-config:
	@if [ ! -f "$(DEVRAIL_CONFIG)" ]; then \
		echo '{"target":"config","status":"error","error":"missing .devrail.yml","exit_code":2}'; \
		exit 2; \
	fi

# --- _lint: language-specific linting ---
_lint: _check-config
	@start_time=$$(date +%s%3N); \
	overall_exit=0; \
	ran_languages=""; \
	failed_languages=""; \
	if [ -n "$(HAS_PYTHON)" ]; then \
		ran_languages="$${ran_languages}\"python\","; \
		ruff check . || { overall_exit=1; failed_languages="$${failed_languages}\"python\","; }; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"lint\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_BASH)" ]; then \
		ran_languages="$${ran_languages}\"bash\","; \
		sh_files=$$(find . -name '*.sh' -not -path './.git/*' -not -path './vendor/*' -not -path './node_modules/*' 2>/dev/null); \
		if [ -n "$$sh_files" ]; then \
			echo "$$sh_files" | xargs shellcheck || { overall_exit=1; failed_languages="$${failed_languages}\"bash\","; }; \
		else \
			echo '{"level":"info","msg":"skipping bash lint: no .sh files found","language":"bash"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"lint\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_TERRAFORM)" ]; then \
		ran_languages="$${ran_languages}\"terraform\","; \
		tf_dirs=$$(find . -name '*.tf' -not -path './.git/*' -not -path './.terraform/*' 2>/dev/null | xargs -I{} dirname {} | sort -u); \
		if [ -n "$$tf_dirs" ]; then \
			for dir in $$tf_dirs; do \
				(cd "$$dir" && tflint) || { overall_exit=1; failed_languages="$${failed_languages}\"terraform\","; }; \
			done; \
		else \
			echo '{"level":"info","msg":"skipping terraform lint: no .tf files found","language":"terraform"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"lint\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_ANSIBLE)" ]; then \
		ran_languages="$${ran_languages}\"ansible\","; \
		ansible-lint || { overall_exit=1; failed_languages="$${failed_languages}\"ansible\","; }; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"lint\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_RUBY)" ]; then \
		ran_languages="$${ran_languages}\"ruby\","; \
		rb_files=$$(find . -name '*.rb' -not -path './.git/*' -not -path './vendor/*' -not -path './node_modules/*' 2>/dev/null); \
		if [ -n "$$rb_files" ]; then \
			rubocop . || { overall_exit=1; failed_languages="$${failed_languages}\"ruby:rubocop\","; }; \
		else \
			echo '{"level":"info","msg":"skipping ruby rubocop lint: no .rb files found","language":"ruby"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"lint\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
		if [ -n "$$rb_files" ]; then \
			reek . || { overall_exit=1; failed_languages="$${failed_languages}\"ruby:reek\","; }; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"lint\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_GO)" ]; then \
		ran_languages="$${ran_languages}\"go\","; \
		go_files=$$(find . -name '*.go' -not -path './.git/*' -not -path './vendor/*' -not -path './node_modules/*' 2>/dev/null); \
		if [ -n "$$go_files" ]; then \
			golangci-lint run ./... || { overall_exit=1; failed_languages="$${failed_languages}\"go\","; }; \
		else \
			echo '{"level":"info","msg":"skipping go lint: no .go files found","language":"go"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"lint\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_JAVASCRIPT)" ]; then \
		ran_languages="$${ran_languages}\"javascript\","; \
		js_files=$$(find . \( -name '*.js' -o -name '*.jsx' -o -name '*.ts' -o -name '*.tsx' -o -name '*.mjs' -o -name '*.cjs' \) -not -path './.git/*' -not -path './vendor/*' -not -path './node_modules/*' -not -path './dist/*' -not -path './build/*' 2>/dev/null); \
		if [ -n "$$js_files" ]; then \
			eslint . || { overall_exit=1; failed_languages="$${failed_languages}\"javascript:eslint\","; }; \
		else \
			echo '{"level":"info","msg":"skipping javascript eslint lint: no JS/TS files found","language":"javascript"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"lint\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
		if [ -f "tsconfig.json" ]; then \
			tsc --noEmit || { overall_exit=1; failed_languages="$${failed_languages}\"javascript:tsc\","; }; \
		else \
			echo '{"level":"info","msg":"skipping tsc type check: no tsconfig.json found","language":"javascript"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"lint\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_RUST)" ]; then \
		ran_languages="$${ran_languages}\"rust\","; \
		rs_files=$$(find . -name '*.rs' -not -path './.git/*' -not -path './vendor/*' -not -path './target/*' 2>/dev/null); \
		if [ -n "$$rs_files" ]; then \
			cargo clippy --all-targets --all-features -- -D warnings || { overall_exit=1; failed_languages="$${failed_languages}\"rust\","; }; \
		else \
			echo '{"level":"info","msg":"skipping rust lint: no .rs files found","language":"rust"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"lint\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	end_time=$$(date +%s%3N); \
	duration=$$((end_time - start_time)); \
	if [ $$overall_exit -eq 0 ]; then \
		echo "{\"target\":\"lint\",\"status\":\"pass\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}]}"; \
	else \
		echo "{\"target\":\"lint\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
	fi; \
	exit $$overall_exit

# --- _format: language-specific format checking ---
_format: _check-config
	@start_time=$$(date +%s%3N); \
	overall_exit=0; \
	ran_languages=""; \
	failed_languages=""; \
	if [ -n "$(HAS_PYTHON)" ]; then \
		ran_languages="$${ran_languages}\"python\","; \
		ruff format --check . || { overall_exit=1; failed_languages="$${failed_languages}\"python\","; }; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"format\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_BASH)" ]; then \
		ran_languages="$${ran_languages}\"bash\","; \
		sh_files=$$(find . -name '*.sh' -not -path './.git/*' -not -path './vendor/*' -not -path './node_modules/*' 2>/dev/null); \
		if [ -n "$$sh_files" ]; then \
			echo "$$sh_files" | xargs shfmt -d || { overall_exit=1; failed_languages="$${failed_languages}\"bash\","; }; \
		else \
			echo '{"level":"info","msg":"skipping bash format: no .sh files found","language":"bash"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"format\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_TERRAFORM)" ]; then \
		ran_languages="$${ran_languages}\"terraform\","; \
		terraform fmt -check -recursive || { overall_exit=1; failed_languages="$${failed_languages}\"terraform\","; }; \
		tg_files=$$(find . -name 'terragrunt.hcl' -not -path './.git/*' -not -path './.terraform/*' 2>/dev/null); \
		if [ -n "$$tg_files" ]; then \
			terragrunt hclfmt --terragrunt-check || { overall_exit=1; failed_languages="$${failed_languages}\"terraform:terragrunt\","; }; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"format\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_ANSIBLE)" ]; then \
		ran_languages="$${ran_languages}\"ansible\","; \
		echo '{"target":"format","language":"ansible","status":"skip","reason":"no formatter configured"}' >&2; \
	fi; \
	if [ -n "$(HAS_RUBY)" ]; then \
		ran_languages="$${ran_languages}\"ruby\","; \
		rb_files=$$(find . -name '*.rb' -not -path './.git/*' -not -path './vendor/*' -not -path './node_modules/*' 2>/dev/null); \
		if [ -n "$$rb_files" ]; then \
			rubocop --check --fail-level error . || { overall_exit=1; failed_languages="$${failed_languages}\"ruby\","; }; \
		else \
			echo '{"level":"info","msg":"skipping ruby format: no .rb files found","language":"ruby"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"format\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_GO)" ]; then \
		ran_languages="$${ran_languages}\"go\","; \
		go_files=$$(find . -name '*.go' -not -path './.git/*' -not -path './vendor/*' -not -path './node_modules/*' 2>/dev/null); \
		if [ -n "$$go_files" ]; then \
			gofumpt -d . || { overall_exit=1; failed_languages="$${failed_languages}\"go\","; }; \
		else \
			echo '{"level":"info","msg":"skipping go format: no .go files found","language":"go"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"format\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_JAVASCRIPT)" ]; then \
		ran_languages="$${ran_languages}\"javascript\","; \
		js_files=$$(find . \( -name '*.js' -o -name '*.jsx' -o -name '*.ts' -o -name '*.tsx' -o -name '*.mjs' -o -name '*.cjs' \) -not -path './.git/*' -not -path './vendor/*' -not -path './node_modules/*' -not -path './dist/*' -not -path './build/*' 2>/dev/null); \
		if [ -n "$$js_files" ]; then \
			prettier --check . || { overall_exit=1; failed_languages="$${failed_languages}\"javascript\","; }; \
		else \
			echo '{"level":"info","msg":"skipping javascript format: no JS/TS files found","language":"javascript"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"format\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_RUST)" ]; then \
		ran_languages="$${ran_languages}\"rust\","; \
		rs_files=$$(find . -name '*.rs' -not -path './.git/*' -not -path './vendor/*' -not -path './target/*' 2>/dev/null); \
		if [ -n "$$rs_files" ]; then \
			cargo fmt --all -- --check || { overall_exit=1; failed_languages="$${failed_languages}\"rust\","; }; \
		else \
			echo '{"level":"info","msg":"skipping rust format: no .rs files found","language":"rust"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"format\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	end_time=$$(date +%s%3N); \
	duration=$$((end_time - start_time)); \
	if [ $$overall_exit -eq 0 ]; then \
		echo "{\"target\":\"format\",\"status\":\"pass\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}]}"; \
	else \
		echo "{\"target\":\"format\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
	fi; \
	exit $$overall_exit

# --- _fix: language-specific format fixing (in-place) ---
_fix: _check-config
	@start_time=$$(date +%s%3N); \
	overall_exit=0; \
	ran_languages=""; \
	failed_languages=""; \
	if [ -n "$(HAS_PYTHON)" ]; then \
		ran_languages="$${ran_languages}\"python\","; \
		ruff format . || { overall_exit=1; failed_languages="$${failed_languages}\"python\","; }; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"fix\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_BASH)" ]; then \
		ran_languages="$${ran_languages}\"bash\","; \
		sh_files=$$(find . -name '*.sh' -not -path './.git/*' -not -path './vendor/*' -not -path './node_modules/*' 2>/dev/null); \
		if [ -n "$$sh_files" ]; then \
			echo "$$sh_files" | xargs shfmt -w || { overall_exit=1; failed_languages="$${failed_languages}\"bash\","; }; \
		else \
			echo '{"level":"info","msg":"skipping bash fix: no .sh files found","language":"bash"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"fix\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_TERRAFORM)" ]; then \
		ran_languages="$${ran_languages}\"terraform\","; \
		terraform fmt -recursive || { overall_exit=1; failed_languages="$${failed_languages}\"terraform\","; }; \
		tg_files=$$(find . -name 'terragrunt.hcl' -not -path './.git/*' -not -path './.terraform/*' 2>/dev/null); \
		if [ -n "$$tg_files" ]; then \
			terragrunt hclfmt || { overall_exit=1; failed_languages="$${failed_languages}\"terraform:terragrunt\","; }; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"fix\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_ANSIBLE)" ]; then \
		ran_languages="$${ran_languages}\"ansible\","; \
		echo '{"target":"fix","language":"ansible","status":"skip","reason":"no formatter configured"}' >&2; \
	fi; \
	if [ -n "$(HAS_RUBY)" ]; then \
		ran_languages="$${ran_languages}\"ruby\","; \
		rb_files=$$(find . -name '*.rb' -not -path './.git/*' -not -path './vendor/*' -not -path './node_modules/*' 2>/dev/null); \
		if [ -n "$$rb_files" ]; then \
			rubocop -a . || { overall_exit=1; failed_languages="$${failed_languages}\"ruby\","; }; \
		else \
			echo '{"level":"info","msg":"skipping ruby fix: no .rb files found","language":"ruby"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"fix\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_GO)" ]; then \
		ran_languages="$${ran_languages}\"go\","; \
		go_files=$$(find . -name '*.go' -not -path './.git/*' -not -path './vendor/*' -not -path './node_modules/*' 2>/dev/null); \
		if [ -n "$$go_files" ]; then \
			gofumpt -w . || { overall_exit=1; failed_languages="$${failed_languages}\"go\","; }; \
		else \
			echo '{"level":"info","msg":"skipping go fix: no .go files found","language":"go"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"fix\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_JAVASCRIPT)" ]; then \
		ran_languages="$${ran_languages}\"javascript\","; \
		js_files=$$(find . \( -name '*.js' -o -name '*.jsx' -o -name '*.ts' -o -name '*.tsx' -o -name '*.mjs' -o -name '*.cjs' \) -not -path './.git/*' -not -path './vendor/*' -not -path './node_modules/*' -not -path './dist/*' -not -path './build/*' 2>/dev/null); \
		if [ -n "$$js_files" ]; then \
			prettier --write . || { overall_exit=1; failed_languages="$${failed_languages}\"javascript\","; }; \
		else \
			echo '{"level":"info","msg":"skipping javascript fix: no JS/TS files found","language":"javascript"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"fix\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_RUST)" ]; then \
		ran_languages="$${ran_languages}\"rust\","; \
		rs_files=$$(find . -name '*.rs' -not -path './.git/*' -not -path './vendor/*' -not -path './target/*' 2>/dev/null); \
		if [ -n "$$rs_files" ]; then \
			cargo fmt --all || { overall_exit=1; failed_languages="$${failed_languages}\"rust\","; }; \
		else \
			echo '{"level":"info","msg":"skipping rust fix: no .rs files found","language":"rust"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"fix\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	end_time=$$(date +%s%3N); \
	duration=$$((end_time - start_time)); \
	if [ $$overall_exit -eq 0 ]; then \
		echo "{\"target\":\"fix\",\"status\":\"pass\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}]}"; \
	else \
		echo "{\"target\":\"fix\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
	fi; \
	exit $$overall_exit

# --- _test: language-specific test runners ---
_test: _check-config
	@start_time=$$(date +%s%3N); \
	overall_exit=0; \
	ran_languages=""; \
	failed_languages=""; \
	skipped_languages=""; \
	if [ -n "$(HAS_PYTHON)" ]; then \
		if [ -d "tests" ] || find . -name '*_test.py' -o -name 'test_*.py' 2>/dev/null | grep -q .; then \
			ran_languages="$${ran_languages}\"python\","; \
			pytest || { overall_exit=1; failed_languages="$${failed_languages}\"python\","; }; \
		else \
			skipped_languages="$${skipped_languages}\"python\","; \
			echo '{"level":"info","msg":"skipping python tests: no test files found","language":"python"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"test\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}],\"skipped\":[$${skipped_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_BASH)" ]; then \
		if find . -name '*.bats' -not -path './.git/*' 2>/dev/null | grep -q .; then \
			ran_languages="$${ran_languages}\"bash\","; \
			bats $$(find . -name '*.bats' -not -path './.git/*') || { overall_exit=1; failed_languages="$${failed_languages}\"bash\","; }; \
		else \
			skipped_languages="$${skipped_languages}\"bash\","; \
			echo '{"level":"info","msg":"skipping bash tests: no .bats files found","language":"bash"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"test\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}],\"skipped\":[$${skipped_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_TERRAFORM)" ]; then \
		if find . -name '*_test.go' -not -path './.git/*' 2>/dev/null | grep -q .; then \
			ran_languages="$${ran_languages}\"terraform\","; \
			(cd tests && go test ./...) || { overall_exit=1; failed_languages="$${failed_languages}\"terraform\","; }; \
		else \
			skipped_languages="$${skipped_languages}\"terraform\","; \
			echo '{"level":"info","msg":"skipping terraform tests: no *_test.go files found","language":"terraform"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"test\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}],\"skipped\":[$${skipped_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_ANSIBLE)" ]; then \
		if [ -d "molecule" ]; then \
			ran_languages="$${ran_languages}\"ansible\","; \
			molecule test || { overall_exit=1; failed_languages="$${failed_languages}\"ansible\","; }; \
		else \
			skipped_languages="$${skipped_languages}\"ansible\","; \
			echo '{"level":"info","msg":"skipping ansible tests: no molecule directory found","language":"ansible"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"test\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}],\"skipped\":[$${skipped_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_RUBY)" ]; then \
		if [ -d "spec" ]; then \
			ran_languages="$${ran_languages}\"ruby\","; \
			rspec || { overall_exit=1; failed_languages="$${failed_languages}\"ruby\","; }; \
		else \
			skipped_languages="$${skipped_languages}\"ruby\","; \
			echo '{"level":"info","msg":"skipping ruby tests: no spec/ directory found","language":"ruby"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"test\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}],\"skipped\":[$${skipped_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_GO)" ]; then \
		if find . -name '*_test.go' -not -path './.git/*' -not -path './vendor/*' 2>/dev/null | grep -q .; then \
			ran_languages="$${ran_languages}\"go\","; \
			go test ./... || { overall_exit=1; failed_languages="$${failed_languages}\"go\","; }; \
		else \
			skipped_languages="$${skipped_languages}\"go\","; \
			echo '{"level":"info","msg":"skipping go tests: no *_test.go files found","language":"go"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"test\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}],\"skipped\":[$${skipped_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_JAVASCRIPT)" ]; then \
		if find . \( -name '*.test.*' -o -name '*.spec.*' \) -not -path './.git/*' -not -path './vendor/*' -not -path './node_modules/*' -not -path './dist/*' -not -path './build/*' 2>/dev/null | grep -q .; then \
			ran_languages="$${ran_languages}\"javascript\","; \
			vitest run || { overall_exit=1; failed_languages="$${failed_languages}\"javascript\","; }; \
		else \
			skipped_languages="$${skipped_languages}\"javascript\","; \
			echo '{"level":"info","msg":"skipping javascript tests: no *.test.* or *.spec.* files found","language":"javascript"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"test\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}],\"skipped\":[$${skipped_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_RUST)" ]; then \
		rs_files=$$(find . -name '*.rs' -not -path './.git/*' -not -path './vendor/*' -not -path './target/*' 2>/dev/null); \
		if [ -n "$$rs_files" ] && [ -f "Cargo.toml" ]; then \
			ran_languages="$${ran_languages}\"rust\","; \
			cargo test --all-targets || { overall_exit=1; failed_languages="$${failed_languages}\"rust\","; }; \
		else \
			skipped_languages="$${skipped_languages}\"rust\","; \
			echo '{"level":"info","msg":"skipping rust tests: no .rs files or Cargo.toml found","language":"rust"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"test\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}],\"skipped\":[$${skipped_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	end_time=$$(date +%s%3N); \
	duration=$$((end_time - start_time)); \
	if [ -z "$${ran_languages}" ] && [ -n "$${skipped_languages}" ]; then \
		echo "{\"target\":\"test\",\"status\":\"skip\",\"reason\":\"no tests found\",\"duration_ms\":$$duration,\"skipped\":[$${skipped_languages%,}]}"; \
	elif [ $$overall_exit -eq 0 ]; then \
		echo "{\"target\":\"test\",\"status\":\"pass\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"skipped\":[$${skipped_languages%,}]}"; \
	else \
		echo "{\"target\":\"test\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}],\"skipped\":[$${skipped_languages%,}]}"; \
	fi; \
	exit $$overall_exit

# --- _security: language-specific security scanners ---
_security: _check-config
	@start_time=$$(date +%s%3N); \
	overall_exit=0; \
	ran_languages=""; \
	failed_languages=""; \
	skipped_languages=""; \
	if [ -n "$(HAS_PYTHON)" ]; then \
		ran_languages="$${ran_languages}\"python\","; \
		bandit -r . -q || { overall_exit=1; failed_languages="$${failed_languages}\"python:bandit\","; }; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"security\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
		semgrep --config auto . --quiet 2>/dev/null || { overall_exit=1; failed_languages="$${failed_languages}\"python:semgrep\","; }; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"security\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_BASH)" ]; then \
		skipped_languages="$${skipped_languages}\"bash\","; \
		echo '{"level":"info","msg":"skipping bash security: no language-specific scanner","language":"bash"}' >&2; \
	fi; \
	if [ -n "$(HAS_TERRAFORM)" ]; then \
		ran_languages="$${ran_languages}\"terraform\","; \
		tfsec . || { overall_exit=1; failed_languages="$${failed_languages}\"terraform:tfsec\","; }; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"security\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
		checkov -d . --quiet || { overall_exit=1; failed_languages="$${failed_languages}\"terraform:checkov\","; }; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"security\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_ANSIBLE)" ]; then \
		skipped_languages="$${skipped_languages}\"ansible\","; \
		echo '{"level":"info","msg":"skipping ansible security: no language-specific scanner","language":"ansible"}' >&2; \
	fi; \
	if [ -n "$(HAS_RUBY)" ]; then \
		ran_languages="$${ran_languages}\"ruby\","; \
		if [ -f "config/application.rb" ]; then \
			brakeman -q || { overall_exit=1; failed_languages="$${failed_languages}\"ruby:brakeman\","; }; \
		else \
			echo '{"level":"info","msg":"skipping brakeman: not a Rails application","language":"ruby"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"security\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
		if [ -f "Gemfile.lock" ]; then \
			bundler-audit check || { overall_exit=1; failed_languages="$${failed_languages}\"ruby:bundler-audit\","; }; \
		else \
			echo '{"level":"info","msg":"skipping bundler-audit: no Gemfile.lock found","language":"ruby"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"security\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_GO)" ]; then \
		if [ -f "go.sum" ]; then \
			ran_languages="$${ran_languages}\"go\","; \
			govulncheck ./... || { overall_exit=1; failed_languages="$${failed_languages}\"go:govulncheck\","; }; \
		else \
			skipped_languages="$${skipped_languages}\"go\","; \
			echo '{"level":"info","msg":"skipping govulncheck: no go.sum found","language":"go"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"security\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_JAVASCRIPT)" ]; then \
		if [ -f "package-lock.json" ]; then \
			ran_languages="$${ran_languages}\"javascript\","; \
			npm audit --audit-level=moderate || { overall_exit=1; failed_languages="$${failed_languages}\"javascript:npm-audit\","; }; \
		else \
			skipped_languages="$${skipped_languages}\"javascript\","; \
			echo '{"level":"info","msg":"skipping npm audit: no package-lock.json found","language":"javascript"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"security\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	if [ -n "$(HAS_RUST)" ]; then \
		if [ -f "Cargo.lock" ]; then \
			ran_languages="$${ran_languages}\"rust\","; \
			cargo audit || { overall_exit=1; failed_languages="$${failed_languages}\"rust:cargo-audit\","; }; \
		else \
			skipped_languages="$${skipped_languages}\"rust:cargo-audit\","; \
			echo '{"level":"info","msg":"skipping cargo audit: no Cargo.lock found","language":"rust"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"security\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
		if [ -f "deny.toml" ]; then \
			cargo deny check || { overall_exit=1; failed_languages="$${failed_languages}\"rust:cargo-deny\","; }; \
		else \
			echo '{"level":"info","msg":"skipping cargo deny: no deny.toml found","language":"rust"}' >&2; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
			end_time=$$(date +%s%3N); \
			duration=$$((end_time - start_time)); \
			echo "{\"target\":\"security\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}]}"; \
			exit $$overall_exit; \
		fi; \
	fi; \
	end_time=$$(date +%s%3N); \
	duration=$$((end_time - start_time)); \
	if [ -z "$${ran_languages}" ] && [ -n "$${skipped_languages}" ]; then \
		echo "{\"target\":\"security\",\"status\":\"skip\",\"reason\":\"no security scanners for declared languages\",\"duration_ms\":$$duration,\"skipped\":[$${skipped_languages%,}]}"; \
	elif [ $$overall_exit -eq 0 ]; then \
		echo "{\"target\":\"security\",\"status\":\"pass\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"skipped\":[$${skipped_languages%,}]}"; \
	else \
		echo "{\"target\":\"security\",\"status\":\"fail\",\"duration_ms\":$$duration,\"languages\":[$${ran_languages%,}],\"failed\":[$${failed_languages%,}],\"skipped\":[$${skipped_languages%,}]}"; \
	fi; \
	exit $$overall_exit

# --- _scan: universal vulnerability and secret scanning ---
_scan: _check-config
	@start_time=$$(date +%s%3N); \
	overall_exit=0; \
	failed_scanners=""; \
	trivy fs --format json --output /tmp/trivy-results.json . 2>/dev/null; \
	trivy_exit=$$?; \
	if [ $$trivy_exit -eq 1 ]; then \
		overall_exit=1; \
		failed_scanners="$${failed_scanners}\"trivy\","; \
	elif [ $$trivy_exit -gt 1 ]; then \
		echo "{\"target\":\"scan\",\"status\":\"error\",\"error\":\"trivy exited with code $$trivy_exit\",\"exit_code\":2}"; \
		exit 2; \
	fi; \
	if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$overall_exit -ne 0 ]; then \
		end_time=$$(date +%s%3N); \
		duration=$$((end_time - start_time)); \
		echo "{\"target\":\"scan\",\"status\":\"fail\",\"duration_ms\":$$duration,\"scanners\":[\"trivy\",\"gitleaks\"],\"failed\":[$${failed_scanners%,}]}"; \
		exit $$overall_exit; \
	fi; \
	gitleaks detect --source . --report-format json --report-path /tmp/gitleaks-results.json 2>/dev/null; \
	gitleaks_exit=$$?; \
	if [ $$gitleaks_exit -eq 1 ]; then \
		overall_exit=1; \
		failed_scanners="$${failed_scanners}\"gitleaks\","; \
	elif [ $$gitleaks_exit -gt 1 ]; then \
		echo "{\"target\":\"scan\",\"status\":\"error\",\"error\":\"gitleaks exited with code $$gitleaks_exit\",\"exit_code\":2}"; \
		exit 2; \
	fi; \
	end_time=$$(date +%s%3N); \
	duration=$$((end_time - start_time)); \
	if [ $$overall_exit -eq 0 ]; then \
		echo "{\"target\":\"scan\",\"status\":\"pass\",\"duration_ms\":$$duration,\"scanners\":[\"trivy\",\"gitleaks\"]}"; \
	else \
		echo "{\"target\":\"scan\",\"status\":\"fail\",\"duration_ms\":$$duration,\"scanners\":[\"trivy\",\"gitleaks\"],\"failed\":[$${failed_scanners%,}]}"; \
	fi; \
	exit $$overall_exit

# --- _docs: documentation generation ---
_docs: _check-config
	@start_time=$$(date +%s%3N); \
	overall_exit=0; \
	generators=""; \
	modules=""; \
	if [ -n "$(HAS_TERRAFORM)" ]; then \
		tf_dirs=$$(find . -name '*.tf' -not -path './.git/*' -not -path './.terraform/*' 2>/dev/null | xargs -I{} dirname {} | sort -u); \
		if [ -n "$$tf_dirs" ]; then \
			for dir in $$tf_dirs; do \
				terraform-docs markdown table --output-file README.md "$$dir" || overall_exit=1; \
				modules="$${modules}\"$$dir\","; \
			done; \
			generators="$${generators}\"terraform-docs\","; \
		else \
			echo '{"level":"info","msg":"skipping terraform-docs: no .tf files found","language":"terraform"}' >&2; \
		fi; \
	fi; \
	mkdir -p .devrail-output; \
	_sep=""; \
	_tv() { _out=$$(eval "$$2" 2>&1 | grep -oE '[0-9]+\.[0-9]+[^ ]*' | head -1); \
		[ -z "$$_out" ] && _out="unknown"; \
		printf '%s"%s":"%s"' "$$_sep" "$$1" "$$_out"; _sep=","; }; \
	{ \
		printf '{"generated_at":"%s","tools":{' "$$(date -u +%Y-%m-%dT%H:%M:%SZ)"; \
		if [ -n "$(HAS_PYTHON)" ]; then \
			_tv ruff "ruff --version"; \
			_tv bandit "bandit --version"; \
			_tv mypy "mypy --version"; \
			_tv pytest "pytest --version"; \
			_tv semgrep "semgrep --version"; \
		fi; \
		if [ -n "$(HAS_BASH)" ]; then \
			_tv shellcheck "shellcheck --version"; \
			_tv shfmt "shfmt --version"; \
			_tv bats "bats --version"; \
		fi; \
		if [ -n "$(HAS_TERRAFORM)" ]; then \
			_tv terraform "terraform version"; \
			_tv tflint "tflint --version"; \
			_tv tfsec "tfsec --version"; \
			_tv checkov "checkov --version"; \
			_tv terraform-docs "terraform-docs --version"; \
			_tv terragrunt "terragrunt --version"; \
		fi; \
		if [ -n "$(HAS_ANSIBLE)" ]; then \
			_tv ansible-lint "ansible-lint --version"; \
			_tv molecule "molecule --version"; \
		fi; \
		if [ -n "$(HAS_RUBY)" ]; then \
			_tv rubocop "rubocop --version"; \
			_tv reek "reek --version"; \
			_tv brakeman "brakeman --version"; \
			_tv bundler-audit "bundler-audit --version"; \
			_tv rspec "rspec --version"; \
			_tv srb "srb --version"; \
		fi; \
		if [ -n "$(HAS_GO)" ]; then \
			_tv go "go version"; \
			_tv golangci-lint "golangci-lint version"; \
			_tv gofumpt "gofumpt --version"; \
			_tv govulncheck "govulncheck -version"; \
		fi; \
		if [ -n "$(HAS_JAVASCRIPT)" ]; then \
			_tv node "node --version"; \
			_tv npm "npm --version"; \
			_tv eslint "eslint --version"; \
			_tv prettier "prettier --version"; \
			_tv tsc "tsc --version"; \
			_tv vitest "vitest --version"; \
		fi; \
		if [ -n "$(HAS_RUST)" ]; then \
			_tv rustc "rustc --version"; \
			_tv cargo "cargo --version"; \
			_tv clippy "cargo clippy --version"; \
			_tv rustfmt "rustfmt --version"; \
			_tv cargo-audit "cargo audit --version"; \
			_tv cargo-deny "cargo deny --version"; \
		fi; \
		_tv trivy "trivy --version"; \
		_tv gitleaks "gitleaks version"; \
		_tv git-cliff "git-cliff --version"; \
		printf '}}\n'; \
	} > .devrail-output/tool-versions.json; \
	generators="$${generators}\"tool-versions\","; \
	end_time=$$(date +%s%3N); \
	duration=$$((end_time - start_time)); \
	if [ $$overall_exit -eq 0 ]; then \
		echo "{\"target\":\"docs\",\"status\":\"pass\",\"duration_ms\":$$duration,\"generators\":[$${generators%,}],\"modules\":[$${modules%,}]}"; \
	else \
		echo "{\"target\":\"docs\",\"status\":\"fail\",\"duration_ms\":$$duration,\"generators\":[$${generators%,}],\"modules\":[$${modules%,}]}"; \
	fi; \
	exit $$overall_exit

# --- _changelog: generate CHANGELOG.md from conventional commits ---
_changelog: _check-config
	@start_time=$$(date +%s%3N); \
	config=""; \
	if [ -f "cliff.toml" ]; then \
		config="cliff.toml"; \
	elif [ -f "/opt/devrail/config/cliff.toml" ]; then \
		config="/opt/devrail/config/cliff.toml"; \
	fi; \
	if [ -z "$$config" ]; then \
		echo '{"target":"changelog","status":"error","error":"no cliff.toml found","exit_code":2}'; \
		exit 2; \
	fi; \
	if ! git rev-parse --git-dir >/dev/null 2>&1; then \
		echo '{"target":"changelog","status":"error","error":"not a git repository","exit_code":2}'; \
		exit 2; \
	fi; \
	git-cliff --config "$$config" --output CHANGELOG.md; \
	cl_exit=$$?; \
	end_time=$$(date +%s%3N); \
	duration=$$((end_time - start_time)); \
	if [ $$cl_exit -eq 0 ]; then \
		echo "{\"target\":\"changelog\",\"status\":\"pass\",\"duration_ms\":$$duration,\"config\":\"$$config\",\"output\":\"CHANGELOG.md\"}"; \
	else \
		echo "{\"target\":\"changelog\",\"status\":\"fail\",\"duration_ms\":$$duration,\"config\":\"$$config\",\"exit_code\":$$cl_exit}"; \
		exit $$cl_exit; \
	fi

# --- _init: scaffold config files for declared languages ---
_init: _check-config
	@created=""; \
	skipped=""; \
	scaffold() { \
	  _f="$$1"; shift; \
	  if [ ! -f "$$_f" ]; then \
	    printf '%s\n' "$$@" > "$$_f"; \
	    created="$${created}\"$$_f\","; \
	  else \
	    skipped="$${skipped}\"$$_f\","; \
	  fi; \
	}; \
	scaffold .editorconfig \
	  'root = true' \
	  '' \
	  '[*]' \
	  'charset = utf-8' \
	  'end_of_line = lf' \
	  'insert_final_newline = true' \
	  'trim_trailing_whitespace = true' \
	  'indent_style = space' \
	  'indent_size = 2' \
	  '' \
	  '[Makefile]' \
	  'indent_style = tab' \
	  '' \
	  '[*.py]' \
	  'indent_size = 4' \
	  '' \
	  '[*.sh]' \
	  'indent_size = 2'; \
	if [ -f "/opt/devrail/config/cliff.toml" ] && [ ! -f "cliff.toml" ]; then \
	  cp /opt/devrail/config/cliff.toml cliff.toml; \
	  created="$${created}\"cliff.toml\","; \
	elif [ -f "cliff.toml" ]; then \
	  skipped="$${skipped}\"cliff.toml\","; \
	fi; \
	if [ -n "$(HAS_PYTHON)" ]; then \
	  scaffold ruff.toml \
	    'line-length = 120' \
	    'target-version = "py311"' \
	    '' \
	    '[lint]' \
	    'select = ["E", "W", "F", "I", "UP", "B", "S", "C4", "SIM"]' \
	    '' \
	    '[format]' \
	    'quote-style = "double"' \
	    'indent-style = "space"'; \
	fi; \
	if [ -n "$(HAS_BASH)" ]; then \
	  scaffold .shellcheckrc \
	    'shell=bash' \
	    'enable=all'; \
	fi; \
	if [ -n "$(HAS_TERRAFORM)" ]; then \
	  scaffold .tflint.hcl \
	    'config {' \
	    '  call_module_type = "local"' \
	    '}' \
	    '' \
	    'plugin "terraform" {' \
	    '  enabled = true' \
	    '  preset  = "recommended"' \
	    '}'; \
	fi; \
	if [ -n "$(HAS_ANSIBLE)" ]; then \
	  scaffold .ansible-lint \
	    'profile: production' \
	    '' \
	    'exclude_paths:' \
	    '  - .cache/' \
	    '  - .github/' \
	    '  - .gitlab/' \
	    '' \
	    'skip_list:' \
	    '  - yaml[truthy]' \
	    '' \
	    'warn_list:' \
	    '  - experimental'; \
	fi; \
	if [ -n "$(HAS_RUBY)" ]; then \
	  scaffold .rubocop.yml \
	    'AllCops:' \
	    '  TargetRubyVersion: 3.1' \
	    '  NewCops: enable' \
	    '  Exclude:' \
	    '    - "db/schema.rb"' \
	    '    - "bin/**/*"' \
	    '    - "vendor/**/*"' \
	    '    - "node_modules/**/*"' \
	    '' \
	    'Style/Documentation:' \
	    '  Enabled: false' \
	    '' \
	    'Metrics/BlockLength:' \
	    '  Exclude:' \
	    '    - "spec/**/*"' \
	    '' \
	    'Layout/LineLength:' \
	    '  Max: 120'; \
	  scaffold .reek.yml \
	    'exclude_paths:' \
	    '  - vendor' \
	    '  - db/schema.rb' \
	    '  - bin' \
	    '' \
	    'detectors:' \
	    '  IrresponsibleModule:' \
	    '    enabled: false'; \
	  scaffold .rspec \
	    '--require spec_helper' \
	    '--format documentation' \
	    '--color'; \
	fi; \
	if [ -n "$(HAS_GO)" ]; then \
	  scaffold .golangci.yml \
	    'version: "2"' \
	    '' \
	    'linters:' \
	    '  enable:' \
	    '    - errcheck' \
	    '    - govet' \
	    '    - staticcheck' \
	    '    - gosec' \
	    '    - ineffassign' \
	    '    - unused' \
	    '    - gocritic' \
	    '    - gofumpt' \
	    '    - misspell' \
	    '    - revive' \
	    '' \
	    'issues:' \
	    '  exclude-dirs:' \
	    '    - vendor' \
	    '    - node_modules'; \
	fi; \
	if [ -n "$(HAS_JAVASCRIPT)" ]; then \
	  scaffold eslint.config.js \
	    'import eslint from "@eslint/js";' \
	    'import tseslint from "typescript-eslint";' \
	    '' \
	    'export default tseslint.config(' \
	    '  eslint.configs.recommended,' \
	    '  tseslint.configs.recommended,' \
	    '  {' \
	    '    ignores: ["node_modules/", "dist/", "build/", "coverage/"],' \
	    '  }' \
	    ');'; \
	  scaffold .prettierrc \
	    '{' \
	    '  "semi": true,' \
	    '  "singleQuote": false,' \
	    '  "trailingComma": "es5",' \
	    '  "printWidth": 80,' \
	    '  "tabWidth": 2' \
	    '}'; \
	  scaffold .prettierignore \
	    'node_modules/' \
	    'dist/' \
	    'build/' \
	    'coverage/'; \
	fi; \
	if [ -n "$(HAS_RUST)" ]; then \
	  scaffold clippy.toml \
	    '# clippy.toml -- DevRail Rust clippy configuration' \
	    '# See: https://doc.rust-lang.org/clippy/lint_configuration.html' \
	    'too-many-arguments-threshold = 7'; \
	  scaffold rustfmt.toml \
	    '# rustfmt.toml -- DevRail Rust formatter configuration' \
	    'edition = "2021"' \
	    'max_width = 100' \
	    'use_field_init_shorthand = true' \
	    'use_try_shorthand = true'; \
	  scaffold deny.toml \
	    '# deny.toml -- DevRail cargo-deny configuration' \
	    '# See: https://embarkstudios.github.io/cargo-deny/' \
	    '' \
	    '[advisories]' \
	    'vulnerability = "deny"' \
	    'unmaintained = "warn"' \
	    'yanked = "warn"' \
	    '' \
	    '[licenses]' \
	    'unlicensed = "deny"' \
	    'allow = [' \
	    '  "MIT",' \
	    '  "Apache-2.0",' \
	    '  "BSD-2-Clause",' \
	    '  "BSD-3-Clause",' \
	    '  "ISC",' \
	    '  "Unicode-3.0",' \
	    '  "Unicode-DFS-2016",' \
	    ']' \
	    '' \
	    '[bans]' \
	    'multiple-versions = "warn"' \
	    '' \
	    '[sources]' \
	    'unknown-registry = "deny"' \
	    'unknown-git = "warn"'; \
	fi; \
	echo "{\"target\":\"init\",\"created\":[$${created%,}],\"skipped\":[$${skipped%,}]}"

# --- _check: orchestrate all targets ---
_check: _check-config
	@overall_exit=0; \
	overall_start=$$(date +%s%3N); \
	results=""; \
	passed=""; \
	failed=""; \
	skipped=""; \
	for target in lint format test security scan docs; do \
		target_start=$$(date +%s%3N); \
		json_output=$$($(MAKE) _$${target} 2>/dev/null); \
		target_exit=$$?; \
		target_end=$$(date +%s%3N); \
		target_duration=$$((target_end - target_start)); \
		status=$$(echo "$$json_output" | grep -o '"status":"[^"]*"' | head -1 | cut -d'"' -f4); \
		if [ -z "$$status" ]; then \
			if [ $$target_exit -eq 0 ]; then status="pass"; \
			elif [ $$target_exit -eq 2 ]; then status="error"; \
			else status="fail"; fi; \
		fi; \
		results="$${results}{\"target\":\"$$target\",\"status\":\"$$status\",\"duration_ms\":$$target_duration},"; \
		case "$$status" in \
			pass) passed="$${passed}\"$$target\","; ;; \
			fail|error) \
				failed="$${failed}\"$$target\","; \
				if [ $$target_exit -eq 2 ]; then overall_exit=2; \
				elif [ $$overall_exit -ne 2 ]; then overall_exit=1; fi; \
				;; \
			skip) skipped="$${skipped}\"$$target\","; ;; \
		esac; \
		if [ "$(DEVRAIL_LOG_FORMAT)" = "human" ]; then \
			case "$$status" in \
				pass) printf '\033[32m%-12s PASS   %s\033[0m\n' "$$target" "$${target_duration}ms" >&2; ;; \
				fail|error) printf '\033[31m%-12s FAIL   %s\033[0m\n' "$$target" "$${target_duration}ms" >&2; ;; \
				skip) printf '\033[33m%-12s SKIP   %s\033[0m\n' "$$target" "$${target_duration}ms" >&2; ;; \
			esac; \
		fi; \
		if [ "$(DEVRAIL_FAIL_FAST)" = "1" ] && [ $$target_exit -ne 0 ]; then \
			for remaining in lint format test security scan docs; do \
				found=0; \
				for done_target in lint format test security scan docs; do \
					if [ "$$done_target" = "$$target" ]; then found=1; break; fi; \
					if [ "$$done_target" = "$$remaining" ]; then break; fi; \
				done; \
			done; \
			break; \
		fi; \
	done; \
	overall_end=$$(date +%s%3N); \
	overall_duration=$$((overall_end - overall_start)); \
	if [ "$(DEVRAIL_LOG_FORMAT)" = "human" ]; then \
		echo "=========================================" >&2; \
		echo "DevRail Check Summary" >&2; \
		echo "=========================================" >&2; \
		if [ $$overall_exit -eq 0 ]; then \
			printf '\033[32mResult: PASS  Total: %sms\033[0m\n' "$$overall_duration" >&2; \
		else \
			printf '\033[31mResult: FAIL  Total: %sms\033[0m\n' "$$overall_duration" >&2; \
		fi; \
		echo "=========================================" >&2; \
	fi; \
	if [ $$overall_exit -eq 0 ]; then \
		check_status="pass"; \
	else \
		check_status="fail"; \
	fi; \
	echo "{\"target\":\"check\",\"status\":\"$$check_status\",\"duration_ms\":$$overall_duration,\"results\":[$${results%,}],\"passed\":[$${passed%,}],\"failed\":[$${failed%,}],\"skipped\":[$${skipped%,}]}"; \
	exit $$overall_exit
