#!/bin/bash

# Test Follow-Up API Response
# This will test the actual API endpoint

echo "🧪 TESTING FOLLOW-UP API RESPONSE"
echo "=================================="

# Get the patient ID and doctor ID from user
echo "Please provide the following information:"
echo "1. Patient ID (clinic_patient_id):"
read PATIENT_ID

echo "2. Doctor ID:"
read DOCTOR_ID

echo "3. Department ID (optional, press Enter to skip):"
read DEPARTMENT_ID

echo "4. Clinic ID:"
read CLINIC_ID

echo "5. API Base URL (e.g., http://localhost:8080):"
read API_BASE

echo ""
echo "🔍 Testing Patient API..."

# Test the patient API with doctor and department filters
if [ -n "$DEPARTMENT_ID" ]; then
    echo "GET $API_BASE/clinic-specific-patients?clinic_id=$CLINIC_ID&doctor_id=$DOCTOR_ID&department_id=$DEPARTMENT_ID"
    curl -s "$API_BASE/clinic-specific-patients?clinic_id=$CLINIC_ID&doctor_id=$DOCTOR_ID&department_id=$DEPARTMENT_ID" | jq '.'
else
    echo "GET $API_BASE/clinic-specific-patients?clinic_id=$CLINIC_ID&doctor_id=$DOCTOR_ID"
    curl -s "$API_BASE/clinic-specific-patients?clinic_id=$CLINIC_ID&doctor_id=$DOCTOR_ID" | jq '.'
fi

echo ""
echo "🎯 Looking for eligible_follow_ups array..."
echo "Expected: Array with entries = FREE (GREEN)"
echo "Expected: Empty array = PAID (ORANGE)"

echo ""
echo "✅ Test complete!"


