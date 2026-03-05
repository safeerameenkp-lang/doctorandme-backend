-- Migration 014: Add clinic_id to session tables for multi-clinic platform
-- This improves query performance and makes clinic-based filtering easier

-- Add clinic_id to doctor_slot_sessions
ALTER TABLE doctor_slot_sessions 
ADD COLUMN clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;

-- Add clinic_id to doctor_individual_slots
ALTER TABLE doctor_individual_slots 
ADD COLUMN clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE;

-- Create indexes on clinic_id for better performance
CREATE INDEX IF NOT EXISTS idx_doctor_slot_sessions_clinic_id ON doctor_slot_sessions(clinic_id);
CREATE INDEX IF NOT EXISTS idx_doctor_individual_slots_clinic_id ON doctor_individual_slots(clinic_id);

-- Create composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_doctor_slot_sessions_clinic_time ON doctor_slot_sessions(clinic_id, start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_doctor_individual_slots_clinic_status ON doctor_individual_slots(clinic_id, status);

-- Add comments
COMMENT ON COLUMN doctor_slot_sessions.clinic_id IS 'Clinic where this session takes place (denormalized for performance)';
COMMENT ON COLUMN doctor_individual_slots.clinic_id IS 'Clinic where this slot is available (denormalized for performance)';

-- Update existing records to populate clinic_id from doctor_time_slots
UPDATE doctor_slot_sessions dss
SET clinic_id = dts.clinic_id
FROM doctor_time_slots dts
WHERE dss.time_slot_id = dts.id;

UPDATE doctor_individual_slots dis
SET clinic_id = dss.clinic_id
FROM doctor_slot_sessions dss
WHERE dis.session_id = dss.id;

-- Make clinic_id NOT NULL after populating existing data
ALTER TABLE doctor_slot_sessions ALTER COLUMN clinic_id SET NOT NULL;
ALTER TABLE doctor_individual_slots ALTER COLUMN clinic_id SET NOT NULL;

