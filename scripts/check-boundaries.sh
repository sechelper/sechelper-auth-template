#!/usr/bin/env bash
set -euo pipefail
if rg -n 'internal/(business|modules/orders)|examples/' api/internal/platform api/internal/modules/authentication api/internal/modules/authorization api/internal/modules/manifest api/internal/modules/audit api/internal/modules/account api/internal/modules/operations 2>/dev/null; then
  echo "framework-to-business dependency detected" >&2; exit 1
fi
if rg -n 'os\.Getenv|viper\.|sql\.Open|fmt\.Print|log\.(Print|Fatal)' api/internal/business 2>/dev/null; then
  echo "business code bypasses framework infrastructure" >&2; exit 1
fi
if rg -n 'orders|order:read|example-order' api/cmd/server web/admin/src/app web/admin/src/navigation 2>/dev/null; then
  echo "example business leaked into the default application" >&2; exit 1
fi
