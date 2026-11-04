#!/usr/bin/env bash
# Purpose: Import an existing Trust certificate UUID without generating a key.
# Usage: import.sh
set -euo pipefail
terraform import opnsense_trust_cert.example 00000000-0000-4000-8000-000000000002
