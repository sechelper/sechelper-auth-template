#!/usr/bin/env bash
set -euo pipefail

check_no_matches() {
  local message="$1"
  shift

  if rg -n "$@"; then
    printf '%s\n' "$message" >&2
    exit 1
  else
    local status=$?
    if [[ "$status" -ne 1 ]]; then
      printf 'boundary scan failed (rg exit %s): %s\n' "$status" "$message" >&2
      exit "$status"
    fi
  fi
}

required_paths=(
  api/internal/platform
  api/internal/modules/authentication
  api/internal/modules/authorization
  api/internal/modules/manifest
  api/internal/modules/audit
  api/internal/modules/account
  api/internal/modules/operations
  api/internal/business
  api/cmd/server
  web/src/framework/app
  web/admin/src/app
)
for path in "${required_paths[@]}"; do
  if [[ ! -d "$path" ]]; then
    printf 'required boundary scan path does not exist: %s\n' "$path" >&2
    exit 2
  fi
done

check_no_matches "framework-to-business dependency detected" \
  'internal/business|examples/' \
  api/internal/platform api/internal/modules/authentication \
  api/internal/modules/authorization api/internal/modules/manifest \
  api/internal/modules/audit api/internal/modules/account api/internal/modules/operations

check_no_matches "business code bypasses framework infrastructure" \
  'os\.Getenv|viper\.|sql\.Open|fmt\.Print|log\.(Print|Fatal)' \
  api/internal/business

check_no_matches "example business leaked into the default application" \
  -g '!business_runtime_example.go' -g '!*.test.*' \
  'orders|order:read|example-order' \
  api/cmd/server web/src/framework/app web/admin/src/app
