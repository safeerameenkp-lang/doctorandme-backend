-- Migration 030: Change token_number to TEXT and update doctor_tokens logic
-- This supports alphanumeric tokens like 'DR01-01'

-- 1. Alter appointments table
ALTER TABLE appointments 
ALTER COLUMN token_number TYPE VARCHAR(50);

-- 2. Update doctor_tokens to support department-specific sequences if needed
-- Drop old constraint
ALTER TABLE doctor_tokens DROP CONSTRAINT IF EXISTS unique_doctor_clinic_date;

-- Add department_id to doctor_tokens (making it nullable if a doctor isn't in a department)
ALTER TABLE doctor_tokens ADD COLUMN IF NOT EXISTS department_id UUID REFERENCES departments(id) ON DELETE SET NULL;

-- Add new unique constraint including department_id
-- We use COALESCE for department_id to handle NULL values in the unique constraint
CREATE UNIQUE INDEX IF NOT EXISTS unique_doctor_clinic_dept_date 
ON doctor_tokens (doctor_id, clinic_id, COALESCE(department_id, '00000000-0000-0000-0000-000000000000'), token_date);

-- Add comment
COMMENT ON COLUMN appointments.token_number IS 'Formatted token number (e.g. DR01-01) for queue management';
