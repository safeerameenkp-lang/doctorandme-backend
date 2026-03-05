# Enhanced Docker Build Script with Network Issue Handling
# This script addresses TLS handshake timeouts and Docker Hub connectivity issues

param(
    [string]$Service = "all",
    [int]$MaxRetries = 5,
    [int]$RetryDelay = 15,
    [switch]$UseMirror = $false,
    [switch]$SkipCache = $false
)

Write-Host "🐳 Enhanced Docker Build Script for DrAndMe Backend" -ForegroundColor Cyan
Write-Host "=================================================" -ForegroundColor Cyan

# Function to configure Docker daemon for better network handling
function Set-DockerDaemonConfig {
    Write-Host "🔧 Configuring Docker daemon for better network handling..." -ForegroundColor Yellow
    
    $dockerConfigPath = "$env:USERPROFILE\.docker\daemon.json"
    $dockerDir = Split-Path $dockerConfigPath -Parent
    
    if (-not (Test-Path $dockerDir)) {
        New-Item -ItemType Directory -Path $dockerDir -Force | Out-Null
    }
    
    $daemonConfig = @{
        "registry-mirrors" = @(
            "https://docker.mirrors.ustc.edu.cn",
            "https://hub-mirror.c.163.com",
            "https://mirror.baidubce.com"
        )
        "max-concurrent-downloads" = 3
        "max-concurrent-uploads" = 3
        "log-driver" = "json-file"
        "log-opts" = @{
            "max-size" = "10m"
            "max-file" = "3"
        }
    }
    
    $daemonConfig | ConvertTo-Json -Depth 3 | Set-Content $dockerConfigPath -Encoding UTF8
    Write-Host "✅ Docker daemon configuration updated" -ForegroundColor Green
}

# Function to test network connectivity
function Test-NetworkConnectivity {
    Write-Host "🌐 Testing network connectivity..." -ForegroundColor Yellow
    
    $testUrls = @(
        "https://registry-1.docker.io",
        "https://auth.docker.io",
        "https://index.docker.io"
    )
    
    foreach ($url in $testUrls) {
        try {
            $response = Invoke-WebRequest -Uri $url -TimeoutSec 10 -UseBasicParsing
            if ($response.StatusCode -eq 200) {
                Write-Host "✅ $url is accessible" -ForegroundColor Green
            }
        } catch {
            Write-Host "❌ $url is not accessible: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
}

# Function to pull base images with multiple strategies
function Pull-BaseImages {
    Write-Host "📥 Pulling base images with multiple strategies..." -ForegroundColor Yellow
    
    $baseImages = @(
        "alpine:3.19",
        "golang:1.21-alpine"
    )
    
    foreach ($image in $baseImages) {
        $attempts = 0
        $success = $false
        
        while ($attempts -lt $MaxRetries -and -not $success) {
            $attempts++
            Write-Host "🔄 Attempting to pull $image (attempt $attempts/$MaxRetries)..." -ForegroundColor Yellow
            
            try {
                # Try with different timeout settings
                $timeout = 300 + ($attempts * 60)  # Increasing timeout
                
                $result = docker pull $image 2>&1
                if ($LASTEXITCODE -eq 0) {
                    Write-Host "✅ Successfully pulled $image" -ForegroundColor Green
                    $success = $true
                } else {
                    Write-Host "❌ Pull failed: $result" -ForegroundColor Red
                }
            } catch {
                Write-Host "❌ Pull exception: $($_.Exception.Message)" -ForegroundColor Red
            }
            
            if (-not $success -and $attempts -lt $MaxRetries) {
                Write-Host "⏳ Waiting $RetryDelay seconds before retry..." -ForegroundColor Yellow
                Start-Sleep -Seconds $RetryDelay
            }
        }
        
        if (-not $success) {
            Write-Host "💥 Failed to pull $image after $MaxRetries attempts" -ForegroundColor Red
            return $false
        }
    }
    
    return $true
}

# Function to build with network optimizations
function Build-ServiceOptimized {
    param([string]$ServiceName, [string]$DockerfilePath)
    
    Write-Host "🚀 Building $ServiceName with network optimizations..." -ForegroundColor Cyan
    
    $buildArgs = @(
        "build"
        "-f", $DockerfilePath
        "-t", "drandme-$ServiceName"
    )
    
    if ($SkipCache) {
        $buildArgs += "--no-cache"
    }
    
    # Add network optimization flags
    $buildArgs += @(
        "--network=host"
        "--build-arg", "BUILDKIT_INLINE_CACHE=1"
    )
    
    $attempts = 0
    while ($attempts -lt $MaxRetries) {
        $attempts++
        Write-Host "🔄 Build attempt $attempts/$MaxRetries for $ServiceName..." -ForegroundColor Yellow
        
        try {
            $result = & docker $buildArgs . 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ Successfully built $ServiceName" -ForegroundColor Green
                return $true
            } else {
                Write-Host "❌ Build failed: $result" -ForegroundColor Red
            }
        } catch {
            Write-Host "❌ Build exception: $($_.Exception.Message)" -ForegroundColor Red
        }
        
        if ($attempts -lt $MaxRetries) {
            Write-Host "⏳ Waiting $RetryDelay seconds before retry..." -ForegroundColor Yellow
            Start-Sleep -Seconds $RetryDelay
            
            # Clear Docker cache between attempts
            Write-Host "🧹 Clearing Docker cache..." -ForegroundColor Yellow
            docker system prune -f | Out-Null
        }
    }
    
    Write-Host "💥 Failed to build $ServiceName after $MaxRetries attempts" -ForegroundColor Red
    return $false
}

# Function to check Docker status
function Test-DockerStatus {
    Write-Host "🔍 Checking Docker status..." -ForegroundColor Yellow
    
    try {
        $result = docker info 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Docker is running and accessible" -ForegroundColor Green
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

# Main execution
Write-Host "Starting enhanced Docker build process..." -ForegroundColor Cyan

# Check Docker status
if (-not (Test-DockerStatus)) {
    Write-Host "❌ Docker is not running. Please start Docker Desktop and try again." -ForegroundColor Red
    exit 1
}

# Configure Docker daemon
Set-DockerDaemonConfig

# Test network connectivity
Test-NetworkConnectivity

# Pull base images first
if (-not (Pull-BaseImages)) {
    Write-Host "❌ Failed to pull base images. Check your network connection." -ForegroundColor Red
    exit 1
}

# Define services
$services = @{
    "auth-service" = "services/auth-service/Dockerfile"
    "organization-service" = "services/organization-service/Dockerfile"
    "appointment-service" = "services/appointment-service/Dockerfile"
}

$failedServices = @()

# Build services
if ($Service -eq "all") {
    foreach ($serviceName in $services.Keys) {
        $dockerfilePath = $services[$serviceName]
        
        if (-not (Build-ServiceOptimized -ServiceName $serviceName -DockerfilePath $dockerfilePath)) {
            $failedServices += $serviceName
        }
        
        Write-Host "" # Empty line for readability
    }
} else {
    if ($services.ContainsKey($Service)) {
        $dockerfilePath = $services[$Service]
        if (-not (Build-ServiceOptimized -ServiceName $Service -DockerfilePath $dockerfilePath)) {
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
    Write-Host "🔧 Additional troubleshooting steps:" -ForegroundColor Yellow
    Write-Host "1. Restart Docker Desktop completely" -ForegroundColor White
    Write-Host "2. Check Windows Firewall settings" -ForegroundColor White
    Write-Host "3. Try using a VPN or different network" -ForegroundColor White
    Write-Host "4. Run: docker system prune -a --volumes" -ForegroundColor White
    Write-Host "5. Try building with: .\docker-build-fixed.ps1 -SkipCache" -ForegroundColor White
}

Write-Host ""
Write-Host "Script completed at $(Get-Date)" -ForegroundColor Gray
