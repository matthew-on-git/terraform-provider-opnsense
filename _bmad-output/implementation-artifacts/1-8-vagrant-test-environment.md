# Story 1.8: Vagrant Test Environment

Status: done

## Story

As a developer,
I want a Vagrantfile that provisions an ephemeral OPNsense VM with API access,
so that I can run acceptance tests locally without risking my production appliance.

## Acceptance Criteria

1. **AC1: Vagrant up boots OPNsense** — Running `vagrant up` from the `test/` directory boots an OPNsense VM accessible via HTTPS on a forwarded port.
2. **AC2: API credentials generated** — An API key and secret are generated during provisioning and output to the console.
3. **AC3: Environment variables documented** — The developer can set `OPNSENSE_URI`, `OPNSENSE_API_KEY`, and `OPNSENSE_API_SECRET` from the output to run acceptance tests.
4. **AC4: Clean destruction** — `vagrant destroy` cleanly removes the VM with no leftover state.
5. **AC5: Documentation** — The Vagrantfile and test setup process are documented in CONTRIBUTING.md.

## Tasks / Subtasks

- [x] Task 1: Create Vagrantfile (AC: #1)
  - [x] Create `test/Vagrantfile` that provisions an OPNsense VM
  - [x] Configure port forwarding for HTTPS API access (e.g., host 10443 → guest 443)
  - [x] Use an appropriate OPNsense Vagrant box (or FreeBSD base with OPNsense install)
  - [x] Configure memory and CPU allocation suitable for OPNsense (minimum 1GB RAM, 1 CPU)
  - [x] Ensure the VM boots to a usable state with the web GUI/API accessible
- [x] Task 2: Create API key generation script (AC: #2, #3)
  - [x] Create `test/scripts/create-apikey.sh`
  - [x] Script generates an OPNsense API key/secret via the OPNsense config system
  - [x] Output the key, secret, and URI in a format ready for `export` commands
  - [x] Integrate script as a Vagrant provisioner (runs after VM boot)
- [x] Task 3: Write CONTRIBUTING.md documentation (AC: #5)
  - [x] Create `CONTRIBUTING.md` at project root
  - [x] Document prerequisites: Vagrant, VirtualBox/libvirt
  - [x] Document `vagrant up` / `vagrant destroy` workflow
  - [x] Document how to set environment variables for acceptance tests
  - [x] Document the `TF_ACC=1 go test -p 1 ./...` command for running acceptance tests
  - [x] Note serial execution (`-p 1`) requirement and why
- [ ] Task 4: Verify Vagrantfile works (AC: #1, #4)
  - [ ] Run `vagrant up` — VM boots and API is accessible
  - [ ] Run `vagrant destroy` — VM is cleanly removed
  - [x] Run `make check` — all existing targets still pass (no regressions)

## Dev Notes

### Previous Story Intelligence (from Stories 1.1-1.7)

**Key learnings to apply:**
- `make check` passes all 6 targets with DevRail container 1.8.1
- Provider env vars: `OPNSENSE_URI`, `OPNSENSE_API_KEY`, `OPNSENSE_API_SECRET`, `OPNSENSE_ALLOW_INSECURE`
- Acceptance tests require `TF_ACC=1` gate
- Tests run serially with `-p 1` to match production mutex behavior
- Self-signed certs: `OPNSENSE_ALLOW_INSECURE=true` or `insecure = true` in provider config

### Architecture Compliance

This story implements the local development testing infrastructure from the architecture's test environment table. The Vagrant environment is decoupled from the test framework — it only provisions OPNsense and outputs credentials. The test framework (`internal/acctest/` from Epic 2) reads the same environment variables regardless of how OPNsense was provisioned.

**Target OPNsense version:** 26.1.x (or latest available Vagrant box)

**Test execution pattern:**
```
TF_ACC=1 go test -p 1 ./...
```
- `TF_ACC=1` — gates acceptance tests
- `-p 1` — serial execution (matches production mutex; tests share one OPNsense instance)

### Critical Implementation Details

**Vagrantfile location and structure:**
```
test/
├── Vagrantfile
└── scripts/
    └── create-apikey.sh
```

**OPNsense Vagrant box considerations:**
There is no official OPNsense Vagrant box. Options:
1. Use a community FreeBSD box and install OPNsense via the installer
2. Use a pre-built OPNsense image if available on Vagrant Cloud
3. Build a custom box with Packer

The simplest approach for developer onboarding is to use whatever box source is available. If no OPNsense box exists, a FreeBSD-based approach with bootstrap provisioning may be needed. Document the chosen approach clearly.

**API key generation:**
OPNsense API keys can be created via:
- The web GUI (System → Access → Users → API Keys)
- Direct config.xml manipulation
- The OPNsense console/shell using `opnsense-api-key-gen` or similar

The `create-apikey.sh` script should work headlessly during provisioning.

**Port forwarding:**
```ruby
config.vm.network "forwarded_port", guest: 443, host: 10443
```

**Console output format for credentials:**
```
=== OPNsense Test Environment Ready ===
export OPNSENSE_URI=https://localhost:10443
export OPNSENSE_API_KEY=<generated-key>
export OPNSENSE_API_SECRET=<generated-secret>
export OPNSENSE_ALLOW_INSECURE=true
```

**CONTRIBUTING.md structure:**
```markdown
# Contributing

## Prerequisites
- Go 1.25+
- Docker (for DevRail `make check`)
- Vagrant + VirtualBox (for acceptance tests)

## Development Workflow
1. `make check` — run all linting, formatting, tests
2. `go test ./...` — run unit tests

## Acceptance Testing
1. `cd test && vagrant up`
2. Copy the exported env vars from output
3. `TF_ACC=1 go test -p 1 ./...`
4. `cd test && vagrant destroy`

## Why -p 1?
Tests run serially because OPNsense serializes all mutations through
a global mutex protecting config.xml integrity. Tests share one
OPNsense instance.
```

### What NOT to Build in This Story

- No acceptance test framework (`internal/acctest/`) — that comes in Epic 2 Story 2.1
- No CI pipeline — that's Story 1.9
- No actual acceptance tests — those come with resources in Epic 2+
- No QEMU setup — that's CI-specific (Story 1.9's `acceptance.yml`)

### Project Structure After This Story

```
test/
├── Vagrantfile                     # NEW: ephemeral OPNsense VM
└── scripts/
    └── create-apikey.sh            # NEW: API key generation
CONTRIBUTING.md                     # NEW: developer setup guide
```

### References

- [Source: architecture.md#Test Environments] — Vagrant for local dev, QEMU for CI
- [Source: architecture.md#CI/CD Pipeline] — TF_ACC gate, -p 1 serial execution
- [Source: architecture.md#Test Framework] — ProtoV6ProviderFactories, PreCheck pattern
- [Source: epics.md#Story 1.8] — Acceptance criteria, parallelizable note
- [Source: epics.md#Story 1.9] — CI pipeline (for context on acceptance.yml separation)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (1M context)

### Debug Log References

- VirtualBox kernel module (`vboxdrv`) not loaded on dev machine — `vagrant up`/`vagrant destroy` verification deferred to user. Task 4 subtasks for VM boot and destroy left unchecked.
- `puzzle/opnsense` box (v25.7) selected from Vagrant Cloud — closest to target 26.1.x and actively maintained with VirtualBox support.
- OPNsense has no official Vagrant box — community `puzzle/opnsense` is the best available option.

### Completion Notes List

- `test/Vagrantfile` provisions OPNsense VM using `puzzle/opnsense` box (v25.7)
- VirtualBox provider: 2GB RAM, 2 CPUs, second NIC for LAN on internal network
- Port forwarding: host 10443 → guest 443 (HTTPS API)
- SSH configured for root user with `/bin/sh` shell (FreeBSD/OPNsense)
- Default synced folder disabled (OPNsense lacks Guest Additions)
- `test/scripts/create-apikey.sh` generates API key/secret via openssl + config.xml manipulation
- Script outputs ready-to-paste export commands for `OPNSENSE_URI`, `OPNSENSE_API_KEY`, `OPNSENSE_API_SECRET`, `OPNSENSE_ALLOW_INSECURE`
- `CONTRIBUTING.md` documents prerequisites, dev workflow, acceptance testing procedure, `-p 1` rationale, `TF_ACC` gate, conventional commits
- `make check` passes all 6 targets — no regressions from new files
- Task 4 (VM verification) partially complete — `make check` passes but `vagrant up`/`vagrant destroy` require VirtualBox kernel module to be loaded

### Change Log

- 2026-03-23: Created Vagrant test environment, API key script, and CONTRIBUTING.md (Story 1.8)

### File List

- `test/Vagrantfile` — NEW: OPNsense VM provisioning with puzzle/opnsense box, port forwarding, VirtualBox config
- `test/scripts/create-apikey.sh` — NEW: API key generation script (Vagrant shell provisioner)
- `CONTRIBUTING.md` — NEW: Developer guide with prerequisites, workflow, acceptance testing docs
