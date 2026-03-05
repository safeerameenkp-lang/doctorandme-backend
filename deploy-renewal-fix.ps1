# Deploy Follow-Up Renewal Fix
# This fixes the "Free follow-up already used" error

Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "  Follow-Up Renewal Error - FIX" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "🔍 Problem: 'Free follow-up already used' error after renewal" -ForegroundColor Yellow
Write-Host "✅ Solution: Fixed conflicting validation logic" -ForegroundColor Green
Write-Host ""

# Step 1: Build the service
Write-Host "Step 1: Building appointment-service..." -ForegroundColor Yellow
docker-compose build appointment-service

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Build successful!" -ForegroundColor Green
Write-Host ""

# Step 2: Deploy the service
Write-Host "Step 2: Deploying appointment-service..." -ForegroundColor Yellow
docker-compose up -d appointment-service

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Deployment failed!" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Deployment successful!" -ForegroundColor Green
Write-Host ""

# Step 3: Wait for service to start
Write-Host "Step 3: Waiting for service to start..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Step 4: Check logs
Write-Host "Step 4: Checking service logs..." -ForegroundColor Yellow
Write-Host ""
docker-compose logs appointment-service --tail=20
Write-Host ""

# Step 5: Test instructions
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "  ✅ Renewal Fix Deployed!" -ForegroundColor Green
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "🧪 Test Your Scenario:" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Book Regular Appointment #1" -ForegroundColor White
Write-Host "2. Book Follow-Up #1 (FREE)" -ForegroundColor White
Write-Host "3. Book Regular Appointment #2 (same doctor+dept)" -ForegroundColor White
Write-Host "4. Book Follow-Up #2 (should be FREE now!)" -ForegroundColor White
Write-Host ""
Write-Host "✅ Expected: Step 4 should work without 'already used' error" -ForegroundColor Green
Write-Host ""
Write-Host "🔍 Check logs for:" -ForegroundColor Yellow
Write-Host "   ✅ Follow-up eligibility: Free=true" -ForegroundColor Gray
Write-Host "   ✅ Free follow-up verified" -ForegroundColor Gray
Write-Host ""
Write-Host "❌ NOT:" -ForegroundColor Yellow
Write-Host "   🚨 FRAUD ATTEMPT: Patient already used" -ForegroundColor Gray
Write-Host ""
Write-Host "📚 Documentation:" -ForegroundColor Yellow
Write-Host "   - FOLLOWUP_RENEWAL_ERROR_FIXED.md" -ForegroundColor Gray
Write-Host ""
Write-Host "✅ Ready to test!" -ForegroundColor Green
Write-Host ""

