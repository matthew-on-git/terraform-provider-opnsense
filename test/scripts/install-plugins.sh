#!/bin/sh
# install-plugins.sh — Install OPNsense plugins required for acceptance testing.
#
# This script runs as a Vagrant provisioner inside the OPNsense VM.
# It installs all plugins that the provider resources depend on.

set -e

echo ""
echo "============================================"
echo "  Installing OPNsense Plugins"
echo "============================================"
echo ""

# Fix networking — VirtualBox NAT adapter needs a default route and DNS.
# The OPNsense base box doesn't configure these automatically.
route add default 10.0.2.2 >/dev/null 2>&1 || true
echo "nameserver 8.8.8.8" > /etc/resolv.conf

# Update package repository.
echo "Updating package repository..."
pkg update -f >/dev/null 2>&1 || true

# List of plugins required by the provider's resources.
PLUGINS="
  os-haproxy
  os-frr
  os-acme-client
  os-ddclient
"

for plugin in $PLUGINS; do
  printf "Installing %s... " "$plugin"
  if pkg install -y "$plugin" >/dev/null 2>&1; then
    echo "OK"
  else
    echo "FAILED (may not be available for this OPNsense version)"
  fi
done

# Restart configd to pick up new plugin models.
echo ""
echo "Restarting configd..."
service configd restart >/dev/null 2>&1 || true

# Wait for configd to be ready.
sleep 3

echo ""
echo "============================================"
echo "  Plugin Installation Complete"
echo "============================================"
echo ""
