# terraform-provider-opnsense

Terraform provider for managing OPNsense appliances through the OPNsense REST/MVC API.

<!-- badges-start -->
[![DevRail compliant](https://devrail.dev/images/badge.svg)](https://devrail.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
<!-- badges-end -->

## Quick Start

```hcl
terraform {
  required_providers {
    opnsense = {
      source = "matthew-on-git/opnsense"
    }
  }
}

provider "opnsense" {
  uri        = "https://opnsense.example.com"
  api_key    = var.opnsense_api_key
  api_secret = var.opnsense_api_secret
}
```

Credentials can also be supplied with `OPNSENSE_URI`, `OPNSENSE_API_KEY`, `OPNSENSE_API_SECRET`, and `OPNSENSE_ALLOW_INSECURE`.

## Support

The current implementation supports 97 resources and 83 data sources across firewall, HAProxy, ACME, DNS/Dnsmasq, DHCP/Kea, VPN, FRR/Quagga, interfaces, auth, trust, syslog, Monit, cron, and traffic shaping.

See [`_bmad-output/planning-artifacts/support-matrix.md`](_bmad-output/planning-artifacts/support-matrix.md) for the current Supported / Coming / Needs research / Upstream-blocked product matrix. Confirmed upstream blockers are summarized in [`docs/upstream-blocked.md`](docs/upstream-blocked.md).

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

## Documentation

Registry documentation is generated with `tfplugindocs` into `docs/`. The provider index documents authentication, minimum OPNsense version, permissions, support status, and migration/import guidance.

For brownfield appliance migration, see [`docs/migration-import.md`](docs/migration-import.md) for dependency-ordered import examples and no-change plan troubleshooting.

## Contributing

See [DEVELOPMENT.md](DEVELOPMENT.md) for development standards, coding conventions, and contribution guidelines.

This project follows [Conventional Commits](https://www.conventionalcommits.org/). All commits use the `type(scope): description` format.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
