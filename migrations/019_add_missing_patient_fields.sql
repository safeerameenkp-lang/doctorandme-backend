-- Migration 019: Add Missing Patient Fields
-- Adds fields from patient form: age, address, district, state, smoking, alcohol, height, weight

-- Add new columns to clinic_patients table
ALTER TABLE clinic_patients 
ADD COLUMN IF NOT EXISTS age INTEGER,
ADD COLUMN IF NOT EXISTS address1 VARCHAR(200),
ADD COLUMN IF NOT EXISTS address2 VARCHAR(200),
ADD COLUMN IF NOT EXISTS district VARCHAR(100),
ADD COLUMN IF NOT EXISTS state VARCHAR(100),
ADD COLUMN IF NOT EXISTS smoking_status VARCHAR(20),
ADD COLUMN IF NOT EXISTS alcohol_use VARCHAR(20),
ADD COLUMN IF NOT EXISTS height_cm INTEGER,
ADD COLUMN IF NOT EXISTS weight_kg INTEGER;

-- Add indexes for new searchable fields
CREATE INDEX IF NOT EXISTS idx_clinic_patients_district ON clinic_patients(district);
CREATE INDEX IF NOT EXISTS idx_clinic_patients_state ON clinic_patients(state);
CREATE INDEX IF NOT EXISTS idx_clinic_patients_age ON clinic_patients(age);

-- Add comments for new fields
COMMENT ON COLUMN clinic_patients.age IS 'Patient age in years';
COMMENT ON COLUMN clinic_patients.address1 IS 'Primary address line';
COMMENT ON COLUMN clinic_patients.address2 IS 'Secondary address line';
COMMENT ON COLUMN clinic_patients.district IS 'Patient district';
COMMENT ON COLUMN clinic_patients.state IS 'Patient state';
COMMENT ON COLUMN clinic_patients.smoking_status IS 'Smoking status (Yes/No)';
COMMENT ON COLUMN clinic_patients.alcohol_use IS 'Alcohol use status (Yes/No)';
COMMENT ON COLUMN clinic_patients.height_cm IS 'Patient height in centimeters';
COMMENT ON COLUMN clinic_patients.weight_kg IS 'Patient weight in kilograms';
