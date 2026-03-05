#!/bin/bash
# Auth Service Migration Runner
# This script runs all migrations in order for the auth-service

set -e

DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres123}"
DB_NAME="${DB_NAME:-drandme}"

export PGPASSWORD="$DB_PASSWORD"

echo "Running auth-service migrations..."
echo "Database: $DB_NAME@$DB_HOST:$DB_PORT"

# Wait for database to be ready
until psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; do
  echo "Waiting for database to be ready..."
  sleep 1
done

# Run migrations in order
for migration in $(ls -1 *.sql | sort); do
  echo "Running migration: $migration"
  psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$migration"
done

echo "Auth service migrations completed successfully!"

