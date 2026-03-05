# Deploy Follow-Up Status Labels Fix
# This script builds and deploys the organization-service with the new status labels

Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "  Follow-Up Status Labels - Deployment" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Build the service
Write-Host "Step 1: Building organization-service..." -ForegroundColor Yellow
docker-compose build organization-service

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Build successful!" -ForegroundColor Green
Write-Host ""

# Step 2: Deploy the service
Write-Host "Step 2: Deploying organization-service..." -ForegroundColor Yellow
docker-compose up -d organization-service

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
docker-compose logs organization-service --tail=20
Write-Host ""

# Step 5: Verification instructions
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "  ✅ Deployment Complete!" -ForegroundColor Green
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "📋 Next Steps:" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Test the API:" -ForegroundColor White
Write-Host "   GET /api/organizations/clinic-specific-patients?clinic_id=xxx&doctor_id=yyy&department_id=zzz" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Verify response has new fields:" -ForegroundColor White
Write-Host "   - status_label: 'free' | 'paid' | 'none' | 'needs_selection'" -ForegroundColor Gray
Write-Host "   - color_code: 'green' | 'orange' | 'gray'" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Update frontend to use status_label for UI logic" -ForegroundColor White
Write-Host ""
Write-Host "📚 Documentation:" -ForegroundColor Yellow
Write-Host "   - FOLLOWUP_STATUS_QUICK_REFERENCE.md (start here)" -ForegroundColor Gray
Write-Host "   - FOLLOWUP_STATUS_LABELS_GUIDE.md (complete guide)" -ForegroundColor Gray
Write-Host "   - TEST_FOLLOWUP_STATUS_LABELS.md (test scenarios)" -ForegroundColor Gray
Write-Host ""
Write-Host "🔍 Check logs:" -ForegroundColor Yellow
Write-Host "   docker-compose logs organization-service --tail=50" -ForegroundColor Gray
Write-Host ""
Write-Host "✅ Ready to test!" -ForegroundColor Green
Write-Host ""

