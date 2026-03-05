# Quick test for clinic doctors API
# Usage: .\test-clinic-doctors-now.ps1

param(
    [Parameter(Mandatory=$true)]
    [string]$Token,
    
    [Parameter(Mandatory=$true)]
    [string]$ClinicId,
    
    [string]$BaseUrl = "http://localhost:8081/api"
)

Write-Host "`n🧪 Testing Clinic Doctors API..." -ForegroundColor Cyan
Write-Host "Clinic ID: $ClinicId`n" -ForegroundColor Gray

try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/doctors/clinic/$ClinicId" `
        -Method GET `
        -Headers @{
            "Authorization" = "Bearer $Token"
            "Content-Type" = "application/json"
        }
    
    Write-Host "✅ SUCCESS! Found $($response.total_doctors) doctors`n" -ForegroundColor Green
    
    if ($response.total_doctors -gt 0) {
        $response.doctors | ForEach-Object {
            Write-Host "👨‍⚕️ Doctor: $($_.full_name)" -ForegroundColor White
            Write-Host "   Specialization: $($_.specialization)"
            Write-Host "   Doctor ID: $($_.doctor_id)"
            Write-Host "   Email: $($_.email)"
            Write-Host "   Phone: $($_.phone)"
            Write-Host ""
            Write-Host "   💰 Clinic Fees:"
            Write-Host "   - Offline Consultation: ₹$($_.clinic_specific_fees.consultation_fee_offline)"
            Write-Host "   - Online Consultation: ₹$($_.clinic_specific_fees.consultation_fee_online)"
            Write-Host "   - Follow-up: ₹$($_.clinic_specific_fees.follow_up_fee) ($($_.clinic_specific_fees.follow_up_days) days)"
            
            if ($_.clinic_specific_fees.notes) {
                Write-Host "   - Notes: $($_.clinic_specific_fees.notes)"
            }
            Write-Host ""
        }
    } else {
        Write-Host "ℹ️  No doctors are linked to this clinic yet." -ForegroundColor Yellow
        Write-Host "   Use POST /api/clinic-doctor-links to link doctors." -ForegroundColor Yellow
    }
    
} catch {
    Write-Host "❌ ERROR: $($_.Exception.Message)`n" -ForegroundColor Red
    
    $statusCode = $_.Exception.Response.StatusCode.value__
    
    switch ($statusCode) {
        404 { 
            Write-Host "⚠️  Possible reasons:" -ForegroundColor Yellow
            Write-Host "   1. Clinic ID not found: $ClinicId"
            Write-Host "   2. Clinic is inactive"
            Write-Host "   3. Wrong URL or endpoint"
        }
        401 { 
            Write-Host "⚠️  Authentication failed:" -ForegroundColor Yellow
            Write-Host "   1. Token is invalid or expired"
            Write-Host "   2. Missing Authorization header"
        }
        500 {
            Write-Host "⚠️  Server error - check logs:" -ForegroundColor Yellow
            Write-Host "   docker-compose logs organization-service"
        }
    }
    
    if ($_.ErrorDetails.Message) {
        Write-Host "`n📄 Server Response:" -ForegroundColor Gray
        Write-Host $_.ErrorDetails.Message
    }
}

Write-Host "`n✅ Correct URL format: $BaseUrl/doctors/clinic/{clinic-id}" -ForegroundColor Green
Write-Host "❌ Wrong URL format: $BaseUrl/organizations/doctors/{clinic-id}/doctors`n" -ForegroundColor Red



