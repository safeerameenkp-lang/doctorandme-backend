#!/bin/bash

# Database migration script
# This script runs the initial database schema

echo "Starting database migration..."

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h localhost -p 5432 -U postgres; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 2
done

echo "PostgreSQL is ready - running migrations"

# Run the migration
psql -h localhost -p 5432 -U postgres -d clinic_management -f migrations/001_initial_schema.sql

echo "Migration completed successfully!"
