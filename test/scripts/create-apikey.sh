#!/bin/sh
# create-apikey.sh — Generate OPNsense API key/secret for acceptance testing.
#
# This script runs as a Vagrant provisioner inside the OPNsense VM.
# It creates an API key for the root user by manipulating config.xml
# and outputs the credentials as shell export commands.

set -e

CONFIG="/conf/config.xml"
API_KEY_DIR="/var/db/api_keys"

# Generate a random API key and secret.
API_KEY=$(openssl rand -hex 20)
API_SECRET=$(openssl rand -hex 40)

# Hash the secret for storage in config (OPNsense uses SHA-512).
API_SECRET_HASH=$(echo -n "$API_SECRET" | openssl dgst -sha512 | awk '{print $NF}')

# Create the API key directory if it doesn't exist.
mkdir -p "$API_KEY_DIR"

# Write the key file in OPNsense's expected format.
# The filename is the key, the content is the hashed secret.
echo "$API_SECRET_HASH" > "${API_KEY_DIR}/${API_KEY}"

# Add the API key reference to the root user in config.xml.
# OPNsense stores API keys under <system><user><apikeys><item>.
# Use sed to inject the key into the root user's config block.
if grep -q "<apikeys>" "$CONFIG"; then
  # apikeys section exists — add our key.
  sed -i '' "/<apikeys>/a\\
<item><key>${API_KEY}</key></item>" "$CONFIG" 2>/dev/null || \
  sed -i "/<apikeys>/a <item><key>${API_KEY}</key></item>" "$CONFIG"
else
  # No apikeys section — create it inside the root user block.
  sed -i '' "/<\/user>/i\\
<apikeys><item><key>${API_KEY}</key></item></apikeys>" "$CONFIG" 2>/dev/null || \
  sed -i "/<\/user>/i <apikeys><item><key>${API_KEY}</key></item></apikeys>" "$CONFIG"
fi

# Reload the config to pick up the new key.
configctl auth restart >/dev/null 2>&1 || true

echo ""
echo "============================================"
echo "  OPNsense Test Environment Ready"
echo "============================================"
echo ""
echo "  Copy and paste these commands:"
echo ""
echo "  export OPNSENSE_URI=https://localhost:10443"
echo "  export OPNSENSE_API_KEY=${API_KEY}"
echo "  export OPNSENSE_API_SECRET=${API_SECRET}"
echo "  export OPNSENSE_ALLOW_INSECURE=true"
echo ""
echo "  Then run acceptance tests:"
echo ""
echo "  TF_ACC=1 go test -p 1 ./..."
echo ""
echo "============================================"
