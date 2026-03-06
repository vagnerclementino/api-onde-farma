#!/usr/bin/env sh
set -eu

if command -v psql >/dev/null 2>&1; then
  exec psql "$@"
fi

exec docker compose exec -T postgres psql "$@"
