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
  ghcr.io/devrail-dev/dev-toolchain:1.12.0 \
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
  ghcr.io/devrail-dev/dev-toolchain:1.12.0 \
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
