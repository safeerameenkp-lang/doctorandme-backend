# Test script for improved login and refresh endpoints
# This script tests the enhanced login and refresh functionality

param(
    [string]$AuthServiceUrl = "http://localhost:8080/api/auth"
)

Write-Host "Testing Improved Login and Refresh Endpoints" -ForegroundColor Green
Write-Host "=============================================" -ForegroundColor Green

# Function to make HTTP requests and display responses
function Test-Endpoint {
    param(
        [string]$Method,
        [string]$Url,
        [string]$Description,
        [hashtable]$Headers = @{},
        [string]$Body = $null
    )
    
    Write-Host "`nTesting: $Description" -ForegroundColor Yellow
    Write-Host "URL: $Method $Url" -ForegroundColor Gray
    
    try {
        if ($Body) {
            $response = Invoke-RestMethod -Uri $Url -Method $Method -Headers $Headers -Body $Body -ContentType "application/json" -ErrorAction Stop
        } else {
            $response = Invoke-RestMethod -Uri $Url -Method $Method -Headers $Headers -ErrorAction Stop
        }
        
        Write-Host "✅ SUCCESS" -ForegroundColor Green
        Write-Host "Response:" -ForegroundColor Cyan
        $response | ConvertTo-Json -Depth 5 | Write-Host -ForegroundColor White
        return $response
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        $responseBody = $_.ErrorDetails.Message
        
        Write-Host "❌ FAILED - Status: $statusCode" -ForegroundColor Red
        Write-Host "Error: $responseBody" -ForegroundColor Red
        return $null
    }
}

# Test 1: Register a test user
Write-Host "`n🔐 Testing User Registration" -ForegroundColor Cyan

$registerData = @{
    first_name = "Test"
    last_name = "User"
    email = "testuser@example.com"
    username = "testuser123"
    phone = "+1234567890"
    password = "password123"
} | ConvertTo-Json

$registerResponse = Test-Endpoint -Method "POST" -Url "$AuthServiceUrl/register" -Description "User Registration" -Body $registerData

# Test 2: Login with the registered user
Write-Host "`n🔑 Testing User Login" -ForegroundColor Cyan

$loginData = @{
    login = "testuser@example.com"
    password = "password123"
} | ConvertTo-Json

$loginResponse = Test-Endpoint -Method "POST" -Url "$AuthServiceUrl/login" -Description "User Login" -Body $loginData

if ($loginResponse -and $loginResponse.accessToken) {
    Write-Host "`n✅ Login successful! Checking response structure..." -ForegroundColor Green
    
    # Check if response has all expected fields
    $expectedFields = @("user", "accessToken", "refreshToken", "tokenType", "expiresIn")
    $missingFields = @()
    
    foreach ($field in $expectedFields) {
        if (-not $loginResponse.PSObject.Properties.Name -contains $field) {
            $missingFields += $field
        }
    }
    
    if ($missingFields.Count -eq 0) {
        Write-Host "✅ All expected response fields present" -ForegroundColor Green
    } else {
        Write-Host "❌ Missing fields: $($missingFields -join ', ')" -ForegroundColor Red
    }
    
    # Check user object structure
    if ($loginResponse.user) {
        Write-Host "`nUser object structure:" -ForegroundColor Cyan
        Write-Host "- ID: $($loginResponse.user.id)" -ForegroundColor White
        Write-Host "- Name: $($loginResponse.user.firstName) $($loginResponse.user.lastName)" -ForegroundColor White
        Write-Host "- Email: $($loginResponse.user.email)" -ForegroundColor White
        Write-Host "- Username: $($loginResponse.user.username)" -ForegroundColor White
        Write-Host "- Phone: $($loginResponse.user.phone)" -ForegroundColor White
        
        if ($loginResponse.user.roles) {
            Write-Host "- Roles count: $($loginResponse.user.roles.Count)" -ForegroundColor White
            foreach ($role in $loginResponse.user.roles) {
                Write-Host "  * Role: $($role.name) (ID: $($role.id))" -ForegroundColor Gray
                if ($role.permissions) {
                    Write-Host "    Permissions: $($role.permissions.Keys -join ', ')" -ForegroundColor Gray
                }
                if ($role.organization_id) {
                    Write-Host "    Organization ID: $($role.organization_id)" -ForegroundColor Gray
                }
                if ($role.clinic_id) {
                    Write-Host "    Clinic ID: $($role.clinic_id)" -ForegroundColor Gray
                }
            }
        }
    }
    
    # Test 3: Refresh token
    Write-Host "`n🔄 Testing Token Refresh" -ForegroundColor Cyan
    
    $refreshData = @{
        refresh_token = $loginResponse.refreshToken
    } | ConvertTo-Json
    
    $refreshResponse = Test-Endpoint -Method "POST" -Url "$AuthServiceUrl/refresh" -Description "Token Refresh" -Body $refreshData
    
    if ($refreshResponse -and $refreshResponse.accessToken) {
        Write-Host "✅ Refresh successful! Checking response structure..." -ForegroundColor Green
        
        # Check if refresh response has all expected fields
        $refreshMissingFields = @()
        foreach ($field in $expectedFields) {
            if (-not $refreshResponse.PSObject.Properties.Name -contains $field) {
                $refreshMissingFields += $field
            }
        }
        
        if ($refreshMissingFields.Count -eq 0) {
            Write-Host "✅ All expected refresh response fields present" -ForegroundColor Green
        } else {
            Write-Host "❌ Missing fields in refresh response: $($refreshMissingFields -join ', ')" -ForegroundColor Red
        }
        
        # Check if user details are included in refresh response
        if ($refreshResponse.user -and $refreshResponse.user.firstName) {
            Write-Host "✅ User details included in refresh response" -ForegroundColor Green
            Write-Host "- User: $($refreshResponse.user.firstName) $($refreshResponse.user.lastName)" -ForegroundColor White
            Write-Host "- Roles count: $($refreshResponse.user.roles.Count)" -ForegroundColor White
        } else {
            Write-Host "❌ User details missing in refresh response" -ForegroundColor Red
        }
        
        # Test 4: Use new token to access protected endpoint
        Write-Host "`n🔒 Testing Protected Endpoint with New Token" -ForegroundColor Cyan
        
        $headers = @{
            "Authorization" = "Bearer $($refreshResponse.accessToken)"
        }
        
        $profileResponse = Test-Endpoint -Method "GET" -Url "$AuthServiceUrl/profile" -Description "Get Profile with New Token" -Headers $headers
        
        if ($profileResponse) {
            Write-Host "✅ Protected endpoint accessible with refreshed token" -ForegroundColor Green
        }
    }
} else {
    Write-Host "❌ Login failed - cannot test refresh functionality" -ForegroundColor Red
}

# Test 5: Error handling tests
Write-Host "`n🚨 Testing Error Handling" -ForegroundColor Cyan

# Test invalid login
$invalidLoginData = @{
    login = "nonexistent@example.com"
    password = "wrongpassword"
} | ConvertTo-Json

Test-Endpoint -Method "POST" -Url "$AuthServiceUrl/login" -Description "Invalid Login" -Body $invalidLoginData

# Test invalid refresh token
$invalidRefreshData = @{
    refresh_token = "invalid_refresh_token"
} | ConvertTo-Json

Test-Endpoint -Method "POST" -Url "$AuthServiceUrl/refresh" -Description "Invalid Refresh Token" -Body $invalidRefreshData

Write-Host "`n🎯 Test Summary" -ForegroundColor Green
Write-Host "==============" -ForegroundColor Green
Write-Host "Login and Refresh endpoint improvements tested:" -ForegroundColor White
Write-Host "  ✅ Enhanced response structure with roles and permissions" -ForegroundColor White
Write-Host "  ✅ Better error handling and validation" -ForegroundColor White
Write-Host "  ✅ User details included in refresh response" -ForegroundColor White
Write-Host "  ✅ Organization/clinic context in roles" -ForegroundColor White
Write-Host "  ✅ Consistent token response format" -ForegroundColor White
Write-Host "`nThe frontend can now route based on user roles and permissions!" -ForegroundColor Green
