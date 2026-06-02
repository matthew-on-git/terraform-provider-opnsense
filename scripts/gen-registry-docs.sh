#!/usr/bin/env bash
# Generate the Terraform Registry documentation (docs/) with tfplugindocs.
#
# tfplugindocs introspects the provider schema by running a Terraform CLI. The
# vendored CLI at test/bin/terraform is put on PATH so tfplugindocs does not try
# to download one (recent hc-install releases fail on an expired HashiCorp GPG
# signing key). Idempotent: re-run any time templates/, schemas, or examples
# change, then commit docs/.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

if [ ! -x "test/bin/terraform" ]; then
  echo "error: vendored terraform not found at test/bin/terraform" >&2
  echo "       install it (see test/README.md) before generating docs." >&2
  exit 1
fi

export PATH="$repo_root/test/bin:$PATH"
exec go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.20.0 \
  generate --provider-name opnsense --rendered-provider-name OPNsense
