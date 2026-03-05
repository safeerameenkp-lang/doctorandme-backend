-- Migration 021: Add capacity tracking to individual slots
-- Allows multiple patients to book the same slot until max_patients reached

-- Add max_patients and available_count columns
ALTER TABLE doctor_individual_slots 
ADD COLUMN max_patients INTEGER DEFAULT 1,
ADD COLUMN available_count INTEGER DEFAULT 1;

-- Update existing slots to have consistent capacity
UPDATE doctor_individual_slots 
SET max_patients = 1, available_count = CASE WHEN is_booked THEN 0 ELSE 1 END;

-- Make columns NOT NULL after setting defaults
ALTER TABLE doctor_individual_slots 
ALTER COLUMN max_patients SET NOT NULL,
ALTER COLUMN available_count SET NOT NULL;

-- Add check constraint
ALTER TABLE doctor_individual_slots
ADD CONSTRAINT check_available_count_valid 
CHECK (available_count >= 0 AND available_count <= max_patients);

-- Create index for availability queries
CREATE INDEX IF NOT EXISTS idx_individual_slots_availability 
ON doctor_individual_slots(session_id, is_booked, available_count) 
WHERE available_count > 0;

-- Add comment
COMMENT ON COLUMN doctor_individual_slots.max_patients IS 'Maximum number of patients that can book this slot';
COMMENT ON COLUMN doctor_individual_slots.available_count IS 'Current number of available spots in this slot';

