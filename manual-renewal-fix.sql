-- Manual Follow-Up Renewal Fix
-- This script manually creates a new active follow-up for testing

-- Step 1: Check current status
SELECT 
    'Current Status' as step,
    status,
    COUNT(*) as count,
    MAX(valid_until) as latest_expiry
FROM follow_ups 
WHERE clinic_patient_id = 'd27a8fa7-b8bc-43e3-837b-87db5dfd4bed'  -- ashiq m
  AND doctor_id = 'ef378478-1091-472e-af40-1655e77985b3'  -- Same doctor
  AND department_id = 'ad958b90-d383-4478-bfe3-08b53b8eeef7'  -- Same department
GROUP BY status;

-- Step 2: Mark existing follow-ups as renewed
UPDATE follow_ups
SET status = 'renewed',
    renewed_at = CURRENT_TIMESTAMP,
    renewed_by_appointment_id = (SELECT id FROM appointments WHERE clinic_patient_id = 'd27a8fa7-b8bc-43e3-837b-87db5dfd4bed' ORDER BY created_at DESC LIMIT 1),
    updated_at = CURRENT_TIMESTAMP
WHERE clinic_patient_id = 'd27a8fa7-b8bc-43e3-837b-87db5dfd4bed'
  AND doctor_id = 'ef378478-1091-472e-af40-1655e77985b3'
  AND department_id = 'ad958b90-d383-4478-bfe3-08b53b8eeef7'
  AND status IN ('active', 'expired');

-- Step 3: Create new active follow-up
INSERT INTO follow_ups (
    clinic_patient_id, 
    clinic_id, 
    doctor_id, 
    department_id,
    source_appointment_id, 
    status, 
    is_free, 
    valid_from, 
    valid_until,
    created_at, 
    updated_at
)
VALUES (
    'd27a8fa7-b8bc-43e3-837b-87db5dfd4bed',  -- clinic_patient_id
    (SELECT clinic_id FROM clinic_patients WHERE id = 'd27a8fa7-b8bc-43e3-837b-87db5dfd4bed'),  -- clinic_id
    'ef378478-1091-472e-af40-1655e77985b3',  -- doctor_id
    'ad958b90-d383-4478-bfe3-08b53b8eeef7',  -- department_id
    (SELECT id FROM appointments WHERE clinic_patient_id = 'd27a8fa7-b8bc-43e3-837b-87db5dfd4bed' ORDER BY created_at DESC LIMIT 1),  -- source_appointment_id
    'active',  -- status
    true,  -- is_free
    CURRENT_DATE,  -- valid_from
    CURRENT_DATE + INTERVAL '5 days',  -- valid_until
    CURRENT_TIMESTAMP,  -- created_at
    CURRENT_TIMESTAMP   -- updated_at
);

-- Step 4: Verify the fix
SELECT 
    'After Fix' as step,
    id,
    status,
    is_free,
    valid_from,
    valid_until,
    renewed_at,
    created_at
FROM follow_ups 
WHERE clinic_patient_id = 'd27a8fa7-b8bc-43e3-837b-87db5dfd4bed'
  AND doctor_id = 'ef378478-1091-472e-af40-1655e77985b3'
  AND department_id = 'ad958b90-d383-4478-bfe3-08b53b8eeef7'
ORDER BY created_at DESC
LIMIT 3;
