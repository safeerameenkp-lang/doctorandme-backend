-- Migration 024: Add updated_at column to appointments table
-- This column is used for tracking when an appointment was last modified (e.g., rescheduled)

DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='appointments' AND column_name='updated_at') THEN
        ALTER TABLE appointments ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
        
        -- Create index for performance on status changes tracking
        CREATE INDEX IF NOT EXISTS idx_appointments_updated_at ON appointments(updated_at);
    END IF;
END $$;

COMMENT ON COLUMN appointments.updated_at IS 'Timestamp of the last update to the appointment record';
