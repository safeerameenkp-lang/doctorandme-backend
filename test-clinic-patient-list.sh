#!/bin/bash

echo "🔍 TESTING CLINIC PATIENT LIST FOLLOW-UP STATUS"
echo "=============================================="

# Test data
CLINIC_ID="f7658c53-72ae-4bd3-9960-741225ebc0a2"
PATIENT_ID="d27a8fa7-b8bc-43e3-837b-87db5dfd4bed"  # ashiq m

echo "📋 Test Data:"
echo "  Clinic ID: $CLINIC_ID"
echo "  Patient ID: $PATIENT_ID"
echo ""

echo "🔍 Step 1: Check follow-ups in database"
echo "---------------------------------------"
docker exec -it drandme-backend-postgres-1 psql -U postgres -d drandme -c "
SELECT 
    'Follow-ups in DB' as type,
    id, clinic_patient_id, doctor_id, department_id, 
    status, is_free, valid_from, valid_until, created_at
FROM follow_ups 
WHERE clinic_patient_id = '$PATIENT_ID' 
  AND clinic_id = '$CLINIC_ID'
ORDER BY created_at DESC
LIMIT 3;
"

echo ""
echo "🔍 Step 2: Check clinic patient record"
echo "-------------------------------------"
docker exec -it drandme-backend-postgres-1 psql -U postgres -d drandme -c "
SELECT 
    'Clinic Patient' as type,
    id, clinic_id, first_name, last_name, phone, is_active
FROM clinic_patients 
WHERE id = '$PATIENT_ID' 
  AND clinic_id = '$CLINIC_ID';
"

echo ""
echo "🔍 Step 3: Test FollowUpHelper query directly"
echo "---------------------------------------------"
docker exec -it drandme-backend-postgres-1 psql -U postgres -d drandme -c "
SELECT 
    'Active Follow-ups Query' as type,
    f.id as follow_up_id,
    f.doctor_id,
    COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name,
    f.department_id,
    dept.name as department_name,
    f.source_appointment_id,
    f.valid_from,
    f.valid_until,
    f.is_free,
    f.status
FROM follow_ups f
JOIN doctors d ON d.id = f.doctor_id
JOIN users u ON u.id = d.user_id
LEFT JOIN departments dept ON dept.id = f.department_id
WHERE f.clinic_patient_id = '$PATIENT_ID'
  AND f.clinic_id = '$CLINIC_ID'
  AND f.status = 'active'
  AND f.valid_until >= CURRENT_DATE
ORDER BY f.valid_until ASC;
"

echo ""
echo "🚀 Step 4: Test clinic patient list API"
echo "--------------------------------------"
echo "This should return follow-up data in the response"
echo ""

# Test the API call (will fail due to auth, but shows the endpoint)
echo "API Endpoint: GET /api/v1/clinic-specific-patients?clinic_id=$CLINIC_ID"
echo "Expected: Should include eligible_follow_ups array with follow-up data"
echo ""

echo "🔧 Step 5: Manual verification"
echo "------------------------------"
echo "1. Check if follow-ups exist in database ✅"
echo "2. Check if clinic patient exists ✅" 
echo "3. Check if FollowUpHelper query works ✅"
echo "4. Test API endpoint (requires auth) ⚠️"
echo ""

echo "🎯 Expected Results:"
echo "==================="
echo "✅ Should see active follow-ups in database"
echo "✅ Should see clinic patient record"
echo "✅ Should see follow-up data in FollowUpHelper query"
echo "✅ API should return eligible_follow_ups array"
echo ""

echo "❌ If API doesn't return follow-up data:"
echo "1. Check organization-service logs"
echo "2. Verify FollowUpHelper is working"
echo "3. Check if clinic patient list is calling populateFullAppointmentHistory"
echo ""

echo "📱 Test in frontend:"
echo "==================="
echo "1. Go to clinic patient list"
echo "2. Look for patient 'ashiq m'"
echo "3. Check if follow-up status is shown"
echo "4. Should show 'Free follow-up available' or similar"
