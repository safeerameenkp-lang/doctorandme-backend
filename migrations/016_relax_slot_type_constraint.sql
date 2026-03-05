-- Migration 016: Relax valid_slot_type_constraint to allow both specific_date and day_of_week
-- For session-based slots, we want both: specific_date (exact date) and day_of_week (auto-calculated)

-- Drop the old restrictive constraint
ALTER TABLE doctor_time_slots
DROP CONSTRAINT IF EXISTS valid_slot_type_constraint;

-- Add new relaxed constraint that allows:
-- 1. Both specific_date AND day_of_week (for session-based slots)
-- 2. Only specific_date (for simple date-specific slots)
-- 3. Only day_of_week (for recurring weekly slots)
-- But NOT: Neither specific_date NOR day_of_week (at least one required)
ALTER TABLE doctor_time_slots
ADD CONSTRAINT valid_slot_type_constraint CHECK (
    specific_date IS NOT NULL OR day_of_week IS NOT NULL
);

-- Add comment explaining the new constraint
COMMENT ON CONSTRAINT valid_slot_type_constraint ON doctor_time_slots IS 
'At least one of specific_date or day_of_week must be set. Both can be set for session-based slots.';

