#!/usr/bin/env sh
set -eu

: "${DATABASE_URL:?DATABASE_URL must be set}"

for file in db/migrations/*.up.sql; do
  [ -e "$file" ] || continue
  sh ./scripts/psql.sh "$DATABASE_URL" -v ON_ERROR_STOP=1 < "$file"
done
