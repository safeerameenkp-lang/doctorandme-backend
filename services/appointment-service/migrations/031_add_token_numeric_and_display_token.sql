-- Migration 031: Add numeric token and display token fields to appointments
-- This aligns with the new requirements for daily resets and doctor-specific prefixes

ALTER TABLE appointments ADD COLUMN IF NOT EXISTS token_numeric INTEGER;
ALTER TABLE appointments ADD COLUMN IF NOT EXISTS display_token VARCHAR(20);

-- Comment for clarity
COMMENT ON COLUMN appointments.token_numeric IS 'Numeric sequential token number (resets daily per doctor)';
COMMENT ON COLUMN appointments.display_token IS 'Human-readable token with doctor prefix (e.g., M1, AR1)';

-- Update existing records if possible (best effort, or leave null)
-- For now, we leave them as is.
