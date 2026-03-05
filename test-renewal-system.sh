#!/bin/bash

echo "🔄 TESTING FOLLOW-UP RENEWAL SYSTEM"
echo "==================================="

# Test data
CLINIC_PATIENT_ID="d27a8fa7-b8bc-43e3-837b-87db5dfd4bed"
CLINIC_ID="f7658c53-72ae-4bd3-9960-741225ebc0a2"
DOCTOR_ID="ef378478-1091-472e-af40-1655e77985b3"
DEPARTMENT_ID="ad958b90-d383-4478-bfe3-08b53b8eeef7"

echo "📋 Test Data:"
echo "  Patient ID: $CLINIC_PATIENT_ID"
echo "  Clinic ID: $CLINIC_ID"
echo "  Doctor ID: $DOCTOR_ID"
echo "  Department ID: $DEPARTMENT_ID"
echo ""

echo "🔍 Step 1: Check current follow-up status"
echo "----------------------------------------"
docker exec -it drandme-backend-postgres-1 psql -U postgres -d drandme -c "
SELECT 
    'Current Follow-ups' as type,
    id, status, is_free, valid_from, valid_until, created_at
FROM follow_ups 
WHERE clinic_patient_id = '$CLINIC_PATIENT_ID' 
  AND doctor_id = '$DOCTOR_ID' 
  AND department_id = '$DEPARTMENT_ID'
ORDER BY created_at DESC
LIMIT 3;
"

echo ""
echo "🔍 Step 2: Check recent appointments"
echo "------------------------------------"
docker exec -it drandme-backend-postgres-1 psql -U postgres -d drandme -c "
SELECT 
    'Recent Appointments' as type,
    id, consultation_type, status, appointment_date, created_at
FROM appointments 
WHERE clinic_patient_id = '$CLINIC_PATIENT_ID' 
  AND doctor_id = '$DOCTOR_ID' 
  AND department_id = '$DEPARTMENT_ID'
ORDER BY created_at DESC
LIMIT 3;
"

echo ""
echo "🚀 Step 3: Create a new regular appointment to test renewal"
echo "----------------------------------------------------------"
echo "This should:"
echo "  1. Mark existing active follow-up as 'renewed'"
echo "  2. Create a new active follow-up with 5 days validity"
echo ""

# Get tomorrow's date for the appointment
TOMORROW=$(date -d "+1 day" +%Y-%m-%d)
echo "📅 Appointment Date: $TOMORROW"

echo ""
echo "📡 Making API call to create regular appointment..."
echo "Note: This will fail due to authentication, but should show debug logs"
echo ""

# Create the request body for a regular appointment
REQUEST_BODY=$(cat <<EOF
{
  "clinic_patient_id": "$CLINIC_PATIENT_ID",
  "doctor_id": "$DOCTOR_ID",
  "clinic_id": "$CLINIC_ID",
  "department_id": "$DEPARTMENT_ID",
  "individual_slot_id": "0d1ed772-114d-41d6-b780-96ab0cd2d6d2",
  "appointment_date": "$TOMORROW",
  "appointment_time": "$TOMORROW 10:00:00",
  "consultation_type": "clinic_visit",
  "payment_method": "pay_now",
  "payment_type": "cash",
  "reason": "Renewal test",
  "notes": "Testing follow-up renewal"
}
EOF
)

echo "Request Body:"
echo "$REQUEST_BODY"
echo ""

# Make the API call (will fail due to auth, but should trigger debug logs)
curl -X POST "http://localhost:8082/api/v1/appointments/simple" \
  -H "Content-Type: application/json" \
  -d "$REQUEST_BODY" \
  -w "\nHTTP Status: %{http_code}\n" || echo "❌ API call failed (expected due to auth)"

echo ""
echo "🔍 Step 4: Check appointment service logs for debug messages"
echo "-----------------------------------------------------------"
echo "Look for these debug messages in the logs:"
echo "  🔄 Creating follow-up for regular appointment: ..."
echo "  🔄 CreateFollowUp called: ..."
echo "  📅 Follow-up validity: From=..., Until=..."
echo "  🔄 Renewed X existing follow-up(s) for Patient=..."
echo "  ✅ Created follow-up eligibility: ..."
echo ""

echo "🎯 Expected Results After Successful API Call:"
echo "============================================="
echo "1. ✅ Existing active follow-up should be marked as 'renewed'"
echo "2. ✅ New active follow-up should be created"
echo "3. ✅ New follow-up should have 5 days validity from appointment date"
echo "4. ✅ Patient should be able to create free follow-ups again"
echo ""

echo "🔧 Manual Steps if API fails:"
echo "============================="
echo "1. Go to your frontend"
echo "2. Create a new regular appointment with:"
echo "   - Patient: ashiq m"
echo "   - Doctor: ef378478-1091-472e-af40-1655e77985b3"
echo "   - Department: ad958b90-d383-4478-bfe3-08b53b8eeef7"
echo "   - Consultation type: clinic_visit"
echo "   - Payment method: pay_now"
echo "3. This should trigger the renewal system"
echo "4. Check logs for debug messages"
echo ""

echo "📱 Then test follow-up creation - should work with new active follow-up!"
