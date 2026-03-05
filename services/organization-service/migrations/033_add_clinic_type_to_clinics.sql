-- Migration to add clinic_type to clinics table
ALTER TABLE clinics ADD COLUMN IF NOT EXISTS clinic_type VARCHAR(50) NOT NULL DEFAULT 'General';

-- Add index on clinic_type since we might want to filter clinics by type in the future
CREATE INDEX IF NOT EXISTS idx_clinics_clinic_type ON clinics(clinic_type);
