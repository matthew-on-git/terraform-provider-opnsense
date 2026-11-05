#!/usr/bin/env bash
# Purpose: Import an existing Trust CA UUID without changing appliance PKI.
# Usage: import.sh
set -euo pipefail
terraform import opnsense_trust_ca.example 00000000-0000-4000-8000-000000000001
