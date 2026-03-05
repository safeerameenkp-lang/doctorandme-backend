-- Migration to make doctor tokens truly global for each doctor
-- This allows tokens to continue increasing regardless of clinic
-- And ensures we have the department_id column

ALTER TABLE doctor_tokens ADD COLUMN IF NOT EXISTS department_id UUID;

-- We want a single sequence per doctor (optionally per department)
-- Drop old clinic-based constraints
ALTER TABLE doctor_tokens DROP CONSTRAINT IF EXISTS unique_doctor_clinic_date;
ALTER TABLE doctor_tokens DROP CONSTRAINT IF EXISTS doctor_tokens_doctor_id_clinic_id_token_date_key;
DROP INDEX IF EXISTS idx_doctor_tokens_unique_composite;

-- Create a truly doctor-global unique index
-- We keep token_date in the table but for global sequences we'll use a fixed value or ignore it
-- Here we make it unique per doctor and department
CREATE UNIQUE INDEX IF NOT EXISTS idx_doctor_tokens_global_doctor_dept
ON doctor_tokens (doctor_id, COALESCE(department_id, '00000000-0000-0000-0000-000000000000'), token_date);

COMMENT ON TABLE doctor_tokens IS 'Stores sequential token counters. For global sequences, token_date is set to 0001-01-01.';
