-- Migration 022: Create doctor_tokens table for token number generation
-- Tracks the current token number for each doctor per clinic per day

CREATE TABLE IF NOT EXISTS doctor_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    token_date DATE NOT NULL,
    current_token INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint: one record per doctor per clinic per date
    CONSTRAINT unique_doctor_clinic_date UNIQUE (doctor_id, clinic_id, token_date)
);

-- Create indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_doctor_tokens_lookup 
ON doctor_tokens(doctor_id, clinic_id, token_date);

CREATE INDEX IF NOT EXISTS idx_doctor_tokens_date 
ON doctor_tokens(token_date);

-- Add comments
COMMENT ON TABLE doctor_tokens IS 'Tracks token numbers for appointments per doctor per clinic per day';
COMMENT ON COLUMN doctor_tokens.current_token IS 'The last assigned token number for this doctor/clinic/date';
COMMENT ON COLUMN doctor_tokens.token_date IS 'Date for which tokens are being tracked';

