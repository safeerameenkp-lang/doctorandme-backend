-- Migration to allow multiple patients with the same phone number in a clinic
-- Dropping the unique constraint on (clinic_id, phone)
ALTER TABLE clinic_patients DROP CONSTRAINT IF EXISTS unique_phone_per_clinic;

-- Optional: Add a comment explaining the change
COMMENT ON COLUMN clinic_patients.phone IS 'Patient phone number. Multiple patients in the same clinic can share the same phone number (e.g., family members).';
