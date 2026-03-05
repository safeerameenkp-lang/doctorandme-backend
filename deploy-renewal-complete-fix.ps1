# Deploy Follow-Up Renewal Fix
# This fixes the department matching issue that prevented renewals

Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "  Follow-Up Renewal System - COMPLETE FIX" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "🔍 Problem: Follow-up not renewing after regular appointments" -ForegroundColor Yellow
Write-Host "✅ Solution: Fixed department matching logic" -ForegroundColor Green
Write-Host ""

# Step 1: Build both services
Write-Host "Step 1: Building services..." -ForegroundColor Yellow
docker-compose build appointment-service organization-service

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Build successful!" -ForegroundColor Green
Write-Host ""

# Step 2: Deploy both services
Write-Host "Step 2: Deploying services..." -ForegroundColor Yellow
docker-compose up -d appointment-service organization-service

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Deployment failed!" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Deployment successful!" -ForegroundColor Green
Write-Host ""

# Step 3: Wait for services to start
Write-Host "Step 3: Waiting for services to start..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Step 4: Check logs
Write-Host "Step 4: Checking service logs..." -ForegroundColor Yellow
Write-Host ""
Write-Host "📋 Appointment Service Logs:" -ForegroundColor Cyan
docker-compose logs appointment-service --tail=10
Write-Host ""
Write-Host "📋 Organization Service Logs:" -ForegroundColor Cyan
docker-compose logs organization-service --tail=10
Write-Host ""

# Step 5: Test instructions
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "  ✅ Renewal Fix Deployed!" -ForegroundColor Green
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "🧪 Test Your Scenario:" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Book Regular Appointment #1" -ForegroundColor White
Write-Host "2. Book Follow-Up #1 (should be GREEN/FREE)" -ForegroundColor White
Write-Host "3. Book Regular Appointment #2 (same doctor+dept)" -ForegroundColor White
Write-Host "4. Book Follow-Up #2 (should be GREEN/FREE again!)" -ForegroundColor White
Write-Host ""
Write-Host "✅ Expected: Step 4 should show GREEN (free) not ORANGE (paid)" -ForegroundColor Green
Write-Host ""
Write-Host "🔍 Check logs for:" -ForegroundColor Yellow
Write-Host "   🔄 Creating follow-up: Patient=xxx, Doctor=yyy, Dept=Cardiology" -ForegroundColor Gray
Write-Host "   🔄 Renewed 1 existing follow-up(s)" -ForegroundColor Gray
Write-Host "   ✅ Created follow-up eligibility" -ForegroundColor Gray
Write-Host ""
Write-Host "❌ NOT:" -ForegroundColor Yellow
Write-Host "   ⚠️ Warning: Failed to renew existing follow-ups" -ForegroundColor Gray
Write-Host ""
Write-Host "📚 Documentation:" -ForegroundColor Yellow
Write-Host "   - FOLLOWUP_RENEWAL_COMPLETE_FIX.md" -ForegroundColor Gray
Write-Host ""
Write-Host "✅ Ready to test!" -ForegroundColor Green
Write-Host ""
