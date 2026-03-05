-- Migration: Fix duplicate free follow-ups
-- Only the FIRST follow-up per patient per doctor should be free
-- All subsequent follow-ups should require payment

-- Step 1: Identify and fix duplicate free follow-ups
-- For each patient + doctor + last_regular_appointment combination,
-- keep only the FIRST free follow-up, mark others as 'pending' (requiring payment)

WITH ranked_followups AS (
    SELECT 
        a.id,
        a.clinic_patient_id,
        a.doctor_id,
        a.appointment_date,
        a.appointment_time,
        ROW_NUMBER() OVER (
            PARTITION BY a.clinic_patient_id, a.doctor_id 
            ORDER BY a.appointment_date ASC, a.appointment_time ASC
        ) as follow_up_rank
    FROM appointments a
    WHERE a.consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
      AND a.payment_status = 'waived'
      AND a.status NOT IN ('cancelled', 'no_show')
)
UPDATE appointments
SET 
    payment_status = 'pending',
    fee_amount = COALESCE(
        (SELECT d.follow_up_fee FROM doctors d WHERE d.id = appointments.doctor_id),
        200.00
    )
FROM ranked_followups rf
WHERE appointments.id = rf.id
  AND rf.follow_up_rank > 1;  -- Keep first (rank 1), mark others as pending

-- Step 2: Add comment
COMMENT ON COLUMN appointments.payment_status IS 'Payment status: paid, pending, waived (only first follow-up within 5 days is free)';

-- Verification query
SELECT 
    'Fixed duplicate free follow-ups' as action,
    COUNT(*) as rows_updated
FROM appointments
WHERE consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
  AND payment_status = 'pending'
  AND updated_at > CURRENT_TIMESTAMP - INTERVAL '1 minute';

