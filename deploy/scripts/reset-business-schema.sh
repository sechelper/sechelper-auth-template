#!/usr/bin/env bash
set -euo pipefail

: "${BUSINESS_DATABASE_URL:?BUSINESS_DATABASE_URL is required}"
: "${BUSINESS_MODULE:?BUSINESS_MODULE is required}"

if [[ ! "$BUSINESS_MODULE" =~ ^[a-z][a-z0-9_]*$ ]]; then
  printf 'BUSINESS_MODULE must be a lowercase module identifier\n' >&2
  exit 2
fi

schema="business_${BUSINESS_MODULE}"
case "$schema" in
  business_configuration|business_framework|business_public|configuration|framework|public)
    printf 'refusing to reset reserved schema: %s\n' "$schema" >&2
    exit 2
    ;;
esac

command -v psql >/dev/null 2>&1 || {
  printf 'psql is required\n' >&2
  exit 2
}

if ! psql "$BUSINESS_DATABASE_URL" --no-password --quiet --tuples-only --no-align --command "SELECT 1 FROM pg_namespace WHERE nspname = 'framework'" | tr -d '[:space:]' | rg -q '^1$'; then
  printf 'refusing reset: target is not an initialized application database with framework schema\n' >&2
  exit 2
fi

if psql "$BUSINESS_DATABASE_URL" --no-password --quiet --tuples-only --no-align --command "SELECT 1 FROM pg_namespace WHERE nspname IN ('configuration', 'framework', 'public') AND nspname = '$schema'" | tr -d '[:space:]' | rg -q '^1$'; then
  printf 'refusing reset of protected schema: %s\n' "$schema" >&2
  exit 2
fi

psql "$BUSINESS_DATABASE_URL" --no-password --set=ON_ERROR_STOP=1 --command "DROP SCHEMA IF EXISTS \"$schema\" CASCADE; CREATE SCHEMA \"$schema\";"
printf 'reset business schema %s\n' "$schema"
