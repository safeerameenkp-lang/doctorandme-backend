# ===================================================================
# Doctor Time Slots API Complete Test Script
# ===================================================================

$BASE_URL = "http://localhost:8081/api"
$AUTH_URL = "http://localhost:8080/api/auth"

Write-Host "=================================================="  -ForegroundColor Cyan
Write-Host "DOCTOR TIME SLOTS API - COMPLETE TEST SUITE" -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host ""

# Login
Write-Host "STEP 1: Authenticating..." -ForegroundColor Cyan
$loginBody = @{
    username = "superadmin"
    password = "Admin@123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$AUTH_URL/login" -Method POST -Body $loginBody -ContentType "application/json"
    $TOKEN = $loginResponse.access_token
    Write-Host "✓ Authentication successful" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "✗ Authentication failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $TOKEN"
    "Content-Type" = "application/json"
}

# Create Time Slots
Write-Host "STEP 2: Creating Time Slots..." -ForegroundColor Cyan
$createBody = @{
    doctor_id = "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d"
    clinic_id = "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2"
    slot_type = "offline"
    date = "2024-10-15"
    slots = @(
        @{
            start_time = "09:00"
            end_time = "12:00"
            max_patients = 10
            notes = "Morning shift"
        },
        @{
            start_time = "14:00"
            end_time = "17:00"
            max_patients = 10
            notes = "Afternoon shift"
        }
    )
} | ConvertTo-Json -Depth 10

try {
    $createResponse = Invoke-RestMethod -Uri "$BASE_URL/doctor-time-slots" -Method POST -Body $createBody -Headers $headers
    Write-Host "✓ Created $($createResponse.total_created) slots" -ForegroundColor Green
    $SLOT_ID = $createResponse.created_slots[0].id
    Write-Host "  First Slot ID: $SLOT_ID" -ForegroundColor Gray
    Write-Host ""
} catch {
    Write-Host "✗ Create failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
}

# List Time Slots
Write-Host "STEP 3: Listing Time Slots..." -ForegroundColor Cyan
try {
    $listUrl = "$BASE_URL/doctor-time-slots?doctor_id=3fd28e6d-7f9a-4dde-8172-d14a74a9b02d&clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&date=2024-10-15"
    $listResponse = Invoke-RestMethod -Uri $listUrl -Method GET -Headers $headers
    Write-Host "✓ Found $($listResponse.total) slots" -ForegroundColor Green
    foreach ($slot in $listResponse.slots) {
        Write-Host "  - $($slot.start_time)-$($slot.end_time): $($slot.status) ($($slot.available_spots)/$($slot.max_patients) available)" -ForegroundColor Gray
    }
    Write-Host ""
} catch {
    Write-Host "✗ List failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
}

# Get Single Slot
if ($SLOT_ID) {
    Write-Host "STEP 4: Getting Single Slot..." -ForegroundColor Cyan
    try {
        $getResponse = Invoke-RestMethod -Uri "$BASE_URL/doctor-time-slots/$SLOT_ID" -Method GET -Headers $headers
        Write-Host "✓ Retrieved slot details" -ForegroundColor Green
        Write-Host "  Time: $($getResponse.slot.start_time) - $($getResponse.slot.end_time)" -ForegroundColor Gray
        Write-Host "  Status: $($getResponse.slot.status)" -ForegroundColor Gray
        Write-Host ""
    } catch {
        Write-Host "✗ Get failed: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host ""
    }

    # Update Slot
    Write-Host "STEP 5: Updating Slot..." -ForegroundColor Cyan
    $updateBody = @{
        slot_type = "online"
        max_patients = 15
        notes = "Updated to online"
    } | ConvertTo-Json

    try {
        $updateResponse = Invoke-RestMethod -Uri "$BASE_URL/doctor-time-slots/$SLOT_ID" -Method PUT -Body $updateBody -Headers $headers
        Write-Host "✓ Slot updated successfully" -ForegroundColor Green
        Write-Host "  New type: $($updateResponse.slot.slot_type)" -ForegroundColor Gray
        Write-Host "  New max: $($updateResponse.slot.max_patients)" -ForegroundColor Gray
        Write-Host ""
    } catch {
        Write-Host "✗ Update failed: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host ""
    }

    # Delete Slot
    Write-Host "STEP 6: Deleting Slot..." -ForegroundColor Cyan
    try {
        $deleteResponse = Invoke-RestMethod -Uri "$BASE_URL/doctor-time-slots/$SLOT_ID" -Method DELETE -Headers $headers
        Write-Host "✓ Slot deleted successfully" -ForegroundColor Green
        Write-Host ""
    } catch {
        Write-Host "✗ Delete failed: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host ""
    }
}

# Summary
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "TEST SUITE COMPLETED" -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Endpoints Tested:" -ForegroundColor Yellow
Write-Host "  1. POST   /doctor-time-slots      - Create (bulk)" -ForegroundColor Green
Write-Host "  2. GET    /doctor-time-slots      - List (filtered)" -ForegroundColor Green
Write-Host "  3. GET    /doctor-time-slots/:id  - Get single" -ForegroundColor Green
Write-Host "  4. PUT    /doctor-time-slots/:id  - Update" -ForegroundColor Green
Write-Host "  5. DELETE /doctor-time-slots/:id  - Delete (soft)" -ForegroundColor Green
Write-Host ""

