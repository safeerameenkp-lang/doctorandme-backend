-- Migration 018: Global Performance Optimization Indexes
-- This migration adds critical composite indexes to support high-concurrency lookup and filtering.

-- 1. Appointments Table: Primary filter combos for lists and summaries
CREATE INDEX IF NOT EXISTS idx_appointments_clinic_date_status 
ON appointments (clinic_id, appointment_date, status);

CREATE INDEX IF NOT EXISTS idx_appointments_doctor_date_status 
ON appointments (doctor_id, appointment_date, status);

CREATE INDEX IF NOT EXISTS idx_appointments_patient_doctor 
ON appointments (patient_id, doctor_id);

-- 2. Doctor Tokens: Explicit index for daily token generation logic
CREATE INDEX IF NOT EXISTS idx_doctor_tokens_lookup 
ON doctor_tokens (doctor_id, COALESCE(department_id, '00000000-0000-0000-0000-000000000000'), token_date);

-- 3. Patient Check-ins: Quick lookup for queue management
CREATE INDEX IF NOT EXISTS idx_patient_checkins_appointment_id 
ON patient_checkins (appointment_id);

-- 4. User and Patient Lookups: Speed up identification during booking
CREATE INDEX IF NOT EXISTS idx_users_phone_active 
ON users (phone) WHERE is_active = true;

CREATE INDEX IF NOT EXISTS idx_patients_mo_id 
ON patients (mo_id) WHERE is_active = true;

-- 5. Individual Slot Availability Lookups
CREATE INDEX IF NOT EXISTS idx_doctor_individual_slots_availability 
ON doctor_individual_slots (doctor_id, clinic_id, status) WHERE status = 'available';

-- 6. Leave Management
CREATE INDEX IF NOT EXISTS idx_doctor_leaves_period 
ON doctor_leaves (doctor_id, clinic_id, from_date, to_date) WHERE status = 'approved';

-- Add comments for maintenance clarity
COMMENT ON INDEX idx_appointments_clinic_date_status IS 'Speeds up dashboard summaries and clinic-level appointment lists';
COMMENT ON INDEX idx_doctor_tokens_lookup IS 'Speeds up daily token generation and avoids full scans during FOR UPDATE locks';
