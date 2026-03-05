# Complete build and run script for microservices (PowerShell)

Write-Host "🚀 Building and Running DrAndMe Microservices" -ForegroundColor Cyan
Write-Host "==============================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Stop existing containers
Write-Host "Step 1: Stopping existing containers..." -ForegroundColor Yellow
docker-compose down

# Step 2: Build all services
Write-Host "`nStep 2: Building all services..." -ForegroundColor Yellow
docker-compose build --no-cache

# Step 3: Start all services
Write-Host "`nStep 3: Starting all services..." -ForegroundColor Yellow
docker-compose up -d

# Step 4: Wait for services to be ready
Write-Host "`nStep 4: Waiting for services to be ready..." -ForegroundColor Yellow
Start-Sleep -Seconds 15

# Step 5: Check service status
Write-Host "`nStep 5: Checking service status..." -ForegroundColor Yellow
docker-compose ps

# Step 6: Health checks
Write-Host "`nStep 6: Running health checks..." -ForegroundColor Yellow

# Check Kong
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8001/status" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "✅ Kong Gateway: Healthy" -ForegroundColor Green
} catch {
    Write-Host "❌ Kong Gateway: Not responding" -ForegroundColor Red
}

# Check Auth Service
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8000/api/auth/health" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "✅ Auth Service: Healthy" -ForegroundColor Green
} catch {
    Write-Host "❌ Auth Service: Not responding" -ForegroundColor Red
}

# Check Organization Service
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8000/api/organizations/health" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "✅ Organization Service: Healthy" -ForegroundColor Green
} catch {
    Write-Host "❌ Organization Service: Not responding" -ForegroundColor Red
}

# Check Appointment Service
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8000/api/v1/health" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "✅ Appointment Service: Healthy" -ForegroundColor Green
} catch {
    Write-Host "❌ Appointment Service: Not responding" -ForegroundColor Red
}

Write-Host ""
Write-Host "🎉 Build and deployment complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Access points:"
Write-Host "  - Kong Gateway: http://localhost:8000"
Write-Host "  - Kong Admin: http://localhost:8001"
Write-Host "  - PgAdmin: http://localhost:5051"
Write-Host ""
Write-Host "View logs: docker-compose logs -f"
Write-Host "Stop services: docker-compose down"

