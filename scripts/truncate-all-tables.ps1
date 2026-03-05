# PowerShell script to truncate all tables and delete all data

Write-Host "=== TRUNCATING ALL DATABASE TABLES ===" -ForegroundColor Red
Write-Host "WARNING: This will delete ALL data from ALL tables!" -ForegroundColor Yellow
Write-Host ""

$confirmation = Read-Host "Are you sure you want to continue? Type 'YES' to confirm"

if ($confirmation -ne "YES") {
    Write-Host "Operation cancelled." -ForegroundColor Yellow
    exit
}

# Database connection details
$DB_USER = "postgres"
$DB_NAME = "drandme"

Write-Host "`nTruncating all tables..." -ForegroundColor Cyan

# Truncate SQL
$truncateSQL = @"
-- Disable triggers and constraints temporarily
SET session_replication_role = 'replica';

-- Truncate all tables
TRUNCATE TABLE refresh_tokens CASCADE;
TRUNCATE TABLE user_roles CASCADE;
TRUNCATE TABLE appointments CASCADE;
TRUNCATE TABLE patient_vitals CASCADE;
TRUNCATE TABLE doctor_leaves CASCADE;
TRUNCATE TABLE doctor_time_slots CASCADE;
TRUNCATE TABLE clinic_doctor_links CASCADE;
TRUNCATE TABLE doctor_clinic_fees CASCADE;
TRUNCATE TABLE patients CASCADE;
TRUNCATE TABLE doctors CASCADE;
TRUNCATE TABLE clinics CASCADE;
TRUNCATE TABLE departments CASCADE;
TRUNCATE TABLE organizations CASCADE;
TRUNCATE TABLE users CASCADE;
TRUNCATE TABLE roles CASCADE;
TRUNCATE TABLE audit_logs CASCADE;

-- Re-enable triggers and constraints
SET session_replication_role = 'origin';

SELECT 'All tables truncated successfully!' as message;
"@

# Execute truncate
docker exec -i drandme-backend-postgres-1 psql -U $DB_USER -d $DB_NAME -c $truncateSQL

Write-Host "`n=== Database Cleaned ===" -ForegroundColor Green
Write-Host "All data has been deleted from all tables." -ForegroundColor Green
Write-Host "`nNext steps:" -ForegroundColor Yellow
Write-Host "1. Run migrations again to recreate initial data (roles, etc.)" -ForegroundColor White
Write-Host "2. Or use init-database.ps1 to reinitialize the database" -ForegroundColor White

