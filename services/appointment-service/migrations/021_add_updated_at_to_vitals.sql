-- Migration 021: Add updated_at column to patient_vitals
-- Supports tracking modifications for audit purposes.

ALTER TABLE patient_vitals 
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Add a comment for clarity
COMMENT ON COLUMN patient_vitals.updated_at IS 'Last modification timestamp';
