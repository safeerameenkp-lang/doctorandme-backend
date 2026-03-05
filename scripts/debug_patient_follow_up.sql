-- Debug script to check patient's appointment history and follow-up eligibility
-- Replace the UUIDs with your actual patient, clinic, doctor, and department IDs

-- Check all appointments for a specific patient
SELECT 
    a.id,
    a.appointment_date,
    a.appointment_time,
    a.status,
    a.consultation_type,
    a.doctor_id,
    d.user_id,
    u.first_name || ' ' || u.last_name as doctor_name,
    a.department_id,
    dept.name as department,
    a.payment_status,
    a.fee_amount,
    CASE 
        WHEN a.appointment_date < CURRENT_DATE THEN 'PAST'
        WHEN a.appointment_date = CURRENT_DATE THEN 'TODAY'
        ELSE 'FUTURE'
    END as timing
FROM appointments a
LEFT JOIN doctors d ON d.id = a.doctor_id
LEFT JOIN users u ON u.id = d.user_id
LEFT JOIN departments dept ON dept.id = a.department_id
WHERE a.clinic_patient_id = 'REPLACE_WITH_PATIENT_ID'  -- ← REPLACE THIS
  AND a.clinic_id = 'REPLACE_WITH_CLINIC_ID'           -- ← REPLACE THIS
ORDER BY a.appointment_date DESC, a.appointment_time DESC;

-- Check what the eligibility query sees
-- This shows what appointment would be used for follow-up eligibility
SELECT 
    a.id,
    a.doctor_id,
    a.department_id,
    a.appointment_date,
    a.status,
    a.consultation_type,
    EXTRACT(DAY FROM (CURRENT_DATE - a.appointment_date)) as days_ago
FROM appointments a
WHERE a.clinic_patient_id = 'REPLACE_WITH_PATIENT_ID'  -- ← REPLACE THIS
  AND a.clinic_id = 'REPLACE_WITH_CLINIC_ID'           -- ← REPLACE THIS
  AND a.status IN ('completed', 'confirmed')           -- ✅ Must be completed or confirmed
  AND a.appointment_date <= CURRENT_DATE               -- ✅ Must be today or past
  -- AND a.doctor_id = 'REPLACE_WITH_DOCTOR_ID'        -- ← Uncomment to filter by doctor
  -- AND a.department_id = 'REPLACE_WITH_DEPT_ID'      -- ← Uncomment to filter by department
ORDER BY a.appointment_date DESC, a.appointment_time DESC
LIMIT 1;

-- Check if free follow-up already used
SELECT 
    a.id,
    a.appointment_date,
    a.consultation_type,
    a.payment_status,
    a.fee_amount
FROM appointments
WHERE a.clinic_patient_id = 'REPLACE_WITH_PATIENT_ID'  -- ← REPLACE THIS
  AND a.clinic_id = 'REPLACE_WITH_CLINIC_ID'           -- ← REPLACE THIS
  AND a.doctor_id = 'REPLACE_WITH_DOCTOR_ID'           -- ← REPLACE THIS
  -- AND a.department_id = 'REPLACE_WITH_DEPT_ID'      -- ← Uncomment to filter by department
  AND a.consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
  AND a.payment_status = 'waived'
  AND a.status NOT IN ('cancelled', 'no_show')
ORDER BY a.appointment_date DESC;

