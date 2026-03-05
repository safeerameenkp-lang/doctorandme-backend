# Test Script for Clinic Doctors List APIs
# This script tests all the APIs for listing doctors associated with clinics

param(
    [string]$BaseUrl = "http://localhost:8080/api",
    [string]$Token = "",
    [string]$ClinicId = "",
    [string]$DoctorId = ""
)

# Colors
$successColor = "Green"
$errorColor = "Red"
$infoColor = "Cyan"
$warningColor = "Yellow"

# Function to print section header
function Write-Section {
    param([string]$Title)
    Write-Host "`n$('='*80)" -ForegroundColor $infoColor
    Write-Host $Title -ForegroundColor $infoColor
    Write-Host "$('='*80)" -ForegroundColor $infoColor
}

# Function to print success message
function Write-Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor $successColor
}

# Function to print error message
function Write-Error-Custom {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor $errorColor
}

# Check if token is provided
if ([string]::IsNullOrEmpty($Token)) {
    Write-Host "Error: Token is required. Please provide a JWT token." -ForegroundColor $errorColor
    Write-Host "Usage: .\test-clinic-doctors-list.ps1 -Token 'your-jwt-token' -ClinicId 'clinic-uuid'" -ForegroundColor $warningColor
    exit 1
}

# Setup headers
$headers = @{
    "Authorization" = "Bearer $Token"
    "Content-Type" = "application/json"
}

# Test 1: Get Doctors by Clinic (Recommended API)
Write-Section "TEST 1: Get Doctors by Clinic (Recommended)"

if ([string]::IsNullOrEmpty($ClinicId)) {
    Write-Host "Skipping: ClinicId not provided" -ForegroundColor $warningColor
} else {
    try {
        $url = "$BaseUrl/doctors/clinic/$ClinicId"
        Write-Host "URL: $url" -ForegroundColor Gray
        
        $response = Invoke-RestMethod -Uri $url -Method GET -Headers $headers
        
        Write-Success "Found $($response.total_doctors) doctors for clinic $($response.clinic_id)"
        
        if ($response.total_doctors -gt 0) {
            Write-Host "`nDoctors List:" -ForegroundColor $infoColor
            $response.doctors | ForEach-Object {
                Write-Host "`n  📋 Doctor Details:" -ForegroundColor $infoColor
                Write-Host "     Full Name: $($_.full_name)"
                Write-Host "     Doctor ID: $($_.doctor_id)"
                Write-Host "     Doctor Code: $($_.doctor_code)"
                Write-Host "     Specialization: $($_.specialization)"
                Write-Host "     License: $($_.license_number)"
                Write-Host "     Email: $($_.email)"
                Write-Host "     Phone: $($_.phone)"
                Write-Host "     Username: $($_.username)"
                Write-Host "     Active: $($_.is_active)"
                
                Write-Host "`n  💰 Clinic-Specific Fees:" -ForegroundColor $infoColor
                Write-Host "     Offline Consultation: ₹$($_.clinic_specific_fees.consultation_fee_offline)"
                Write-Host "     Online Consultation: ₹$($_.clinic_specific_fees.consultation_fee_online)"
                Write-Host "     Follow-up Fee: ₹$($_.clinic_specific_fees.follow_up_fee)"
                Write-Host "     Follow-up Days: $($_.clinic_specific_fees.follow_up_days)"
                if ($_.clinic_specific_fees.notes) {
                    Write-Host "     Notes: $($_.clinic_specific_fees.notes)"
                }
                
                Write-Host "`n  📊 Default Fees:" -ForegroundColor $infoColor
                Write-Host "     Consultation Fee: ₹$($_.default_fees.consultation_fee)"
                Write-Host "     Follow-up Fee: ₹$($_.default_fees.follow_up_fee)"
                Write-Host "     Follow-up Days: $($_.default_fees.follow_up_days)"
                
                Write-Host "`n  $('-'*70)" -ForegroundColor Gray
            }
        } else {
            Write-Host "`nNo doctors found for this clinic." -ForegroundColor $warningColor
        }
        
    } catch {
        Write-Error-Custom "Failed to get doctors by clinic"
        Write-Host "Error: $($_.Exception.Message)" -ForegroundColor $errorColor
        if ($_.ErrorDetails.Message) {
            Write-Host "Details: $($_.ErrorDetails.Message)" -ForegroundColor $errorColor
        }
    }
}

# Test 2: Get Clinic-Doctor Links (Alternative API)
Write-Section "TEST 2: Get Clinic-Doctor Links"

if ([string]::IsNullOrEmpty($ClinicId)) {
    Write-Host "Skipping: ClinicId not provided" -ForegroundColor $warningColor
} else {
    try {
        $url = "$BaseUrl/clinic-doctor-links?clinic_id=$ClinicId"
        Write-Host "URL: $url" -ForegroundColor Gray
        
        $response = Invoke-RestMethod -Uri $url -Method GET -Headers $headers
        
        Write-Success "Found $($response.count) clinic-doctor links"
        
        if ($response.count -gt 0) {
            Write-Host "`nLinks List:" -ForegroundColor $infoColor
            $response.links | ForEach-Object {
                Write-Host "`n  🔗 Link Information:" -ForegroundColor $infoColor
                Write-Host "     Link ID: $($_.link_id)"
                Write-Host "     Active: $($_.is_active)"
                Write-Host "     Created: $($_.created_at)"
                
                Write-Host "`n  🏥 Clinic:" -ForegroundColor $infoColor
                Write-Host "     Name: $($_.clinic.name)"
                Write-Host "     Code: $($_.clinic.clinic_code)"
                Write-Host "     ID: $($_.clinic.clinic_id)"
                
                Write-Host "`n  👨‍⚕️ Doctor:" -ForegroundColor $infoColor
                Write-Host "     Name: $($_.doctor.first_name) $($_.doctor.last_name)"
                Write-Host "     Code: $($_.doctor.doctor_code)"
                Write-Host "     Specialization: $($_.doctor.specialization)"
                Write-Host "     License: $($_.doctor.license_number)"
                Write-Host "     Email: $($_.doctor.email)"
                
                Write-Host "`n  💰 Fees:" -ForegroundColor $infoColor
                Write-Host "     Offline: ₹$($_.fees.consultation_fee_offline)"
                Write-Host "     Online: ₹$($_.fees.consultation_fee_online)"
                Write-Host "     Follow-up: ₹$($_.fees.follow_up_fee) ($($_.fees.follow_up_days) days)"
                
                if ($_.notes) {
                    Write-Host "`n  📝 Notes: $($_.notes)"
                }
                
                Write-Host "`n  $('-'*70)" -ForegroundColor Gray
            }
        } else {
            Write-Host "`nNo links found for this clinic." -ForegroundColor $warningColor
        }
        
    } catch {
        Write-Error-Custom "Failed to get clinic-doctor links"
        Write-Host "Error: $($_.Exception.Message)" -ForegroundColor $errorColor
        if ($_.ErrorDetails.Message) {
            Write-Host "Details: $($_.ErrorDetails.Message)" -ForegroundColor $errorColor
        }
    }
}

# Test 3: Get Clinics for a Doctor
Write-Section "TEST 3: Get Clinics for a Doctor"

if ([string]::IsNullOrEmpty($DoctorId)) {
    Write-Host "Skipping: DoctorId not provided" -ForegroundColor $warningColor
} else {
    try {
        $url = "$BaseUrl/clinic-doctor-links?doctor_id=$DoctorId"
        Write-Host "URL: $url" -ForegroundColor Gray
        
        $response = Invoke-RestMethod -Uri $url -Method GET -Headers $headers
        
        Write-Success "Found $($response.count) clinics for doctor $DoctorId"
        
        if ($response.count -gt 0) {
            Write-Host "`nClinics List:" -ForegroundColor $infoColor
            $response.links | ForEach-Object {
                Write-Host "`n  🏥 Clinic: $($_.clinic.name)"
                Write-Host "     Code: $($_.clinic.clinic_code)"
                Write-Host "     Offline Fee: ₹$($_.fees.consultation_fee_offline)"
                Write-Host "     Online Fee: ₹$($_.fees.consultation_fee_online)"
                Write-Host "     Follow-up: ₹$($_.fees.follow_up_fee) ($($_.fees.follow_up_days) days)"
                Write-Host "`n  $('-'*70)" -ForegroundColor Gray
            }
        } else {
            Write-Host "`nNo clinics found for this doctor." -ForegroundColor $warningColor
        }
        
    } catch {
        Write-Error-Custom "Failed to get clinics for doctor"
        Write-Host "Error: $($_.Exception.Message)" -ForegroundColor $errorColor
        if ($_.ErrorDetails.Message) {
            Write-Host "Details: $($_.ErrorDetails.Message)" -ForegroundColor $errorColor
        }
    }
}

# Test 4: Get All Doctors (Generic)
Write-Section "TEST 4: Get All Doctors (Generic)"

try {
    $url = "$BaseUrl/doctors"
    Write-Host "URL: $url" -ForegroundColor Gray
    
    $response = Invoke-RestMethod -Uri $url -Method GET -Headers $headers
    
    $totalDoctors = if ($response -is [Array]) { $response.Length } else { 1 }
    Write-Success "Found $totalDoctors doctors in the system"
    
    if ($totalDoctors -gt 0) {
        Write-Host "`nShowing first 5 doctors:" -ForegroundColor $infoColor
        $response | Select-Object -First 5 | ForEach-Object {
            Write-Host "`n  👨‍⚕️ Doctor:" -ForegroundColor $infoColor
            Write-Host "     Name: $($_.user.first_name) $($_.user.last_name)"
            Write-Host "     Code: $($_.doctor.doctor_code)"
            Write-Host "     Specialization: $($_.doctor.specialization)"
            Write-Host "     Email: $($_.user.email)"
            Write-Host "     Default Fee: ₹$($_.doctor.consultation_fee)"
            Write-Host "`n  $('-'*70)" -ForegroundColor Gray
        }
        
        if ($totalDoctors -gt 5) {
            Write-Host "`n... and $($totalDoctors - 5) more doctors" -ForegroundColor Gray
        }
    } else {
        Write-Host "`nNo doctors found in the system." -ForegroundColor $warningColor
    }
    
} catch {
    Write-Error-Custom "Failed to get all doctors"
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor $errorColor
    if ($_.ErrorDetails.Message) {
        Write-Host "Details: $($_.ErrorDetails.Message)" -ForegroundColor $errorColor
    }
}

# Test 5: Get Doctors Filtered by Clinic (Generic API with filter)
Write-Section "TEST 5: Get Doctors Filtered by Clinic (Generic API)"

if ([string]::IsNullOrEmpty($ClinicId)) {
    Write-Host "Skipping: ClinicId not provided" -ForegroundColor $warningColor
} else {
    try {
        $url = "$BaseUrl/doctors?clinic_id=$ClinicId"
        Write-Host "URL: $url" -ForegroundColor Gray
        
        $response = Invoke-RestMethod -Uri $url -Method GET -Headers $headers
        
        $totalDoctors = if ($response -is [Array]) { $response.Length } else { 1 }
        Write-Success "Found $totalDoctors doctors for clinic (using generic API)"
        
        if ($totalDoctors -gt 0) {
            Write-Host "`nDoctors List:" -ForegroundColor $infoColor
            $response | ForEach-Object {
                Write-Host "`n  👨‍⚕️ $($_.user.first_name) $($_.user.last_name)"
                Write-Host "     Specialization: $($_.doctor.specialization)"
                Write-Host "     Default Fee: ₹$($_.doctor.consultation_fee)"
                Write-Host "`n  $('-'*70)" -ForegroundColor Gray
            }
        }
        
    } catch {
        Write-Error-Custom "Failed to get doctors by clinic (generic API)"
        Write-Host "Error: $($_.Exception.Message)" -ForegroundColor $errorColor
    }
}

# Summary
Write-Section "SUMMARY"

Write-Host @"

📊 API Comparison Summary:

1. GET /doctors/clinic/:clinic_id (RECOMMENDED)
   ✓ Clinic-specific fees
   ✓ Default fees
   ✓ Full doctor details
   ✓ Clean response format
   ✓ Only active records

2. GET /clinic-doctor-links?clinic_id=...
   ✓ Clinic-specific fees
   ✓ Link metadata (ID, status, created date)
   ✓ Clinic information
   ✓ Bidirectional filtering (clinic or doctor)

3. GET /doctors?clinic_id=...
   ✓ Default fees only
   ✓ Simple doctor listing
   - No clinic-specific fees
   - No link metadata

📋 Use the appropriate API based on your needs:
   - Appointment booking → Use API #1
   - Admin link management → Use API #2
   - General doctor listing → Use API #3

"@

Write-Host "`nTest completed!" -ForegroundColor $successColor
Write-Host "For more information, see CLINIC_DOCTORS_LIST_API_COMPLETE.md" -ForegroundColor $infoColor

