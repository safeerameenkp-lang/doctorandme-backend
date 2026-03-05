-- Manual Renewal Test
-- This script creates a new regular appointment and tests the renewal logic

-- Step 1: Create a new regular appointment
INSERT INTO appointments (
    clinic_patient_id, clinic_id, doctor_id, department_id,
    booking_number, token_number, appointment_date, appointment_time,
    duration_minutes, consultation_type, reason, notes,
    fee_amount, payment_mode, payment_status, status,
    individual_slot_id, created_at, updated_at
)
VALUES (
    'd27a8fa7-b8bc-43e3-837b-87db5dfd4bed',  -- clinic_patient_id
    'f7658c53-72ae-4bd3-9960-741225ebc0a2',  -- clinic_id
    'ef378478-1091-472e-af40-1655e77985b3',  -- doctor_id
    'ad958b90-d383-4478-bfe3-08b53b8eeef7',  -- department_id
    'TEST-RENEWAL-001',  -- booking_number
    1,  -- token_number
    CURRENT_DATE + INTERVAL '1 day',  -- appointment_date (tomorrow)
    '10:00:00',  -- appointment_time
    5,  -- duration_minutes
    'clinic_visit',  -- consultation_type
    'Renewal test',  -- reason
    'Testing follow-up renewal',  -- notes
    100.00,  -- fee_amount
    'pay_now',  -- payment_mode
    'paid',  -- payment_status
    'confirmed',  -- status
    '0d1ed772-114d-41d6-b780-96ab0cd2d6d2',  -- individual_slot_id
    CURRENT_TIMESTAMP,  -- created_at
    CURRENT_TIMESTAMP   -- updated_at
);

-- Step 2: Get the appointment ID
SELECT 
    'New Appointment Created' as step,
    id,
    consultation_type,
    appointment_date,
    created_at
FROM appointments 
WHERE clinic_patient_id = 'd27a8fa7-b8bc-43e3-837b-87db5dfd4bed'
  AND doctor_id = 'ef378478-1091-472e-af40-1655e77985b3'
  AND department_id = 'ad958b90-d383-4478-bfe3-08b53b8eeef7'
  AND booking_number = 'TEST-RENEWAL-001';

-- Step 3: Check current follow-up status before renewal
SELECT 
    'Before Renewal' as step,
    id, status, is_free, valid_from, valid_until, created_at
FROM follow_ups 
WHERE clinic_patient_id = 'd27a8fa7-b8bc-43e3-837b-87db5dfd4bed'
  AND doctor_id = 'ef378478-1091-472e-af40-1655e77985b3'
  AND department_id = 'ad958b90-d383-4478-bfe3-08b53b8eeef7'
ORDER BY created_at DESC
LIMIT 3;
