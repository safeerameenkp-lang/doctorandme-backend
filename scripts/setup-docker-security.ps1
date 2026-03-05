# Complete Docker Build and Database Setup for Healthcare SaaS Security System
# This script builds all services and ensures the security tables are properly created

param(
    [switch]$Clean = $false,
    [switch]$SkipBuild = $false,
    [switch]$SkipMigration = $false,
    [switch]$TestSecurity = $false
)

Write-Host "🐳 Healthcare SaaS Security System - Complete Docker Setup" -ForegroundColor Green
Write-Host "=========================================================" -ForegroundColor Green

# Function to check if Docker is running
function Test-DockerRunning {
    try {
        docker version | Out-Null
        return $true
    }
    catch {
        Write-Host "❌ Docker is not running. Please start Docker Desktop." -ForegroundColor Red
        return $false
    }
}

# Function to clean Docker environment
function Clear-DockerEnvironment {
    Write-Host "🧹 Cleaning Docker environment..." -ForegroundColor Yellow
    
    # Stop and remove containers
    docker-compose down -v --remove-orphans
    
    # Remove images if clean flag is set
    if ($Clean) {
        Write-Host "Removing all project images..." -ForegroundColor Yellow
        try {
            docker rmi $(docker images "drandme-backend*" -q) 2>$null
        } catch {
            # Ignore errors if no images found
        }
        docker system prune -f
    }
}

# Function to build services
function Build-Services {
    if ($SkipBuild) {
        Write-Host "⏭️  Skipping build (--SkipBuild)" -ForegroundColor Yellow
        return
    }
    
    Write-Host "🔨 Building services..." -ForegroundColor Cyan
    
    $services = @("auth-service", "organization-service", "appointment-service")
    
    foreach ($service in $services) {
        Write-Host "Building $service..." -ForegroundColor Yellow
        try {
            docker-compose build --no-cache $service
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ $service built successfully" -ForegroundColor Green
            } else {
                Write-Host "❌ $service build failed" -ForegroundColor Red
                return $false
            }
        }
        catch {
            Write-Host "❌ Exception building $service`: $($_.Exception.Message)" -ForegroundColor Red
            return $false
        }
    }
    
    return $true
}

# Function to start PostgreSQL and wait for it
function Start-PostgreSQL {
    Write-Host "🗄️  Starting PostgreSQL..." -ForegroundColor Cyan
    docker-compose up -d postgres
    
    Write-Host "Waiting for PostgreSQL to be ready..." -ForegroundColor Yellow
    $maxWait = 60
    $waited = 0
    
    do {
        try {
            $result = docker-compose exec -T postgres pg_isready -U postgres
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ PostgreSQL is ready!" -ForegroundColor Green
                return $true
            }
        }
        catch {
            # Continue waiting
        }
        
        Start-Sleep -Seconds 2
        $waited += 2
        Write-Host "." -NoNewline -ForegroundColor Gray
        
    } while ($waited -lt $maxWait)
    
    Write-Host "`n❌ PostgreSQL failed to start within $maxWait seconds" -ForegroundColor Red
    return $false
}

# Function to verify security tables
function Test-SecurityTables {
    Write-Host "🔐 Verifying security tables..." -ForegroundColor Cyan
    
    try {
        $query = @"
SELECT 
    table_name,
    CASE 
        WHEN table_name IN ('failed_login_attempts', 'account_lockouts', 'blocked_ips', 'security_audit_log') 
        THEN 'Security Table'
        ELSE 'Regular Table'
    END as table_type
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name IN ('failed_login_attempts', 'account_lockouts', 'blocked_ips', 'security_audit_log', 'users', 'roles')
ORDER BY table_name;
"@
        
        $result = docker-compose exec -T postgres psql -U postgres -d drandme -c $query
        Write-Host "Security tables verification:" -ForegroundColor White
        Write-Host $result -ForegroundColor Gray
        
        # Check if security tables exist
        $securityTables = @('failed_login_attempts', 'account_lockouts', 'blocked_ips', 'security_audit_log')
        $allPresent = $true
        
        foreach ($table in $securityTables) {
            if ($result -match $table) {
                Write-Host "✅ $table exists" -ForegroundColor Green
            } else {
                Write-Host "❌ $table missing" -ForegroundColor Red
                $allPresent = $false
            }
        }
        
        return $allPresent
    }
    catch {
        Write-Host "❌ Failed to verify security tables: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

# Function to start all services
function Start-AllServices {
    Write-Host "🚀 Starting all services..." -ForegroundColor Cyan
    docker-compose up -d
    
    Write-Host "Waiting for services to start..." -ForegroundColor Yellow
    Start-Sleep -Seconds 15
    
    # Check service status
    Write-Host "`nService Status:" -ForegroundColor Cyan
    docker-compose ps
}

# Function to test security system
function Test-SecuritySystem {
    if (-not $TestSecurity) {
        return
    }
    
    Write-Host "`n🧪 Testing security system..." -ForegroundColor Cyan
    
    try {
        # Test auth service health
        $healthResponse = Invoke-RestMethod -Uri "http://localhost:8080/auth/health" -Method Get -TimeoutSec 10
        Write-Host "✅ Auth service is healthy" -ForegroundColor Green
        
        # Test security stats endpoint
        $statsResponse = Invoke-RestMethod -Uri "http://localhost:8080/auth/security/stats" -Method Get -TimeoutSec 10
        Write-Host "✅ Security stats endpoint working" -ForegroundColor Green
        
        Write-Host "Security system test results:" -ForegroundColor White
        Write-Host ($statsResponse | ConvertTo-Json -Depth 2) -ForegroundColor Gray
        
    }
    catch {
        Write-Host "⚠️  Security system test failed: $($_.Exception.Message)" -ForegroundColor Yellow
    }
}

# Main execution
try {
    # Check Docker
    if (-not (Test-DockerRunning)) {
        exit 1
    }
    
    # Clean environment if requested
    if ($Clean) {
        Clear-DockerEnvironment
    }
    
    # Build services
    if (-not (Build-Services)) {
        Write-Host "❌ Service build failed" -ForegroundColor Red
        exit 1
    }
    
    # Start PostgreSQL
    if (-not (Start-PostgreSQL)) {
        Write-Host "❌ PostgreSQL startup failed" -ForegroundColor Red
        exit 1
    }
    
    # Verify security tables
    if (-not (Test-SecurityTables)) {
        Write-Host "❌ Security tables verification failed" -ForegroundColor Red
        Write-Host "The security migration may not have run properly." -ForegroundColor Yellow
        Write-Host "Try running: docker-compose down -v && docker-compose up -d" -ForegroundColor Cyan
        exit 1
    }
    
    # Start all services
    Start-AllServices
    
    # Test security system
    Test-SecuritySystem
    
    # Final summary
    Write-Host "`n🎉 Healthcare SaaS Security System Setup Complete!" -ForegroundColor Green
    Write-Host "===============================================" -ForegroundColor Green
    
    Write-Host "`n🌐 Service URLs:" -ForegroundColor Cyan
    Write-Host "  - Auth Service: http://localhost:8080" -ForegroundColor White
    Write-Host "  - Organization Service: http://localhost:8081" -ForegroundColor White
    Write-Host "  - Appointment Service: http://localhost:8082" -ForegroundColor White
    Write-Host "  - PostgreSQL: localhost:5432" -ForegroundColor White
    Write-Host "  - pgAdmin: http://localhost:5050" -ForegroundColor White
    
    Write-Host "`n🔐 Security Features Active:" -ForegroundColor Magenta
    Write-Host "  - Role-based authentication" -ForegroundColor White
    Write-Host "  - Failed login attempt tracking" -ForegroundColor White
    Write-Host "  - Account lockout system (5 attempts, 15 min)" -ForegroundColor White
    Write-Host "  - IP blocking" -ForegroundColor White
    Write-Host "  - Comprehensive audit logging" -ForegroundColor White
    
    Write-Host "`n📋 Next Steps:" -ForegroundColor Cyan
    Write-Host "1. Test the system: ./scripts/test-security-system.ps1" -ForegroundColor White
    Write-Host "2. Access pgAdmin: http://localhost:5050 (admin@drandme.com / admin123)" -ForegroundColor White
    Write-Host "3. Check logs: docker-compose logs [service-name]" -ForegroundColor White
    Write-Host "4. View security stats: GET http://localhost:8080/auth/security/stats" -ForegroundColor White
    
}
catch {
    Write-Host "`n❌ Setup failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Check the error messages above and try again." -ForegroundColor Yellow
    exit 1
}
