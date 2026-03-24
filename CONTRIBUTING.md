# Contributing to terraform-provider-opnsense

## Prerequisites

- **Go 1.25+** — for building and testing the provider
- **Docker** — for running DevRail linting/formatting checks (`make check`)
- **Vagrant + VirtualBox** — for running acceptance tests against a real OPNsense instance

## Development Workflow

### Running Checks

```bash
make check    # Run all linting, formatting, tests, security, scan, docs
go test ./... # Run unit tests only (no OPNsense instance needed)
go build ./...# Verify the provider builds
```

All linting, formatting, and security tools run inside the DevRail container.
You do not need to install them locally.

### Project Structure

```
pkg/opnsense/       # API client library (independent of Terraform)
internal/provider/  # Terraform provider implementation
test/               # Vagrant test environment
```

## Acceptance Testing

Acceptance tests run against a real OPNsense instance. A Vagrantfile is
provided to spin up an ephemeral VM for local testing.

### 1. Start the OPNsense VM

```bash
cd test
vagrant up
```

After provisioning completes, the output displays export commands with the
API credentials:

```
export OPNSENSE_URI=https://localhost:10443
export OPNSENSE_API_KEY=<generated-key>
export OPNSENSE_API_SECRET=<generated-secret>
export OPNSENSE_ALLOW_INSECURE=true
```

Copy and paste these into your shell.

### 2. Run Acceptance Tests

```bash
TF_ACC=1 go test -p 1 ./...
```

### 3. Destroy the VM

```bash
cd test
vagrant destroy -f
```

### Why `-p 1`?

Acceptance tests run serially (`-p 1`) because OPNsense serializes all
mutation operations through a global mutex that protects `config.xml`
integrity. All tests share one OPNsense instance, so concurrent writes
would risk XML corruption. The `-p 1` flag mirrors this production
constraint.

### Why `TF_ACC=1`?

The `TF_ACC` environment variable gates acceptance tests. Without it, only
unit tests run. This prevents accidental execution of tests that require a
live OPNsense instance.

## Commit Messages

This project uses [conventional commits](https://www.conventionalcommits.org/):

```
type(scope): description

feat(provider): add insecure TLS option
fix(haproxy): correct server weight validation
test(firewall): add alias resource acceptance tests
docs(readme): update provider configuration example
```
