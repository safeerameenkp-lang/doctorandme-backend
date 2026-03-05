#!/bin/bash

# =====================================================
# FRONTEND FOLLOW-UP ISSUE DEBUG SCRIPT
# Tests the specific scenario mentioned by user
# =====================================================

BASE_URL="http://localhost:8080"
APPOINTMENT_URL="http://localhost:8081"

echo "🔍 DEBUGGING FRONTEND FOLLOW-UP ISSUE"
echo "======================================"

# Test the specific patient and doctor mentioned in the logs
PATIENT_ID="your-patient-id"
CLINIC_ID="your-clinic-id"
DOCTOR_ID="ef378478-1091-472e-af40-1655e77985b3"
DEPARTMENT_ID="ad958b90-d383-4478-bfe3-08b53b8eeef7"

echo ""
echo "📋 STEP 1: Check Patient Details with Follow-Up Data"
echo "---------------------------------------------------"
PATIENT_DETAILS=$(curl -s -X GET "$BASE_URL/clinic-specific-patients/$PATIENT_ID?doctor_id=$DOCTOR_ID&department_id=$DEPARTMENT_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "Patient Details Response:"
echo "$PATIENT_DETAILS" | jq '.'

# Check if eligible_follow_ups is populated
ELIGIBLE_COUNT=$(echo $PATIENT_DETAILS | jq '.patient.eligible_follow_ups | length')
echo ""
echo "📊 Eligible Follow-ups Count: $ELIGIBLE_COUNT"

if [ "$ELIGIBLE_COUNT" -eq 0 ]; then
    echo "❌ ISSUE FOUND: eligible_follow_ups is empty!"
    echo "   This is why frontend shows 'eligibleFollowUps.length: 0'"
else
    echo "✅ Eligible follow-ups found: $ELIGIBLE_COUNT"
fi

echo ""
echo "🔍 STEP 2: Check Follow-Up Eligibility API Directly"
echo "----------------------------------------------------"
ELIGIBILITY_API=$(curl -s -X GET "$APPOINTMENT_URL/appointments/followup-eligibility?clinic_patient_id=$PATIENT_ID&clinic_id=$CLINIC_ID&doctor_id=$DOCTOR_ID&department_id=$DEPARTMENT_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "Eligibility API Response:"
echo "$ELIGIBILITY_API" | jq '.'

echo ""
echo "📋 STEP 3: Check Active Follow-Ups API"
echo "------------------------------------"
ACTIVE_FOLLOWUPS=$(curl -s -X GET "$APPOINTMENT_URL/appointments/followup-eligibility/active?clinic_patient_id=$PATIENT_ID&clinic_id=$CLINIC_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "Active Follow-ups API Response:"
echo "$ACTIVE_FOLLOWUPS" | jq '.'

echo ""
echo "🔍 STEP 4: Check Follow-Ups Table Directly"
echo "------------------------------------------"
echo "Checking follow_ups table for this patient..."

# This would need to be run directly on the database
echo "Run this SQL query on your database:"
echo ""
echo "SELECT * FROM follow_ups WHERE clinic_patient_id = '$PATIENT_ID' ORDER BY created_at DESC;"
echo ""
echo "Expected: Should see follow-up records with status 'active' for recent appointments"

echo ""
echo "🔍 STEP 5: Check Recent Appointments"
echo "------------------------------------"
echo "Run this SQL query to check recent appointments:"
echo ""
echo "SELECT a.id, a.appointment_date, a.consultation_type, a.status, a.doctor_id, a.department_id"
echo "FROM appointments a"
echo "WHERE a.clinic_patient_id = '$PATIENT_ID'"
echo "  AND a.clinic_id = '$CLINIC_ID'"
echo "  AND a.doctor_id = '$DOCTOR_ID'"
echo "  AND a.department_id = '$DEPARTMENT_ID'"
echo "ORDER BY a.appointment_date DESC"
echo "LIMIT 5;"

echo ""
echo "🎯 DIAGNOSIS SUMMARY"
echo "==================="
echo ""
echo "Based on the frontend logs, the issue is:"
echo "1. ✅ Patient has 10 appointments"
echo "2. ✅ Latest appointment is on 2025-10-27 (future appointment)"
echo "3. ✅ Frontend calculates eligibility correctly"
echo "4. ❌ Backend returns eligibleFollowUps.length: 0"
echo ""
echo "🔧 LIKELY CAUSES:"
echo "1. follow_ups table not populated for recent appointments"
echo "2. FollowUpHelper.GetActiveFollowUps() not working correctly"
echo "3. Database connection issue in organization service"
echo "4. Migration not run (follow_ups table doesn't exist)"
echo ""
echo "🚀 SOLUTIONS:"
echo "1. Run migration: psql -f migrations/025_create_follow_ups_table.sql"
echo "2. Check if follow_ups table exists and has data"
echo "3. Verify FollowUpHelper is working correctly"
echo "4. Test the appointment creation to ensure follow-ups are created"

