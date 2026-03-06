#!/usr/bin/env sh
set -eu

: "${DATABASE_URL:?DATABASE_URL must be set}"

sh ./scripts/psql.sh "$DATABASE_URL" -v ON_ERROR_STOP=1 -f scripts/import_csv.sql
