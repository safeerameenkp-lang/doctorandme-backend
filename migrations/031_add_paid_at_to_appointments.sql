-- Add paid_at column to appointments table
ALTER TABLE appointments ADD COLUMN IF NOT EXISTS paid_at TIMESTAMP;

-- Add updated_at column to appointments table (missing in initial schema)
ALTER TABLE appointments ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Create index for paid_at
CREATE INDEX IF NOT EXISTS idx_appointments_paid_at ON appointments(paid_at);

-- Create index for updated_at
CREATE INDEX IF NOT EXISTS idx_appointments_updated_at ON appointments(updated_at);

-- Backfill paid_at for already paid appointments (using updated_at or created_at as fallback)
UPDATE appointments 
SET paid_at = COALESCE(updated_at, created_at)
WHERE payment_status = 'paid' AND paid_at IS NULL;
