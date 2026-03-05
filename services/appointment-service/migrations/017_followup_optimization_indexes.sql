-- Migration 017: Optimized Indexes for Follow-ups and Appointment Lookups
-- These indexes are designed to reduce CPU usage and improve response times for eligibility checks

-- Index for follow-up lookups
CREATE INDEX IF NOT EXISTS idx_followups_lookup 
ON follow_ups (clinic_patient_id, clinic_id, doctor_id, department_id, status);

-- Index for appointment lookups during eligibility checks
CREATE INDEX IF NOT EXISTS idx_appointments_lookup 
ON appointments (clinic_patient_id, clinic_id, doctor_id, department_id, status);

-- Add comments for maintenance
COMMENT ON INDEX idx_followups_lookup IS 'Optimizes follow-up eligibility API lookups';
COMMENT ON INDEX idx_appointments_lookup IS 'Optimizes appointment status checks for follow-up eligibility';
