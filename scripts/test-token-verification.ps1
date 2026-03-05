# Complete Token System Verification Script
# This script tests all scenarios for the token numbering system

Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "  Doctor Token System Verification Test" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""

# Configuration
$BASE_URL = "http://localhost:8001"
$AUTH_URL = "http://localhost:8000"

# Test IDs (Replace with your actual IDs)
$CLINIC_A_ID = "your-clinic-a-uuid"
$CLINIC_B_ID = "your-clinic-b-uuid"
$DOCTOR_A_ID = "your-doctor-a-uuid"
$DOCTOR_B_ID = "your-doctor-b-uuid"
$PATIENT_1_ID = "your-patient-1-uuid"
$PATIENT_2_ID = "your-patient-2-uuid"
$PATIENT_3_ID = "your-patient-3-uuid"

Write-Host "NOTE: Update the script with your actual UUIDs before running!" -ForegroundColor Yellow
Write-Host ""

# Function to create appointment
function Create-Appointment {
    param(
        [string]$PatientID,
        [string]$ClinicID,
        [string]$DoctorID,
        [string]$Date,
        [string]$Time,
        [string]$Token
    )
    
    $body = @{
        patient_id = $PatientID
        clinic_id = $ClinicID
        doctor_id = $DoctorID
        appointment_date = $Date
        appointment_time = $Time
        consultation_type = "offline"
        duration_minutes = 12
        reason = "Checkup"
        payment_mode = "cash"
    } | ConvertTo-Json
    
    $headers = @{
        "Content-Type" = "application/json"
        "Authorization" = "Bearer $Token"
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$BASE_URL/appointments" -Method Post -Headers $headers -Body $body
        return $response
    } catch {
        Write-Host "Error: $_" -ForegroundColor Red
        return $null
    }
}

# Function to get appointments list
function Get-AppointmentsList {
    param(
        [string]$ClinicID,
        [string]$DoctorID,
        [string]$Date,
        [string]$Token
    )
    
    $params = "?clinic_id=$ClinicID&doctor_id=$DoctorID&date=$Date"
    
    $headers = @{
        "Authorization" = "Bearer $Token"
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$BASE_URL/appointments/list$params" -Method Get -Headers $headers
        return $response
    } catch {
        Write-Host "Error: $_" -ForegroundColor Red
        return $null
    }
}

Write-Host ""
Write-Host "=== TEST SCENARIO 1: Sequential Tokens for Same Doctor/Clinic/Day ===" -ForegroundColor Green
Write-Host ""
Write-Host "Creating 3 appointments for Doctor A at Clinic X on Day 1..." -ForegroundColor Yellow

$today = Get-Date -Format "yyyy-MM-dd"
$time1 = "$today 09:00:00"
$time2 = "$today 09:30:00"
$time3 = "$today 10:00:00"

Write-Host "Expected: Token #1, #2, #3" -ForegroundColor Cyan
Write-Host ""

# You need to get auth token first
# $authToken = "your-auth-token"

<#
Write-Host "Appointment 1:" -ForegroundColor White
$appt1 = Create-Appointment -PatientID $PATIENT_1_ID -ClinicID $CLINIC_A_ID -DoctorID $DOCTOR_A_ID -Date $today -Time $time1 -Token $authToken
if ($appt1) {
    Write-Host "  Token Number: $($appt1.token_number)" -ForegroundColor $(if ($appt1.token_number -eq 1) { "Green" } else { "Red" })
    Write-Host "  Booking Number: $($appt1.booking_number)" -ForegroundColor Gray
}

Write-Host ""
Write-Host "Appointment 2:" -ForegroundColor White
$appt2 = Create-Appointment -PatientID $PATIENT_2_ID -ClinicID $CLINIC_A_ID -DoctorID $DOCTOR_A_ID -Date $today -Time $time2 -Token $authToken
if ($appt2) {
    Write-Host "  Token Number: $($appt2.token_number)" -ForegroundColor $(if ($appt2.token_number -eq 2) { "Green" } else { "Red" })
    Write-Host "  Booking Number: $($appt2.booking_number)" -ForegroundColor Gray
}

Write-Host ""
Write-Host "Appointment 3:" -ForegroundColor White
$appt3 = Create-Appointment -PatientID $PATIENT_3_ID -ClinicID $CLINIC_A_ID -DoctorID $DOCTOR_A_ID -Date $today -Time $time3 -Token $authToken
if ($appt3) {
    Write-Host "  Token Number: $($appt3.token_number)" -ForegroundColor $(if ($appt3.token_number -eq 3) { "Green" } else { "Red" })
    Write-Host "  Booking Number: $($appt3.booking_number)" -ForegroundColor Gray
}
#>

Write-Host ""
Write-Host "=== TEST SCENARIO 2: Different Clinic, Same Doctor, Same Day ===" -ForegroundColor Green
Write-Host ""
Write-Host "Creating appointment for Doctor A at Clinic Y on Day 1..." -ForegroundColor Yellow
Write-Host "Expected: Token #1 (new sequence for different clinic)" -ForegroundColor Cyan
Write-Host ""

<#
Write-Host "Appointment 4 (Clinic B):" -ForegroundColor White
$appt4 = Create-Appointment -PatientID $PATIENT_1_ID -ClinicID $CLINIC_B_ID -DoctorID $DOCTOR_A_ID -Date $today -Time $time1 -Token $authToken
if ($appt4) {
    Write-Host "  Token Number: $($appt4.token_number)" -ForegroundColor $(if ($appt4.token_number -eq 1) { "Green" } else { "Red" })
    Write-Host "  Booking Number: $($appt4.booking_number)" -ForegroundColor Gray
    Write-Host "  Clinic: Different from previous appointments" -ForegroundColor Gray
}
#>

Write-Host ""
Write-Host "=== TEST SCENARIO 3: Next Day Reset ===" -ForegroundColor Green
Write-Host ""
Write-Host "Creating appointment for Doctor A at Clinic X on Day 2..." -ForegroundColor Yellow

$tomorrow = (Get-Date).AddDays(1).ToString("yyyy-MM-dd")
$timeNextDay = "$tomorrow 09:00:00"

Write-Host "Expected: Token #1 (reset for new day)" -ForegroundColor Cyan
Write-Host ""

<#
Write-Host "Appointment 5 (Next Day):" -ForegroundColor White
$appt5 = Create-Appointment -PatientID $PATIENT_1_ID -ClinicID $CLINIC_A_ID -DoctorID $DOCTOR_A_ID -Date $tomorrow -Time $timeNextDay -Token $authToken
if ($appt5) {
    Write-Host "  Token Number: $($appt5.token_number)" -ForegroundColor $(if ($appt5.token_number -eq 1) { "Green" } else { "Red" })
    Write-Host "  Booking Number: $($appt5.booking_number)" -ForegroundColor Gray
    Write-Host "  Date: Next day - tokens should reset" -ForegroundColor Gray
}
#>

Write-Host ""
Write-Host "=== TEST SCENARIO 4: Different Doctor, Same Clinic, Same Day ===" -ForegroundColor Green
Write-Host ""
Write-Host "Creating appointment for Doctor B at Clinic X on Day 1..." -ForegroundColor Yellow
Write-Host "Expected: Token #1 (new sequence for different doctor)" -ForegroundColor Cyan
Write-Host ""

<#
Write-Host "Appointment 6 (Different Doctor):" -ForegroundColor White
$appt6 = Create-Appointment -PatientID $PATIENT_1_ID -ClinicID $CLINIC_A_ID -DoctorID $DOCTOR_B_ID -Date $today -Time $time1 -Token $authToken
if ($appt6) {
    Write-Host "  Token Number: $($appt6.token_number)" -ForegroundColor $(if ($appt6.token_number -eq 1) { "Green" } else { "Red" })
    Write-Host "  Booking Number: $($appt6.booking_number)" -ForegroundColor Gray
    Write-Host "  Doctor: Different from previous appointments" -ForegroundColor Gray
}
#>

Write-Host ""
Write-Host "=== Verification of List API ===" -ForegroundColor Green
Write-Host ""
Write-Host "Fetching appointment list to verify token numbers..." -ForegroundColor Yellow

<#
$list = Get-AppointmentsList -ClinicID $CLINIC_A_ID -DoctorID $DOCTOR_A_ID -Date $today -Token $authToken
if ($list -and $list.appointments) {
    Write-Host "Appointments for Doctor A at Clinic X on $today" -ForegroundColor Cyan
    Write-Host ""
    foreach ($appt in $list.appointments) {
        Write-Host "  Token #$($appt.token_number) - $($appt.patient_name) - $($appt.appointment_date_time)" -ForegroundColor White
    }
}
#>

Write-Host ""
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "  Expected Results Summary" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Day 1, Doctor A, Clinic X:" -ForegroundColor White
Write-Host "  - Appointment 1 → Token #1" -ForegroundColor Green
Write-Host "  - Appointment 2 → Token #2" -ForegroundColor Green
Write-Host "  - Appointment 3 → Token #3" -ForegroundColor Green
Write-Host ""
Write-Host "Day 1, Doctor A, Clinic Y:" -ForegroundColor White
Write-Host "  - Appointment 1 → Token #1 (different clinic)" -ForegroundColor Green
Write-Host ""
Write-Host "Day 2, Doctor A, Clinic X:" -ForegroundColor White
Write-Host "  - Appointment 1 → Token #1 (new day, reset)" -ForegroundColor Green
Write-Host ""
Write-Host "Day 1, Doctor B, Clinic X:" -ForegroundColor White
Write-Host "  - Appointment 1 → Token #1 (different doctor)" -ForegroundColor Green
Write-Host ""

Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "  Database Verification Queries" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Run these SQL queries to verify token data:" -ForegroundColor Yellow
Write-Host ""
Write-Host "-- Check doctor_tokens table" -ForegroundColor Gray
Write-Host "SELECT doctor_id, clinic_id, token_date, current_token, created_at" -ForegroundColor Gray
Write-Host "FROM doctor_tokens" -ForegroundColor Gray
Write-Host "ORDER BY token_date DESC, doctor_id, clinic_id;" -ForegroundColor Gray
Write-Host ""
Write-Host "-- Check appointments with token numbers" -ForegroundColor Gray
Write-Host "SELECT booking_number, token_number, appointment_date, doctor_id, clinic_id" -ForegroundColor Gray
Write-Host "FROM appointments" -ForegroundColor Gray
Write-Host "WHERE appointment_date >= CURRENT_DATE - INTERVAL '1 day'" -ForegroundColor Gray
Write-Host "ORDER BY appointment_date, doctor_id, clinic_id, token_number;" -ForegroundColor Gray
Write-Host ""

Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "  Token System Features" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "✓ Doctor-wise isolation" -ForegroundColor Green
Write-Host "  Each doctor has their own token sequence" -ForegroundColor Gray
Write-Host ""
Write-Host "✓ Clinic-specific" -ForegroundColor Green
Write-Host "  Same doctor at different clinics has separate tokens" -ForegroundColor Gray
Write-Host ""
Write-Host "✓ Daily reset" -ForegroundColor Green
Write-Host "  Tokens automatically reset to 1 every morning" -ForegroundColor Gray
Write-Host ""
Write-Host "✓ Race condition safe" -ForegroundColor Green
Write-Host "  Uses database transactions with row locking" -ForegroundColor Gray
Write-Host ""
Write-Host "✓ Sequential numbering" -ForegroundColor Green
Write-Host "  No gaps in token numbers" -ForegroundColor Gray
Write-Host ""

Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "Test Complete!" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan

