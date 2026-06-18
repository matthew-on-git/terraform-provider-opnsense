# Project Context

## Validation Rules

- Run `make check` before considering implementation, review patching, or story work complete.
- For Terraform provider resource changes, run targeted package tests first, then run focused Vagrant acceptance when the local appliance/API is available.
- Prefer the project acceptance runner for live OPNsense validation:

```sh
OPNSENSE_URI=https://localhost:10445 \
OPNSENSE_API_KEY=... \
OPNSENSE_API_SECRET=... \
OPNSENSE_ALLOW_INSECURE=true \
TEST_TIMEOUT=40m \
test/scripts/run-vagrant-acceptance.sh --package ./internal/service/<package>
```

- Acceptance tests must run serially because many resources mutate global appliance configuration and service reload state.
- Do not destroy, reset, or reprovision the Vagrant appliance during review unless the user explicitly approves it.
- Do not commit or document live API secrets. Use environment variables from the local shell or ask the user when credentials are unavailable.
- If host `go generate ./tools` tries to download Terraform and fails with an OpenPGP key error, run the generator in the dev-toolchain container instead.

## Code Review Expectations

- When code review covers provider resource behavior, include Vagrant-backed acceptance evidence when feasible, not only unit tests.
- Record acceptance result directories or command summaries in the story/review notes when they are part of completion evidence.
- Treat missing endpoint behavior as an explicit skip/precheck concern, not an obscure acceptance failure.
