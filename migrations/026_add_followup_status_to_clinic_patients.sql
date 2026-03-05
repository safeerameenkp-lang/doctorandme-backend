-- Migration 026: Add Follow-Up Status Tracking to Clinic Patients
-- Adds fields to track follow-up status lifecycle per clinic patient

-- Add status tracking columns to clinic_patients table
ALTER TABLE clinic_patients 
ADD COLUMN IF NOT EXISTS current_followup_status VARCHAR(20) DEFAULT 'none' CHECK (current_followup_status IN ('none', 'active', 'used', 'expired', 'renewed')),
ADD COLUMN IF NOT EXISTS last_appointment_id UUID REFERENCES appointments(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS last_followup_id UUID REFERENCES follow_ups(id) ON DELETE SET NULL;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_clinic_patients_followup_status ON clinic_patients(current_followup_status);
CREATE INDEX IF NOT EXISTS idx_clinic_patients_last_appointment ON clinic_patients(last_appointment_id);
CREATE INDEX IF NOT EXISTS idx_clinic_patients_last_followup ON clinic_patients(last_followup_id);

-- Add comments
COMMENT ON COLUMN clinic_patients.current_followup_status IS 'Current follow-up status: none, active (has valid follow-up), used, expired, or renewed';
COMMENT ON COLUMN clinic_patients.last_appointment_id IS 'Reference to the last appointment for this patient';
COMMENT ON COLUMN clinic_patients.last_followup_id IS 'Reference to the last follow-up record for this patient';

-- Update existing records to have 'none' status (should already be default, but ensure consistency)
UPDATE clinic_patients SET current_followup_status = 'none' WHERE current_followup_status IS NULL;

