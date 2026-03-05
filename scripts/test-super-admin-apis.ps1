# Super Admin API Test Script
# This script tests all Super Admin User and Role Management APIs

# Configuration
$BASE_URL = "http://localhost:8000/api/v1/auth"
$SUPER_ADMIN_USERNAME = "superadmin"
$SUPER_ADMIN_PASSWORD = "SuperAdmin123"

# Colors for output
function Write-Success { param($message) Write-Host $message -ForegroundColor Green }
function Write-Error { param($message) Write-Host $message -ForegroundColor Red }
function Write-Info { param($message) Write-Host $message -ForegroundColor Cyan }
function Write-Warning { param($message) Write-Host $message -ForegroundColor Yellow }

# Global variables for storing test data
$global:AccessToken = ""
$global:TestUserID = ""
$global:TestRoleID = ""

# Helper function to make API calls
function Invoke-APICall {
    param(
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null,
        [bool]$RequireAuth = $true
    )

    $headers = @{
        "Content-Type" = "application/json"
    }

    if ($RequireAuth -and $global:AccessToken) {
        $headers["Authorization"] = "Bearer $global:AccessToken"
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
        return @{ Success = $true; Data = $response }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        $errorBody = $_.ErrorDetails.Message | ConvertFrom-Json
        return @{ Success = $false; StatusCode = $statusCode; Error = $errorBody }
    }
}

# Test 1: Login as Super Admin
function Test-SuperAdminLogin {
    Write-Info "`n========== Test 1: Super Admin Login =========="
    
    $body = @{
        login = $SUPER_ADMIN_USERNAME
        password = $SUPER_ADMIN_PASSWORD
    }

    $result = Invoke-APICall -Method POST -Endpoint "/login" -Body $body -RequireAuth $false

    if ($result.Success) {
        $global:AccessToken = $result.Data.accessToken
        Write-Success "✓ Login successful"
        Write-Host "  User ID: $($result.Data.id)"
        Write-Host "  Roles: $($result.Data.roles | ForEach-Object { $_.name } | Join-String -Separator ', ')"
        return $true
    } else {
        Write-Error "✗ Login failed: $($result.Error.error)"
        return $false
    }
}

# Test 2: List All Users
function Test-ListUsers {
    Write-Info "`n========== Test 2: List All Users =========="
    
    $result = Invoke-APICall -Method GET -Endpoint "/admin/users?page=1&page_size=10"

    if ($result.Success) {
        Write-Success "✓ Users listed successfully"
        Write-Host "  Total users: $($result.Data.pagination.total_count)"
        Write-Host "  Page: $($result.Data.pagination.page) of $($result.Data.pagination.total_pages)"
        return $true
    } else {
        Write-Error "✗ Failed to list users: $($result.Error.message)"
        return $false
    }
}

# Test 3: Create New User
function Test-CreateUser {
    Write-Info "`n========== Test 3: Create New User =========="
    
    $timestamp = Get-Date -Format "yyyyMMddHHmmss"
    $body = @{
        first_name = "Test"
        last_name = "User"
        username = "testuser_$timestamp"
        email = "testuser_$timestamp@example.com"
        phone = "+1234567890"
        password = "TestPassword123"
        is_active = $true
    }

    $result = Invoke-APICall -Method POST -Endpoint "/admin/users" -Body $body

    if ($result.Success) {
        $global:TestUserID = $result.Data.user.id
        Write-Success "✓ User created successfully"
        Write-Host "  User ID: $($global:TestUserID)"
        Write-Host "  Username: $($result.Data.user.username)"
        return $true
    } else {
        Write-Error "✗ Failed to create user: $($result.Error.message)"
        return $false
    }
}

# Test 4: Get Single User
function Test-GetUser {
    Write-Info "`n========== Test 4: Get Single User =========="
    
    if (-not $global:TestUserID) {
        Write-Warning "! Skipping: No test user ID available"
        return $false
    }

    $result = Invoke-APICall -Method GET -Endpoint "/admin/users/$global:TestUserID"

    if ($result.Success) {
        Write-Success "✓ User retrieved successfully"
        Write-Host "  Username: $($result.Data.username)"
        Write-Host "  Email: $($result.Data.email)"
        Write-Host "  Active: $($result.Data.is_active)"
        Write-Host "  Blocked: $($result.Data.is_blocked)"
        return $true
    } else {
        Write-Error "✗ Failed to get user: $($result.Error.message)"
        return $false
    }
}

# Test 5: Update User
function Test-UpdateUser {
    Write-Info "`n========== Test 5: Update User =========="
    
    if (-not $global:TestUserID) {
        Write-Warning "! Skipping: No test user ID available"
        return $false
    }

    $body = @{
        first_name = "Updated"
        last_name = "User"
        gender = "male"
    }

    $result = Invoke-APICall -Method PUT -Endpoint "/admin/users/$global:TestUserID" -Body $body

    if ($result.Success) {
        Write-Success "✓ User updated successfully"
        return $true
    } else {
        Write-Error "✗ Failed to update user: $($result.Error.message)"
        return $false
    }
}

# Test 6: Block User
function Test-BlockUser {
    Write-Info "`n========== Test 6: Block User =========="
    
    if (-not $global:TestUserID) {
        Write-Warning "! Skipping: No test user ID available"
        return $false
    }

    $body = @{
        reason = "Test blocking - automated test script"
    }

    $result = Invoke-APICall -Method POST -Endpoint "/admin/users/$global:TestUserID/block" -Body $body

    if ($result.Success) {
        Write-Success "✓ User blocked successfully"
        return $true
    } else {
        Write-Error "✗ Failed to block user: $($result.Error.message)"
        return $false
    }
}

# Test 7: Unblock User
function Test-UnblockUser {
    Write-Info "`n========== Test 7: Unblock User =========="
    
    if (-not $global:TestUserID) {
        Write-Warning "! Skipping: No test user ID available"
        return $false
    }

    $result = Invoke-APICall -Method POST -Endpoint "/admin/users/$global:TestUserID/unblock"

    if ($result.Success) {
        Write-Success "✓ User unblocked successfully"
        return $true
    } else {
        Write-Error "✗ Failed to unblock user: $($result.Error.message)"
        return $false
    }
}

# Test 8: Deactivate User
function Test-DeactivateUser {
    Write-Info "`n========== Test 8: Deactivate User =========="
    
    if (-not $global:TestUserID) {
        Write-Warning "! Skipping: No test user ID available"
        return $false
    }

    $result = Invoke-APICall -Method POST -Endpoint "/admin/users/$global:TestUserID/deactivate"

    if ($result.Success) {
        Write-Success "✓ User deactivated successfully"
        return $true
    } else {
        Write-Error "✗ Failed to deactivate user: $($result.Error.message)"
        return $false
    }
}

# Test 9: Activate User
function Test-ActivateUser {
    Write-Info "`n========== Test 9: Activate User =========="
    
    if (-not $global:TestUserID) {
        Write-Warning "! Skipping: No test user ID available"
        return $false
    }

    $result = Invoke-APICall -Method POST -Endpoint "/admin/users/$global:TestUserID/activate"

    if ($result.Success) {
        Write-Success "✓ User activated successfully"
        return $true
    } else {
        Write-Error "✗ Failed to activate user: $($result.Error.message)"
        return $false
    }
}

# Test 10: Admin Change Password
function Test-AdminChangePassword {
    Write-Info "`n========== Test 10: Admin Change Password =========="
    
    if (-not $global:TestUserID) {
        Write-Warning "! Skipping: No test user ID available"
        return $false
    }

    $body = @{
        new_password = "NewTestPassword123"
    }

    $result = Invoke-APICall -Method POST -Endpoint "/admin/users/$global:TestUserID/change-password" -Body $body

    if ($result.Success) {
        Write-Success "✓ Password changed successfully"
        return $true
    } else {
        Write-Error "✗ Failed to change password: $($result.Error.message)"
        return $false
    }
}

# Test 11: List All Roles
function Test-ListRoles {
    Write-Info "`n========== Test 11: List All Roles =========="
    
    $result = Invoke-APICall -Method GET -Endpoint "/admin/roles?page=1&page_size=20"

    if ($result.Success) {
        Write-Success "✓ Roles listed successfully"
        Write-Host "  Total roles: $($result.Data.pagination.total_count)"
        Write-Host "  Available roles:"
        $result.Data.roles | ForEach-Object {
            Write-Host "    - $($_.name) (Users: $($_.users_count), System: $($_.is_system_role))"
        }
        return $true
    } else {
        Write-Error "✗ Failed to list roles: $($result.Error.message)"
        return $false
    }
}

# Test 12: Create New Role
function Test-CreateRole {
    Write-Info "`n========== Test 12: Create New Role =========="
    
    $timestamp = Get-Date -Format "yyyyMMddHHmmss"
    $body = @{
        name = "Test Role $timestamp"
        description = "Test role created by automated script"
        permissions = @{
            users = @("read", "update")
            reports = @("read", "create")
        }
    }

    $result = Invoke-APICall -Method POST -Endpoint "/admin/roles" -Body $body

    if ($result.Success) {
        $global:TestRoleID = $result.Data.role_id
        Write-Success "✓ Role created successfully"
        Write-Host "  Role ID: $($global:TestRoleID)"
        Write-Host "  Role Name: $($result.Data.name)"
        return $true
    } else {
        Write-Error "✗ Failed to create role: $($result.Error.message)"
        return $false
    }
}

# Test 13: Get Single Role
function Test-GetRole {
    Write-Info "`n========== Test 13: Get Single Role =========="
    
    if (-not $global:TestRoleID) {
        Write-Warning "! Skipping: No test role ID available"
        return $false
    }

    $result = Invoke-APICall -Method GET -Endpoint "/admin/roles/$global:TestRoleID"

    if ($result.Success) {
        Write-Success "✓ Role retrieved successfully"
        Write-Host "  Role Name: $($result.Data.name)"
        Write-Host "  Description: $($result.Data.description)"
        Write-Host "  System Role: $($result.Data.is_system_role)"
        Write-Host "  Active: $($result.Data.is_active)"
        return $true
    } else {
        Write-Error "✗ Failed to get role: $($result.Error.message)"
        return $false
    }
}

# Test 14: Update Role Permissions
function Test-UpdateRolePermissions {
    Write-Info "`n========== Test 14: Update Role Permissions =========="
    
    if (-not $global:TestRoleID) {
        Write-Warning "! Skipping: No test role ID available"
        return $false
    }

    $body = @{
        permissions = @{
            users = @("read", "create", "update")
            reports = @("read", "create", "update")
            dashboard = @("read")
        }
    }

    $result = Invoke-APICall -Method PUT -Endpoint "/admin/roles/$global:TestRoleID/permissions" -Body $body

    if ($result.Success) {
        Write-Success "✓ Role permissions updated successfully"
        return $true
    } else {
        Write-Error "✗ Failed to update permissions: $($result.Error.message)"
        return $false
    }
}

# Test 15: Assign Role to User
function Test-AssignRole {
    Write-Info "`n========== Test 15: Assign Role to User =========="
    
    if (-not $global:TestUserID -or -not $global:TestRoleID) {
        Write-Warning "! Skipping: No test user or role ID available"
        return $false
    }

    $body = @{
        role_id = $global:TestRoleID
    }

    $result = Invoke-APICall -Method POST -Endpoint "/admin/users/$global:TestUserID/roles" -Body $body

    if ($result.Success) {
        Write-Success "✓ Role assigned to user successfully"
        return $true
    } else {
        Write-Error "✗ Failed to assign role: $($result.Error.message)"
        return $false
    }
}

# Test 16: Get User Activity Logs
function Test-GetUserActivityLogs {
    Write-Info "`n========== Test 16: Get User Activity Logs =========="
    
    if (-not $global:TestUserID) {
        Write-Warning "! Skipping: No test user ID available"
        return $false
    }

    $result = Invoke-APICall -Method GET -Endpoint "/admin/users/$global:TestUserID/activity-logs?page=1&page_size=10"

    if ($result.Success) {
        Write-Success "✓ Activity logs retrieved successfully"
        Write-Host "  Total logs: $($result.Data.pagination.total_count)"
        if ($result.Data.logs.Count -gt 0) {
            Write-Host "  Recent activities:"
            $result.Data.logs | Select-Object -First 5 | ForEach-Object {
                Write-Host "    - $($_.action_type): $($_.action_description)"
            }
        }
        return $true
    } else {
        Write-Error "✗ Failed to get activity logs: $($result.Error.message)"
        return $false
    }
}

# Test 17: Get Role Users
function Test-GetRoleUsers {
    Write-Info "`n========== Test 17: Get Role Users =========="
    
    if (-not $global:TestRoleID) {
        Write-Warning "! Skipping: No test role ID available"
        return $false
    }

    $result = Invoke-APICall -Method GET -Endpoint "/admin/roles/$global:TestRoleID/users"

    if ($result.Success) {
        Write-Success "✓ Role users retrieved successfully"
        Write-Host "  Total users with this role: $($result.Data.pagination.total_count)"
        return $true
    } else {
        Write-Error "✗ Failed to get role users: $($result.Error.message)"
        return $false
    }
}

# Test 18: Remove Role from User
function Test-RemoveRole {
    Write-Info "`n========== Test 18: Remove Role from User =========="
    
    if (-not $global:TestUserID -or -not $global:TestRoleID) {
        Write-Warning "! Skipping: No test user or role ID available"
        return $false
    }

    $result = Invoke-APICall -Method DELETE -Endpoint "/admin/users/$global:TestUserID/roles/$global:TestRoleID"

    if ($result.Success) {
        Write-Success "✓ Role removed from user successfully"
        return $true
    } else {
        Write-Error "✗ Failed to remove role: $($result.Error.message)"
        return $false
    }
}

# Test 19: Get Permission Templates
function Test-GetPermissionTemplates {
    Write-Info "`n========== Test 19: Get Permission Templates =========="
    
    $result = Invoke-APICall -Method GET -Endpoint "/admin/permission-templates"

    if ($result.Success) {
        Write-Success "✓ Permission templates retrieved successfully"
        Write-Host "  Available templates:"
        $result.Data.templates | ForEach-Object {
            Write-Host "    - $($_.name): $($_.description)"
        }
        return $true
    } else {
        Write-Error "✗ Failed to get permission templates: $($result.Error.message)"
        return $false
    }
}

# Test 20: Search Users
function Test-SearchUsers {
    Write-Info "`n========== Test 20: Search Users =========="
    
    $result = Invoke-APICall -Method GET -Endpoint "/admin/users?search=test&page=1&page_size=10"

    if ($result.Success) {
        Write-Success "✓ Users searched successfully"
        Write-Host "  Found users: $($result.Data.pagination.total_count)"
        return $true
    } else {
        Write-Error "✗ Failed to search users: $($result.Error.message)"
        return $false
    }
}

# Test 21: Filter Users by Role
function Test-FilterUsersByRole {
    Write-Info "`n========== Test 21: Filter Users by Role =========="
    
    $result = Invoke-APICall -Method GET -Endpoint "/admin/users?role=patient&page=1&page_size=10"

    if ($result.Success) {
        Write-Success "✓ Users filtered by role successfully"
        Write-Host "  Found users with patient role: $($result.Data.pagination.total_count)"
        return $true
    } else {
        Write-Error "✗ Failed to filter users: $($result.Error.message)"
        return $false
    }
}

# Test 22: Delete Role
function Test-DeleteRole {
    Write-Info "`n========== Test 22: Delete Role =========="
    
    if (-not $global:TestRoleID) {
        Write-Warning "! Skipping: No test role ID available"
        return $false
    }

    $result = Invoke-APICall -Method DELETE -Endpoint "/admin/roles/$global:TestRoleID"

    if ($result.Success) {
        Write-Success "✓ Role deleted successfully"
        return $true
    } else {
        Write-Error "✗ Failed to delete role: $($result.Error.message)"
        return $false
    }
}

# Test 23: Delete User
function Test-DeleteUser {
    Write-Info "`n========== Test 23: Delete User =========="
    
    if (-not $global:TestUserID) {
        Write-Warning "! Skipping: No test user ID available"
        return $false
    }

    $result = Invoke-APICall -Method DELETE -Endpoint "/admin/users/$global:TestUserID"

    if ($result.Success) {
        Write-Success "✓ User deleted successfully"
        return $true
    } else {
        Write-Error "✗ Failed to delete user: $($result.Error.message)"
        return $false
    }
}

# Main execution
Write-Host "`n" + "="*70
Write-Host "  Super Admin API Test Suite" -ForegroundColor Yellow
Write-Host "="*70

# Track test results
$testResults = @()

# Run tests
$testResults += @{ Name = "Super Admin Login"; Result = (Test-SuperAdminLogin) }

if ($global:AccessToken) {
    $testResults += @{ Name = "List All Users"; Result = (Test-ListUsers) }
    $testResults += @{ Name = "Create New User"; Result = (Test-CreateUser) }
    $testResults += @{ Name = "Get Single User"; Result = (Test-GetUser) }
    $testResults += @{ Name = "Update User"; Result = (Test-UpdateUser) }
    $testResults += @{ Name = "Block User"; Result = (Test-BlockUser) }
    $testResults += @{ Name = "Unblock User"; Result = (Test-UnblockUser) }
    $testResults += @{ Name = "Deactivate User"; Result = (Test-DeactivateUser) }
    $testResults += @{ Name = "Activate User"; Result = (Test-ActivateUser) }
    $testResults += @{ Name = "Admin Change Password"; Result = (Test-AdminChangePassword) }
    $testResults += @{ Name = "List All Roles"; Result = (Test-ListRoles) }
    $testResults += @{ Name = "Create New Role"; Result = (Test-CreateRole) }
    $testResults += @{ Name = "Get Single Role"; Result = (Test-GetRole) }
    $testResults += @{ Name = "Update Role Permissions"; Result = (Test-UpdateRolePermissions) }
    $testResults += @{ Name = "Assign Role to User"; Result = (Test-AssignRole) }
    $testResults += @{ Name = "Get User Activity Logs"; Result = (Test-GetUserActivityLogs) }
    $testResults += @{ Name = "Get Role Users"; Result = (Test-GetRoleUsers) }
    $testResults += @{ Name = "Remove Role from User"; Result = (Test-RemoveRole) }
    $testResults += @{ Name = "Get Permission Templates"; Result = (Test-GetPermissionTemplates) }
    $testResults += @{ Name = "Search Users"; Result = (Test-SearchUsers) }
    $testResults += @{ Name = "Filter Users by Role"; Result = (Test-FilterUsersByRole) }
    $testResults += @{ Name = "Delete Role"; Result = (Test-DeleteRole) }
    $testResults += @{ Name = "Delete User"; Result = (Test-DeleteUser) }
} else {
    Write-Error "`nCannot proceed without authentication. Please check credentials.`n"
}

# Summary
Write-Host "`n" + "="*70
Write-Host "  Test Summary" -ForegroundColor Yellow
Write-Host "="*70

$passed = ($testResults | Where-Object { $_.Result -eq $true }).Count
$failed = ($testResults | Where-Object { $_.Result -eq $false }).Count
$total = $testResults.Count

Write-Host "`nTotal Tests: $total"
Write-Success "Passed: $passed"
if ($failed -gt 0) {
    Write-Error "Failed: $failed"
}

Write-Host "`nDetailed Results:"
$testResults | ForEach-Object {
    $status = if ($_.Result) { "[✓]" } else { "[✗]" }
    $color = if ($_.Result) { "Green" } else { "Red" }
    Write-Host "  $status $($_.Name)" -ForegroundColor $color
}

Write-Host "`n" + "="*70 + "`n"

# Exit with appropriate code
if ($failed -eq 0 -and $passed -gt 0) {
    Write-Success "All tests passed successfully!`n"
    exit 0
} else {
    Write-Error "Some tests failed. Please review the output above.`n"
    exit 1
}

