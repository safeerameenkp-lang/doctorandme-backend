# Test Script for Session-Based Time Slots API
# This script demonstrates creating and listing session-based slots

$ErrorActionPreference = "Stop"

# Configuration
$baseUrl = "http://localhost:8081/organizations"
$token = "your-token-here"  # Replace with your actual token
$doctorId = "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d"
$clinicId = "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2"
$date = "2025-10-20"  # Monday

Write-Host "`n╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  SESSION-BASED TIME SLOTS API TEST                             ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# ============================================================================
# TEST 1: Create Session-Based Slots
# ============================================================================

Write-Host "`n📋 TEST 1: Creating Session-Based Time Slots" -ForegroundColor Yellow
Write-Host "─────────────────────────────────────────────────────────────" -ForegroundColor Gray

$createRequest = @{
    doctor_id = $doctorId
    clinic_id = $clinicId
    slot_type = "offline"
    slot_duration = 5
    date = $date
    is_available = $true
    notes = "Regular clinic hours"
    sessions = @(
        @{
            session_name = "Morning Session"
            start_time = "09:00"
            end_time = "12:00"
            max_patients = 36
            slot_interval_minutes = 5
            notes = "Morning consultations"
        },
        @{
            session_name = "Afternoon Session"
            start_time = "14:00"
            end_time = "17:00"
            max_patients = 36
            slot_interval_minutes = 5
            notes = "Afternoon consultations"
        }
    )
} | ConvertTo-Json -Depth 10

Write-Host "`n📤 Request:" -ForegroundColor Cyan
Write-Host $createRequest -ForegroundColor Gray

try {
    $createResponse = Invoke-RestMethod -Uri "$baseUrl/doctor-session-slots" -Method Post -Headers $headers -Body $createRequest -ContentType "application/json"
    
    Write-Host "`n✅ SUCCESS! Slots Created" -ForegroundColor Green
    Write-Host "─────────────────────────────────────────────────────────────" -ForegroundColor Gray
    
    Write-Host "`n📊 Summary:" -ForegroundColor Cyan
    Write-Host "  Time Slot ID: $($createResponse.data.id)" -ForegroundColor White
    Write-Host "  Date: $($createResponse.data.date)" -ForegroundColor White
    Write-Host "  Day of Week: $($createResponse.data.day_of_week) (Auto-calculated)" -ForegroundColor Green
    Write-Host "  Total Sessions: $($createResponse.data.sessions.Count)" -ForegroundColor White
    
    foreach ($session in $createResponse.data.sessions) {
        Write-Host "`n  📅 $($session.session_name):" -ForegroundColor Yellow
        Write-Host "     Time: $($session.start_time) - $($session.end_time)" -ForegroundColor White
        Write-Host "     Generated Slots: $($session.generated_slots)" -ForegroundColor Green
        Write-Host "     Available: $($session.available_slots)" -ForegroundColor Green
        Write-Host "     Booked: $($session.booked_slots)" -ForegroundColor White
        
        if ($session.slots.Count -le 5) {
            Write-Host "     Individual Slots:" -ForegroundColor Cyan
            foreach ($slot in $session.slots) {
                Write-Host "       - $($slot.slot_start) to $($slot.slot_end) [$($slot.status)]" -ForegroundColor Gray
            }
        } else {
            Write-Host "     Showing first 3 slots:" -ForegroundColor Cyan
            for ($i = 0; $i -lt 3 -and $i -lt $session.slots.Count; $i++) {
                $slot = $session.slots[$i]
                Write-Host "       - $($slot.slot_start) to $($slot.slot_end) [$($slot.status)]" -ForegroundColor Gray
            }
            Write-Host "       ... and $($session.slots.Count - 3) more" -ForegroundColor DarkGray
        }
    }
    
    Write-Host "`n✅ Full Response:" -ForegroundColor Cyan
    $createResponse | ConvertTo-Json -Depth 10
    
} catch {
    Write-Host "`n❌ ERROR Creating Slots" -ForegroundColor Red
    Write-Host "Status: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        $errorJson = $_.ErrorDetails.Message | ConvertFrom-Json
        Write-Host "Error: $($errorJson.error)" -ForegroundColor Red
        if ($errorJson.message) {
            Write-Host "Message: $($errorJson.message)" -ForegroundColor Red
        }
    } else {
        Write-Host "Message: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# ============================================================================
# TEST 2: List Session-Based Slots
# ============================================================================

Write-Host "`n`n📋 TEST 2: Listing Session-Based Time Slots (With Filters)" -ForegroundColor Yellow
Write-Host "─────────────────────────────────────────────────────────────" -ForegroundColor Gray

$listUrl = "$baseUrl/doctor-session-slots?doctor_id=$doctorId&date=$date&clinic_id=$clinicId&slot_type=offline"
Write-Host "`n📍 Endpoint: $listUrl" -ForegroundColor Cyan
Write-Host "   Filters:" -ForegroundColor Gray
Write-Host "   - doctor_id: $doctorId" -ForegroundColor Gray
Write-Host "   - clinic_id: $clinicId" -ForegroundColor Gray
Write-Host "   - date: $date" -ForegroundColor Gray
Write-Host "   - slot_type: offline" -ForegroundColor Green

try {
    $listResponse = Invoke-RestMethod -Uri $listUrl -Method Get -Headers $headers
    
    Write-Host "`n✅ SUCCESS! Retrieved Slots" -ForegroundColor Green
    Write-Host "─────────────────────────────────────────────────────────────" -ForegroundColor Gray
    
    Write-Host "`n📊 Summary:" -ForegroundColor Cyan
    Write-Host "  Total Days: $($listResponse.total)" -ForegroundColor White
    Write-Host "  Date Filter: $($listResponse.date)" -ForegroundColor White
    
    foreach ($daySlot in $listResponse.slots) {
        Write-Host "`n📅 Date: $($daySlot.date) (Day $($daySlot.day_of_week))" -ForegroundColor Yellow
        Write-Host "   Slot Type: $($daySlot.slot_type)" -ForegroundColor White
        Write-Host "   Available: $($daySlot.is_available)" -ForegroundColor White
        Write-Host "   Sessions: $($daySlot.sessions.Count)" -ForegroundColor White
        
        foreach ($session in $daySlot.sessions) {
            Write-Host "`n   📊 $($session.session_name):" -ForegroundColor Cyan
            Write-Host "      Time: $($session.start_time) - $($session.end_time)" -ForegroundColor White
            Write-Host "      Total Slots: $($session.generated_slots)" -ForegroundColor White
            Write-Host "      Available: $($session.available_slots)" -ForegroundColor Green
            Write-Host "      Booked: $($session.booked_slots)" -ForegroundColor $(if ($session.booked_slots -gt 0) { "Yellow" } else { "White" })
            
            $availableSlots = $session.slots | Where-Object { $_.status -eq "available" }
            $bookedSlots = $session.slots | Where-Object { $_.is_booked -eq $true }
            
            if ($availableSlots.Count -gt 0) {
                Write-Host "`n      ✅ First 3 Available Slots:" -ForegroundColor Green
                for ($i = 0; $i -lt 3 -and $i -lt $availableSlots.Count; $i++) {
                    $slot = $availableSlots[$i]
                    Write-Host "         $($slot.slot_start) - $($slot.slot_end)" -ForegroundColor Gray
                }
                if ($availableSlots.Count -gt 3) {
                    Write-Host "         ... and $($availableSlots.Count - 3) more available" -ForegroundColor DarkGray
                }
            }
            
            if ($bookedSlots.Count -gt 0) {
                Write-Host "`n      📌 Booked Slots: $($bookedSlots.Count)" -ForegroundColor Yellow
                foreach ($slot in $bookedSlots) {
                    Write-Host "         $($slot.slot_start) - $($slot.slot_end) [Patient: $($slot.booked_patient_id)]" -ForegroundColor DarkYellow
                }
            }
        }
    }
    
    Write-Host "`n✅ Full Response:" -ForegroundColor Cyan
    $listResponse | ConvertTo-Json -Depth 10
    
} catch {
    Write-Host "`n❌ ERROR Listing Slots" -ForegroundColor Red
    Write-Host "Status: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        $errorJson = $_.ErrorDetails.Message | ConvertFrom-Json
        Write-Host "Error: $($errorJson.error)" -ForegroundColor Red
    } else {
        Write-Host "Message: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# ============================================================================
# SUMMARY
# ============================================================================

Write-Host "`n`n╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  TEST SUMMARY                                                  ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

Write-Host "`n📋 API Endpoints Tested:" -ForegroundColor Yellow
Write-Host "  1. POST /doctor-session-slots - Create session-based slots" -ForegroundColor White
Write-Host "  2. GET  /doctor-session-slots - List session-based slots" -ForegroundColor White

Write-Host "`n💡 Key Features Demonstrated:" -ForegroundColor Yellow
Write-Host "  ✅ Auto-generation of individual slots" -ForegroundColor Green
Write-Host "  ✅ Auto-calculation of day_of_week" -ForegroundColor Green
Write-Host "  ✅ Session-based organization" -ForegroundColor Green
Write-Host "  ✅ Real-time availability tracking" -ForegroundColor Green

Write-Host "`n📖 For full documentation, see:" -ForegroundColor Cyan
Write-Host "  SESSION_BASED_SLOTS_COMPLETE_GUIDE.md" -ForegroundColor White

Write-Host "`n════════════════════════════════════════════════════════════════`n" -ForegroundColor Cyan

