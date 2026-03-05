# Test Reschedule Appointment API
# This script tests the reschedule appointment functionality

$baseUrl = "http://localhost:8080/api/appointment-service"
$headers = @{
    "Content-Type" = "application/json"
    "Authorization" = "Bearer YOUR_JWT_TOKEN_HERE"  # Replace with actual token
}

Write-Host "=== Testing Reschedule Appointment API ===" -ForegroundColor Green

# Test 1: Reschedule appointment with valid data
Write-Host "`n1. Testing reschedule appointment with valid data..." -ForegroundColor Yellow

$appointmentId = "YOUR_APPOINTMENT_ID_HERE"  # Replace with actual appointment ID
$rescheduleData = @{
    department_id = "YOUR_DEPARTMENT_ID_HERE"  # Optional, replace with actual department ID
    doctor_id = "YOUR_DOCTOR_ID_HERE"  # Replace with actual doctor ID
    clinic_id = "YOUR_CLINIC_ID_HERE"  # Replace with actual clinic ID
    individual_slot_id = "YOUR_SLOT_ID_HERE"  # Replace with actual slot ID
    appointment_date = "2024-07-20"  # New date
    appointment_time = "2024-07-20 10:30:00"  # New time
    reason = "Patient requested time change"
    notes = "Rescheduled due to patient availability"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/appointments/$appointmentId/reschedule" -Method POST -Headers $headers -Body $rescheduleData
    Write-Host "✅ Reschedule successful!" -ForegroundColor Green
    Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Reschedule failed!" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response Body: $responseBody" -ForegroundColor Red
    }
}

# Test 2: Reschedule with invalid appointment ID
Write-Host "`n2. Testing reschedule with invalid appointment ID..." -ForegroundColor Yellow

$invalidAppointmentId = "00000000-0000-0000-0000-000000000000"
$invalidRescheduleData = @{
    doctor_id = "YOUR_DOCTOR_ID_HERE"
    clinic_id = "YOUR_CLINIC_ID_HERE"
    individual_slot_id = "YOUR_SLOT_ID_HERE"
    appointment_date = "2024-07-20"
    appointment_time = "2024-07-20 10:30:00"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/appointments/$invalidAppointmentId/reschedule" -Method POST -Headers $headers -Body $invalidRescheduleData
    Write-Host "❌ Should have failed!" -ForegroundColor Red
} catch {
    Write-Host "✅ Correctly failed with invalid appointment ID" -ForegroundColor Green
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Cyan
}

# Test 3: Reschedule with invalid slot (fully booked)
Write-Host "`n3. Testing reschedule with invalid slot..." -ForegroundColor Yellow

$rescheduleDataInvalidSlot = @{
    doctor_id = "YOUR_DOCTOR_ID_HERE"
    clinic_id = "YOUR_CLINIC_ID_HERE"
    individual_slot_id = "INVALID_SLOT_ID_HERE"  # Replace with actual invalid slot ID
    appointment_date = "2024-07-20"
    appointment_time = "2024-07-20 10:30:00"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/appointments/$appointmentId/reschedule" -Method POST -Headers $headers -Body $rescheduleDataInvalidSlot
    Write-Host "❌ Should have failed!" -ForegroundColor Red
} catch {
    Write-Host "✅ Correctly failed with invalid slot" -ForegroundColor Green
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Cyan
}

# Test 4: Reschedule with missing required fields
Write-Host "`n4. Testing reschedule with missing required fields..." -ForegroundColor Yellow

$incompleteRescheduleData = @{
    doctor_id = "YOUR_DOCTOR_ID_HERE"
    # Missing clinic_id, individual_slot_id, appointment_date, appointment_time
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/appointments/$appointmentId/reschedule" -Method POST -Headers $headers -Body $incompleteRescheduleData
    Write-Host "❌ Should have failed!" -ForegroundColor Red
} catch {
    Write-Host "✅ Correctly failed with missing required fields" -ForegroundColor Green
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Cyan
}

Write-Host "`n=== Reschedule Appointment API Testing Complete ===" -ForegroundColor Green
Write-Host "`n📝 Instructions:" -ForegroundColor Yellow
Write-Host "1. Replace 'YOUR_JWT_TOKEN_HERE' with a valid JWT token" -ForegroundColor White
Write-Host "2. Replace 'YOUR_APPOINTMENT_ID_HERE' with a valid appointment ID" -ForegroundColor White
Write-Host "3. Replace 'YOUR_DOCTOR_ID_HERE' with a valid doctor ID" -ForegroundColor White
Write-Host "4. Replace 'YOUR_CLINIC_ID_HERE' with a valid clinic ID" -ForegroundColor White
Write-Host "5. Replace 'YOUR_SLOT_ID_HERE' with a valid available slot ID" -ForegroundColor White
Write-Host "6. Replace 'YOUR_DEPARTMENT_ID_HERE' with a valid department ID (optional)" -ForegroundColor White
Write-Host "7. Make sure your appointment service is running on localhost:8080" -ForegroundColor White
