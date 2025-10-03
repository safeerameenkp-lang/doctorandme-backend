# Test script to verify authentication and authorization error handling
# This script tests all endpoints to ensure proper error responses

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$AuthServiceUrl = "http://localhost:8080",
    [string]$OrgServiceUrl = "http://localhost:8081", 
    [string]$AppointmentServiceUrl = "http://localhost:8082"
)

Write-Host "Testing Authentication and Authorization Error Handling" -ForegroundColor Green
Write-Host "=====================================================" -ForegroundColor Green

# Function to test endpoint and validate error response
function Test-Endpoint {
    param(
        [string]$Method,
        [string]$Url,
        [string]$Description,
        [string]$ExpectedErrorCode,
        [hashtable]$Headers = @{},
        [string]$Body = $null
    )
    
    Write-Host "`nTesting: $Description" -ForegroundColor Yellow
    Write-Host "URL: $Method $Url" -ForegroundColor Gray
    
    try {
        $response = Invoke-RestMethod -Uri $Url -Method $Method -Headers $Headers -Body $Body -ContentType "application/json" -ErrorAction Stop
        Write-Host "❌ FAIL: Expected error but got success response" -ForegroundColor Red
        Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Red
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        $responseBody = $_.ErrorDetails.Message
        
        try {
            $errorObj = $responseBody | ConvertFrom-Json
            $actualErrorCode = $errorObj.code
            $errorMessage = $errorObj.error
            $detailedMessage = $errorObj.message
            
            if ($actualErrorCode -eq $ExpectedErrorCode) {
                Write-Host "✅ PASS: Status $statusCode, Code: $actualErrorCode" -ForegroundColor Green
                Write-Host "   Error: $errorMessage" -ForegroundColor Gray
                Write-Host "   Details: $detailedMessage" -ForegroundColor Gray
            } else {
                Write-Host "❌ FAIL: Expected code '$ExpectedErrorCode' but got '$actualErrorCode'" -ForegroundColor Red
                Write-Host "   Status: $statusCode" -ForegroundColor Red
                Write-Host "   Response: $responseBody" -ForegroundColor Red
            }
        }
        catch {
            Write-Host "❌ FAIL: Invalid JSON response" -ForegroundColor Red
            Write-Host "   Status: $statusCode" -ForegroundColor Red
            Write-Host "   Response: $responseBody" -ForegroundColor Red
        }
    }
}

# Test 1: Missing Token Tests
Write-Host "`n🔐 Testing Missing Token Scenarios" -ForegroundColor Cyan

# Auth Service - Protected endpoints without token
Test-Endpoint -Method "GET" -Url "$AuthServiceUrl/api/profile" -Description "Get Profile without token" -ExpectedErrorCode "MISSING_TOKEN"

# Organization Service - Protected endpoints without token  
Test-Endpoint -Method "GET" -Url "$OrgServiceUrl/api/organizations" -Description "Get Organizations without token" -ExpectedErrorCode "MISSING_TOKEN"
Test-Endpoint -Method "POST" -Url "$OrgServiceUrl/api/organizations" -Description "Create Organization without token" -ExpectedErrorCode "MISSING_TOKEN" -Body '{"name":"Test Org"}'

# Appointment Service - Protected endpoints without token
Test-Endpoint -Method "GET" -Url "$AppointmentServiceUrl/api/appointments" -Description "Get Appointments without token" -ExpectedErrorCode "MISSING_TOKEN"
Test-Endpoint -Method "POST" -Url "$AppointmentServiceUrl/api/appointments" -Description "Create Appointment without token" -ExpectedErrorCode "MISSING_TOKEN" -Body '{"patient_id":"123","clinic_id":"456","doctor_id":"789","appointment_time":"2024-01-01 10:00:00","consultation_type":"new"}'

# Test 2: Invalid Token Tests
Write-Host "`n🔑 Testing Invalid Token Scenarios" -ForegroundColor Cyan

$invalidTokenHeaders = @{
    "Authorization" = "Bearer invalid_token_here"
}

Test-Endpoint -Method "GET" -Url "$AuthServiceUrl/api/profile" -Description "Get Profile with invalid token" -ExpectedErrorCode "INVALID_TOKEN" -Headers $invalidTokenHeaders
Test-Endpoint -Method "GET" -Url "$OrgServiceUrl/api/organizations" -Description "Get Organizations with invalid token" -ExpectedErrorCode "INVALID_TOKEN" -Headers $invalidTokenHeaders
Test-Endpoint -Method "GET" -Url "$AppointmentServiceUrl/api/appointments" -Description "Get Appointments with invalid token" -ExpectedErrorCode "INVALID_TOKEN" -Headers $invalidTokenHeaders

# Test 3: Malformed Token Tests
Write-Host "`n🔧 Testing Malformed Token Scenarios" -ForegroundColor Cyan

$malformedTokenHeaders = @{
    "Authorization" = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid"
}

Test-Endpoint -Method "GET" -Url "$AuthServiceUrl/api/profile" -Description "Get Profile with malformed token" -ExpectedErrorCode "INVALID_TOKEN" -Headers $malformedTokenHeaders
Test-Endpoint -Method "GET" -Url "$OrgServiceUrl/api/organizations" -Description "Get Organizations with malformed token" -ExpectedErrorCode "INVALID_TOKEN" -Headers $malformedTokenHeaders

# Test 4: Insufficient Permissions Tests (requires valid token with limited role)
Write-Host "`n🚫 Testing Insufficient Permissions Scenarios" -ForegroundColor Cyan
Write-Host "Note: These tests require a valid token with limited permissions" -ForegroundColor Yellow

# For these tests, you would need to:
# 1. Create a user with limited role (e.g., patient role)
# 2. Login to get a valid token
# 3. Test endpoints that require higher privileges

# Example test structure (commented out as it requires setup):
# $patientTokenHeaders = @{
#     "Authorization" = "Bearer $patientToken"
# }
# Test-Endpoint -Method "POST" -Url "$OrgServiceUrl/api/organizations" -Description "Create Organization as patient" -ExpectedErrorCode "INSUFFICIENT_PERMISSIONS" -Headers $patientTokenHeaders -Body '{"name":"Test Org"}'

# Test 5: Health Check Endpoints (should work without auth)
Write-Host "`n🏥 Testing Health Check Endpoints (should work without auth)" -ForegroundColor Cyan

try {
    $response = Invoke-RestMethod -Uri "$AuthServiceUrl/api/health" -Method "GET"
    Write-Host "✅ Auth Service Health Check: PASS" -ForegroundColor Green
} catch {
    Write-Host "❌ Auth Service Health Check: FAIL" -ForegroundColor Red
}

try {
    $response = Invoke-RestMethod -Uri "$OrgServiceUrl/api/health" -Method "GET"
    Write-Host "✅ Organization Service Health Check: PASS" -ForegroundColor Green
} catch {
    Write-Host "❌ Organization Service Health Check: FAIL" -ForegroundColor Red
}

try {
    $response = Invoke-RestMethod -Uri "$AppointmentServiceUrl/api/health" -Method "GET"
    Write-Host "✅ Appointment Service Health Check: PASS" -ForegroundColor Green
} catch {
    Write-Host "❌ Appointment Service Health Check: FAIL" -ForegroundColor Red
}

# Test 6: Public Auth Endpoints (should work without auth)
Write-Host "`n🔓 Testing Public Auth Endpoints (should work without auth)" -ForegroundColor Cyan

try {
    $response = Invoke-RestMethod -Uri "$AuthServiceUrl/api/login" -Method "POST" -Body '{"login":"test","password":"test"}' -ContentType "application/json"
    Write-Host "✅ Login endpoint accessible: PASS" -ForegroundColor Green
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    if ($statusCode -eq 401) {
        Write-Host "✅ Login endpoint accessible (expected auth failure): PASS" -ForegroundColor Green
    } else {
        Write-Host "❌ Login endpoint test: FAIL" -ForegroundColor Red
    }
}

try {
    $response = Invoke-RestMethod -Uri "$AuthServiceUrl/api/register" -Method "POST" -Body '{"first_name":"Test","last_name":"User","username":"testuser","password":"testpass123"}' -ContentType "application/json"
    Write-Host "✅ Register endpoint accessible: PASS" -ForegroundColor Green
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    if ($statusCode -eq 400 -or $statusCode -eq 409) {
        Write-Host "✅ Register endpoint accessible (expected validation/conflict): PASS" -ForegroundColor Green
    } else {
        Write-Host "❌ Register endpoint test: FAIL" -ForegroundColor Red
    }
}

Write-Host "`n🎯 Test Summary" -ForegroundColor Green
Write-Host "==============" -ForegroundColor Green
Write-Host "All authentication and authorization error handling tests completed." -ForegroundColor White
Write-Host "Check the results above to ensure all endpoints properly handle:" -ForegroundColor White
Write-Host "  • Missing tokens (MISSING_TOKEN)" -ForegroundColor White
Write-Host "  • Invalid tokens (INVALID_TOKEN)" -ForegroundColor White
Write-Host "  • Malformed tokens (INVALID_TOKEN)" -ForegroundColor White
Write-Host "  • Insufficient permissions (INSUFFICIENT_PERMISSIONS)" -ForegroundColor White
Write-Host "  • Health checks work without authentication" -ForegroundColor White
Write-Host "  • Public endpoints are accessible" -ForegroundColor White
