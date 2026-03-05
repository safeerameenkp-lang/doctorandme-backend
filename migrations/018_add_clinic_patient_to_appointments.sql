-- Migration 018: Add clinic_patient_id to appointments table
-- Supports appointments with clinic-specific patients (isolated per clinic)

-- Add clinic_patient_id column
ALTER TABLE appointments 
ADD COLUMN clinic_patient_id UUID REFERENCES clinic_patients(id) ON DELETE SET NULL;

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_appointments_clinic_patient ON appointments(clinic_patient_id);

-- Add comment
COMMENT ON COLUMN appointments.clinic_patient_id IS 'Reference to clinic-specific patient (alternative to global patient_id)';

-- Note: An appointment can have EITHER patient_id (global) OR clinic_patient_id (clinic-specific)
-- Both can be NULL for walk-in appointments

