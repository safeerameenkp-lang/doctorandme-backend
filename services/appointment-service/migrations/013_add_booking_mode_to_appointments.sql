-- Add booking_mode column to appointments table
-- Default is 'slot' for backward compatibility
ALTER TABLE appointments ADD COLUMN IF NOT EXISTS booking_mode VARCHAR(20) DEFAULT 'slot';

-- Create index for better filtering performance
CREATE INDEX IF NOT EXISTS idx_appointments_booking_mode ON appointments(booking_mode);
