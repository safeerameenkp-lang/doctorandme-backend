#!/bin/sh
export PGPASSWORD="$DB_PASSWORD"
export PGSSLMODE="${DB_SSLMODE:-require}"
if [ -n "$DB_SSLROOTCERT" ]; then
  export PGSSLROOTCERT="$DB_SSLROOTCERT"
fi

echo "Waiting for database..."
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"; do
  echo "Postgres not ready, sleeping 2s..."
  sleep 2
done

echo "Database ready. Running organization migrations..."

for f in /migrations/*.sql; do
  if [ -f "$f" ]; then
    echo "Running $f"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$f"
  fi
done

echo "Organization migrations completed!"
