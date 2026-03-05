# Script to delete time slots for a specific date
# This helps clear overlapping slots before creating new ones

$ErrorActionPreference = "Stop"

# Configuration
$baseUrl = "http://localhost:8081/organizations"
$token = "your-token-here"  # Replace with your actual token
$doctorId = "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d"
$clinicId = "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2"
$date = "2025-10-20"

Write-Host "`n╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║     DELETE TIME SLOTS BY DATE                                  ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

Write-Host "`n📅 Date: $date" -ForegroundColor Yellow
Write-Host "👨‍⚕️ Doctor ID: $doctorId" -ForegroundColor Yellow
Write-Host "🏥 Clinic ID: $clinicId" -ForegroundColor Yellow

# Step 1: Get all slots for this date
Write-Host "`n📋 Step 1: Fetching existing slots..." -ForegroundColor Cyan

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

$listUrl = "$baseUrl/doctor-time-slots?doctor_id=$doctorId&clinic_id=$clinicId&date=$date"

try {
    $response = Invoke-RestMethod -Uri $listUrl -Method Get -Headers $headers
    $slots = $response.slots
    
    if ($null -eq $slots -or $slots.Count -eq 0) {
        Write-Host "✅ No slots found for this date" -ForegroundColor Green
        exit 0
    }
    
    Write-Host "📊 Found $($slots.Count) slots" -ForegroundColor Yellow
    
    # Step 2: Delete each slot
    Write-Host "`n🗑️ Step 2: Deleting slots..." -ForegroundColor Cyan
    
    $deletedCount = 0
    $failedCount = 0
    
    foreach ($slot in $slots) {
        Write-Host "  Deleting slot $($slot.id) ($($ slot.start_time) - $($slot.end_time))..." -ForegroundColor Gray
        
        try {
            $deleteUrl = "$baseUrl/doctor-time-slots/$($slot.id)"
            $deleteResponse = Invoke-RestMethod -Uri $deleteUrl -Method Delete -Headers $headers
            Write-Host "    ✅ Deleted" -ForegroundColor Green
            $deletedCount++
        }
        catch {
            Write-Host "    ❌ Failed: $($_.Exception.Message)" -ForegroundColor Red
            $failedCount++
        }
    }
    
    # Summary
    Write-Host "`n╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "║     DELETION SUMMARY                                           ║" -ForegroundColor Cyan
    Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
    Write-Host "✅ Deleted: $deletedCount slots" -ForegroundColor Green
    if ($failedCount -gt 0) {
        Write-Host "❌ Failed: $failedCount slots" -ForegroundColor Red
    }
    
} catch {
    Write-Host "❌ Error: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "Details: $($_.ErrorDetails.Message)" -ForegroundColor Red
    }
    exit 1
}

