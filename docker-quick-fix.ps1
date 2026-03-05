# Quick Docker Build Fix Script
# Addresses TLS handshake timeout issues

param(
    [string]$Service = "appointment-service",
    [int]$MaxRetries = 3
)

Write-Host "🐳 Quick Docker Build Fix" -ForegroundColor Cyan
Write-Host "=========================" -ForegroundColor Cyan

# Check Docker status
Write-Host "🔍 Checking Docker status..." -ForegroundColor Yellow
try {
    docker info | Out-Null
    Write-Host "✅ Docker is running" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}

# Clear Docker cache
Write-Host "🧹 Clearing Docker cache..." -ForegroundColor Yellow
docker system prune -f | Out-Null

# Try to pull Alpine image with retry
Write-Host "📥 Pulling Alpine 3.19 image..." -ForegroundColor Yellow
$attempts = 0
$success = $false

while ($attempts -lt $MaxRetries -and -not $success) {
    $attempts++
    Write-Host "🔄 Attempt $attempts/$MaxRetries" -ForegroundColor Yellow
    
    try {
        $result = docker pull alpine:3.19 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Successfully pulled Alpine 3.19" -ForegroundColor Green
            $success = $true
        } else {
            Write-Host "❌ Pull failed: $result" -ForegroundColor Red
        }
    } catch {
        Write-Host "❌ Pull exception: $($_.Exception.Message)" -ForegroundColor Red
    }
    
    if (-not $success -and $attempts -lt $MaxRetries) {
        Write-Host "⏳ Waiting 10 seconds before retry..." -ForegroundColor Yellow
        Start-Sleep -Seconds 10
    }
}

if (-not $success) {
    Write-Host "❌ Failed to pull Alpine image. Trying alternative approach..." -ForegroundColor Red
    
    # Try building with --network=host
    Write-Host "🌐 Trying build with host network..." -ForegroundColor Yellow
    try {
        $result = docker build --network=host -f "services/$Service/Dockerfile" -t "drandme-$Service" . 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Successfully built $Service with host network" -ForegroundColor Green
            $success = $true
        } else {
            Write-Host "❌ Host network build failed: $result" -ForegroundColor Red
        }
    } catch {
        Write-Host "❌ Host network build exception: $($_.Exception.Message)" -ForegroundColor Red
    }
}

if ($success) {
    Write-Host "🎉 Build completed successfully!" -ForegroundColor Green
    Write-Host "🚀 You can now run: docker-compose up" -ForegroundColor Green
} else {
    Write-Host "💥 Build failed after all attempts" -ForegroundColor Red
    Write-Host ""
    Write-Host "🔧 Troubleshooting suggestions:" -ForegroundColor Yellow
    Write-Host "1. Restart Docker Desktop" -ForegroundColor White
    Write-Host "2. Check your internet connection" -ForegroundColor White
    Write-Host "3. Try using a VPN" -ForegroundColor White
    Write-Host "4. Run: docker system prune -a" -ForegroundColor White
}

Write-Host ""
Write-Host "Script completed at $(Get-Date)" -ForegroundColor Gray
