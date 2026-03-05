#!/bin/bash

echo "🔧 FIXING RENEWAL ISSUE - DIRECT TEST"
echo "====================================="

# Test data
CLINIC_PATIENT_ID="d27a8fa7-b8bc-43e3-837b-87db5dfd4bed"
CLINIC_ID="f7658c53-72ae-4bd3-9960-741225ebc0a2"
DOCTOR_ID="ef378478-1091-472e-af40-1655e77985b3"
DEPARTMENT_ID="ad958b90-d383-4478-bfe3-08b53b8eeef7"

echo "📋 Current Status:"
echo "  Patient: ashiq m"
echo "  Doctor: ef378478-1091-472e-af40-1655e77985b3"
echo "  Department: ad958b90-d383-4478-bfe3-08b53b8eeef7"
echo ""

echo "🔍 Step 1: Check current follow-ups"
echo "-----------------------------------"
docker exec -it drandme-backend-postgres-1 psql -U postgres -d drandme -c "
SELECT 
    id, status, is_free, valid_from, valid_until, created_at
FROM follow_ups 
WHERE clinic_patient_id = '$CLINIC_PATIENT_ID' 
  AND doctor_id = '$DOCTOR_ID' 
  AND department_id = '$DEPARTMENT_ID'
ORDER BY created_at DESC
LIMIT 2;
"

echo ""
echo "🔧 Step 2: Manually trigger renewal by creating follow-up"
echo "--------------------------------------------------------"
echo "This simulates what should happen when a regular appointment is created"
echo ""

# Get tomorrow's date
TOMORROW=$(date -d "+1 day" +%Y-%m-%d)
echo "📅 Creating follow-up for appointment date: $TOMORROW"

# First, mark existing follow-ups as renewed
docker exec -it drandme-backend-postgres-1 psql -U postgres -d drandme -c "
UPDATE follow_ups
SET status = 'renewed',
    renewed_at = CURRENT_TIMESTAMP,
    renewed_by_appointment_id = (SELECT id FROM appointments WHERE clinic_patient_id = '$CLINIC_PATIENT_ID' ORDER BY created_at DESC LIMIT 1),
    updated_at = CURRENT_TIMESTAMP
WHERE clinic_patient_id = '$CLINIC_PATIENT_ID'
  AND doctor_id = '$DOCTOR_ID'
  AND department_id = '$DEPARTMENT_ID'
  AND status = 'active';
"

# Then create new active follow-up
docker exec -it drandme-backend-postgres-1 psql -U postgres -d drandme -c "
INSERT INTO follow_ups (
    clinic_patient_id, clinic_id, doctor_id, department_id,
    source_appointment_id, status, is_free, valid_from, valid_until,
    created_at, updated_at
)
VALUES (
    '$CLINIC_PATIENT_ID',
    '$CLINIC_ID',
    '$DOCTOR_ID',
    '$DEPARTMENT_ID',
    (SELECT id FROM appointments WHERE clinic_patient_id = '$CLINIC_PATIENT_ID' ORDER BY created_at DESC LIMIT 1),
    'active',
    true,
    '$TOMORROW',
    '$TOMORROW'::date + INTERVAL '5 days',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);
"

echo ""
echo "🔍 Step 3: Verify renewal worked"
echo "-------------------------------"
docker exec -it drandme-backend-postgres-1 psql -U postgres -d drandme -c "
SELECT 
    id, status, is_free, valid_from, valid_until, renewed_at, created_at
FROM follow_ups 
WHERE clinic_patient_id = '$CLINIC_PATIENT_ID' 
  AND doctor_id = '$DOCTOR_ID' 
  AND department_id = '$DEPARTMENT_ID'
ORDER BY created_at DESC
LIMIT 3;
"

echo ""
echo "✅ Step 4: Test follow-up eligibility"
echo "------------------------------------"
docker exec -it drandme-backend-postgres-1 psql -U postgres -d drandme -c "
SELECT 
    'Active Follow-up Check' as test,
    COUNT(*) as active_count,
    MAX(valid_until) as latest_expiry
FROM follow_ups 
WHERE clinic_patient_id = '$CLINIC_PATIENT_ID' 
  AND doctor_id = '$DOCTOR_ID' 
  AND department_id = '$DEPARTMENT_ID'
  AND status = 'active'
  AND valid_until >= CURRENT_DATE;
"

echo ""
echo "🎯 Expected Results:"
echo "==================="
echo "✅ Should see 1 'renewed' follow-up (old one)"
echo "✅ Should see 1 'active' follow-up (new one)"
echo "✅ New follow-up should be valid for 5 days from tomorrow"
echo "✅ Frontend should now show eligibleFollowUps.length: 1"
echo ""
echo "📱 Now test follow-up creation in your frontend - should work!"
