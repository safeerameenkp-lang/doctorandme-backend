-- Follow-Up Renewal Diagnostic Script
-- Run this to check the current state of follow-ups

-- 1. Check all follow-ups for a specific patient
SELECT 
    f.id,
    cp.first_name || ' ' || cp.last_name as patient_name,
    u.first_name || ' ' || u.last_name as doctor_name,
    dept.name as department,
    f.status,
    f.is_free,
    f.valid_from,
    f.valid_until,
    f.source_appointment_id,
    f.renewed_at,
    f.renewed_by_appointment_id,
    f.created_at,
    CASE 
        WHEN f.valid_until >= CURRENT_DATE THEN 'ACTIVE'
        ELSE 'EXPIRED'
    END as validity_status
FROM follow_ups f
JOIN clinic_patients cp ON cp.id = f.clinic_patient_id
JOIN doctors d ON d.id = f.doctor_id
JOIN users u ON u.id = d.user_id
LEFT JOIN departments dept ON dept.id = f.department_id
WHERE f.clinic_patient_id = 'REPLACE_WITH_PATIENT_ID'
ORDER BY f.created_at DESC;

-- 2. Check appointments for the same patient
SELECT 
    a.id,
    a.appointment_date,
    a.consultation_type,
    a.payment_status,
    a.status,
    u.first_name || ' ' || u.last_name as doctor_name,
    dept.name as department,
    CASE 
        WHEN a.consultation_type IN ('clinic_visit', 'video_consultation') THEN 'REGULAR'
        WHEN a.consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video') THEN 'FOLLOW-UP'
        ELSE 'OTHER'
    END as appointment_category
FROM appointments a
JOIN doctors d ON d.id = a.doctor_id
JOIN users u ON u.id = d.user_id
LEFT JOIN departments dept ON dept.id = a.department_id
WHERE a.clinic_patient_id = 'REPLACE_WITH_PATIENT_ID'
ORDER BY a.appointment_date DESC, a.created_at DESC;

-- 3. Check if there are any active follow-ups that should be renewed
SELECT 
    f.id,
    cp.first_name || ' ' || cp.last_name as patient_name,
    f.status,
    f.is_free,
    f.valid_until,
    f.created_at,
    'SHOULD_BE_RENEWED' as issue
FROM follow_ups f
JOIN clinic_patients cp ON cp.id = f.clinic_patient_id
WHERE f.clinic_patient_id = 'REPLACE_WITH_PATIENT_ID'
  AND f.status = 'active'
  AND f.valid_until >= CURRENT_DATE
  AND EXISTS (
      -- Check if there's a newer regular appointment
      SELECT 1 FROM appointments a
      WHERE a.clinic_patient_id = f.clinic_patient_id
        AND a.doctor_id = f.doctor_id
        AND a.department_id = f.department_id
        AND a.consultation_type IN ('clinic_visit', 'video_consultation')
        AND a.status IN ('completed', 'confirmed')
        AND a.appointment_date > f.valid_from
        AND a.created_at > f.created_at
  );

-- 4. Check for duplicate active follow-ups (should not happen)
SELECT 
    clinic_patient_id,
    doctor_id,
    department_id,
    COUNT(*) as active_count
FROM follow_ups
WHERE status = 'active'
  AND valid_until >= CURRENT_DATE
GROUP BY clinic_patient_id, doctor_id, department_id
HAVING COUNT(*) > 1;

-- 5. Check recent follow-up creation activity
SELECT 
    f.id,
    cp.first_name || ' ' || cp.last_name as patient_name,
    f.status,
    f.is_free,
    f.valid_from,
    f.valid_until,
    f.created_at,
    a.appointment_date as source_appointment_date,
    a.consultation_type as source_appointment_type
FROM follow_ups f
JOIN clinic_patients cp ON cp.id = f.clinic_patient_id
LEFT JOIN appointments a ON a.id = f.source_appointment_id
WHERE f.clinic_patient_id = 'REPLACE_WITH_PATIENT_ID'
ORDER BY f.created_at DESC
LIMIT 10;
