#!/bin/sh
# wait-for-postgres.sh

set -e

>&2 echo "sleeping for 5 seconds"

sleep 5

## the container doesn't have psql so can't use this nice script...
# until PGPASSWORD=$POSTGRES_PASSWORD psql -h "$LANTERN_DBHOST" -U "$POSTGRES_USER" -c '\q'; do
#   >&2 echo "Postgres is unavailable - sleeping"
#   sleep 1
# done

>&2 echo "executing startup command"

exec /prometheus-postgresql-adapter $@