-- Migration 012: Make day_of_week nullable to support specific_date slots
-- This fixes the issue where specific_date slots fail because day_of_week is still NOT NULL

-- Make day_of_week nullable so it can be NULL when specific_date is set
ALTER TABLE doctor_time_slots 
ALTER COLUMN day_of_week DROP NOT NULL;

-- Add comment for clarity
COMMENT ON COLUMN doctor_time_slots.day_of_week IS 'For recurring weekly slots (0=Sunday to 6=Saturday). Either day_of_week or specific_date must be set (enforced by valid_slot_type_constraint), but not both. This column is nullable.';

