# Acceptance test environment

A Vagrant-provisioned OPNsense VM for running the provider's `TF_ACC` acceptance
tests locally.

## Prerequisites

- VirtualBox + Vagrant.
- **Terraform (or OpenTofu) CLI on `PATH`.** The test framework will try to
  auto-download Terraform otherwise, which can fail (`openpgp: key expired`).
  Install `terraform`/`tofu`, or set `TF_ACC_TERRAFORM_PATH=/path/to/terraform`.
- If `vagrant` fails with a `bcrypt_pbkdf` gem error (e.g. when a version
  manager like mise shadows Vagrant's Ruby), invoke it with Vagrant's embedded
  Ruby:
  ```sh
  env -u GEM_HOME -u GEM_PATH -u RUBYLIB -u RUBYOPT \
    PATH="/opt/vagrant/embedded/bin:/usr/bin:/bin" vagrant <cmd>
  ```
  If your shell defines the `vg` alias, use that instead:
  ```sh
  alias vg='env -u GEM_HOME -u GEM_PATH -u RUBYLIB -u RUBYOPT PATH="/opt/vagrant/embedded/bin:/usr/bin:/bin" vagrant'
  vg status
  ```

## Bring up the VM

```sh
cd test
vagrant up        # boots OPNsense, installs plugins, prints export commands
```

`create-apikey.sh` provisions a root API key (key + bcrypt-hashed secret written
via OPNsense's config framework — see OPNsense/Auth/API.php `password_verify`)
and prints the `export OPNSENSE_*` lines.

## Connectivity

The Vagrantfile forwards guest `443` → host `10443`, but **TLS over the
VirtualBox NAT port-forward is unreliable** (handshake hangs). The robust path
is an SSH tunnel straight to the VM's API:

```sh
# host:10444 -> VM 127.0.0.1:443  (run in its own shell; keep it open)
sshpass -p opnsense ssh -p 2222 \
  -o PreferredAuthentications=password -o PubkeyAuthentication=no \
  -N -L 10444:127.0.0.1:443 root@127.0.0.1

export OPNSENSE_URI=https://localhost:10444   # override the printed :10443
```

If `localhost:10443` resets TLS during `curl` or provider credential validation,
keep the VM running and use the SSH tunnel path instead. This was the confirmed
working method for HAProxy acceptance testing on 2026-06-10.

## Run the tests

```sh
export OPNSENSE_API_KEY=...      # from `vagrant up` output
export OPNSENSE_API_SECRET=...
export OPNSENSE_ALLOW_INSECURE=true
TF_ACC=1 go test -p 1 ./internal/service/...
```

For whole-provider Vagrant validation, prefer the Docker-backed runner. It runs
packages serially, keeps the API tunnel reachable through host networking, and
writes a per-package summary under `test/acceptance-results/`:

```sh
export OPNSENSE_URI=https://localhost:10444
export OPNSENSE_API_KEY=...
export OPNSENSE_API_SECRET=...
export OPNSENSE_ALLOW_INSECURE=true

test/scripts/run-vagrant-acceptance.sh
```

To run a smaller slice while debugging:

```sh
test/scripts/run-vagrant-acceptance.sh --package ./internal/service/firewall
OPNSENSE_ACCEPTANCE_PACKAGES='./internal/service/iface ./internal/service/haproxy' \
  test/scripts/run-vagrant-acceptance.sh
```

The full suite is intentionally serial (`-p 1`) because many resources mutate
global appliance configuration and call service reconfigure endpoints. A few
tests are optional and self-skip unless extra environment variables are set, for
example live ACME issuance (`OPNSENSE_ACME_ISSUE=1` plus ACME UUID/domain envs)
and HAProxy certificate binding (`OPNSENSE_HAPROXY_CERT_REFID`).

## Firewall and NAT parity tests

Story 31.0a tests are separately opt-in and must target only a disposable
OPNsense 26.7.7 or newer appliance. Filter and DNAT objects are disabled. The
suite proves advanced read-back, import/no-change, alias and filter drift
repair, multi-interface DNAT membership update, data-source parity, and exact
test-object cleanup.

```sh
export OPNSENSE_FIREWALL_PARITY=1
export OPNSENSE_FIREWALL_TEST_INTERFACE=lan
export OPNSENSE_FIREWALL_TEST_INTERFACES=wan,lan

TEST_REGEX='TestAccFirewall(Alias|FilterRule|NatPortForward)_advancedParity' \
  test/scripts/run-vagrant-acceptance.sh --package ./internal/service/firewall
```

Do not point these tests at production or management objects.

## Kea field and surface parity tests

Story 31.0b tests are separately opt-in and require a disposable OPNsense
26.7.7 or newer appliance plus an isolated interface. The preflight refuses to
run unless Kea DHCPv4, DHCPv6, and both HA configurations are disabled. The
suite proves advanced v4/v6 subnet read-back, typed DHCPv6 peer and PD-pool
lifecycle/import/update/drift repair, dependency-ordered cleanup, data-source
parity, stable plans, and four exact zero-match cleanup searches.

```sh
export OPNSENSE_KEA_PARITY=1
export OPNSENSE_KEA_TEST_INTERFACE=opt3

TEST_REGEX='TestAcc(Kea.*Parity|DHCPv6(Peer|PDPool)_basic)' \
  test/scripts/run-vagrant-acceptance.sh --package ./internal/service/kea
```

Keep all four Kea service/HA flags (DHCPv4, DHCPv4 HA, DHCPv6, and DHCPv6 HA)
disabled throughout the run. Never point this test
at a production appliance or a management interface.

## OpenVPN field and key parity tests

Story 31.0c uses a dedicated digest-bound runner against exactly OPNsense
26.7.7. It creates only disabled, collision-tagged instances and client
overwrites and delegates the focused Go tests to the existing DevRail/Vagrant
acceptance container. The runner requires an external executable that queries
the hypervisor/inventory and prints independently observed `name=` and `uuid=`
lines; a caller-supplied disposable flag alone is not accepted.

Set the normal API variables, the expected VM name/UUID, the absolute identity
probe path, the output of `test/scripts/story-31-0c-code-digest.sh`, and valid
existing reference IDs for the CA, certificate, CRL, static key, local group,
and CARP VIP. Then run:

```sh
test/scripts/run-story-31-0c-acceptance.sh
```

The runner requires exact non-skipped PASS lines for advanced CRUD/import,
static-key lookup, and the Terraform CLI ephemeral protocol test. It parses two
marker-bound zero-count cleanup lines, scans temporary logs for key/secret
material, verifies the candidate digest before and after, and removes the
temporary artifacts. Do not run it against a production appliance.

## Interface assignment/IP tests

The interface-assignment lifecycle test is opt-in because it changes addressing.
It requires OPNsense 26.7.6 or newer and refuses `wan` or `lan`; provide an
existing disposable `optN` plus two unused test addresses on its isolated
network. The test captures the complete assignment before the first apply and
restores it after Terraform's state-only destroy.

```sh
export OPNSENSE_INTERFACE_ASSIGNMENT=opt3
export OPNSENSE_INTERFACE_ASSIGNMENT_IPV4_ADDRESS=192.0.2.10
export OPNSENSE_INTERFACE_ASSIGNMENT_UPDATED_IPV4_ADDRESS=192.0.2.11
export OPNSENSE_INTERFACE_ASSIGNMENT_IPV4_PREFIX_LENGTH=24

TEST_REGEX=TestAccInterfaceAssignment \
  test/scripts/run-vagrant-acceptance.sh --package ./internal/service/iface
```

Never point these variables at a management interface or a production subnet.
The bundled Vagrant box is currently OPNsense 25.7 and cannot run this test.

## Interface LAGG tests

The Vagrant VM includes two dedicated, isolated VirtualBox NICs on
`opnsense-lagg-test` for LAGG member validation. They are intentionally placed
on later adapters, which normally appear as `em4` and `em5`, because earlier
adapters may already be assigned as WAN/OPT interfaces by the base box. Keep
these NICs separate from the WAN/LAN management path so creating or deleting a
LAGG does not break API access.

Before running LAGG acceptance tests, verify the extra NICs appear as selectable
members in the appliance. Keep the SSH tunnel from the connectivity section open
if direct `:10443` TLS is unreliable, then run the focused test serially:

```sh
docker run --rm --network host \
  -v "$(pwd)/..:/workspace" \
  -w /workspace \
  -e TF_ACC=1 \
  -e OPNSENSE_URI=https://localhost:10444 \
  -e OPNSENSE_API_KEY \
  -e OPNSENSE_API_SECRET \
  -e OPNSENSE_ALLOW_INSECURE=true \
  ghcr.io/devrail-dev/dev-toolchain:1.12.24 \
  go test -run TestAccLagg_basic -count=1 -p 1 ./internal/service/iface
```

When host-side Terraform auto-install fails with `openpgp: key expired`, run the
focused acceptance test in the dev-toolchain container and reach the tunnel with
host networking:

```sh
docker run --rm --network host \
  -v "$(pwd)/..:/workspace" \
  -w /workspace \
  -e TF_ACC=1 \
  -e OPNSENSE_URI=https://localhost:10444 \
  -e OPNSENSE_API_KEY \
  -e OPNSENSE_API_SECRET \
  -e OPNSENSE_ALLOW_INSECURE=true \
  ghcr.io/devrail-dev/dev-toolchain:1.12.24 \
  go test -run TestAccLagg_basic -count=1 -p 1 ./internal/service/iface
```

## Tear down

```sh
vagrant destroy -f
```

## Debugging the VM

`vagrant ssh -c` does **not** work (the box uses password-only SSH). Use:

```sh
sshpass -p opnsense ssh -p 2222 \
  -o PreferredAuthentications=password -o PubkeyAuthentication=no root@127.0.0.1
```

The root shell is `tcsh`; pipe a script to `/bin/sh` for POSIX commands:
`... root@127.0.0.1 /bin/sh < script.sh`.

### Advanced FRR candidate target bindings

The digest-bound Story 31.0h runner requires a newly reserved isolated guest and
independent identity, TLS, oracle and IPv6-readiness probes. Historical guest IDs
and versions are not defaults. Supply `OPNSENSE_FRR_EXPECTED_CORE`,
`OPNSENSE_FRR_EXPECTED_PLUGIN`, and `OPNSENSE_FRR_EXPECTED_DAEMON` from independent
installed-version observations, along with the existing disposable VM name/UUID,
identity/guest probe, loopback API/oracle and expected candidate digest bindings.
Compare installed model/controller/template bytes to the released source evidence
before acceptance; a numeric plugin revision is eligibility, not appliance proof.
The matrix adds an isolated IPv6 advanced neighbor, sensitive config-owned input,
password-free lookup, source enums, strict listen ranges, import/no-change and
advanced drift repair with generated AF/transport observations. It retains the
collector-positive-prefix, exact restoration and zero-leftover gates. Legacy
appliance/OpenTofu rows are not synthesized by these tests. Raw artifacts remain
private; only sanitized versions/digests/counts belong in published evidence.
