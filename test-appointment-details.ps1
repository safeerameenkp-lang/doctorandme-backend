# Test script for Get Single Appointment Details API
# GET /appointments/simple/:id

# Configuration
$BASE_URL = "http://localhost:8082/api/v1"
$TOKEN = "your-auth-token-here"

# Test 1: Get appointment details by ID
Write-Host "`n=== Test 1: Get Appointment Details ===" -ForegroundColor Cyan
$appointmentId = "your-appointment-id-here"

$headers = @{
    "Authorization" = "Bearer $TOKEN"
    "Content-Type" = "application/json"
}

try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/appointments/simple/$appointmentId" `
        -Method GET `
        -Headers $headers

    Write-Host "✓ Success!" -ForegroundColor Green
    Write-Host "`nAppointment Details:" -ForegroundColor Yellow
    Write-Host "ID: $($response.appointment.id)"
    Write-Host "Booking Number: $($response.appointment.booking_number)"
    Write-Host "Token Number: $($response.appointment.token_number)"
    Write-Host "Patient: $($response.appointment.patient.name) ($($response.appointment.mo_id))"
    Write-Host "Doctor: $($response.appointment.doctor.name)"
    Write-Host "Department: $($response.appointment.department.name)"
    Write-Host "Date/Time: $($response.appointment.appointment_date_time)"
    Write-Host "Status: $($response.appointment.status)"
    Write-Host "Payment: $($response.appointment.payment_status)"
    
    if ($response.appointment.slot_details) {
        Write-Host "`nSlot Details:" -ForegroundColor Cyan
        Write-Host "Slot ID: $($response.appointment.slot_details.slot_id)"
        Write-Host "Slot Time: $($response.appointment.slot_details.slot_full_time)"
        Write-Host "Session: $($response.appointment.slot_details.session_name)"
        Write-Host "Slot Status: $($response.appointment.slot_details.slot_status)"
    } else {
        Write-Host "`nNo slot details (appointment not booked via slot system)" -ForegroundColor Yellow
    }
    
    Write-Host "`nFull Response:" -ForegroundColor Gray
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Host "✗ Error:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    if ($_.ErrorDetails.Message) {
        Write-Host $_.ErrorDetails.Message
    }
}

# Test 2: Get non-existent appointment (should return 404)
Write-Host "`n=== Test 2: Get Non-Existent Appointment ===" -ForegroundColor Cyan
$invalidId = "00000000-0000-0000-0000-000000000000"

try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/appointments/simple/$invalidId" `
        -Method GET `
        -Headers $headers

    Write-Host "Response:" -ForegroundColor Yellow
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Host "✓ Expected error (404):" -ForegroundColor Green
    Write-Host $_.Exception.Message
}

Write-Host "`n=== Tests Complete ===" -ForegroundColor Cyan

