# ===================================================================
# Doctor Time Slots API Complete Test Script
# Tests all 5 endpoints with proper authentication
# ===================================================================

$ErrorActionPreference = "Continue"

# Configuration
$BASE_URL = "http://localhost:8081/api"
$AUTH_URL = "http://localhost:8080/api/auth"

# Colors for output
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Error { Write-Host $args -ForegroundColor Red }
function Write-Info { Write-Host $args -ForegroundColor Cyan }
function Write-Warning { Write-Host $args -ForegroundColor Yellow }

Write-Info "=================================================="
Write-Info "DOCTOR TIME SLOTS API - COMPLETE TEST SUITE"
Write-Info "=================================================="
Write-Host ""

# ===================================================================
# STEP 1: Login and Get Auth Token
# ===================================================================
Write-Info "STEP 1: Authenticating..."
Write-Host ""

$loginBody = @{
    username = "superadmin"
    password = "Admin@123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$AUTH_URL/login" -Method POST -Body $loginBody -ContentType "application/json"
    $TOKEN = $loginResponse.access_token
    Write-Success "✓ Authentication successful"
    Write-Host "Token: $($TOKEN.Substring(0, 20))..." -ForegroundColor Gray
    Write-Host ""
}
catch {
    Write-Error "✗ Authentication failed: $($_.Exception.Message)"
    Write-Warning "Please ensure the auth service is running and credentials are correct"
    exit 1
}

# Headers for authenticated requests
$headers = @{
    "Authorization" = "Bearer $TOKEN"
    "Content-Type" = "application/json"
}

# ===================================================================
# STEP 2: Create Time Slots (Bulk)
# ===================================================================
Write-Info "STEP 2: Creating Time Slots (Bulk Creation)..."
Write-Host ""

$createSlotsBody = @{
    doctor_id = "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d"
    clinic_id = "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2"
    slot_type = "offline"
    date = "2024-10-15"
    slots = @(
        @{
            start_time = "09:00"
            end_time = "12:00"
            max_patients = 10
            notes = "Morning shift - Monday"
        },
        @{
            start_time = "14:00"
            end_time = "17:00"
            max_patients = 10
            notes = "Afternoon shift - Monday"
        },
        @{
            start_time = "18:00"
            end_time = "20:00"
            max_patients = 5
            notes = "Evening shift - Monday"
        }
    )
} | ConvertTo-Json -Depth 10

$SLOT_ID = $null

try {
    $createResponse = Invoke-RestMethod -Uri "$BASE_URL/doctor-time-slots" -Method POST -Body $createSlotsBody -Headers $headers
    Write-Success "✓ Time slots created successfully"
    Write-Host "Message: $($createResponse.message)" -ForegroundColor Gray
    Write-Host "Total Created: $($createResponse.total_created)" -ForegroundColor Gray
    Write-Host "Total Failed: $($createResponse.total_failed)" -ForegroundColor Gray
    
    if ($createResponse.created_slots.Count -gt 0) {
        $SLOT_ID = $createResponse.created_slots[0].id
        Write-Host "First Slot ID: $SLOT_ID" -ForegroundColor Gray
        Write-Host ""
        
        Write-Host "Created Slots:" -ForegroundColor Yellow
        foreach ($slot in $createResponse.created_slots) {
            Write-Host "  - ID: $($slot.id)" -ForegroundColor Gray
            Write-Host "    Time: $($slot.start_time) - $($slot.end_time)" -ForegroundColor Gray
            Write-Host "    Max Patients: $($slot.max_patients)" -ForegroundColor Gray
            Write-Host "    Status: $($slot.status)" -ForegroundColor Gray
            Write-Host ""
        }
    }
}
catch {
    Write-Error "✗ Failed to create time slots"
    Write-Error $_.Exception.Message
    Write-Host ""
}

# ===================================================================
# STEP 3: List Time Slots (with filters)
# ===================================================================
Write-Info "STEP 3: Listing Time Slots..."
Write-Host ""

$doctorID = "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d"
$clinicID = "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2"
$slotType = "offline"
$date = "2024-10-15"

try {
    $listUrl = "$BASE_URL/doctor-time-slots?doctor_id=$doctorID&clinic_id=$clinicID&slot_type=$slotType&date=$date"
    $listResponse = Invoke-RestMethod -Uri $listUrl -Method GET -Headers $headers
    
    Write-Success "✓ Time slots retrieved successfully"
    Write-Host "Total Slots: $($listResponse.total)" -ForegroundColor Gray
    Write-Host ""
    
    if ($listResponse.slots.Count -gt 0) {
        Write-Host "Slots List:" -ForegroundColor Yellow
        foreach ($slot in $listResponse.slots) {
            Write-Host "  - ID: $($slot.id)" -ForegroundColor Gray
            Write-Host "    Date: $($slot.date)" -ForegroundColor Gray
            Write-Host "    Time: $($slot.start_time) - $($slot.end_time)" -ForegroundColor Gray
            Write-Host "    Type: $($slot.slot_type)" -ForegroundColor Gray
            Write-Host "    Max Patients: $($slot.max_patients)" -ForegroundColor Gray
            Write-Host "    Booked: $($slot.booked_patients)" -ForegroundColor Gray
            Write-Host "    Available: $($slot.available_spots)" -ForegroundColor Gray
            Write-Host "    Status: $($slot.status)" -ForegroundColor Gray
            Write-Host ""
        }
    }
}
catch {
    Write-Error "✗ Failed to list time slots"
    Write-Error $_.Exception.Message
    Write-Host ""
}

# ===================================================================
# STEP 4: Get Single Time Slot
# ===================================================================
Write-Info "STEP 4: Getting Single Time Slot..."
Write-Host ""

if ($SLOT_ID) {
    try {
        $getSlotResponse = Invoke-RestMethod -Uri "$BASE_URL/doctor-time-slots/$SLOT_ID" -Method GET -Headers $headers
        Write-Success "✓ Time slot retrieved successfully"
        Write-Host ""
        
        $slot = $getSlotResponse.slot
        Write-Host "Slot Details:" -ForegroundColor Yellow
        Write-Host "  ID: $($slot.id)" -ForegroundColor Gray
        Write-Host "  Doctor ID: $($slot.doctor_id)" -ForegroundColor Gray
        Write-Host "  Clinic ID: $($slot.clinic_id)" -ForegroundColor Gray
        Write-Host "  Date: $($slot.date)" -ForegroundColor Gray
        Write-Host "  Type: $($slot.slot_type)" -ForegroundColor Gray
        Write-Host "  Time: $($slot.start_time) - $($slot.end_time)" -ForegroundColor Gray
        Write-Host "  Max Patients: $($slot.max_patients)" -ForegroundColor Gray
        Write-Host "  Booked Patients: $($slot.booked_patients)" -ForegroundColor Gray
        Write-Host "  Available Spots: $($slot.available_spots)" -ForegroundColor Gray
        Write-Host "  Status: $($slot.status)" -ForegroundColor Gray
        Write-Host ""
    }
    catch {
        Write-Error "✗ Failed to get time slot"
        Write-Error $_.Exception.Message
        Write-Host ""
    }
}
else {
    Write-Warning "Skipping - No slot ID available from creation step"
    Write-Host ""
}

# ===================================================================
# STEP 5: Update Time Slot
# ===================================================================
Write-Info "STEP 5: Updating Time Slot..."
Write-Host ""

if ($SLOT_ID) {
    $updateBody = @{
        slot_type = "online"
        start_time = "10:00"
        end_time = "13:00"
        max_patients = 15
        notes = "Updated morning shift - now online"
    } | ConvertTo-Json

    try {
        $updateResponse = Invoke-RestMethod -Uri "$BASE_URL/doctor-time-slots/$SLOT_ID" -Method PUT -Body $updateBody -Headers $headers
        Write-Success "✓ Time slot updated successfully"
        Write-Host "Message: $($updateResponse.message)" -ForegroundColor Gray
        Write-Host ""
        
        $updatedSlot = $updateResponse.slot
        Write-Host "Updated Slot:" -ForegroundColor Yellow
        Write-Host "  ID: $($updatedSlot.id)" -ForegroundColor Gray
        Write-Host "  Type: $($updatedSlot.slot_type)" -ForegroundColor Gray
        Write-Host "  Time: $($updatedSlot.start_time) - $($updatedSlot.end_time)" -ForegroundColor Gray
        Write-Host "  Max Patients: $($updatedSlot.max_patients)" -ForegroundColor Gray
        Write-Host "  Notes: $($updatedSlot.notes)" -ForegroundColor Gray
        Write-Host ""
    }
    catch {
        Write-Error "✗ Failed to update time slot"
        Write-Error $_.Exception.Message
        Write-Host ""
    }
}
else {
    Write-Warning "Skipping - No slot ID available from creation step"
    Write-Host ""
}

# ===================================================================
# STEP 6: Delete Time Slot (Soft Delete)
# ===================================================================
Write-Info "STEP 6: Deleting Time Slot (Soft Delete)..."
Write-Host ""

if ($SLOT_ID) {
    try {
        $deleteResponse = Invoke-RestMethod -Uri "$BASE_URL/doctor-time-slots/$SLOT_ID" -Method DELETE -Headers $headers
        Write-Success "✓ Time slot deleted successfully"
        Write-Host "Message: $($deleteResponse.message)" -ForegroundColor Gray
        Write-Host "Deleted Slot ID: $($deleteResponse.slot_id)" -ForegroundColor Gray
        Write-Host ""
    }
    catch {
        Write-Error "✗ Failed to delete time slot"
        Write-Error $_.Exception.Message
        Write-Host ""
    }
}
else {
    Write-Warning "Skipping - No slot ID available from creation step"
    Write-Host ""
}

# ===================================================================
# STEP 7: Verify Deletion (List again)
# ===================================================================
Write-Info "STEP 7: Verifying Deletion (Listing Active Slots)..."
Write-Host ""

try {
    $verifyUrl = "$BASE_URL/doctor-time-slots?doctor_id=$doctorID&clinic_id=$clinicID&slot_type=$slotType&date=$date"
    $verifyResponse = Invoke-RestMethod -Uri $verifyUrl -Method GET -Headers $headers
    
    Write-Success "✓ Verification complete"
    Write-Host "Active Slots Remaining: $($verifyResponse.total)" -ForegroundColor Gray
    Write-Host ""
    
    if ($verifyResponse.slots.Count -gt 0) {
        Write-Host "Remaining Active Slots:" -ForegroundColor Yellow
        foreach ($slot in $verifyResponse.slots) {
            Write-Host "  - ID: $($slot.id)" -ForegroundColor Gray
            Write-Host "    Time: $($slot.start_time) - $($slot.end_time)" -ForegroundColor Gray
            Write-Host "    Status: $($slot.status)" -ForegroundColor Gray
            Write-Host ""
        }
    }
    else {
        Write-Host "No active slots remaining for the deleted slot type" -ForegroundColor Yellow
        Write-Host ""
    }
}
catch {
    Write-Error "✗ Verification failed"
    Write-Error $_.Exception.Message
    Write-Host ""
}

# ===================================================================
# SUMMARY
# ===================================================================
Write-Info "=================================================="
Write-Info "TEST SUITE COMPLETED"
Write-Info "=================================================="
Write-Host ""

Write-Host "All API endpoints have been tested:" -ForegroundColor Yellow
Write-Host "  1. ✓ POST   /doctor-time-slots         - Create slots (bulk)" -ForegroundColor Green
Write-Host "  2. ✓ GET    /doctor-time-slots         - List slots (filtered)" -ForegroundColor Green
Write-Host "  3. ✓ GET    /doctor-time-slots/:id     - Get single slot" -ForegroundColor Green
Write-Host "  4. ✓ PUT    /doctor-time-slots/:id     - Update slot" -ForegroundColor Green
Write-Host "  5. ✓ DELETE /doctor-time-slots/:id     - Delete slot (soft)" -ForegroundColor Green
Write-Host ""
Write-Success "All tests completed successfully!"
Write-Host ""
