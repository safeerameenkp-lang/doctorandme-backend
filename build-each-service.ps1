# Build Each Microservice Individually (PowerShell)

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("auth", "organization", "appointment", "all")]
    [string]$Service = "all"
)

Write-Host "🔨 Building Microservices Individually" -ForegroundColor Cyan
Write-Host "=======================================" -ForegroundColor Cyan
Write-Host ""

function Build-Service {
    param([string]$ServiceName, [string]$DisplayName)
    
    Write-Host "Building $DisplayName..." -ForegroundColor Yellow
    docker-compose build $ServiceName
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ $DisplayName built successfully!" -ForegroundColor Green
        Write-Host ""
        return $true
    } else {
        Write-Host "❌ $DisplayName build failed!" -ForegroundColor Red
        Write-Host ""
        return $false
    }
}

# Build based on parameter
switch ($Service) {
    "auth" {
        Build-Service -ServiceName "auth-service" -DisplayName "Auth Service"
    }
    "organization" {
        Build-Service -ServiceName "organization-service" -DisplayName "Organization Service"
    }
    "appointment" {
        Build-Service -ServiceName "appointment-service" -DisplayName "Appointment Service"
    }
    "all" {
        Write-Host "Building all services..." -ForegroundColor Cyan
        Write-Host ""
        
        $results = @()
        $results += Build-Service -ServiceName "auth-service" -DisplayName "Auth Service"
        $results += Build-Service -ServiceName "organization-service" -DisplayName "Organization Service"
        $results += Build-Service -ServiceName "appointment-service" -DisplayName "Appointment Service"
        
        Write-Host "=======================================" -ForegroundColor Cyan
        Write-Host "Build Summary:" -ForegroundColor Cyan
        Write-Host ""
        
        if ($results -contains $false) {
            Write-Host "⚠️  Some services failed to build" -ForegroundColor Yellow
        } else {
            Write-Host "✅ All services built successfully!" -ForegroundColor Green
        }
        
        Write-Host ""
        Write-Host "Built images:" -ForegroundColor Cyan
        docker images | Select-String "drandme"
    }
}

Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "  - Start services: docker-compose up -d"
Write-Host "  - View logs: docker-compose logs -f"
Write-Host "  - Check status: docker-compose ps"

