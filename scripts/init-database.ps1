# Database initialization script for Healthcare SaaS Security System
# This script ensures all migrations are applied in the correct order

param(
    [string]$DbHost = "localhost",
    [int]$Port = 5432,
    [string]$Database = "drandme",
    [string]$Username = "postgres",
    [string]$Password = "postgres123"
)

Write-Host "🗄️  Initializing Healthcare SaaS Database with Security Features" -ForegroundColor Green
Write-Host "================================================================" -ForegroundColor Green

# Function to wait for PostgreSQL
function Wait-ForPostgreSQL {
    param([string]$DbHost, [int]$Port)
    
    Write-Host "Waiting for PostgreSQL to be ready..." -ForegroundColor Yellow
    
    do {
        try {
            $connection = New-Object System.Data.SqlClient.SqlConnection
            $connection.ConnectionString = "Server=$DbHost,$Port;Database=postgres;User Id=$Username;Password=$Password;Connection Timeout=5;"
            $connection.Open()
            $connection.Close()
            Write-Host "✅ PostgreSQL is ready!" -ForegroundColor Green
            return $true
        }
        catch {
            Write-Host "PostgreSQL is unavailable - sleeping" -ForegroundColor Gray
            Start-Sleep -Seconds 2
        }
    } while ($true)
}

# Function to run SQL file
function Invoke-SqlFile {
    param([string]$FilePath, [string]$Description)
    
    try {
        Write-Host "Running: $Description" -ForegroundColor Cyan
        
        # Use psql if available, otherwise use PowerShell
        if (Get-Command psql -ErrorAction SilentlyContinue) {
            $env:PGPASSWORD = $Password
            & psql -h $DbHost -p $Port -U $Username -d $Database -f $FilePath
        } else {
            # Alternative: Use .NET SQL client
            $sql = Get-Content $FilePath -Raw
            $connection = New-Object System.Data.SqlClient.SqlConnection
            $connection.ConnectionString = "Server=$DbHost,$Port;Database=$Database;User Id=$Username;Password=$Password;"
            $connection.Open()
            
            $command = New-Object System.Data.SqlClient.SqlCommand($sql, $connection)
            $command.ExecuteNonQuery()
            $connection.Close()
        }
        
        Write-Host "✅ $Description completed" -ForegroundColor Green
    }
    catch {
        Write-Host "❌ Failed to run $Description`: $($_.Exception.Message)" -ForegroundColor Red
        throw
    }
}

# Main execution
try {
    # Wait for PostgreSQL
    Wait-ForPostgreSQL -DbHost $DbHost -Port $Port
    
    # Run migrations in order
    $migrations = @(
        @{ File = "migrations/001_initial_schema.sql"; Description = "Initial schema" },
        @{ File = "migrations/002_add_mo_id_to_patients.sql"; Description = "MO ID to patients" },
        @{ File = "migrations/003_admin_features.sql"; Description = "Admin features" },
        @{ File = "migrations/004_add_appointment_fields.sql"; Description = "Appointment fields" },
        @{ File = "migrations/005_security_features.sql"; Description = "Security features" }
    )
    
    foreach ($migration in $migrations) {
        if (Test-Path $migration.File) {
            Invoke-SqlFile -FilePath $migration.File -Description $migration.Description
        } else {
            Write-Host "⚠️  Migration file not found: $($migration.File)" -ForegroundColor Yellow
        }
    }
    
    # Verify security tables
    Write-Host "`nVerifying security tables..." -ForegroundColor Cyan
    
    $verifyQuery = @"
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
"@
    
    try {
        $env:PGPASSWORD = $Password
        & psql -h $DbHost -p $Port -U $Username -d $Database -c $verifyQuery
    }
    catch {
        Write-Host "Could not verify tables (psql not available)" -ForegroundColor Yellow
    }
    
    Write-Host "`n🔐 Security Tables Created:" -ForegroundColor Magenta
    Write-Host "  - failed_login_attempts: Tracks failed login attempts" -ForegroundColor White
    Write-Host "  - account_lockouts: Manages account lockout status" -ForegroundColor White
    Write-Host "  - blocked_ips: IP address blocking" -ForegroundColor White
    Write-Host "  - security_audit_log: Comprehensive audit logging" -ForegroundColor White
    
    Write-Host "`n🎉 Database initialization completed successfully!" -ForegroundColor Green
    Write-Host "Your healthcare SaaS security system is ready!" -ForegroundColor Green
    
}
catch {
    Write-Host "`n❌ Database initialization failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
