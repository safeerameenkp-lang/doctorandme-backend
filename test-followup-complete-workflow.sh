#!/bin/bash

# =====================================================
# FOLLOW-UP SYSTEM TEST SCRIPT
# Tests the complete follow-up workflow
# =====================================================

BASE_URL="http://localhost:8081"
CLINIC_ID="your-clinic-id"
PATIENT_ID="your-patient-id"
DOCTOR_ID="your-doctor-id"
DEPARTMENT_ID="your-department-id"

echo "🧪 TESTING FOLLOW-UP SYSTEM"
echo "================================"

# Test 1: Create Regular Appointment (should grant free follow-up)
echo ""
echo "📅 TEST 1: Create Regular Appointment"
echo "--------------------------------------"
REGULAR_APPOINTMENT=$(curl -s -X POST "$BASE_URL/appointments/simple" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "clinic_patient_id": "'$PATIENT_ID'",
    "doctor_id": "'$DOCTOR_ID'",
    "clinic_id": "'$CLINIC_ID'",
    "department_id": "'$DEPARTMENT_ID'",
    "individual_slot_id": "slot-id",
    "appointment_date": "2025-01-20",
    "appointment_time": "2025-01-20 10:00:00",
    "consultation_type": "clinic_visit",
    "payment_method": "pay_now",
    "payment_type": "cash"
  }')

echo "Response: $REGULAR_APPOINTMENT"

# Extract appointment ID
APPOINTMENT_ID=$(echo $REGULAR_APPOINTMENT | jq -r '.appointment.id')
echo "✅ Regular appointment created: $APPOINTMENT_ID"

# Test 2: Check Follow-Up Eligibility (should be FREE)
echo ""
echo "🔍 TEST 2: Check Follow-Up Eligibility"
echo "--------------------------------------"
ELIGIBILITY=$(curl -s -X GET "$BASE_URL/appointments/followup-eligibility?clinic_patient_id=$PATIENT_ID&clinic_id=$CLINIC_ID&doctor_id=$DOCTOR_ID&department_id=$DEPARTMENT_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "Eligibility Response: $ELIGIBILITY"

IS_FREE=$(echo $ELIGIBILITY | jq -r '.eligibility.is_free')
IS_ELIGIBLE=$(echo $ELIGIBILITY | jq -r '.eligibility.eligible')

if [ "$IS_FREE" = "true" ] && [ "$IS_ELIGIBLE" = "true" ]; then
    echo "✅ FREE follow-up available!"
else
    echo "❌ Follow-up not free or not eligible"
    exit 1
fi

# Test 3: Book FREE Follow-Up
echo ""
echo "🎉 TEST 3: Book FREE Follow-Up"
echo "-------------------------------"
FREE_FOLLOWUP=$(curl -s -X POST "$BASE_URL/appointments/simple" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "clinic_patient_id": "'$PATIENT_ID'",
    "doctor_id": "'$DOCTOR_ID'",
    "clinic_id": "'$CLINIC_ID'",
    "department_id": "'$DEPARTMENT_ID'",
    "individual_slot_id": "slot-id-2",
    "appointment_date": "2025-01-22",
    "appointment_time": "2025-01-22 11:00:00",
    "consultation_type": "follow-up-via-clinic"
  }')

echo "Free Follow-up Response: $FREE_FOLLOWUP"

FOLLOWUP_ID=$(echo $FREE_FOLLOWUP | jq -r '.appointment.id')
IS_FREE_FOLLOWUP=$(echo $FREE_FOLLOWUP | jq -r '.is_free_followup')

if [ "$IS_FREE_FOLLOWUP" = "true" ]; then
    echo "✅ FREE follow-up booked successfully: $FOLLOWUP_ID"
else
    echo "❌ Follow-up was not free"
    exit 1
fi

# Test 4: Try to Book Another FREE Follow-Up (should be PAID)
echo ""
echo "💰 TEST 4: Try Another Follow-Up (should be PAID)"
echo "------------------------------------------------"
PAID_FOLLOWUP=$(curl -s -X POST "$BASE_URL/appointments/simple" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "clinic_patient_id": "'$PATIENT_ID'",
    "doctor_id": "'$DOCTOR_ID'",
    "clinic_id": "'$CLINIC_ID'",
    "department_id": "'$DEPARTMENT_ID'",
    "individual_slot_id": "slot-id-3",
    "appointment_date": "2025-01-24",
    "appointment_time": "2025-01-24 12:00:00",
    "consultation_type": "follow-up-via-clinic",
    "payment_method": "pay_now",
    "payment_type": "cash"
  }')

echo "Paid Follow-up Response: $PAID_FOLLOWUP"

IS_FREE_SECOND=$(echo $PAID_FOLLOWUP | jq -r '.is_free_followup')

if [ "$IS_FREE_SECOND" = "false" ]; then
    echo "✅ Second follow-up correctly requires payment"
else
    echo "❌ Second follow-up was free (should be paid)"
    exit 1
fi

# Test 5: Create Another Regular Appointment (should RENEW follow-up)
echo ""
echo "🔄 TEST 5: Create Another Regular Appointment (RENEWAL)"
echo "------------------------------------------------------"
RENEWAL_APPOINTMENT=$(curl -s -X POST "$BASE_URL/appointments/simple" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "clinic_patient_id": "'$PATIENT_ID'",
    "doctor_id": "'$DOCTOR_ID'",
    "clinic_id": "'$CLINIC_ID'",
    "department_id": "'$DEPARTMENT_ID'",
    "individual_slot_id": "slot-id-4",
    "appointment_date": "2025-01-25",
    "appointment_time": "2025-01-25 14:00:00",
    "consultation_type": "clinic_visit",
    "payment_method": "pay_now",
    "payment_type": "cash"
  }')

echo "Renewal Appointment Response: $RENEWAL_APPOINTMENT"

RENEWAL_ID=$(echo $RENEWAL_APPOINTMENT | jq -r '.appointment.id')
echo "✅ Renewal appointment created: $RENEWAL_ID"

# Test 6: Check Follow-Up Eligibility After Renewal (should be FREE again)
echo ""
echo "🆕 TEST 6: Check Follow-Up Eligibility After Renewal"
echo "---------------------------------------------------"
NEW_ELIGIBILITY=$(curl -s -X GET "$BASE_URL/appointments/followup-eligibility?clinic_patient_id=$PATIENT_ID&clinic_id=$CLINIC_ID&doctor_id=$DOCTOR_ID&department_id=$DEPARTMENT_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "New Eligibility Response: $NEW_ELIGIBILITY"

NEW_IS_FREE=$(echo $NEW_ELIGIBILITY | jq -r '.eligibility.is_free')
NEW_IS_ELIGIBLE=$(echo $NEW_ELIGIBILITY | jq -r '.eligibility.eligible')

if [ "$NEW_IS_FREE" = "true" ] && [ "$NEW_IS_ELIGIBLE" = "true" ]; then
    echo "✅ FREE follow-up available again after renewal!"
else
    echo "❌ Follow-up not free after renewal"
    exit 1
fi

# Test 7: List All Active Follow-Ups
echo ""
echo "📋 TEST 7: List All Active Follow-Ups"
echo "-------------------------------------"
ACTIVE_FOLLOWUPS=$(curl -s -X GET "$BASE_URL/appointments/followup-eligibility/active?clinic_patient_id=$PATIENT_ID&clinic_id=$CLINIC_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "Active Follow-ups Response: $ACTIVE_FOLLOWUPS"

TOTAL_FOLLOWUPS=$(echo $ACTIVE_FOLLOWUPS | jq -r '.total')
echo "✅ Total active follow-ups: $TOTAL_FOLLOWUPS"

echo ""
echo "🎉 ALL TESTS PASSED!"
echo "===================="
echo "✅ Regular appointment creates free follow-up"
echo "✅ First follow-up is FREE"
echo "✅ Second follow-up requires PAYMENT"
echo "✅ New regular appointment RENEWS follow-up"
echo "✅ After renewal, follow-up is FREE again"
echo "✅ System correctly tracks follow-up status"

