# Script to apply migration 012 - Make day_of_week nullable
# This fixes the issue where creating time slots with specific_date fails

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Applying Migration 012" -ForegroundColor Cyan
Write-Host "Making day_of_week nullable" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Database connection parameters
$dbHost = "localhost"
$dbPort = "5432"
$dbUser = "postgres"
$dbPassword = "postgres123"
$dbName = "drandme"

# Set PostgreSQL password environment variable
$env:PGPASSWORD = $dbPassword

Write-Host "`nApplying migration..." -ForegroundColor Yellow

# Run the migration using psql
$migrationFile = "migrations/012_make_day_of_week_nullable.sql"

try {
    # Execute the migration
    psql -h $dbHost -p $dbPort -U $dbUser -d $dbName -f $migrationFile
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "`n✅ Migration applied successfully!" -ForegroundColor Green
        Write-Host "The day_of_week column is now nullable." -ForegroundColor Green
        Write-Host "You can now create time slots with specific dates." -ForegroundColor Green
    } else {
        Write-Host "`n❌ Migration failed with exit code: $LASTEXITCODE" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "`n❌ Error applying migration: $_" -ForegroundColor Red
    exit 1
} finally {
    # Clear the password from environment
    Remove-Item Env:\PGPASSWORD
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Migration Complete" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

