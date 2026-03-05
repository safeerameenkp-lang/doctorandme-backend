#!/bin/bash

# Test script to fix the follow-up renewal issue
# This creates a new regular appointment to trigger follow-up renewal

echo "🔧 FOLLOW-UP RENEWAL FIX TEST"
echo "=============================="

# Test data (replace with actual values from your system)
CLINIC_PATIENT_ID="d27a8fa7-b8bc-43e3-837b-87db5dfd4bed"  # ashiq m
CLINIC_ID="your-clinic-id"
DOCTOR_ID="ef378478-1091-472e-af40-1655e77985b3"  # Same doctor as expired follow-up
DEPARTMENT_ID="ad958b90-d383-4478-bfe3-08b53b8eeef7"  # Same department as expired follow-up

echo "📋 Test Data:"
echo "  Patient ID: $CLINIC_PATIENT_ID"
echo "  Doctor ID: $DOCTOR_ID" 
echo "  Department ID: $DEPARTMENT_ID"
echo ""

echo "🔍 Step 1: Check current follow-up status"
echo "----------------------------------------"
docker exec -it drandme-backend-postgres-1 psql -U postgres -d drandme -c "
SELECT 
    status,
    COUNT(*) as count,
    MAX(valid_until) as latest_expiry
FROM follow_ups 
WHERE clinic_patient_id = '$CLINIC_PATIENT_ID' 
  AND doctor_id = '$DOCTOR_ID' 
  AND department_id = '$DEPARTMENT_ID'
GROUP BY status;
"

echo ""
echo "🔍 Step 2: Check if there are any active follow-ups"
echo "---------------------------------------------------"
docker exec -it drandme-backend-postgres-1 psql -U postgres -d drandme -c "
SELECT 
    id,
    status,
    is_free,
    valid_from,
    valid_until,
    created_at
FROM follow_ups 
WHERE clinic_patient_id = '$CLINIC_PATIENT_ID' 
  AND doctor_id = '$DOCTOR_ID' 
  AND department_id = '$DEPARTMENT_ID'
  AND status = 'active'
ORDER BY created_at DESC;
"

echo ""
echo "📝 Step 3: Create a new regular appointment to trigger renewal"
echo "--------------------------------------------------------------"
echo "This should:"
echo "  1. Mark existing expired follow-ups as 'renewed'"
echo "  2. Create a new 'active' follow-up for 5 days"
echo ""

# Get tomorrow's date for the appointment
TOMORROW=$(date -d "+1 day" +%Y-%m-%d)
echo "📅 Appointment Date: $TOMORROW"

echo ""
echo "🚀 Step 4: Make API call to create regular appointment"
echo "-----------------------------------------------------"

# Create the appointment via API
curl -X POST "http://localhost:8080/appointments/simple" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d "{
    \"clinic_patient_id\": \"$CLINIC_PATIENT_ID\",
    \"clinic_id\": \"$CLINIC_ID\",
    \"doctor_id\": \"$DOCTOR_ID\",
    \"department_id\": \"$DEPARTMENT_ID\",
    \"individual_slot_id\": \"YOUR_SLOT_ID_HERE\",
    \"appointment_date\": \"$TOMORROW\",
    \"appointment_time\": \"10:00\",
    \"consultation_type\": \"clinic_visit\",
    \"payment_method\": \"pay_now\",
    \"payment_type\": \"cash\",
    \"reason\": \"Follow-up renewal test\"
  }" || echo "❌ API call failed - please check your token and slot ID"

echo ""
echo "🔍 Step 5: Verify follow-up renewal worked"
echo "------------------------------------------"
docker exec -it drandme-backend-postgres-1 psql -U postgres -d drandme -c "
SELECT 
    id,
    status,
    is_free,
    valid_from,
    valid_until,
    renewed_at,
    created_at
FROM follow_ups 
WHERE clinic_patient_id = '$CLINIC_PATIENT_ID' 
  AND doctor_id = '$DOCTOR_ID' 
  AND department_id = '$DEPARTMENT_ID'
ORDER BY created_at DESC
LIMIT 3;
"

echo ""
echo "✅ Step 6: Test follow-up eligibility API"
echo "----------------------------------------"
curl -X GET "http://localhost:8080/appointments/followup-eligibility?clinic_patient_id=$CLINIC_PATIENT_ID&clinic_id=$CLINIC_ID&doctor_id=$DOCTOR_ID&department_id=$DEPARTMENT_ID" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" || echo "❌ API call failed"

echo ""
echo "🎯 Expected Results:"
echo "==================="
echo "✅ Should see 1 'renewed' follow-up (old expired one)"
echo "✅ Should see 1 'active' follow-up (newly created)"
echo "✅ Frontend should now show eligibleFollowUps.length: 1"
echo "✅ Follow-up should be FREE for 5 days from appointment date"
echo ""
echo "🔧 Manual Steps if API fails:"
echo "============================="
echo "1. Go to your frontend"
echo "2. Create a new regular appointment with:"
echo "   - Same patient: ashiq m"
echo "   - Same doctor: ef378478-1091-472e-af40-1655e77985b3"
echo "   - Same department: ad958b90-d383-4478-bfe3-08b53b8eeef7"
echo "   - Consultation type: clinic_visit"
echo "3. This should automatically renew the follow-up"
echo ""
echo "📱 Then test follow-up creation - should now work!"
