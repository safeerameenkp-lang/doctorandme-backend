-- Migration to fix doctor_tokens table
-- Add department_id and update unique constraints

ALTER TABLE doctor_tokens ADD COLUMN IF NOT EXISTS department_id UUID;

-- Update unique constraint to include department_id
-- We first drop the old constraint if it exists
-- The constraint name might be unique_doctor_clinic_date or similar from previous manual creations
ALTER TABLE doctor_tokens DROP CONSTRAINT IF EXISTS unique_doctor_clinic_date;
ALTER TABLE doctor_tokens DROP CONSTRAINT IF EXISTS doctor_tokens_doctor_id_clinic_id_token_date_key;

-- Create a new unique constraint including department_id
-- We use a dummy UUID '00000000-0000-0000-0000-000000000000' for cases where department is null
CREATE UNIQUE INDEX IF NOT EXISTS idx_doctor_tokens_unique_composite 
ON doctor_tokens (doctor_id, clinic_id, token_date, COALESCE(department_id, '00000000-0000-0000-0000-000000000000'));

COMMENT ON COLUMN doctor_tokens.department_id IS 'Optional department ID for separate token sequences per department';
