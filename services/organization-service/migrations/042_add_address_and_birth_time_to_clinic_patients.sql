-- Migration 042: Add Address and Birth Time to Clinic Patients
-- Adds dedicated address and birth_time fields

ALTER TABLE clinic_patients 
ADD COLUMN IF NOT EXISTS address TEXT,
ADD COLUMN IF NOT EXISTS birth_time VARCHAR(20);

-- Add comments
COMMENT ON COLUMN clinic_patients.address IS 'Full address of the patient';
COMMENT ON COLUMN clinic_patients.birth_time IS 'Time of birth (optional)';
