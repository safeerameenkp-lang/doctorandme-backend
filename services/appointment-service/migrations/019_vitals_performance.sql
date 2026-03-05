-- Migration 019: Vitals Performance Optimization
-- This migration adds indexes to patient_vitals for faster history lookups and queue management.

-- 1. Patient Vitals: Quick lookup for specific appointments
CREATE INDEX IF NOT EXISTS idx_patient_vitals_appointment_id 
ON patient_vitals (appointment_id);

-- 2. Patient Vitals: Historical sorting and filtering
CREATE INDEX IF NOT EXISTS idx_patient_vitals_recorded_at 
ON patient_vitals (recorded_at DESC);

-- 3. Patient Vitals: Patient-centric lookup
CREATE INDEX IF NOT EXISTS idx_patient_vitals_clinic_patient_id 
ON patient_vitals (clinic_patient_id);

-- 4. Appointments: Status-specific sorting for queue
CREATE INDEX IF NOT EXISTS idx_appointments_queue 
ON appointments (clinic_id, status, appointment_date, appointment_time);

COMMENT ON INDEX idx_patient_vitals_appointment_id IS 'Speeds up check-in/vitals flow';
COMMENT ON INDEX idx_patient_vitals_recorded_at IS 'Speeds up vitals history display';
COMMENT ON INDEX idx_appointments_queue IS 'Speeds up doctor queue and clinic dashboard lists';
