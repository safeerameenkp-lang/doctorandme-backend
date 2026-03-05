#!/bin/bash

# Comprehensive Follow-Up Reset Test
# This script will test the complete follow-up reset functionality

echo "🧪 COMPREHENSIVE FOLLOW-UP RESET TEST"
echo "====================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to test API endpoint
test_api() {
    local url=$1
    local description=$2
    
    print_status $BLUE "🔍 Testing: $description"
    print_status $BLUE "URL: $url"
    
    response=$(curl -s -w "\n%{http_code}" "$url")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "200" ]; then
        print_status $GREEN "✅ HTTP 200 OK"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        print_status $RED "❌ HTTP $http_code"
        echo "$body"
    fi
    echo ""
}

# Get test parameters
echo "Please provide the following information:"
echo "1. API Base URL (e.g., http://localhost:8080):"
read API_BASE

echo "2. Clinic ID:"
read CLINIC_ID

echo "3. Doctor ID:"
read DOCTOR_ID

echo "4. Department ID (optional, press Enter to skip):"
read DEPARTMENT_ID

echo "5. Patient ID (clinic_patient_id):"
read PATIENT_ID

echo ""
print_status $YELLOW "🚀 Starting comprehensive test..."

# Test 1: Check patient data without filters
print_status $YELLOW "📋 TEST 1: Patient Data (No Filters)"
test_api "$API_BASE/clinic-specific-patients?clinic_id=$CLINIC_ID&search=$(echo $PATIENT_ID | cut -c1-8)" "Get patient data without doctor/department filters"

# Test 2: Check patient data with doctor filter
print_status $YELLOW "📋 TEST 2: Patient Data (With Doctor Filter)"
if [ -n "$DEPARTMENT_ID" ]; then
    test_api "$API_BASE/clinic-specific-patients?clinic_id=$CLINIC_ID&doctor_id=$DOCTOR_ID&department_id=$DEPARTMENT_ID&search=$(echo $PATIENT_ID | cut -c1-8)" "Get patient data with doctor+department filters"
else
    test_api "$API_BASE/clinic-specific-patients?clinic_id=$CLINIC_ID&doctor_id=$DOCTOR_ID&search=$(echo $PATIENT_ID | cut -c1-8)" "Get patient data with doctor filter"
fi

# Test 3: Check appointment list
print_status $YELLOW "📋 TEST 3: Appointment List"
test_api "$API_BASE/appointments/simple-list?clinic_id=$CLINIC_ID&patient_id=$PATIENT_ID" "Get appointment list for patient"

# Test 4: Direct database check
print_status $YELLOW "📋 TEST 4: Direct Database Check"
echo "Checking database directly..."

# Get latest regular appointment date
if [ -n "$DEPARTMENT_ID" ]; then
    LATEST_DATE=$(psql -d drandme_db -t -c "
    SELECT a.appointment_date
    FROM appointments a
    WHERE a.clinic_patient_id = '$PATIENT_ID'
      AND a.clinic_id = '$CLINIC_ID'
      AND a.doctor_id = '$DOCTOR_ID'
      AND a.department_id = '$DEPARTMENT_ID'
      AND a.consultation_type IN ('clinic_visit', 'video_consultation')
      AND a.status IN ('completed', 'confirmed')
    ORDER BY a.appointment_date DESC, a.appointment_time DESC
    LIMIT 1;
    " | xargs)
else
    LATEST_DATE=$(psql -d drandme_db -t -c "
    SELECT a.appointment_date
    FROM appointments a
    WHERE a.clinic_patient_id = '$PATIENT_ID'
      AND a.clinic_id = '$CLINIC_ID'
      AND a.doctor_id = '$DOCTOR_ID'
      AND a.consultation_type IN ('clinic_visit', 'video_consultation')
      AND a.status IN ('completed', 'confirmed')
    ORDER BY a.appointment_date DESC, a.appointment_time DESC
    LIMIT 1;
    " | xargs)
fi

print_status $BLUE "Latest regular appointment date: $LATEST_DATE"

# Count free follow-ups from latest date
if [ -n "$DEPARTMENT_ID" ]; then
    FREE_COUNT=$(psql -d drandme_db -t -c "
    SELECT COUNT(*)
    FROM appointments
    WHERE clinic_patient_id = '$PATIENT_ID'
      AND clinic_id = '$CLINIC_ID'
      AND doctor_id = '$DOCTOR_ID'
      AND department_id = '$DEPARTMENT_ID'
      AND consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
      AND payment_status = 'waived'
      AND appointment_date >= '$LATEST_DATE'
      AND status NOT IN ('cancelled', 'no_show');
    " | xargs)
else
    FREE_COUNT=$(psql -d drandme_db -t -c "
    SELECT COUNT(*)
    FROM appointments
    WHERE clinic_patient_id = '$PATIENT_ID'
      AND clinic_id = '$CLINIC_ID'
      AND doctor_id = '$DOCTOR_ID'
      AND consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
      AND payment_status = 'waived'
      AND appointment_date >= '$LATEST_DATE'
      AND status NOT IN ('cancelled', 'no_show');
    " | xargs)
fi

print_status $BLUE "Free follow-up count from latest date: $FREE_COUNT"

# Analysis
echo ""
print_status $YELLOW "📊 ANALYSIS:"
if [ "$FREE_COUNT" = "0" ]; then
    print_status $GREEN "✅ Should show FREE (GREEN) - No free follow-ups used since latest regular appointment"
else
    print_status $RED "❌ Should show PAID (ORANGE) - $FREE_COUNT free follow-up(s) already used since latest regular appointment"
fi

echo ""
print_status $YELLOW "🔍 EXPECTED API RESPONSE:"
if [ "$FREE_COUNT" = "0" ]; then
    print_status $GREEN "✅ eligible_follow_ups array should contain 1 entry"
    print_status $GREEN "✅ Card should show GREEN avatar"
    print_status $GREEN "✅ Status should be 'free'"
else
    print_status $RED "❌ eligible_follow_ups array should be empty"
    print_status $RED "❌ Card should show ORANGE avatar"
    print_status $RED "❌ Status should be 'paid_expired'"
fi

echo ""
print_status $YELLOW "🎯 NEXT STEPS:"
echo "1. Check the API responses above"
echo "2. Compare with your frontend console output"
echo "3. If API shows FREE but frontend shows PAID, it's a frontend issue"
echo "4. If API shows PAID but you expect FREE, it's a backend logic issue"

echo ""
print_status $GREEN "✅ Test complete!"


