# Test Reschedule Appointment with Slot Re-enabling
# This script tests that when an appointment is rescheduled, the old slot becomes available again

$baseUrl = "http://localhost:8080/api/appointment-service"
$organizationUrl = "http://localhost:8080/api/organization-service"
$headers = @{
    "Content-Type" = "application/json"
    "Authorization" = "Bearer YOUR_JWT_TOKEN_HERE"  # Replace with actual token
}

Write-Host "=== Testing Reschedule with Slot Re-enabling ===" -ForegroundColor Green

# Step 1: Get current slots for a doctor to see available slots
Write-Host "`n1. Getting current doctor slots..." -ForegroundColor Yellow

$doctorId = "YOUR_DOCTOR_ID_HERE"  # Replace with actual doctor ID
$clinicId = "YOUR_CLINIC_ID_HERE"  # Replace with actual clinic ID
$date = "2024-07-20"  # Replace with actual date

try {
    $slotsResponse = Invoke-RestMethod -Uri "$organizationUrl/doctor-session-slots?doctor_id=$doctorId&clinic_id=$clinicId&date=$date" -Method GET -Headers $headers
    Write-Host "✅ Slots retrieved successfully!" -ForegroundColor Green
    
    # Find an available slot
    $availableSlot = $null
    foreach ($timeSlot in $slotsResponse) {
        foreach ($session in $timeSlot.sessions) {
            foreach ($slot in $session.slots) {
                if ($slot.is_bookable -eq $true -and $slot.status -eq "available") {
                    $availableSlot = $slot
                    break
                }
            }
            if ($availableSlot) { break }
        }
        if ($availableSlot) { break }
    }
    
    if ($availableSlot) {
        Write-Host "Found available slot: $($availableSlot.id) - $($availableSlot.slot_start) to $($availableSlot.slot_end)" -ForegroundColor Cyan
        Write-Host "Available count: $($availableSlot.available_count)/$($availableSlot.max_patients)" -ForegroundColor Cyan
    } else {
        Write-Host "❌ No available slots found!" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "❌ Failed to get slots!" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Step 2: Create a test appointment
Write-Host "`n2. Creating test appointment..." -ForegroundColor Yellow

$appointmentData = @{
    clinic_patient_id = "YOUR_CLINIC_PATIENT_ID_HERE"  # Replace with actual clinic patient ID
    doctor_id = $doctorId
    clinic_id = $clinicId
    department_id = "YOUR_DEPARTMENT_ID_HERE"  # Optional
    individual_slot_id = $availableSlot.id
    appointment_date = $date
    appointment_time = "$date $($availableSlot.slot_start):00"
    consultation_type = "offline"
    reason = "Test appointment for reschedule"
    notes = "Testing slot re-enabling"
    payment_method = "pay_now"
    payment_type = "cash"
} | ConvertTo-Json

try {
    $appointmentResponse = Invoke-RestMethod -Uri "$baseUrl/appointments/simple" -Method POST -Headers $headers -Body $appointmentData
    Write-Host "✅ Test appointment created!" -ForegroundColor Green
    Write-Host "Appointment ID: $($appointmentResponse.appointment.id)" -ForegroundColor Cyan
    
    $appointmentId = $appointmentResponse.appointment.id
    $originalSlotId = $availableSlot.id
} catch {
    Write-Host "❌ Failed to create test appointment!" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Step 3: Verify slot is now booked
Write-Host "`n3. Verifying slot is now booked..." -ForegroundColor Yellow

Start-Sleep -Seconds 2  # Wait a moment for database updates

try {
    $updatedSlotsResponse = Invoke-RestMethod -Uri "$organizationUrl/doctor-session-slots?doctor_id=$doctorId&clinic_id=$clinicId&date=$date" -Method GET -Headers $headers
    
    $bookedSlot = $null
    foreach ($timeSlot in $updatedSlotsResponse) {
        foreach ($session in $timeSlot.sessions) {
            foreach ($slot in $session.slots) {
                if ($slot.id -eq $originalSlotId) {
                    $bookedSlot = $slot
                    break
                }
            }
            if ($bookedSlot) { break }
        }
        if ($bookedSlot) { break }
    }
    
    if ($bookedSlot) {
        Write-Host "Slot status after booking:" -ForegroundColor Cyan
        Write-Host "  - Available count: $($bookedSlot.available_count)/$($bookedSlot.max_patients)" -ForegroundColor Cyan
        Write-Host "  - Is bookable: $($bookedSlot.is_bookable)" -ForegroundColor Cyan
        Write-Host "  - Status: $($bookedSlot.status)" -ForegroundColor Cyan
        Write-Host "  - Display message: $($bookedSlot.display_message)" -ForegroundColor Cyan
        
        if ($bookedSlot.available_count -lt $availableSlot.available_count) {
            Write-Host "✅ Slot correctly shows reduced availability!" -ForegroundColor Green
        } else {
            Write-Host "❌ Slot availability not updated!" -ForegroundColor Red
        }
    } else {
        Write-Host "❌ Could not find the booked slot!" -ForegroundColor Red
    }
} catch {
    Write-Host "❌ Failed to verify slot booking!" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Step 4: Find a different available slot for rescheduling
Write-Host "`n4. Finding different slot for rescheduling..." -ForegroundColor Yellow

$newAvailableSlot = $null
try {
    $slotsForReschedule = Invoke-RestMethod -Uri "$organizationUrl/doctor-session-slots?doctor_id=$doctorId&clinic_id=$clinicId&date=$date" -Method GET -Headers $headers
    
    foreach ($timeSlot in $slotsForReschedule) {
        foreach ($session in $timeSlot.sessions) {
            foreach ($slot in $session.slots) {
                if ($slot.id -ne $originalSlotId -and $slot.is_bookable -eq $true -and $slot.status -eq "available") {
                    $newAvailableSlot = $slot
                    break
                }
            }
            if ($newAvailableSlot) { break }
        }
        if ($newAvailableSlot) { break }
    }
    
    if ($newAvailableSlot) {
        Write-Host "Found new available slot: $($newAvailableSlot.id) - $($newAvailableSlot.slot_start) to $($newAvailableSlot.slot_end)" -ForegroundColor Cyan
    } else {
        Write-Host "❌ No other available slots found for rescheduling!" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "❌ Failed to find new slot!" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Step 5: Reschedule the appointment
Write-Host "`n5. Rescheduling appointment to new slot..." -ForegroundColor Yellow

$rescheduleData = @{
    doctor_id = $doctorId
    clinic_id = $clinicId
    department_id = "YOUR_DEPARTMENT_ID_HERE"  # Optional
    individual_slot_id = $newAvailableSlot.id
    appointment_date = $date
    appointment_time = "$date $($newAvailableSlot.slot_start):00"
    reason = "Rescheduled to different time"
    notes = "Testing slot re-enabling functionality"
} | ConvertTo-Json

try {
    $rescheduleResponse = Invoke-RestMethod -Uri "$baseUrl/appointments/$appointmentId/reschedule" -Method POST -Headers $headers -Body $rescheduleData
    Write-Host "✅ Appointment rescheduled successfully!" -ForegroundColor Green
    Write-Host "Response: $($rescheduleResponse | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
    
    if ($rescheduleResponse.slot_re_enabled) {
        Write-Host "✅ Slot re-enabling information included in response!" -ForegroundColor Green
        Write-Host "Old slot ID: $($rescheduleResponse.slot_re_enabled.old_slot_id)" -ForegroundColor Cyan
    }
} catch {
    Write-Host "❌ Failed to reschedule appointment!" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response Body: $responseBody" -ForegroundColor Red
    }
    exit 1
}

# Step 6: Verify old slot is now available again
Write-Host "`n6. Verifying old slot is now available again..." -ForegroundColor Yellow

Start-Sleep -Seconds 2  # Wait for database updates

try {
    $finalSlotsResponse = Invoke-RestMethod -Uri "$organizationUrl/doctor-session-slots?doctor_id=$doctorId&clinic_id=$clinicId&date=$date" -Method GET -Headers $headers
    
    $reEnabledSlot = $null
    $newBookedSlot = $null
    
    foreach ($timeSlot in $finalSlotsResponse) {
        foreach ($session in $timeSlot.sessions) {
            foreach ($slot in $session.slots) {
                if ($slot.id -eq $originalSlotId) {
                    $reEnabledSlot = $slot
                } elseif ($slot.id -eq $newAvailableSlot.id) {
                    $newBookedSlot = $slot
                }
            }
        }
    }
    
    Write-Host "`n=== SLOT RE-ENABLING VERIFICATION ===" -ForegroundColor Green
    
    if ($reEnabledSlot) {
        Write-Host "Old slot (should be available again):" -ForegroundColor Cyan
        Write-Host "  - Slot ID: $($reEnabledSlot.id)" -ForegroundColor Cyan
        Write-Host "  - Available count: $($reEnabledSlot.available_count)/$($reEnabledSlot.max_patients)" -ForegroundColor Cyan
        Write-Host "  - Is bookable: $($reEnabledSlot.is_bookable)" -ForegroundColor Cyan
        Write-Host "  - Status: $($reEnabledSlot.status)" -ForegroundColor Cyan
        Write-Host "  - Display message: $($reEnabledSlot.display_message)" -ForegroundColor Cyan
        
        if ($reEnabledSlot.is_bookable -eq $true -and $reEnabledSlot.status -eq "available") {
            Write-Host "✅ OLD SLOT SUCCESSFULLY RE-ENABLED!" -ForegroundColor Green
        } else {
            Write-Host "❌ Old slot not properly re-enabled!" -ForegroundColor Red
        }
    } else {
        Write-Host "❌ Could not find old slot!" -ForegroundColor Red
    }
    
    if ($newBookedSlot) {
        Write-Host "`nNew slot (should be booked now):" -ForegroundColor Cyan
        Write-Host "  - Slot ID: $($newBookedSlot.id)" -ForegroundColor Cyan
        Write-Host "  - Available count: $($newBookedSlot.available_count)/$($newBookedSlot.max_patients)" -ForegroundColor Cyan
        Write-Host "  - Is bookable: $($newBookedSlot.is_bookable)" -ForegroundColor Cyan
        Write-Host "  - Status: $($newBookedSlot.status)" -ForegroundColor Cyan
        Write-Host "  - Display message: $($newBookedSlot.display_message)" -ForegroundColor Cyan
        
        if ($newBookedSlot.available_count -lt $newAvailableSlot.available_count) {
            Write-Host "✅ NEW SLOT CORRECTLY BOOKED!" -ForegroundColor Green
        } else {
            Write-Host "❌ New slot not properly booked!" -ForegroundColor Red
        }
    } else {
        Write-Host "❌ Could not find new slot!" -ForegroundColor Red
    }
    
} catch {
    Write-Host "❌ Failed to verify slot re-enabling!" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== SLOT RE-ENABLING TEST COMPLETE ===" -ForegroundColor Green
Write-Host "`n📝 Instructions:" -ForegroundColor Yellow
Write-Host "1. Replace 'YOUR_JWT_TOKEN_HERE' with a valid JWT token" -ForegroundColor White
Write-Host "2. Replace 'YOUR_DOCTOR_ID_HERE' with a valid doctor ID" -ForegroundColor White
Write-Host "3. Replace 'YOUR_CLINIC_ID_HERE' with a valid clinic ID" -ForegroundColor White
Write-Host "4. Replace 'YOUR_CLINIC_PATIENT_ID_HERE' with a valid clinic patient ID" -ForegroundColor White
Write-Host "5. Replace 'YOUR_DEPARTMENT_ID_HERE' with a valid department ID (optional)" -ForegroundColor White
Write-Host "6. Update the date variable to a date with available slots" -ForegroundColor White
Write-Host "7. Make sure both appointment-service and organization-service are running" -ForegroundColor White
