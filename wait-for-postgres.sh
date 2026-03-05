#!/bin/sh
set -e

host="$1"
shift
cmd="$@"

until pg_isready -h "$host" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}"; do
  echo "Waiting for postgres at $host..."
  sleep 2
done

exec $cmd
