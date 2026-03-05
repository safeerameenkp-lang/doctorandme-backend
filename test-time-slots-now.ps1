# Quick Test for Doctor Time Slots API
# Usage: .\test-time-slots-now.ps1 -Token "your-token" [-DoctorId "uuid"] [-ClinicId "uuid"]

param(
    [Parameter(Mandatory=$true)]
    [string]$Token,
    
    [string]$DoctorId = "",
    [string]$ClinicId = "",
    [string]$BaseUrl = "http://localhost:8081/api"
)

$headers = @{
    "Authorization" = "Bearer $Token"
    "Content-Type" = "application/json"
}

function Show-TimeSlots {
    param($Slots, $Title)
    
    if ($Slots.total_count -eq 0) {
        Write-Host "No time slots found." -ForegroundColor Yellow
        return
    }
    
    Write-Host "`n$Title" -ForegroundColor Cyan
    Write-Host "Total: $($Slots.total_count) slots`n" -ForegroundColor Gray
    
    $Slots.time_slots | ForEach-Object {
        $status = if ($_.is_active) { "✅" } else { "❌" }
        $typeIcon = if ($_.slot_type -eq "offline") { "🏥" } else { "💻" }
        
        Write-Host "$status $typeIcon $($_.day_name) - $($_.start_time) to $($_.end_time)" -ForegroundColor White
        Write-Host "   Doctor: $($_.doctor_name)" -ForegroundColor Gray
        Write-Host "   Clinic: $($_.clinic_name)" -ForegroundColor Gray
        Write-Host "   Type: $($_.slot_type) | Max Patients: $($_.max_patients)" -ForegroundColor Gray
        if ($_.notes) {
            Write-Host "   Notes: $($_.notes)" -ForegroundColor Gray
        }
        Write-Host ""
    }
}

Write-Host "`n🔍 Testing Doctor Time Slots API" -ForegroundColor Cyan
Write-Host "=" * 60

# Test 1: Get slots based on provided filters
try {
    $url = "$BaseUrl/doctor-time-slots"
    $queryParams = @()
    
    if ($DoctorId) {
        $queryParams += "doctor_id=$DoctorId"
        Write-Host "📌 Filter: Doctor ID = $DoctorId" -ForegroundColor Yellow
    }
    if ($ClinicId) {
        $queryParams += "clinic_id=$ClinicId"
        Write-Host "📌 Filter: Clinic ID = $ClinicId" -ForegroundColor Yellow
    }
    
    if ($queryParams.Count -gt 0) {
        $url += "?" + ($queryParams -join "&")
    }
    
    Write-Host "`n🌐 URL: $url" -ForegroundColor Gray
    
    $response = Invoke-RestMethod -Uri $url -Method GET -Headers $headers
    Show-TimeSlots $response "📅 Time Slots"
    
} catch {
    Write-Host "`n❌ Error: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "Details: $($_.ErrorDetails.Message)" -ForegroundColor Red
    }
    exit 1
}

# Show summary by day
if ($response.total_count -gt 0) {
    Write-Host "`n📊 Summary by Day:" -ForegroundColor Cyan
    $response.time_slots | Group-Object day_name | Sort-Object {
        switch ($_.Name) {
            "Sunday" { 0 }
            "Monday" { 1 }
            "Tuesday" { 2 }
            "Wednesday" { 3 }
            "Thursday" { 4 }
            "Friday" { 5 }
            "Saturday" { 6 }
        }
    } | ForEach-Object {
        Write-Host "  $($_.Name): $($_.Count) slots" -ForegroundColor White
    }
    
    # Show summary by type
    Write-Host "`n🎯 Summary by Type:" -ForegroundColor Cyan
    $response.time_slots | Group-Object slot_type | ForEach-Object {
        $icon = if ($_.Name -eq "offline") { "🏥" } else { "💻" }
        Write-Host "  $icon $($_.Name): $($_.Count) slots" -ForegroundColor White
    }
}

# Additional Tests
if ($DoctorId) {
    Write-Host "`n" + ("=" * 60)
    Write-Host "📋 Additional Tests for Doctor`n" -ForegroundColor Cyan
    
    # Test: Get Monday slots
    try {
        Write-Host "🔍 Getting Monday slots..." -ForegroundColor Yellow
        $mondayUrl = "$BaseUrl/doctor-time-slots?doctor_id=$DoctorId&day_of_week=1"
        $mondaySlots = Invoke-RestMethod -Uri $mondayUrl -Method GET -Headers $headers
        Write-Host "✓ Found $($mondaySlots.total_count) Monday slots" -ForegroundColor Green
    } catch {
        Write-Host "✗ Could not fetch Monday slots" -ForegroundColor Red
    }
    
    # Test: Get offline slots
    try {
        Write-Host "🔍 Getting offline slots..." -ForegroundColor Yellow
        $offlineUrl = "$BaseUrl/doctor-time-slots?doctor_id=$DoctorId&slot_type=offline"
        $offlineSlots = Invoke-RestMethod -Uri $offlineUrl -Method GET -Headers $headers
        Write-Host "✓ Found $($offlineSlots.total_count) offline slots" -ForegroundColor Green
    } catch {
        Write-Host "✗ Could not fetch offline slots" -ForegroundColor Red
    }
}

Write-Host "`n✅ Test completed!" -ForegroundColor Green
Write-Host "`n📚 For more details, see: DOCTOR_TIME_SLOTS_API_GUIDE.md" -ForegroundColor Cyan


