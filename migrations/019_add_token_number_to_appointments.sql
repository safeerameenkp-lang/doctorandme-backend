-- Migration 019: Add token_number to appointments table
-- Token number is auto-generated for queue management

-- Add token_number column
ALTER TABLE appointments 
ADD COLUMN token_number INTEGER;

-- Create index for token queries
CREATE INDEX IF NOT EXISTS idx_appointments_token_number ON appointments(token_number);

-- Create composite index for clinic + date + token
CREATE INDEX IF NOT EXISTS idx_appointments_clinic_date_token ON appointments(clinic_id, appointment_date, token_number);

-- Add comment
COMMENT ON COLUMN appointments.token_number IS 'Auto-generated token number for queue management (per doctor per clinic per date)';

