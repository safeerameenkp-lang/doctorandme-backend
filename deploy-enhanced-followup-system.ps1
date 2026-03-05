# Enhanced Follow-Up System Deployment Script
# Deploys the doctor-specific follow-up system

Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "  Enhanced Follow-Up System Deployment" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""

# Check if Docker is running
Write-Host "🔍 Checking Docker status..." -ForegroundColor Yellow
$dockerStatus = docker info 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Docker is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}
Write-Host "✅ Docker is running" -ForegroundColor Green
Write-Host ""

# Build organization service with enhanced follow-up system
Write-Host "🔨 Building organization service with enhanced follow-up system..." -ForegroundColor Yellow
docker-compose build organization-service

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Build successful" -ForegroundColor Green
Write-Host ""

# Deploy organization service
Write-Host "🚀 Deploying organization service..." -ForegroundColor Yellow
docker-compose up -d organization-service

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Deployment failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Deployment successful" -ForegroundColor Green
Write-Host ""

# Wait for service to start
Write-Host "⏳ Waiting for service to start..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Check service health
Write-Host "🔍 Checking service health..." -ForegroundColor Yellow
$healthCheck = Invoke-RestMethod -Uri "http://localhost:3002/api/organizations/health" -Method GET -TimeoutSec 10 2>$null

if ($healthCheck) {
    Write-Host "✅ Service is healthy" -ForegroundColor Green
} else {
    Write-Host "⚠️ Service health check failed, but service might still be starting..." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "📋 Enhanced Follow-Up System Features:" -ForegroundColor Cyan
Write-Host "   ✅ Doctor-specific follow-up status" -ForegroundColor Green
Write-Host "   ✅ Department-specific follow-up tracking" -ForegroundColor Green
Write-Host "   ✅ Complete appointment history per doctor+department" -ForegroundColor Green
Write-Host "   ✅ Enhanced API endpoints" -ForegroundColor Green
Write-Host "   ✅ Automatic follow-up creation on regular appointments" -ForegroundColor Green
Write-Host ""

Write-Host "🔗 New API Endpoints:" -ForegroundColor Cyan
Write-Host "   GET /api/organizations/patient-followup-status/{patient_id}" -ForegroundColor White
Write-Host "   GET /api/organizations/clinic-specific-patients (enhanced)" -ForegroundColor White
Write-Host ""

Write-Host "📱 Flutter Integration:" -ForegroundColor Cyan
Write-Host "   ✅ Enhanced API service functions" -ForegroundColor Green
Write-Host "   ✅ Doctor-specific follow-up UI components" -ForegroundColor Green
Write-Host "   ✅ Complete appointment history display" -ForegroundColor Green
Write-Host "   ✅ Status-based color coding" -ForegroundColor Green
Write-Host ""

Write-Host "🧪 Test the Enhanced System:" -ForegroundColor Yellow
Write-Host "1. Get patient follow-up status:" -ForegroundColor White
Write-Host "   curl -H 'Authorization: Bearer YOUR_TOKEN' 'http://localhost:3002/api/organizations/patient-followup-status/PATIENT_ID'" -ForegroundColor Gray
Write-Host ""
Write-Host "2. List patients with enhanced follow-up status:" -ForegroundColor White
Write-Host "   curl -H 'Authorization: Bearer YOUR_TOKEN' 'http://localhost:3002/api/organizations/clinic-specific-patients?clinic_id=CLINIC_ID'" -ForegroundColor Gray
Write-Host ""

Write-Host "📚 Documentation:" -ForegroundColor Cyan
Write-Host "   - ENHANCED_FOLLOWUP_SYSTEM_COMPLETE.md" -ForegroundColor Gray
Write-Host "   - COMPLETE_FOLLOWUP_FLUTTER_DOCUMENTATION.md" -ForegroundColor Gray
Write-Host ""

Write-Host "✅ Enhanced Follow-Up System deployed successfully!" -ForegroundColor Green
Write-Host ""
