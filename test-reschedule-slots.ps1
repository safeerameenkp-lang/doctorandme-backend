# Test script for Reschedule Slot Fix
# This script tests that slots show as available during reschedule

# Configuration
$BASE_URL = "http://localhost:8082/api/v1"
$TOKEN = "your-auth-token-here"

Write-Host "=== Testing Reschedule Slot Fix ===" -ForegroundColor Cyan

# Test 1: Get slots for new appointment (should show booked slots as red)
Write-Host "`n=== Test 1: New Appointment (Normal Mode) ===" -ForegroundColor Yellow
$newAppointmentUrl = "$BASE_URL/doctor-session-slots?doctor_id=DOCTOR_ID&clinic_id=CLINIC_ID&date=2024-10-17&slot_type=offline"

$headers = @{
    "Authorization" = "Bearer $TOKEN"
    "Content-Type" = "application/json"
}

try {
    Write-Host "Calling: $newAppointmentUrl" -ForegroundColor Gray
    $response = Invoke-RestMethod -Uri $newAppointmentUrl -Method GET -Headers $headers
    
    Write-Host "✓ New Appointment Response:" -ForegroundColor Green
    Write-Host "Total slots found: $($response.total)"
    
    if ($response.slots -and $response.slots.Count -gt 0) {
        $firstSession = $response.slots[0].sessions[0]
        if ($firstSession.slots -and $firstSession.slots.Count -gt 0) {
            $firstSlot = $firstSession.slots[0]
            Write-Host "First slot: $($firstSlot.slot_start) - $($firstSlot.slot_end)"
            Write-Host "  Is Booked: $($firstSlot.is_booked)"
            Write-Host "  Is Bookable: $($firstSlot.is_bookable)"
            Write-Host "  Display Message: $($firstSlot.display_message)"
            Write-Host "  Available Count: $($firstSlot.available_count)/$($firstSlot.max_patients)"
        }
    }
} catch {
    Write-Host "✗ Error in new appointment test:" -ForegroundColor Red
    Write-Host $_.Exception.Message
}

# Test 2: Get slots for reschedule (should show current slot as available)
Write-Host "`n=== Test 2: Reschedule Mode (With appointment_id) ===" -ForegroundColor Yellow
$appointmentId = "APPOINTMENT_ID_TO_RESCHEDULE"
$rescheduleUrl = "$BASE_URL/doctor-session-slots?doctor_id=DOCTOR_ID&clinic_id=CLINIC_ID&date=2024-10-17&slot_type=offline&appointment_id=$appointmentId"

try {
    Write-Host "Calling: $rescheduleUrl" -ForegroundColor Gray
    $response = Invoke-RestMethod -Uri $rescheduleUrl -Method GET -Headers $headers
    
    Write-Host "✓ Reschedule Response:" -ForegroundColor Green
    Write-Host "Total slots found: $($response.total)"
    Write-Host "Appointment ID: $($response.appointment_id)"
    
    if ($response.slots -and $response.slots.Count -gt 0) {
        $firstSession = $response.slots[0].sessions[0]
        if ($firstSession.slots -and $firstSession.slots.Count -gt 0) {
            $firstSlot = $firstSession.slots[0]
            Write-Host "First slot: $($firstSlot.slot_start) - $($firstSlot.slot_end)"
            Write-Host "  Is Booked: $($firstSlot.is_booked)"
            Write-Host "  Is Bookable: $($firstSlot.is_bookable)"
            Write-Host "  Display Message: $($firstSlot.display_message)"
            Write-Host "  Available Count: $($firstSlot.available_count)/$($firstSlot.max_patients)"
            
            # Check if this is the current appointment's slot
            if ($firstSlot.booked_appointment_id -eq $appointmentId) {
                Write-Host "  🎯 This is the current appointment's slot - should be available!" -ForegroundColor Cyan
                if ($firstSlot.is_bookable -eq $true) {
                    Write-Host "  ✅ SUCCESS: Current slot is now bookable!" -ForegroundColor Green
                } else {
                    Write-Host "  ❌ FAIL: Current slot is still not bookable" -ForegroundColor Red
                }
            }
        }
    }
} catch {
    Write-Host "✗ Error in reschedule test:" -ForegroundColor Red
    Write-Host $_.Exception.Message
}

# Test 3: Compare the same slot in both modes
Write-Host "`n=== Test 3: Compare Same Slot in Both Modes ===" -ForegroundColor Yellow

try {
    # Get normal mode
    $normalResponse = Invoke-RestMethod -Uri $newAppointmentUrl -Method GET -Headers $headers
    
    # Get reschedule mode
    $rescheduleResponse = Invoke-RestMethod -Uri $rescheduleUrl -Method GET -Headers $headers
    
    if ($normalResponse.slots -and $rescheduleResponse.slots -and 
        $normalResponse.slots.Count -gt 0 -and $rescheduleResponse.slots.Count -gt 0) {
        
        $normalSlot = $normalResponse.slots[0].sessions[0].slots[0]
        $rescheduleSlot = $rescheduleResponse.slots[0].sessions[0].slots[0]
        
        Write-Host "Comparing same slot in both modes:" -ForegroundColor Gray
        Write-Host "  Normal Mode - Is Bookable: $($normalSlot.is_bookable), Available: $($normalSlot.available_count)"
        Write-Host "  Reschedule Mode - Is Bookable: $($rescheduleSlot.is_bookable), Available: $($rescheduleSlot.available_count)"
        
        if ($normalSlot.available_count -lt $rescheduleSlot.available_count) {
            Write-Host "  ✅ SUCCESS: Reschedule mode shows more availability!" -ForegroundColor Green
        } elseif ($normalSlot.available_count -eq $rescheduleSlot.available_count) {
            Write-Host "  ⚠️  INFO: Same availability in both modes (slot might not be the current appointment's slot)" -ForegroundColor Yellow
        } else {
            Write-Host "  ❌ ISSUE: Reschedule mode shows less availability" -ForegroundColor Red
        }
    }
} catch {
    Write-Host "✗ Error in comparison test:" -ForegroundColor Red
    Write-Host $_.Exception.Message
}

Write-Host "`n=== Test Complete ===" -ForegroundColor Cyan
Write-Host "`nExpected Results:" -ForegroundColor Yellow
Write-Host "1. Normal mode: Shows actual slot availability" -ForegroundColor White
Write-Host "2. Reschedule mode: Shows current appointment's slot as available" -ForegroundColor White
Write-Host "3. Same slot should show higher availability in reschedule mode" -ForegroundColor White
