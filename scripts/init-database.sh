#!/bin/bash

# Database initialization script for Healthcare SaaS Security System
# This script ensures all migrations are applied in the correct order

set -e

echo "🗄️  Initializing Healthcare SaaS Database with Security Features"
echo "================================================================"

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h postgres -p 5432 -U postgres; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 2
done

echo "✅ PostgreSQL is ready!"

# Connect to database and run migrations
echo "Running database migrations..."

# Run migrations in order
psql -h postgres -U postgres -d drandme -f /docker-entrypoint-initdb.d/001_initial_schema.sql
echo "✅ Initial schema created"

psql -h postgres -U postgres -d drandme -f /docker-entrypoint-initdb.d/002_add_mo_id_to_patients.sql
echo "✅ MO ID added to patients"

psql -h postgres -U postgres -d drandme -f /docker-entrypoint-initdb.d/003_admin_features.sql
echo "✅ Admin features added"

psql -h postgres -U postgres -d drandme -f /docker-entrypoint-initdb.d/004_add_appointment_fields.sql
echo "✅ Appointment fields added"

psql -h postgres -U postgres -d drandme -f /docker-entrypoint-initdb.d/005_security_features.sql
echo "✅ Security features added"

# Verify security tables were created
echo "Verifying security tables..."
psql -h postgres -U postgres -d drandme -c "
SELECT 
    table_name,
    CASE 
        WHEN table_name IN ('failed_login_attempts', 'account_lockouts', 'blocked_ips', 'security_audit_log') 
        THEN '✅ Security Table'
        ELSE '📋 Regular Table'
    END as table_type
FROM information_schema.tables 
WHERE table_schema = 'public' 
ORDER BY table_name;
"

echo ""
echo "🔐 Security Tables Created:"
echo "  - failed_login_attempts: Tracks failed login attempts"
echo "  - account_lockouts: Manages account lockout status"
echo "  - blocked_ips: IP address blocking"
echo "  - security_audit_log: Comprehensive audit logging"

echo ""
echo "🎉 Database initialization completed successfully!"
echo "Your healthcare SaaS security system is ready!"
