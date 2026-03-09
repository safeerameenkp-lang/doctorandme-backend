-- Relax constraints to allow multiple clinics with same code
-- This supports the request to remove "waste" validations
ALTER TABLE clinics DROP CONSTRAINT IF EXISTS clinics_clinic_code_key CASCADE;
ALTER TABLE clinics DROP CONSTRAINT IF EXISTS unique_clinic_code CASCADE;

-- Also remove unique index if it exists explicitly
DROP INDEX IF EXISTS idx_clinics_clinic_code;






