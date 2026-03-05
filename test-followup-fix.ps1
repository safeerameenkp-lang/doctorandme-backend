# Test script to verify follow-up eligibility fix
# This script tests the patient search API with the problematic patient data

$clinicId = "f7658c53-72ae-4bd3-9960-741225ebc0a2"
$doctorId = "ef378478-1091-472e-af40-1655e77985b3"
$departmentId = "ad958b90-d383-4478-bfe3-08b53b8eeef7"
$searchTerm = "qw"

Write-Host "🔍 Testing Follow-Up Eligibility Fix" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Green

# Test the patient search API with doctor and department context
$url = "http://localhost:8080/api/organizations/clinic-specific-patients?clinic_id=$clinicId&search=$searchTerm&doctor_id=$doctorId&department_id=$departmentId"

Write-Host "📡 Testing API: $url" -ForegroundColor Yellow

try {
    $response = Invoke-RestMethod -Uri $url -Method GET -ContentType "application/json"
    
    Write-Host "✅ API Response received" -ForegroundColor Green
    Write-Host "📊 Total patients found: $($response.total)" -ForegroundColor Cyan
    
    if ($response.patients.Count -gt 0) {
        $patient = $response.patients[0]
        Write-Host "👤 Patient: $($patient.first_name) $($patient.last_name)" -ForegroundColor Cyan
        Write-Host "📱 Phone: $($patient.phone)" -ForegroundColor Cyan
        
        if ($patient.follow_up_eligibility) {
            Write-Host "🎯 Follow-up Eligibility:" -ForegroundColor Yellow
            Write-Host "   - Eligible: $($patient.follow_up_eligibility.eligible)" -ForegroundColor White
            Write-Host "   - Is Free: $($patient.follow_up_eligibility.is_free)" -ForegroundColor White
            Write-Host "   - Status Label: $($patient.follow_up_eligibility.status_label)" -ForegroundColor White
            Write-Host "   - Color Code: $($patient.follow_up_eligibility.color_code)" -ForegroundColor White
            Write-Host "   - Message: $($patient.follow_up_eligibility.message)" -ForegroundColor White
            
            if ($patient.follow_up_eligibility.days_remaining) {
                Write-Host "   - Days Remaining: $($patient.follow_up_eligibility.days_remaining)" -ForegroundColor White
            }
        } else {
            Write-Host "❌ No follow-up eligibility data found" -ForegroundColor Red
        }
        
        Write-Host "📋 Appointment History:" -ForegroundColor Yellow
        Write-Host "   - Total Appointments: $($patient.total_appointments)" -ForegroundColor White
        
        if ($patient.appointments) {
            foreach ($apt in $patient.appointments) {
                Write-Host "   - Appointment: $($apt.appointment_id)" -ForegroundColor White
                Write-Host "     Date: $($apt.appointment_date)" -ForegroundColor White
                Write-Host "     Status: $($apt.status)" -ForegroundColor White
                Write-Host "     Days Since: $($apt.days_since)" -ForegroundColor White
                Write-Host "     Follow-up Eligible: $($apt.follow_up_eligible)" -ForegroundColor White
                Write-Host "     Free Follow-up Used: $($apt.free_follow_up_used)" -ForegroundColor White
                Write-Host ""
            }
        }
        
        Write-Host "🎯 Eligible Follow-ups:" -ForegroundColor Yellow
        if ($patient.eligible_follow_ups) {
            Write-Host "   - Count: $($patient.eligible_follow_ups.Count)" -ForegroundColor White
            foreach ($followup in $patient.eligible_follow_ups) {
                Write-Host "   - Doctor: $($followup.doctor_name)" -ForegroundColor White
                Write-Host "     Department: $($followup.department)" -ForegroundColor White
                Write-Host "     Remaining Days: $($followup.remaining_days)" -ForegroundColor White
                Write-Host "     Expiry: $($followup.next_followup_expiry)" -ForegroundColor White
            }
        } else {
            Write-Host "   - No eligible follow-ups found" -ForegroundColor White
        }
        
    } else {
        Write-Host "❌ No patients found" -ForegroundColor Red
    }
    
} catch {
    Write-Host "❌ Error testing API: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Response: $($_.Exception.Response)" -ForegroundColor Red
}

Write-Host ""
Write-Host "🔍 Expected Results:" -ForegroundColor Green
Write-Host "- Patient should be found" -ForegroundColor White
Write-Host "- Follow-up eligibility should show 'eligible: true' and 'is_free: true'" -ForegroundColor White
Write-Host "- Status label should be 'free' with green color" -ForegroundColor White
Write-Host "- Should show remaining days for free follow-up" -ForegroundColor White
