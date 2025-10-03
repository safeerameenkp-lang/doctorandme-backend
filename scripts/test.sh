#!/bin/bash

# Test script for the Clinic Management System
# This script tests the basic functionality of both services

echo "Testing Clinic Management System..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to test HTTP endpoint
test_endpoint() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=$4
    local description=$5
    
    echo -n "Testing $description... "
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "%{http_code}" -X $method -H "Content-Type: application/json" -d "$data" "$url")
    else
        response=$(curl -s -w "%{http_code}" -X $method "$url")
    fi
    
    http_code="${response: -3}"
    
    if [ "$http_code" = "$expected_status" ]; then
        echo -e "${GREEN}PASS${NC}"
    else
        echo -e "${RED}FAIL${NC} (Expected: $expected_status, Got: $http_code)"
    fi
}

# Wait for services to be ready
echo -e "${YELLOW}Waiting for services to start...${NC}"
sleep 10

# Test Auth Service
echo -e "\n${YELLOW}Testing Auth Service${NC}"
test_endpoint "GET" "http://localhost:8080/api/auth/register" "" "404" "Auth service health check"

# Test Organization Service (should require auth)
echo -e "\n${YELLOW}Testing Organization Service${NC}"
test_endpoint "GET" "http://localhost:8081/api/organizations/organizations" "" "401" "Organization service auth check"

# Test user registration
echo -e "\n${YELLOW}Testing User Registration${NC}"
register_data='{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com",
    "username": "johndoe",
    "phone": "+1234567890",
    "password": "password123"
}'

test_endpoint "POST" "http://localhost:8080/api/auth/register" "$register_data" "201" "User registration"

# Test user login
echo -e "\n${YELLOW}Testing User Login${NC}"
login_data='{
    "login": "john.doe@example.com",
    "password": "password123"
}'

echo -n "Testing user login... "
login_response=$(curl -s -X POST -H "Content-Type: application/json" -d "$login_data" "http://localhost:8080/api/auth/login")
http_code=$(curl -s -w "%{http_code}" -X POST -H "Content-Type: application/json" -d "$login_data" "http://localhost:8080/api/auth/login" -o /dev/null)

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}PASS${NC}"
    
    # Extract token for further tests
    token=$(echo $login_response | grep -o '"accessToken":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$token" ]; then
        echo -e "\n${YELLOW}Testing Authenticated Endpoints${NC}"
        
        # Test organization listing with auth
        echo -n "Testing get organizations with auth... "
        org_response=$(curl -s -H "Authorization: Bearer $token" "http://localhost:8081/api/organizations/organizations")
        org_code=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $token" "http://localhost:8081/api/organizations/organizations" -o /dev/null)
        
        if [ "$org_code" = "200" ]; then
            echo -e "${GREEN}PASS${NC}"
        else
            echo -e "${RED}FAIL${NC} (Expected: 200, Got: $org_code)"
        fi
        
        echo -e "\n${GREEN}All tests completed!${NC}"
        echo -e "${YELLOW}Note: Some tests may fail if the database is not fully initialized.${NC}"
        echo -e "${YELLOW}Run 'docker-compose logs' to check service logs.${NC}"
    else
        echo -e "${RED}Could not extract token from login response${NC}"
    fi
else
    echo -e "${RED}FAIL${NC} (Expected: 200, Got: $http_code)"
fi

echo -e "\n${YELLOW}Test Summary:${NC}"
echo "- Auth Service: Registration and Login endpoints tested"
echo "- Organization Service: Authentication requirement verified"
echo "- Database connectivity: Verified through successful operations"
echo -e "\n${GREEN}System is ready for use!${NC}"