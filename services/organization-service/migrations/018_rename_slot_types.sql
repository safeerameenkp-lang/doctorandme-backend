-- Migration: Rename slot types from offline/online to clinic_visit/video_consultation
-- This migration updates existing records to use the new naming convention

-- Step 1: Drop the old constraint
ALTER TABLE doctor_time_slots DROP CONSTRAINT IF EXISTS doctor_time_slots_slot_type_check;

-- Step 2: Update the data
UPDATE doctor_time_slots 
SET slot_type = 'clinic_visit' 
WHERE slot_type = 'offline';

UPDATE doctor_time_slots 
SET slot_type = 'video_consultation' 
WHERE slot_type = 'online';

-- Step 3: Add new constraint with updated values
ALTER TABLE doctor_time_slots 
ADD CONSTRAINT doctor_time_slots_slot_type_check 
CHECK (slot_type IN ('clinic_visit', 'video_consultation'));

-- Add comments
COMMENT ON COLUMN doctor_time_slots.slot_type IS 'Type of consultation: clinic_visit (in-person) or video_consultation (online)';

-- Verification query (optional, for manual checking)
-- SELECT slot_type, COUNT(*) FROM doctor_time_slots GROUP BY slot_type;

