-- Migration 022: Add unique constraint to appointment vitals
-- This enables atomic UPSERT operations (Insert or Update if exists)

ALTER TABLE patient_vitals 
ADD CONSTRAINT unique_appointment_vitals UNIQUE (appointment_id);

COMMENT ON CONSTRAINT unique_appointment_vitals ON patient_vitals IS 'Ensures only one vitals record exists per appointment, enabling atomic UPSERT logic.';
