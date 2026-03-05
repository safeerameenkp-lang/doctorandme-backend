-- Migration 015: Create doctor_tokens table for persistent sequence management
-- This table tracks the current token sequence per doctor/clinic/department

CREATE TABLE IF NOT EXISTS doctor_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    doctor_id UUID NOT NULL,
    clinic_id UUID NOT NULL,
    department_id UUID,
    token_date DATE NOT NULL,
    current_token INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure unique sequence per doctor/clinic/dept/date
    -- We use a placeholder UUID '00000000-0000-0000-0000-000000000000' for NULL departments in logic
    UNIQUE (doctor_id, clinic_id, department_id, token_date)
);

-- Index for fast lookup used in GenerateTokenNumber
CREATE INDEX IF NOT EXISTS idx_doctor_tokens_lookup 
ON doctor_tokens(doctor_id, clinic_id, token_date);

-- Comment
COMMENT ON TABLE doctor_tokens IS 'Manages sequential token numbers per doctor/clinic/department';
COMMENT ON COLUMN doctor_tokens.token_date IS 'Set to 0001-01-01 for persistent global sequences (no daily reset)';
