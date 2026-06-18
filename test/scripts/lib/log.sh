#!/usr/bin/env sh

log_info() {
  printf '[INFO] %s\n' "$*"
}

log_warn() {
  printf '[WARN] %s\n' "$*" >&2
}

log_error() {
  printf '[ERROR] %s\n' "$*" >&2
}

log_debug() {
  if [ "${DEBUG:-}" = "1" ]; then
    printf '[DEBUG] %s\n' "$*" >&2
  fi
}

die() {
  log_error "$*"
  exit 1
}
