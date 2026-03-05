# Docker Build Retry Script for Windows PowerShell
# This script handles Docker Hub connectivity issues and provides multiple fallback options

param(
    [string]$Service = "all",
    [int]$MaxRetries = 3,
    [int]$RetryDelay = 10
)

Write-Host "🐳 Starting Docker build with retry mechanisms..." -ForegroundColor Cyan

# Function to test Docker connectivity
function Test-DockerConnectivity {
    Write-Host "🔍 Testing Docker connectivity..." -ForegroundColor Yellow
    
    try {
        $result = docker info 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Docker is running" -ForegroundColor Green
            return $true
        } else {
            Write-Host "❌ Docker is not running: $result" -ForegroundColor Red
            return $false
        }
    } catch {
        Write-Host "❌ Docker is not accessible: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

# Function to pull Alpine image with retry
function Pull-AlpineImage {
    param([string]$ImageTag = "3.18")
    
    Write-Host "📥 Attempting to pull Alpine $ImageTag image..." -ForegroundColor Yellow
    
    $attempts = 0
    while ($attempts -lt $MaxRetries) {
        $attempts++
        Write-Host "🔄 Attempt $attempts/$MaxRetries" -ForegroundColor Yellow
        
        try {
            $result = docker pull "alpine:$ImageTag" 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ Successfully pulled Alpine $ImageTag" -ForegroundColor Green
                return $true
            } else {
                Write-Host "❌ Pull failed: $result" -ForegroundColor Red
            }
        } catch {
            Write-Host "❌ Pull exception: $($_.Exception.Message)" -ForegroundColor Red
        }
        
        if ($attempts -lt $MaxRetries) {
            Write-Host "⏳ Waiting $RetryDelay seconds before retry..." -ForegroundColor Yellow
            Start-Sleep -Seconds $RetryDelay
        }
    }
    
    return $false
}

# Function to try alternative registries
function Try-AlternativeRegistries {
    param([string]$ServiceName, [string]$DockerfilePath)
    
    Write-Host "🌐 Trying alternative registries for $ServiceName..." -ForegroundColor Yellow
    
    $alternativeRegistries = @(
        "registry.k8s.io/alpine:3.18",
        "quay.io/alpine/alpine:3.18",
        "ghcr.io/alpine/alpine:3.18"
    )
    
    foreach ($registry in $alternativeRegistries) {
        Write-Host "🔄 Trying registry: $registry" -ForegroundColor Yellow
        
        # Create temporary Dockerfile with alternative registry
        $tempDockerfile = "$DockerfilePath.alt"
        Copy-Item $DockerfilePath $tempDockerfile
        
        # Replace the FROM line
        $content = Get-Content $tempDockerfile
        $content = $content -replace "FROM alpine:3.18", "FROM $registry"
        Set-Content $tempDockerfile $content
        
        try {
            $result = docker build -f $tempDockerfile -t "drandme-$ServiceName" . 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ Successfully built $ServiceName with $registry" -ForegroundColor Green
                Remove-Item $tempDockerfile -Force
                return $true
            }
        } catch {
            Write-Host "❌ Build failed with $registry" -ForegroundColor Red
        }
        
        Remove-Item $tempDockerfile -Force -ErrorAction SilentlyContinue
    }
    
    return $false
}

# Function to build service with multiple strategies
function Build-Service {
    param([string]$ServiceName, [string]$DockerfilePath)
    
    Write-Host "🚀 Building $ServiceName service..." -ForegroundColor Cyan
    
    # Strategy 1: Try to pull Alpine image first
    if (Pull-AlpineImage) {
        Write-Host "📦 Building $ServiceName with pre-pulled Alpine image..." -ForegroundColor Yellow
        try {
            $result = docker build -f $DockerfilePath -t "drandme-$ServiceName" . 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ Successfully built $ServiceName" -ForegroundColor Green
                return $true
            } else {
                Write-Host "❌ Build failed: $result" -ForegroundColor Red
            }
        } catch {
            Write-Host "❌ Build exception: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
    
    # Strategy 2: Try alternative registries
    if (Try-AlternativeRegistries -ServiceName $ServiceName -DockerfilePath $DockerfilePath) {
        return $true
    }
    
    # Strategy 3: Try building without pulling (use cached layers)
    Write-Host "🔄 Trying build with cached layers..." -ForegroundColor Yellow
    try {
        $result = docker build --no-cache -f $DockerfilePath -t "drandme-$ServiceName" . 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Successfully built $ServiceName with no-cache" -ForegroundColor Green
            return $true
        }
    } catch {
        Write-Host "❌ No-cache build failed" -ForegroundColor Red
    }
    
    Write-Host "💥 Failed to build $ServiceName with all strategies" -ForegroundColor Red
    return $false
}

# Function to clear Docker cache and try again
function Clear-DockerCache {
    Write-Host "🧹 Clearing Docker cache..." -ForegroundColor Yellow
    
    try {
        docker system prune -f
        docker builder prune -f
        Write-Host "✅ Docker cache cleared" -ForegroundColor Green
    } catch {
        Write-Host "⚠️ Could not clear Docker cache: $($_.Exception.Message)" -ForegroundColor Yellow
    }
}

# Main execution
Write-Host "🔧 Docker Build Retry Script for DrAndMe Backend" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan

# Check Docker connectivity
if (-not (Test-DockerConnectivity)) {
    Write-Host "❌ Docker is not running. Please start Docker Desktop and try again." -ForegroundColor Red
    exit 1
}

# Define services
$services = @{
    "auth-service" = "services/auth-service/Dockerfile"
    "organization-service" = "services/organization-service/Dockerfile"
    "appointment-service" = "services/appointment-service/Dockerfile"
}

$failedServices = @()

# Clear Docker cache first
Clear-DockerCache

# Build services
if ($Service -eq "all") {
    foreach ($serviceName in $services.Keys) {
        $dockerfilePath = $services[$serviceName]
        
        if (-not (Build-Service -ServiceName $serviceName -DockerfilePath $dockerfilePath)) {
            $failedServices += $serviceName
        }
        
        Write-Host "" # Empty line for readability
    }
} else {
    if ($services.ContainsKey($Service)) {
        $dockerfilePath = $services[$Service]
        if (-not (Build-Service -ServiceName $Service -DockerfilePath $dockerfilePath)) {
            $failedServices += $Service
        }
    } else {
        Write-Host "❌ Unknown service: $Service" -ForegroundColor Red
        Write-Host "Available services: $($services.Keys -join ', ')" -ForegroundColor Yellow
        exit 1
    }
}

# Summary
Write-Host "📊 Build Summary:" -ForegroundColor Cyan
Write-Host "================" -ForegroundColor Cyan

if ($failedServices.Count -eq 0) {
    Write-Host "✅ All services built successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "🚀 You can now run: docker-compose up" -ForegroundColor Green
} else {
    Write-Host "❌ Failed services: $($failedServices -join ', ')" -ForegroundColor Red
    Write-Host ""
    Write-Host "🔧 Troubleshooting suggestions:" -ForegroundColor Yellow
    Write-Host "1. Check your internet connection" -ForegroundColor White
    Write-Host "2. Try running: docker system prune -a" -ForegroundColor White
    Write-Host "3. Restart Docker Desktop" -ForegroundColor White
    Write-Host "4. Try building individual services:" -ForegroundColor White
    Write-Host ""
    
    foreach ($serviceName in $services.Keys) {
        $dockerfilePath = $services[$serviceName]
        Write-Host "   .\docker-build-retry.ps1 -Service $serviceName" -ForegroundColor Cyan
    }
    
    Write-Host ""
    Write-Host "🌐 Alternative: Try using a VPN or different network" -ForegroundColor Yellow
    Write-Host "📞 If issues persist, Docker Hub might be experiencing outages" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Script completed at $(Get-Date)" -ForegroundColor Gray
