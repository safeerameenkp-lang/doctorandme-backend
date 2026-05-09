-- Migration 032: Production-level Token System Hardening
-- This migration aligns the system with large-scale hospital requirements

-- 1. Ensure doctor_tokens table is robust and matches recommended schema
CREATE TABLE IF NOT EXISTS doctor_tokens (
    id SERIAL PRIMARY KEY,
    clinic_id UUID NOT NULL,
    doctor_id UUID NOT NULL,
    token_date DATE NOT NULL,
    current_token INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(clinic_id, doctor_id, token_date)
);

-- 2. Add high-performance lookup index to appointments
CREATE INDEX IF NOT EXISTS idx_token_lookup
ON appointments (
    clinic_id,
    doctor_id,
    appointment_date,
    token_numeric
);

-- 3. Add doctor_prefix column to appointments for fast retrieval
ALTER TABLE appointments ADD COLUMN IF NOT EXISTS doctor_prefix VARCHAR(5);

COMMENT ON TABLE doctor_tokens IS 'Atomic sequence tracker for daily doctor tokens with concurrency protection';
