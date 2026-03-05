#!/bin/bash

echo "🧪 TESTING FOLLOW-UP ELIGIBILITY DEBUG"
echo "======================================"

# Test data
CLINIC_PATIENT_ID="d27a8fa7-b8bc-43e3-837b-87db5dfd4bed"
CLINIC_ID="f7658c53-72ae-4bd3-9960-741225ebc0a2"
DOCTOR_ID="ef378478-1091-472e-af40-1655e77985b3"
DEPARTMENT_ID="ad958b90-d383-4478-bfe3-08b53b8eeef7"
SLOT_ID="0d1ed772-114d-41d6-b780-96ab0cd2d6d2"

echo "📋 Test Data:"
echo "  Patient ID: $CLINIC_PATIENT_ID"
echo "  Clinic ID: $CLINIC_ID"
echo "  Doctor ID: $DOCTOR_ID"
echo "  Department ID: $DEPARTMENT_ID"
echo "  Slot ID: $SLOT_ID"
echo ""

echo "🔍 Step 1: Check database directly"
echo "----------------------------------"
docker exec -it drandme-backend-postgres-1 psql -U postgres -d drandme -c "
SELECT 
    'Active Follow-up' as type,
    id, status, is_free, valid_from, valid_until
FROM follow_ups 
WHERE clinic_patient_id = '$CLINIC_PATIENT_ID' 
  AND clinic_id = '$CLINIC_ID'
  AND doctor_id = '$DOCTOR_ID' 
  AND department_id = '$DEPARTMENT_ID'
  AND status = 'active'
  AND valid_until >= CURRENT_DATE;
"

echo ""
echo "🔍 Step 2: Check previous appointments"
echo "--------------------------------------"
docker exec -it drandme-backend-postgres-1 psql -U postgres -d drandme -c "
SELECT 
    'Previous Appointments' as type,
    id, consultation_type, status, appointment_date, created_at
FROM appointments 
WHERE clinic_patient_id = '$CLINIC_PATIENT_ID' 
  AND clinic_id = '$CLINIC_ID'
  AND doctor_id = '$DOCTOR_ID' 
  AND department_id = '$DEPARTMENT_ID'
  AND consultation_type IN ('clinic_visit', 'video_consultation')
  AND status IN ('completed', 'confirmed')
ORDER BY created_at DESC
LIMIT 3;
"

echo ""
echo "🚀 Step 3: Simulate follow-up creation request"
echo "----------------------------------------------"
echo "This will trigger the debug logs in the appointment service..."
echo ""

# Create the request body
REQUEST_BODY=$(cat <<EOF
{
  "clinic_patient_id": "$CLINIC_PATIENT_ID",
  "doctor_id": "$DOCTOR_ID",
  "clinic_id": "$CLINIC_ID",
  "department_id": "$DEPARTMENT_ID",
  "individual_slot_id": "$SLOT_ID",
  "appointment_date": "2025-10-27",
  "appointment_time": "2025-10-27 14:37:00",
  "consultation_type": "follow-up-via-clinic",
  "reason": "Test follow-up",
  "notes": "Debug test"
}
EOF
)

echo "Request Body:"
echo "$REQUEST_BODY"
echo ""

echo "📡 Making API call..."
echo "Note: This will fail due to authentication, but should show debug logs"
echo ""

# Make the API call (will fail due to auth, but should trigger debug logs)
curl -X POST "http://localhost:8082/api/v1/appointments/simple" \
  -H "Content-Type: application/json" \
  -d "$REQUEST_BODY" \
  -w "\nHTTP Status: %{http_code}\n" || echo "❌ API call failed (expected due to auth)"

echo ""
echo "🔍 Step 4: Check appointment service logs"
echo "----------------------------------------"
echo "Look for debug messages starting with 🔍 in the logs above"
echo ""

echo "✅ Expected Debug Messages:"
echo "  🔍 CheckFollowUpEligibility: Patient=..., Clinic=..., Doctor=..., Dept=..."
echo "  🔍 GetActiveFollowUp: Patient=..., Clinic=..., Doctor=..., Dept=..."
echo "  🔍 Executing query: ... with args: ..."
echo "  ✅ Found active free follow-up: X days remaining"
echo ""

echo "❌ If you see '⚠️ No active follow-up found', there's a query issue"
echo "❌ If you see '❌ GetActiveFollowUp error:', there's a database issue"
echo ""

echo "🎯 Next Steps:"
echo "1. Check the logs above for debug messages"
echo "2. If debug shows active follow-up found, the issue is elsewhere"
echo "3. If debug shows no follow-up found, there's a database query issue"
echo "4. Test with proper authentication token if needed"
