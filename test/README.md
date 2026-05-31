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

## Run the tests

```sh
export OPNSENSE_API_KEY=...      # from `vagrant up` output
export OPNSENSE_API_SECRET=...
export OPNSENSE_ALLOW_INSECURE=true
TF_ACC=1 go test -p 1 ./internal/service/...
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
