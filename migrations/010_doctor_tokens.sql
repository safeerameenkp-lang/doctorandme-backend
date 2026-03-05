-- Migration: Doctor Token Management System
-- Purpose: Track token numbers for each doctor per clinic per day
-- Tokens reset daily and are doctor-specific

-- Create doctor_tokens table to track current token for each doctor/clinic/day
CREATE TABLE IF NOT EXISTS doctor_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    token_date DATE NOT NULL,
    current_token INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(doctor_id, clinic_id, token_date)
);

-- Create index for fast lookups
CREATE INDEX idx_doctor_tokens_lookup ON doctor_tokens(doctor_id, clinic_id, token_date);

-- Add token_number column to appointments table
ALTER TABLE appointments 
ADD COLUMN IF NOT EXISTS token_number INTEGER;

-- Create index on appointments for token lookup
CREATE INDEX IF NOT EXISTS idx_appointments_token ON appointments(doctor_id, clinic_id, appointment_date, token_number);

-- Add comment for documentation
COMMENT ON TABLE doctor_tokens IS 'Tracks the current token number for each doctor per clinic per day. Tokens reset daily.';
COMMENT ON COLUMN doctor_tokens.current_token IS 'Last assigned token number for this doctor/clinic/date combination';
COMMENT ON COLUMN appointments.token_number IS 'Sequential token number assigned to patient, resets daily per doctor';

