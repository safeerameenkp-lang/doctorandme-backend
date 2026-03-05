# Test script for Doctor Token System
# This demonstrates token generation behavior

$BASE_URL = "http://localhost:8001"

Write-Host "=== Testing Doctor Token System ===" -ForegroundColor Cyan
Write-Host ""

# You'll need to replace these with actual IDs from your database
$DOCTOR_A_ID = "your-doctor-a-uuid"
$CLINIC_X_ID = "your-clinic-x-uuid"
$CLINIC_Y_ID = "your-clinic-y-uuid"
$PATIENT_1_ID = "your-patient-1-uuid"
$PATIENT_2_ID = "your-patient-2-uuid"
$PATIENT_3_ID = "your-patient-3-uuid"
$PATIENT_4_ID = "your-patient-4-uuid"

Write-Host "Replace the UUIDs in this script with real IDs from your database" -ForegroundColor Yellow
Write-Host ""

# Get auth token first
$LOGIN_URL = "http://localhost:8000/api/auth/login"
$LOGIN_BODY = @{
    email = "admin@example.com"
    password = "your-password"
} | ConvertTo-Json

Write-Host "Step 1: Login to get auth token..." -ForegroundColor Green
try {
    $loginResponse = Invoke-RestMethod -Uri $LOGIN_URL -Method POST -Body $LOGIN_BODY -ContentType "application/json"
    $TOKEN = $loginResponse.access_token
    Write-Host "✓ Login successful" -ForegroundColor Green
} catch {
    Write-Host "✗ Login failed: $_" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $TOKEN"
    "Content-Type" = "application/json"
}

Write-Host ""
Write-Host "=== Test 1: Same Day, Same Doctor, Same Clinic ===" -ForegroundColor Cyan
Write-Host "Expected: Token #1, #2, #3" -ForegroundColor Yellow
Write-Host ""

# Appointment 1
Write-Host "Creating Appointment 1..." -ForegroundColor White
$body1 = @{
    doctor_id = $DOCTOR_A_ID
    clinic_id = $CLINIC_X_ID
    patient_id = $PATIENT_1_ID
    appointment_date = "2025-01-15"
    appointment_time = "2025-01-15 09:00:00"
    consultation_type = "offline"
    reason = "Checkup"
} | ConvertTo-Json

try {
    $response1 = Invoke-RestMethod -Uri "$BASE_URL/appointments" -Method POST -Headers $headers -Body $body1
    Write-Host "✓ Appointment 1 created - Token #$($response1.token_number)" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 1

# Appointment 2
Write-Host "Creating Appointment 2..." -ForegroundColor White
$body2 = @{
    doctor_id = $DOCTOR_A_ID
    clinic_id = $CLINIC_X_ID
    patient_id = $PATIENT_2_ID
    appointment_date = "2025-01-15"
    appointment_time = "2025-01-15 09:30:00"
    consultation_type = "offline"
    reason = "Follow-up"
} | ConvertTo-Json

try {
    $response2 = Invoke-RestMethod -Uri "$BASE_URL/appointments" -Method POST -Headers $headers -Body $body2
    Write-Host "✓ Appointment 2 created - Token #$($response2.token_number)" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 1

# Appointment 3
Write-Host "Creating Appointment 3..." -ForegroundColor White
$body3 = @{
    doctor_id = $DOCTOR_A_ID
    clinic_id = $CLINIC_X_ID
    patient_id = $PATIENT_3_ID
    appointment_date = "2025-01-15"
    appointment_time = "2025-01-15 10:00:00"
    consultation_type = "offline"
    reason = "Consultation"
} | ConvertTo-Json

try {
    $response3 = Invoke-RestMethod -Uri "$BASE_URL/appointments" -Method POST -Headers $headers -Body $body3
    Write-Host "✓ Appointment 3 created - Token #$($response3.token_number)" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Test 2: Same Day, Same Doctor, Different Clinic ===" -ForegroundColor Cyan
Write-Host "Expected: Token #1 (starts fresh for different clinic)" -ForegroundColor Yellow
Write-Host ""

# Appointment 4 - Different Clinic
Write-Host "Creating Appointment at different clinic..." -ForegroundColor White
$body4 = @{
    doctor_id = $DOCTOR_A_ID
    clinic_id = $CLINIC_Y_ID  # Different clinic
    patient_id = $PATIENT_4_ID
    appointment_date = "2025-01-15"  # Same day
    appointment_time = "2025-01-15 14:00:00"
    consultation_type = "offline"
    reason = "Checkup"
} | ConvertTo-Json

try {
    $response4 = Invoke-RestMethod -Uri "$BASE_URL/appointments" -Method POST -Headers $headers -Body $body4
    Write-Host "✓ Appointment at Clinic Y created - Token #$($response4.token_number)" -ForegroundColor Green
    Write-Host "   (Should be Token #1 - separate sequence)" -ForegroundColor Yellow
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Test 3: Different Day, Same Doctor, Same Clinic ===" -ForegroundColor Cyan
Write-Host "Expected: Token #1 (reset for new day)" -ForegroundColor Yellow
Write-Host ""

# Appointment 5 - Next Day
Write-Host "Creating Appointment on next day..." -ForegroundColor White
$body5 = @{
    doctor_id = $DOCTOR_A_ID
    clinic_id = $CLINIC_X_ID  # Same clinic as Test 1
    patient_id = $PATIENT_1_ID
    appointment_date = "2025-01-16"  # Next day
    appointment_time = "2025-01-16 09:00:00"
    consultation_type = "offline"
    reason = "Follow-up"
} | ConvertTo-Json

try {
    $response5 = Invoke-RestMethod -Uri "$BASE_URL/appointments" -Method POST -Headers $headers -Body $body5
    Write-Host "✓ Appointment on Day 2 created - Token #$($response5.token_number)" -ForegroundColor Green
    Write-Host "   (Should be Token #1 - daily reset)" -ForegroundColor Yellow
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Test Complete ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Summary:" -ForegroundColor White
Write-Host "- Same doctor, same clinic, same day: Sequential tokens (1, 2, 3, ...)" -ForegroundColor White
Write-Host "- Same doctor, different clinic, same day: Separate sequence (starts at 1)" -ForegroundColor White
Write-Host "- Same doctor, same clinic, different day: Reset to 1" -ForegroundColor White

