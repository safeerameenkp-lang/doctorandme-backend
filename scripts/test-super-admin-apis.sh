#!/bin/bash

# Super Admin API Test Script (Bash version)
# This script tests all Super Admin User and Role Management APIs

# Configuration
BASE_URL="http://localhost:8000/api/v1/auth"
SUPER_ADMIN_USERNAME="superadmin"
SUPER_ADMIN_PASSWORD="SuperAdmin123"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Global variables
ACCESS_TOKEN=""
TEST_USER_ID=""
TEST_ROLE_ID=""
PASSED_TESTS=0
FAILED_TESTS=0

# Helper functions
write_success() {
    echo -e "${GREEN}$1${NC}"
}

write_error() {
    echo -e "${RED}$1${NC}"
}

write_info() {
    echo -e "${CYAN}$1${NC}"
}

write_warning() {
    echo -e "${YELLOW}$1${NC}"
}

# API call helper
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    local require_auth=${4:-true}
    
    local headers=(-H "Content-Type: application/json")
    
    if [ "$require_auth" = true ] && [ -n "$ACCESS_TOKEN" ]; then
        headers+=(-H "Authorization: Bearer $ACCESS_TOKEN")
    fi
    
    if [ -n "$data" ]; then
        curl -s -X "$method" "${BASE_URL}${endpoint}" "${headers[@]}" -d "$data"
    else
        curl -s -X "$method" "${BASE_URL}${endpoint}" "${headers[@]}"
    fi
}

# Test 1: Login as Super Admin
test_super_admin_login() {
    write_info "\n========== Test 1: Super Admin Login =========="
    
    local data="{\"login\":\"$SUPER_ADMIN_USERNAME\",\"password\":\"$SUPER_ADMIN_PASSWORD\"}"
    local response=$(api_call "POST" "/login" "$data" false)
    
    if echo "$response" | jq -e '.accessToken' > /dev/null 2>&1; then
        ACCESS_TOKEN=$(echo "$response" | jq -r '.accessToken')
        write_success "✓ Login successful"
        echo "  User ID: $(echo "$response" | jq -r '.id')"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Login failed"
        echo "$response" | jq -r '.error' 2>/dev/null || echo "$response"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Test 2: List All Users
test_list_users() {
    write_info "\n========== Test 2: List All Users =========="
    
    local response=$(api_call "GET" "/admin/users?page=1&page_size=10")
    
    if echo "$response" | jq -e '.users' > /dev/null 2>&1; then
        write_success "✓ Users listed successfully"
        echo "  Total users: $(echo "$response" | jq -r '.pagination.total_count')"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Failed to list users"
        echo "$response" | jq -r '.message' 2>/dev/null || echo "$response"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Test 3: Create New User
test_create_user() {
    write_info "\n========== Test 3: Create New User =========="
    
    local timestamp=$(date +%s)
    local data="{
        \"first_name\":\"Test\",
        \"last_name\":\"User\",
        \"username\":\"testuser_$timestamp\",
        \"email\":\"testuser_$timestamp@example.com\",
        \"phone\":\"+1234567890\",
        \"password\":\"TestPassword123\",
        \"is_active\":true
    }"
    
    local response=$(api_call "POST" "/admin/users" "$data")
    
    if echo "$response" | jq -e '.user.id' > /dev/null 2>&1; then
        TEST_USER_ID=$(echo "$response" | jq -r '.user.id')
        write_success "✓ User created successfully"
        echo "  User ID: $TEST_USER_ID"
        echo "  Username: $(echo "$response" | jq -r '.user.username')"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Failed to create user"
        echo "$response" | jq -r '.message' 2>/dev/null || echo "$response"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Test 4: Get Single User
test_get_user() {
    write_info "\n========== Test 4: Get Single User =========="
    
    if [ -z "$TEST_USER_ID" ]; then
        write_warning "! Skipping: No test user ID available"
        return 1
    fi
    
    local response=$(api_call "GET" "/admin/users/$TEST_USER_ID")
    
    if echo "$response" | jq -e '.id' > /dev/null 2>&1; then
        write_success "✓ User retrieved successfully"
        echo "  Username: $(echo "$response" | jq -r '.username')"
        echo "  Email: $(echo "$response" | jq -r '.email')"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Failed to get user"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Test 5: Update User
test_update_user() {
    write_info "\n========== Test 5: Update User =========="
    
    if [ -z "$TEST_USER_ID" ]; then
        write_warning "! Skipping: No test user ID available"
        return 1
    fi
    
    local data="{\"first_name\":\"Updated\",\"last_name\":\"User\",\"gender\":\"male\"}"
    local response=$(api_call "PUT" "/admin/users/$TEST_USER_ID" "$data")
    
    if echo "$response" | jq -e '.message' > /dev/null 2>&1; then
        write_success "✓ User updated successfully"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Failed to update user"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Test 6: Block User
test_block_user() {
    write_info "\n========== Test 6: Block User =========="
    
    if [ -z "$TEST_USER_ID" ]; then
        write_warning "! Skipping: No test user ID available"
        return 1
    fi
    
    local data="{\"reason\":\"Test blocking - automated test script\"}"
    local response=$(api_call "POST" "/admin/users/$TEST_USER_ID/block" "$data")
    
    if echo "$response" | jq -e '.message' > /dev/null 2>&1; then
        write_success "✓ User blocked successfully"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Failed to block user"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Test 7: Unblock User
test_unblock_user() {
    write_info "\n========== Test 7: Unblock User =========="
    
    if [ -z "$TEST_USER_ID" ]; then
        write_warning "! Skipping: No test user ID available"
        return 1
    fi
    
    local response=$(api_call "POST" "/admin/users/$TEST_USER_ID/unblock")
    
    if echo "$response" | jq -e '.message' > /dev/null 2>&1; then
        write_success "✓ User unblocked successfully"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Failed to unblock user"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Test 8: Deactivate User
test_deactivate_user() {
    write_info "\n========== Test 8: Deactivate User =========="
    
    if [ -z "$TEST_USER_ID" ]; then
        write_warning "! Skipping: No test user ID available"
        return 1
    fi
    
    local response=$(api_call "POST" "/admin/users/$TEST_USER_ID/deactivate")
    
    if echo "$response" | jq -e '.message' > /dev/null 2>&1; then
        write_success "✓ User deactivated successfully"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Failed to deactivate user"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Test 9: Activate User
test_activate_user() {
    write_info "\n========== Test 9: Activate User =========="
    
    if [ -z "$TEST_USER_ID" ]; then
        write_warning "! Skipping: No test user ID available"
        return 1
    fi
    
    local response=$(api_call "POST" "/admin/users/$TEST_USER_ID/activate")
    
    if echo "$response" | jq -e '.message' > /dev/null 2>&1; then
        write_success "✓ User activated successfully"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Failed to activate user"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Test 10: Admin Change Password
test_admin_change_password() {
    write_info "\n========== Test 10: Admin Change Password =========="
    
    if [ -z "$TEST_USER_ID" ]; then
        write_warning "! Skipping: No test user ID available"
        return 1
    fi
    
    local data="{\"new_password\":\"NewTestPassword123\"}"
    local response=$(api_call "POST" "/admin/users/$TEST_USER_ID/change-password" "$data")
    
    if echo "$response" | jq -e '.message' > /dev/null 2>&1; then
        write_success "✓ Password changed successfully"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Failed to change password"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Test 11: List All Roles
test_list_roles() {
    write_info "\n========== Test 11: List All Roles =========="
    
    local response=$(api_call "GET" "/admin/roles?page=1&page_size=20")
    
    if echo "$response" | jq -e '.roles' > /dev/null 2>&1; then
        write_success "✓ Roles listed successfully"
        echo "  Total roles: $(echo "$response" | jq -r '.pagination.total_count')"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Failed to list roles"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Test 12: Create New Role
test_create_role() {
    write_info "\n========== Test 12: Create New Role =========="
    
    local timestamp=$(date +%s)
    local data="{
        \"name\":\"Test Role $timestamp\",
        \"description\":\"Test role created by automated script\",
        \"permissions\":{
            \"users\":[\"read\",\"update\"],
            \"reports\":[\"read\",\"create\"]
        }
    }"
    
    local response=$(api_call "POST" "/admin/roles" "$data")
    
    if echo "$response" | jq -e '.role_id' > /dev/null 2>&1; then
        TEST_ROLE_ID=$(echo "$response" | jq -r '.role_id')
        write_success "✓ Role created successfully"
        echo "  Role ID: $TEST_ROLE_ID"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Failed to create role"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Test 13: Get Single Role
test_get_role() {
    write_info "\n========== Test 13: Get Single Role =========="
    
    if [ -z "$TEST_ROLE_ID" ]; then
        write_warning "! Skipping: No test role ID available"
        return 1
    fi
    
    local response=$(api_call "GET" "/admin/roles/$TEST_ROLE_ID")
    
    if echo "$response" | jq -e '.id' > /dev/null 2>&1; then
        write_success "✓ Role retrieved successfully"
        echo "  Role Name: $(echo "$response" | jq -r '.name')"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Failed to get role"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Test 14: Assign Role to User
test_assign_role() {
    write_info "\n========== Test 14: Assign Role to User =========="
    
    if [ -z "$TEST_USER_ID" ] || [ -z "$TEST_ROLE_ID" ]; then
        write_warning "! Skipping: No test user or role ID available"
        return 1
    fi
    
    local data="{\"role_id\":\"$TEST_ROLE_ID\"}"
    local response=$(api_call "POST" "/admin/users/$TEST_USER_ID/roles" "$data")
    
    if echo "$response" | jq -e '.message' > /dev/null 2>&1; then
        write_success "✓ Role assigned to user successfully"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Failed to assign role"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Test 15: Get Permission Templates
test_get_permission_templates() {
    write_info "\n========== Test 15: Get Permission Templates =========="
    
    local response=$(api_call "GET" "/admin/permission-templates")
    
    if echo "$response" | jq -e '.templates' > /dev/null 2>&1; then
        write_success "✓ Permission templates retrieved successfully"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Failed to get permission templates"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Test 16: Delete Role
test_delete_role() {
    write_info "\n========== Test 16: Delete Role =========="
    
    if [ -z "$TEST_ROLE_ID" ]; then
        write_warning "! Skipping: No test role ID available"
        return 1
    fi
    
    # First remove role from user
    if [ -n "$TEST_USER_ID" ]; then
        api_call "DELETE" "/admin/users/$TEST_USER_ID/roles/$TEST_ROLE_ID" > /dev/null 2>&1
    fi
    
    local response=$(api_call "DELETE" "/admin/roles/$TEST_ROLE_ID")
    
    if echo "$response" | jq -e '.message' > /dev/null 2>&1; then
        write_success "✓ Role deleted successfully"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Failed to delete role"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Test 17: Delete User
test_delete_user() {
    write_info "\n========== Test 17: Delete User =========="
    
    if [ -z "$TEST_USER_ID" ]; then
        write_warning "! Skipping: No test user ID available"
        return 1
    fi
    
    local response=$(api_call "DELETE" "/admin/users/$TEST_USER_ID")
    
    if echo "$response" | jq -e '.message' > /dev/null 2>&1; then
        write_success "✓ User deleted successfully"
        ((PASSED_TESTS++))
        return 0
    else
        write_error "✗ Failed to delete user"
        ((FAILED_TESTS++))
        return 1
    fi
}

# Main execution
echo ""
echo "======================================================================"
echo -e "${YELLOW}  Super Admin API Test Suite${NC}"
echo "======================================================================"

# Check for jq
if ! command -v jq &> /dev/null; then
    write_error "Error: jq is not installed. Please install jq to run this script."
    echo "  Ubuntu/Debian: sudo apt-get install jq"
    echo "  MacOS: brew install jq"
    echo "  CentOS/RHEL: sudo yum install jq"
    exit 1
fi

# Run tests
test_super_admin_login

if [ -n "$ACCESS_TOKEN" ]; then
    test_list_users
    test_create_user
    test_get_user
    test_update_user
    test_block_user
    test_unblock_user
    test_deactivate_user
    test_activate_user
    test_admin_change_password
    test_list_roles
    test_create_role
    test_get_role
    test_assign_role
    test_get_permission_templates
    test_delete_role
    test_delete_user
else
    write_error "\nCannot proceed without authentication. Please check credentials.\n"
fi

# Summary
echo ""
echo "======================================================================"
echo -e "${YELLOW}  Test Summary${NC}"
echo "======================================================================"
echo ""
echo "Total Tests: $((PASSED_TESTS + FAILED_TESTS))"
write_success "Passed: $PASSED_TESTS"
if [ $FAILED_TESTS -gt 0 ]; then
    write_error "Failed: $FAILED_TESTS"
fi

echo ""
echo "======================================================================"
echo ""

# Exit with appropriate code
if [ $FAILED_TESTS -eq 0 ] && [ $PASSED_TESTS -gt 0 ]; then
    write_success "All tests passed successfully!\n"
    exit 0
else
    write_error "Some tests failed. Please review the output above.\n"
    exit 1
fi

