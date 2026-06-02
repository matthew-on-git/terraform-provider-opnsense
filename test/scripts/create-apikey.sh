#!/bin/sh
# create-apikey.sh — Generate an OPNsense API key/secret for acceptance testing.
#
# Runs as a Vagrant provisioner inside the OPNsense VM. OPNsense stores API
# credentials under <system><user><apikeys><item> in config.xml, with the
# secret hashed via PHP password_hash() (verified with password_verify() — see
# OPNsense/Auth/API.php). We therefore:
#   1. generate a random key + secret,
#   2. bcrypt-hash the secret,
#   3. write the apikeys item via OPNsense's own config framework (idempotent —
#      replaces any existing keys for root), and
#   4. reload the GUI/API so the new key is active.

set -e

KEY=$(openssl rand -hex 20)
SECRET=$(openssl rand -hex 32)
HASH=$(php -r 'echo password_hash($argv[1], PASSWORD_DEFAULT);' "$SECRET")

# Inject the apikey into config.xml using the OPNsense config framework so the
# structure and revision metadata are correct. Idempotent: clears prior keys.
# Key and hash are passed via argv (NOT interpolated) — the bcrypt hash contains
# '$' sequences that PHP would otherwise treat as variables.
php -r '
require_once("config.inc");
require_once("util.inc");
global $config;
foreach ($config["system"]["user"] as &$u) {
    if ($u["name"] == "root") {
        unset($u["apikeys"]);
        $u["apikeys"] = array("item" => array(array("key" => $argv[1], "secret" => $argv[2])));
    }
}
write_config("acceptance test api key");
' "$KEY" "$HASH"

# Reload so the new key is picked up by the API.
configctl webgui restart >/dev/null 2>&1 || true

echo ""
echo "============================================"
echo "  OPNsense Test Environment Ready"
echo "============================================"
echo ""
echo "  Copy and paste these commands:"
echo ""
echo "  export OPNSENSE_URI=https://localhost:10443"
echo "  export OPNSENSE_API_KEY=${KEY}"
echo "  export OPNSENSE_API_SECRET=${SECRET}"
echo "  export OPNSENSE_ALLOW_INSECURE=true"
echo ""
echo "  Then run acceptance tests:"
echo ""
echo "  TF_ACC=1 go test -p 1 ./..."
echo ""
echo "============================================"
