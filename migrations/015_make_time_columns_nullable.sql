-- Migration 015: Make start_time and end_time nullable in doctor_time_slots
-- For session-based slots, times are stored at session level, not day level

-- Make start_time and end_time nullable
ALTER TABLE doctor_time_slots 
ALTER COLUMN start_time DROP NOT NULL;

ALTER TABLE doctor_time_slots 
ALTER COLUMN end_time DROP NOT NULL;

-- Update comments
COMMENT ON COLUMN doctor_time_slots.start_time IS 'Start time (nullable for session-based slots, where times are at session level)';
COMMENT ON COLUMN doctor_time_slots.end_time IS 'End time (nullable for session-based slots, where times are at session level)';

-- Note: The valid_time_range constraint will still work when both values are present
-- It will be ignored when either is NULL

