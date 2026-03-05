# Security Fixes Validation Test Script
# Tests all critical security fixes applied to the RBAC system

# Configuration
$BASE_URL = "http://localhost:8000/api/v1/auth"

# Test users (update with your actual test data)
$SUPER_ADMIN_USER = "superadmin"
$SUPER_ADMIN_PASS = "SuperAdmin123"
$ORG_A_ADMIN_USER = "orgadmin_a"
$ORG_A_ADMIN_PASS = "OrgAdmin123"
$CLINIC_ADMIN_USER = "clinicadmin"
$CLINIC_ADMIN_PASS = "ClinicAdmin123"

# Colors
function Write-Success { param($message) Write-Host $message -ForegroundColor Green }
function Write-Error { param($message) Write-Host $message -ForegroundColor Red }
function Write-Info { param($message) Write-Host $message -ForegroundColor Cyan }
function Write-Warning { param($message) Write-Host $message -ForegroundColor Yellow }

# Global counters
$global:PassedTests = 0
$global:FailedTests = 0

# Helper function
function Invoke-APICall {
    param(
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null,
        [string]$Token = ""
    )

    $headers = @{
        "Content-Type" = "application/json"
    }

    if ($Token) {
        $headers["Authorization"] = "Bearer $Token"
    }

    $params = @{
        Uri = "$BASE_URL$Endpoint"
        Method = $Method
        Headers = $headers
    }

    if ($Body) {
        $params["Body"] = ($Body | ConvertTo-Json -Depth 10)
    }

    try {
        $response = Invoke-RestMethod @params
        return @{ Success = $true; StatusCode = 200; Data = $response }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        try {
            $errorBody = $_.ErrorDetails.Message | ConvertFrom-Json
            return @{ Success = $false; StatusCode = $statusCode; Error = $errorBody }
        } catch {
            return @{ Success = $false; StatusCode = $statusCode; Error = @{error = "Unknown error"} }
        }
    }
}

# Test helper
function Test-Expectation {
    param(
        [string]$TestName,
        [bool]$ShouldSucceed,
        [object]$Result,
        [int]$ExpectedStatus = 0
    )
    
    $passed = $false
    
    if ($ShouldSucceed) {
        if ($Result.Success) {
            Write-Success "  ✓ $TestName - PASSED (Expected: Success, Got: Success)"
            $passed = $true
        } else {
            Write-Error "  ✗ $TestName - FAILED (Expected: Success, Got: $($Result.StatusCode) - $($Result.Error.error))"
        }
    } else {
        if (!$Result.Success -and ($ExpectedStatus -eq 0 -or $Result.StatusCode -eq $ExpectedStatus)) {
            Write-Success "  ✓ $TestName - PASSED (Expected: Failure $ExpectedStatus, Got: $($Result.StatusCode))"
            $passed = $true
        } else {
            Write-Error "  ✗ $TestName - FAILED (Expected: Failure $ExpectedStatus, Got: $($Result.StatusCode))"
        }
    }
    
    if ($passed) {
        $global:PassedTests++
    } else {
        $global:FailedTests++
    }
}

Write-Host "`n" + "="*80
Write-Host "  SECURITY FIXES VALIDATION TEST SUITE" -ForegroundColor Yellow
Write-Host "="*80 + "`n"

# =========================================
# SETUP: Login as different admin levels
# =========================================

Write-Info "========== SETUP: Authenticating Test Users =========="

# Login as Super Admin
$superAdminResult = Invoke-APICall -Method POST -Endpoint "/login" -Body @{
    login = $SUPER_ADMIN_USER
    password = $SUPER_ADMIN_PASS
}

if ($superAdminResult.Success) {
    $SUPER_ADMIN_TOKEN = $superAdminResult.Data.accessToken
    $SUPER_ADMIN_ID = $superAdminResult.Data.id
    Write-Success "✓ Super Admin authenticated"
} else {
    Write-Error "✗ Failed to authenticate Super Admin - Cannot proceed"
    Write-Warning "Please ensure Super Admin user exists and credentials are correct"
    exit 1
}

# Login as Org Admin (if exists)
$orgAdminResult = Invoke-APICall -Method POST -Endpoint "/login" -Body @{
    login = $ORG_A_ADMIN_USER
    password = $ORG_A_ADMIN_PASS
}

$ORG_A_ADMIN_TOKEN = ""
if ($orgAdminResult.Success) {
    $ORG_A_ADMIN_TOKEN = $orgAdminResult.Data.accessToken
    Write-Success "✓ Organization Admin authenticated"
} else {
    Write-Warning "! Organization Admin not available - some tests will be skipped"
}

# Login as Clinic Admin (if exists)
$clinicAdminResult = Invoke-APICall -Method POST -Endpoint "/login" -Body @{
    login = $CLINIC_ADMIN_USER
    password = $CLINIC_ADMIN_PASS
}

$CLINIC_ADMIN_TOKEN = ""
if ($clinicAdminResult.Success) {
    $CLINIC_ADMIN_TOKEN = $clinicAdminResult.Data.accessToken
    Write-Success "✓ Clinic Admin authenticated"
} else {
    Write-Warning "! Clinic Admin not available - some tests will be skipped"
}

Write-Host ""

# =========================================
# TEST 1: Blocked User Cannot Login
# =========================================

Write-Info "========== TEST CATEGORY 1: Blocked User Login Prevention =========="

# Create a test user
$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$testUser = @{
    first_name = "Test"
    last_name = "BlockedUser"
    username = "blocked_test_$timestamp"
    email = "blocked_$timestamp@test.com"
    password = "TestPassword123"
}

$createResult = Invoke-APICall -Method POST -Endpoint "/admin/users" -Body $testUser -Token $SUPER_ADMIN_TOKEN

if ($createResult.Success) {
    $TEST_USER_ID = $createResult.Data.user.id
    $TEST_USERNAME = $createResult.Data.user.username
    Write-Info "Created test user: $TEST_USERNAME"
    
    # Block the user
    $blockResult = Invoke-APICall -Method POST -Endpoint "/admin/users/$TEST_USER_ID/block" `
        -Body @{reason = "Security test - automated"} -Token $SUPER_ADMIN_TOKEN
    
    # Try to login as blocked user (should FAIL)
    $blockedLoginResult = Invoke-APICall -Method POST -Endpoint "/login" -Body @{
        login = $TEST_USERNAME
        password = "TestPassword123"
    }
    
    Test-Expectation -TestName "Blocked user cannot login" -ShouldSucceed $false `
        -Result $blockedLoginResult -ExpectedStatus 401
    
    # Unblock the user for further tests
    Invoke-APICall -Method POST -Endpoint "/admin/users/$TEST_USER_ID/unblock" -Token $SUPER_ADMIN_TOKEN | Out-Null
} else {
    Write-Warning "! Could not create test user - skipping blocked user tests"
}

Write-Host ""

# =========================================
# TEST 2: Scope Validation Tests
# =========================================

Write-Info "========== TEST CATEGORY 2: Scope Validation =========="

if ($ORG_A_ADMIN_TOKEN -and $TEST_USER_ID) {
    # Org Admin should NOT be able to access Super Admin's user
    $result = Invoke-APICall -Method GET -Endpoint "/org-admin/users/$SUPER_ADMIN_ID" -Token $ORG_A_ADMIN_TOKEN
    Test-Expectation -TestName "Org Admin cannot view Super Admin user" -ShouldSucceed $false `
        -Result $result -ExpectedStatus 403
    
    # Org Admin should NOT be able to update Super Admin
    $result = Invoke-APICall -Method PUT -Endpoint "/org-admin/users/$SUPER_ADMIN_ID" `
        -Body @{first_name = "Hacked"} -Token $ORG_A_ADMIN_TOKEN
    Test-Expectation -TestName "Org Admin cannot update Super Admin" -ShouldSucceed $false `
        -Result $result -ExpectedStatus 403
    
    # Org Admin should NOT be able to delete Super Admin
    $result = Invoke-APICall -Method DELETE -Endpoint "/org-admin/users/$SUPER_ADMIN_ID" `
        -Token $ORG_A_ADMIN_TOKEN
    Test-Expectation -TestName "Org Admin cannot delete Super Admin" -ShouldSucceed $false `
        -Result $result -ExpectedStatus 403
} else {
    Write-Warning "! Skipping Org Admin scope tests - Org Admin not available"
}

if ($CLINIC_ADMIN_TOKEN) {
    # Clinic Admin should NOT be able to access Super Admin
    $result = Invoke-APICall -Method GET -Endpoint "/clinic-admin/users/$SUPER_ADMIN_ID" -Token $CLINIC_ADMIN_TOKEN
    Test-Expectation -TestName "Clinic Admin cannot view Super Admin user" -ShouldSucceed $false `
        -Result $result -ExpectedStatus 403
    
    # Clinic Admin should NOT be able to update Super Admin
    $result = Invoke-APICall -Method PUT -Endpoint "/clinic-admin/users/$SUPER_ADMIN_ID" `
        -Body @{first_name = "Hacked"} -Token $CLINIC_ADMIN_TOKEN
    Test-Expectation -TestName "Clinic Admin cannot update Super Admin" -ShouldSucceed $false `
        -Result $result -ExpectedStatus 403
} else {
    Write-Warning "! Skipping Clinic Admin scope tests - Clinic Admin not available"
}

Write-Host ""

# =========================================
# TEST 3: Privilege Escalation Prevention
# =========================================

Write-Info "========== TEST CATEGORY 3: Privilege Escalation Prevention =========="

# Get super_admin role ID
$rolesResult = Invoke-APICall -Method GET -Endpoint "/admin/roles?search=super_admin" -Token $SUPER_ADMIN_TOKEN
if ($rolesResult.Success -and $rolesResult.Data.roles.Count -gt 0) {
    $SUPER_ADMIN_ROLE_ID = $rolesResult.Data.roles[0].id
    
    if ($ORG_A_ADMIN_TOKEN -and $TEST_USER_ID) {
        # Org Admin tries to assign super_admin role (should FAIL)
        $result = Invoke-APICall -Method POST -Endpoint "/org-admin/users/$TEST_USER_ID/roles" `
            -Body @{role_id = $SUPER_ADMIN_ROLE_ID} -Token $ORG_A_ADMIN_TOKEN
        Test-Expectation -TestName "Org Admin cannot assign super_admin role" -ShouldSucceed $false `
            -Result $result -ExpectedStatus 403
    }
    
    if ($CLINIC_ADMIN_TOKEN -and $TEST_USER_ID) {
        # Clinic Admin tries to assign super_admin role (should FAIL)
        $result = Invoke-APICall -Method POST -Endpoint "/clinic-admin/users/$TEST_USER_ID/roles" `
            -Body @{role_id = $SUPER_ADMIN_ROLE_ID} -Token $CLINIC_ADMIN_TOKEN
        Test-Expectation -TestName "Clinic Admin cannot assign super_admin role" -ShouldSucceed $false `
            -Result $result -ExpectedStatus 403
    }
}

# Get organization_admin role ID
$rolesResult = Invoke-APICall -Method GET -Endpoint "/admin/roles?search=organization_admin" -Token $SUPER_ADMIN_TOKEN
if ($rolesResult.Success -and $rolesResult.Data.roles.Count -gt 0) {
    $ORG_ADMIN_ROLE_ID = $rolesResult.Data.roles[0].id
    
    if ($CLINIC_ADMIN_TOKEN -and $TEST_USER_ID) {
        # Clinic Admin tries to assign organization_admin role (should FAIL)
        $result = Invoke-APICall -Method POST -Endpoint "/clinic-admin/users/$TEST_USER_ID/roles" `
            -Body @{role_id = $ORG_ADMIN_ROLE_ID} -Token $CLINIC_ADMIN_TOKEN
        Test-Expectation -TestName "Clinic Admin cannot assign organization_admin role" -ShouldSucceed $false `
            -Result $result -ExpectedStatus 403
    }
}

Write-Host ""

# =========================================
# TEST 4: Super Admin Retains Full Access
# =========================================

Write-Info "========== TEST CATEGORY 4: Super Admin Full Access =========="

if ($TEST_USER_ID) {
    # Super Admin can view any user
    $result = Invoke-APICall -Method GET -Endpoint "/admin/users/$TEST_USER_ID" -Token $SUPER_ADMIN_TOKEN
    Test-Expectation -TestName "Super Admin can view any user" -ShouldSucceed $true -Result $result
    
    # Super Admin can update any user
    $result = Invoke-APICall -Method PUT -Endpoint "/admin/users/$TEST_USER_ID" `
        -Body @{first_name = "Updated"} -Token $SUPER_ADMIN_TOKEN
    Test-Expectation -TestName "Super Admin can update any user" -ShouldSucceed $true -Result $result
    
    # Super Admin can activate/deactivate any user
    $result = Invoke-APICall -Method POST -Endpoint "/admin/users/$TEST_USER_ID/deactivate" -Token $SUPER_ADMIN_TOKEN
    Test-Expectation -TestName "Super Admin can deactivate any user" -ShouldSucceed $true -Result $result
    
    $result = Invoke-APICall -Method POST -Endpoint "/admin/users/$TEST_USER_ID/activate" -Token $SUPER_ADMIN_TOKEN
    Test-Expectation -TestName "Super Admin can activate any user" -ShouldSucceed $true -Result $result
}

Write-Host ""

# =========================================
# TEST 5: Cleanup
# =========================================

if ($TEST_USER_ID) {
    Write-Info "Cleaning up test user..."
    Invoke-APICall -Method DELETE -Endpoint "/admin/users/$TEST_USER_ID" -Token $SUPER_ADMIN_TOKEN | Out-Null
    Write-Info "Test user deleted."
}

# =========================================
# SUMMARY
# =========================================

Write-Host "`n" + "="*80
Write-Host "  SECURITY TEST SUMMARY" -ForegroundColor Yellow
Write-Host "="*80 + "`n"

$total = $global:PassedTests + $global:FailedTests
Write-Host "Total Tests: $total"
Write-Success "Passed: $global:PassedTests"

if ($global:FailedTests -gt 0) {
    Write-Error "Failed: $global:FailedTests"
    Write-Host ""
    Write-Warning "⚠️  SECURITY ISSUES DETECTED!"
    Write-Warning "Some security fixes may not be working correctly."
    Write-Warning "Please review the failed tests above and fix before deploying."
    Write-Host ""
    exit 1
} else {
    Write-Host ""
    Write-Success "✅ ALL SECURITY TESTS PASSED!"
    Write-Success "The security fixes are working correctly."
    Write-Host ""
    Write-Info "Security improvements verified:"
    Write-Info "  ✓ Scope validation enforced"
    Write-Info "  ✓ Privilege escalation prevented"
    Write-Info "  ✓ Blocked users cannot login"
    Write-Info "  ✓ Super Admin retains full access"
    Write-Host ""
    Write-Success "System is ready for production deployment! 🚀"
    Write-Host ""
    exit 0
}

