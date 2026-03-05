# Follow-Up Renewal Diagnostic Script
# Run this to check if the renewal system is working

Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "  Follow-Up Renewal Diagnostic" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""

# Check if services are running
Write-Host "🔍 Checking service status..." -ForegroundColor Yellow
$appointmentStatus = docker-compose ps appointment-service --format "table {{.State}}"
$organizationStatus = docker-compose ps organization-service --format "table {{.State}}"

Write-Host "Appointment Service: $appointmentStatus" -ForegroundColor White
Write-Host "Organization Service: $organizationStatus" -ForegroundColor White
Write-Host ""

# Check recent logs
Write-Host "📋 Recent Appointment Service Logs:" -ForegroundColor Cyan
docker-compose logs appointment-service --tail=20 | Select-String -Pattern "Creating follow-up|Renewed|Created follow-up eligibility|Warning"
Write-Host ""

Write-Host "📋 Recent Organization Service Logs:" -ForegroundColor Cyan
docker-compose logs organization-service --tail=20 | Select-String -Pattern "status_label|color_code|Follow-up eligibility"
Write-Host ""

# Test API endpoints
Write-Host "🧪 Testing API endpoints..." -ForegroundColor Yellow
Write-Host ""

Write-Host "1. Test Appointment Service:" -ForegroundColor White
Write-Host "   curl -X GET http://localhost:3001/api/appointments/health" -ForegroundColor Gray
Write-Host ""

Write-Host "2. Test Organization Service:" -ForegroundColor White
Write-Host "   curl -X GET http://localhost:3002/api/organizations/health" -ForegroundColor Gray
Write-Host ""

Write-Host "3. Test Patient List API:" -ForegroundColor White
Write-Host "   curl -H 'Authorization: Bearer YOUR_TOKEN' 'http://localhost:3002/api/organizations/clinic-specific-patients?clinic_id=xxx&doctor_id=yyy&department_id=zzz'" -ForegroundColor Gray
Write-Host ""

# Database check instructions
Write-Host "🗄️ Database Check:" -ForegroundColor Yellow
Write-Host "Run the SQL queries in debug-followup-renewal.sql to check:" -ForegroundColor White
Write-Host "   - Active follow-ups" -ForegroundColor Gray
Write-Host "   - Appointment history" -ForegroundColor Gray
Write-Host "   - Renewal status" -ForegroundColor Gray
Write-Host ""

# Success indicators
Write-Host "✅ Success Indicators:" -ForegroundColor Green
Write-Host "   - Services show 'Up' status" -ForegroundColor Gray
Write-Host "   - Logs show 'Creating follow-up' messages" -ForegroundColor Gray
Write-Host "   - Logs show 'Renewed X existing follow-up(s)'" -ForegroundColor Gray
Write-Host "   - No 'Warning: Failed to renew' messages" -ForegroundColor Gray
Write-Host "   - API returns status_label and color_code fields" -ForegroundColor Gray
Write-Host ""

# Failure indicators
Write-Host "❌ Failure Indicators:" -ForegroundColor Red
Write-Host "   - Services show 'Exit' or 'Error' status" -ForegroundColor Gray
Write-Host "   - Logs show 'Warning: Failed to renew existing follow-ups'" -ForegroundColor Gray
Write-Host "   - API returns old format without status_label" -ForegroundColor Gray
Write-Host "   - Follow-ups not being created" -ForegroundColor Gray
Write-Host ""

Write-Host "📚 Documentation:" -ForegroundColor Yellow
Write-Host "   - FOLLOWUP_RENEWAL_COMPLETE_FIX.md" -ForegroundColor Gray
Write-Host "   - debug-followup-renewal.sql" -ForegroundColor Gray
Write-Host ""

Write-Host "✅ Diagnostic complete!" -ForegroundColor Green
Write-Host ""
