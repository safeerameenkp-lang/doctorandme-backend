#!/bin/sh
echo "Waiting for database..."
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER"; do
  echo "Postgres not ready, sleeping 2s..."
  sleep 2
done

echo "Database ready. Running appointment migrations..."
export PGPASSWORD="$DB_PASSWORD"

for f in /migrations/*.sql; do
  if [ -f "$f" ]; then
    echo "Running $f"
    psql "host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME sslmode=$DB_SSLMODE sslrootcert=$DB_SSLROOTCERT" -f "$f"
  fi
done

echo "Appointment migrations completed!"
