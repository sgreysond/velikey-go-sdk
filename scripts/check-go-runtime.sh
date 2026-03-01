#!/usr/bin/env bash
set -euo pipefail

MIN_VERSION="${GO_SDK_MIN_RUNTIME_VERSION:-1.25.7}"

log() {
  printf '[go-runtime-check] %s\n' "$*"
}

strip_prefix() {
  local value="$1"
  value="${value#go}"
  printf '%s' "$value"
}

normalize_version() {
  local value
  value="$(strip_prefix "$1")"
  IFS='.' read -r major minor patch <<<"$value"
  major="${major:-0}"
  minor="${minor:-0}"
  patch="${patch:-0}"
  printf '%s.%s.%s' "$major" "$minor" "$patch"
}

version_ge() {
  local left right
  left="$(normalize_version "$1")"
  right="$(normalize_version "$2")"

  IFS='.' read -r l1 l2 l3 <<<"$left"
  IFS='.' read -r r1 r2 r3 <<<"$right"

  if (( l1 > r1 )); then return 0; fi
  if (( l1 < r1 )); then return 1; fi
  if (( l2 > r2 )); then return 0; fi
  if (( l2 < r2 )); then return 1; fi
  if (( l3 >= r3 )); then return 0; fi
  return 1
}

if ! command -v go >/dev/null 2>&1; then
  log "FAIL: go command not found"
  exit 1
fi

GO_VERSION_RAW="$(go version)"
GO_VERSION_FIELD="$(printf '%s' "$GO_VERSION_RAW" | awk '{print $3}')"
GO_VERSION="$(normalize_version "$GO_VERSION_FIELD")"
MIN_NORMALIZED="$(normalize_version "$MIN_VERSION")"

if version_ge "$GO_VERSION" "$MIN_NORMALIZED"; then
  log "PASS: current Go runtime $GO_VERSION >= required $MIN_NORMALIZED"
  exit 0
fi

log "FAIL: current Go runtime $GO_VERSION is below required $MIN_NORMALIZED"
exit 1
