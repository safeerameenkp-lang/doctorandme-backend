# Test Script for Auto MO ID Generation
# Tests the automatic MO ID generation feature for clinic patients

Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "Testing Auto MO ID Generation Feature" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""

# Configuration
$baseUrl = "http://localhost:8080"
$clinicId = Read-Host "Enter Clinic ID (UUID)"

if ([string]::IsNullOrWhiteSpace($clinicId)) {
    Write-Host "Error: Clinic ID is required!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Using Clinic ID: $clinicId" -ForegroundColor Yellow
Write-Host ""

# Test 1: Create patient WITHOUT mo_id (auto-generate)
Write-Host "===============================================" -ForegroundColor Green
Write-Host "Test 1: Auto-Generate MO ID" -ForegroundColor Green
Write-Host "===============================================" -ForegroundColor Green
Write-Host "Creating patient WITHOUT mo_id..." -ForegroundColor Yellow

$patient1 = @{
    clinic_id = $clinicId
    first_name = "AutoTest"
    last_name = "Patient1"
    phone = "+91 9876543210"
    email = "autotest1@example.com"
    age = 30
    gender = "Male"
} | ConvertTo-Json

Write-Host "Request:" -ForegroundColor Cyan
Write-Host $patient1
Write-Host ""

try {
    $response1 = Invoke-RestMethod -Uri "$baseUrl/clinic-specific-patients" `
        -Method Post `
        -ContentType "application/json" `
        -Body $patient1
    
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Response:" -ForegroundColor Cyan
    $response1 | ConvertTo-Json -Depth 5
    Write-Host ""
    Write-Host "Generated MO ID: $($response1.patient.mo_id)" -ForegroundColor Magenta
    Write-Host ""
    
    $generatedMoId = $response1.patient.mo_id
    $patient1Id = $response1.patient.id
} catch {
    Write-Host "❌ Failed!" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ""
}

Start-Sleep -Seconds 2

# Test 2: Create second patient WITHOUT mo_id (should increment)
Write-Host "===============================================" -ForegroundColor Green
Write-Host "Test 2: Sequential Auto-Generation" -ForegroundColor Green
Write-Host "===============================================" -ForegroundColor Green
Write-Host "Creating second patient WITHOUT mo_id..." -ForegroundColor Yellow

$patient2 = @{
    clinic_id = $clinicId
    first_name = "AutoTest"
    last_name = "Patient2"
    phone = "+91 8765432109"
    email = "autotest2@example.com"
    age = 35
    gender = "Female"
} | ConvertTo-Json

Write-Host "Request:" -ForegroundColor Cyan
Write-Host $patient2
Write-Host ""

try {
    $response2 = Invoke-RestMethod -Uri "$baseUrl/clinic-specific-patients" `
        -Method Post `
        -ContentType "application/json" `
        -Body $patient2
    
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Response:" -ForegroundColor Cyan
    $response2 | ConvertTo-Json -Depth 5
    Write-Host ""
    Write-Host "Generated MO ID: $($response2.patient.mo_id)" -ForegroundColor Magenta
    Write-Host ""
    
    $generatedMoId2 = $response2.patient.mo_id
    $patient2Id = $response2.patient.id
    
    # Verify sequential increment
    Write-Host "Verification:" -ForegroundColor Yellow
    Write-Host "First MO ID:  $generatedMoId" -ForegroundColor White
    Write-Host "Second MO ID: $generatedMoId2" -ForegroundColor White
    
    if ($generatedMoId2 -gt $generatedMoId) {
        Write-Host "✅ Sequential increment verified!" -ForegroundColor Green
    } else {
        Write-Host "⚠️ Sequential increment might not be working correctly" -ForegroundColor Yellow
    }
    Write-Host ""
} catch {
    Write-Host "❌ Failed!" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ""
}

Start-Sleep -Seconds 2

# Test 3: Create patient WITH custom mo_id
Write-Host "===============================================" -ForegroundColor Green
Write-Host "Test 3: Custom MO ID" -ForegroundColor Green
Write-Host "===============================================" -ForegroundColor Green
Write-Host "Creating patient WITH custom mo_id..." -ForegroundColor Yellow

$customMoId = "CUSTOM$(Get-Random -Minimum 1000 -Maximum 9999)"

$patient3 = @{
    clinic_id = $clinicId
    first_name = "CustomTest"
    last_name = "Patient3"
    phone = "+91 7654321098"
    email = "customtest@example.com"
    age = 40
    gender = "Male"
    mo_id = $customMoId
} | ConvertTo-Json

Write-Host "Request:" -ForegroundColor Cyan
Write-Host $patient3
Write-Host ""

try {
    $response3 = Invoke-RestMethod -Uri "$baseUrl/clinic-specific-patients" `
        -Method Post `
        -ContentType "application/json" `
        -Body $patient3
    
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Response:" -ForegroundColor Cyan
    $response3 | ConvertTo-Json -Depth 5
    Write-Host ""
    Write-Host "Custom MO ID: $($response3.patient.mo_id)" -ForegroundColor Magenta
    
    $patient3Id = $response3.patient.id
    
    # Verify custom MO ID
    if ($response3.patient.mo_id -eq $customMoId) {
        Write-Host "✅ Custom MO ID applied correctly!" -ForegroundColor Green
    } else {
        Write-Host "⚠️ Custom MO ID not applied correctly" -ForegroundColor Yellow
    }
    Write-Host ""
} catch {
    Write-Host "❌ Failed!" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ""
}

Start-Sleep -Seconds 2

# Test 4: Try to create patient with duplicate mo_id (should fail)
Write-Host "===============================================" -ForegroundColor Green
Write-Host "Test 4: Duplicate MO ID Validation" -ForegroundColor Green
Write-Host "===============================================" -ForegroundColor Green
Write-Host "Attempting to create patient with duplicate mo_id..." -ForegroundColor Yellow

if ($generatedMoId) {
    $patient4 = @{
        clinic_id = $clinicId
        first_name = "DuplicateTest"
        last_name = "Patient4"
        phone = "+91 6543210987"
        email = "duplicate@example.com"
        mo_id = $generatedMoId
    } | ConvertTo-Json
    
    Write-Host "Request (using existing MO ID: $generatedMoId):" -ForegroundColor Cyan
    Write-Host $patient4
    Write-Host ""
    
    try {
        $response4 = Invoke-RestMethod -Uri "$baseUrl/clinic-specific-patients" `
            -Method Post `
            -ContentType "application/json" `
            -Body $patient4
        
        Write-Host "⚠️ Unexpected: Request succeeded when it should have failed!" -ForegroundColor Yellow
        Write-Host "Response:" -ForegroundColor Cyan
        $response4 | ConvertTo-Json -Depth 5
        Write-Host ""
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 409) {
            Write-Host "✅ Success! Duplicate MO ID validation working correctly!" -ForegroundColor Green
            Write-Host "Error Message: $($_.Exception.Message)" -ForegroundColor Yellow
        } else {
            Write-Host "⚠️ Unexpected error code: $statusCode" -ForegroundColor Yellow
            Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
        }
        Write-Host ""
    }
} else {
    Write-Host "⚠️ Skipping Test 4 (no generated MO ID from Test 1)" -ForegroundColor Yellow
    Write-Host ""
}

# Test 5: Create patient after custom ID (should continue sequence)
Write-Host "===============================================" -ForegroundColor Green
Write-Host "Test 5: Sequence Continuation After Custom ID" -ForegroundColor Green
Write-Host "===============================================" -ForegroundColor Green
Write-Host "Creating patient after custom MO ID (verify sequence continues)..." -ForegroundColor Yellow

$patient5 = @{
    clinic_id = $clinicId
    first_name = "SequenceTest"
    last_name = "Patient5"
    phone = "+91 5432109876"
    email = "sequence@example.com"
    age = 28
    gender = "Female"
} | ConvertTo-Json

Write-Host "Request:" -ForegroundColor Cyan
Write-Host $patient5
Write-Host ""

try {
    $response5 = Invoke-RestMethod -Uri "$baseUrl/clinic-specific-patients" `
        -Method Post `
        -ContentType "application/json" `
        -Body $patient5
    
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Response:" -ForegroundColor Cyan
    $response5 | ConvertTo-Json -Depth 5
    Write-Host ""
    Write-Host "Generated MO ID: $($response5.patient.mo_id)" -ForegroundColor Magenta
    Write-Host ""
    
    $patient5Id = $response5.patient.id
} catch {
    Write-Host "❌ Failed!" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ""
}

# Summary
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Clinic ID: $clinicId" -ForegroundColor White
Write-Host ""
Write-Host "Created Patients:" -ForegroundColor Yellow
if ($patient1Id) {
    Write-Host "  1. $patient1Id - MO ID: $generatedMoId (Auto)" -ForegroundColor White
}
if ($patient2Id) {
    Write-Host "  2. $patient2Id - MO ID: $generatedMoId2 (Auto)" -ForegroundColor White
}
if ($patient3Id) {
    Write-Host "  3. $patient3Id - MO ID: $customMoId (Custom)" -ForegroundColor White
}
if ($patient5Id) {
    Write-Host "  4. $($patient5Id) - MO ID: $($response5.patient.mo_id) (Auto)" -ForegroundColor White
}
Write-Host ""

# Cleanup option
Write-Host "===============================================" -ForegroundColor Yellow
$cleanup = Read-Host "Do you want to delete the test patients? (y/N)"

if ($cleanup -eq "y" -or $cleanup -eq "Y") {
    Write-Host ""
    Write-Host "Cleaning up test patients..." -ForegroundColor Yellow
    
    $patientIds = @($patient1Id, $patient2Id, $patient3Id, $patient5Id) | Where-Object { $_ }
    
    foreach ($patientId in $patientIds) {
        try {
            Invoke-RestMethod -Uri "$baseUrl/clinic-specific-patients/$patientId" `
                -Method Delete | Out-Null
            Write-Host "✅ Deleted patient: $patientId" -ForegroundColor Green
        } catch {
            Write-Host "❌ Failed to delete patient: $patientId" -ForegroundColor Red
        }
    }
    
    Write-Host ""
    Write-Host "Cleanup complete!" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "Test patients were not deleted." -ForegroundColor Yellow
    Write-Host "You can manually delete them later if needed." -ForegroundColor Gray
}

Write-Host ""
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "Testing Complete!" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan

