-- Add paid_at column to appointments table
ALTER TABLE appointments ADD COLUMN paid_at TIMESTAMP;

-- Create index for paid_at
CREATE INDEX idx_appointments_paid_at ON appointments(paid_at);

-- Backfill paid_at for already paid appointments (using updated_at or created_at as fallback)
UPDATE appointments 
SET paid_at = updated_at 
WHERE payment_status = 'paid' AND paid_at IS NULL;
