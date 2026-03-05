-- Migration 009: Add specific date support to doctor time slots
-- Allows doctors to create both recurring weekly slots AND one-time date-specific slots

-- Add specific_date column to support one-time date-specific slots
ALTER TABLE doctor_time_slots 
ADD COLUMN specific_date DATE;

-- Update the check constraint to ensure either day_of_week or specific_date is used, but not both
ALTER TABLE doctor_time_slots
DROP CONSTRAINT IF EXISTS valid_slot_type_constraint;

ALTER TABLE doctor_time_slots
ADD CONSTRAINT valid_slot_type_constraint CHECK (
    (day_of_week IS NOT NULL AND specific_date IS NULL) OR 
    (day_of_week IS NULL AND specific_date IS NOT NULL)
);

-- Add index for specific_date queries
CREATE INDEX IF NOT EXISTS idx_doctor_time_slots_specific_date ON doctor_time_slots(specific_date);

-- Add composite index for specific date overlap checking
CREATE INDEX IF NOT EXISTS idx_doctor_time_slots_specific_date_overlap ON doctor_time_slots(doctor_id, specific_date, start_time, end_time, is_active);

-- Update comments
COMMENT ON COLUMN doctor_time_slots.specific_date IS 'For one-time date-specific slots. Either day_of_week or specific_date must be set, but not both.';
COMMENT ON COLUMN doctor_time_slots.day_of_week IS 'For recurring weekly slots (0=Sunday to 6=Saturday). Either day_of_week or specific_date must be set, but not both.';

-- Drop and recreate the overlap checking function to handle specific dates
DROP FUNCTION IF EXISTS check_doctor_time_slot_overlap() CASCADE;

CREATE OR REPLACE FUNCTION check_doctor_time_slot_overlap()
RETURNS TRIGGER AS $$
BEGIN
    -- Check for overlaps based on whether it's a recurring or specific date slot
    
    -- Case 1: NEW slot is a recurring weekly slot (day_of_week is set)
    IF NEW.day_of_week IS NOT NULL THEN
        -- Check against other recurring slots on the same day
        IF EXISTS (
            SELECT 1 FROM doctor_time_slots dts
            WHERE dts.doctor_id = NEW.doctor_id
            AND dts.day_of_week = NEW.day_of_week
            AND dts.specific_date IS NULL  -- Only check against other recurring slots
            AND dts.is_active = TRUE
            AND dts.id != COALESCE(NEW.id, '00000000-0000-0000-0000-000000000000'::uuid)
            AND (
                (NEW.start_time >= dts.start_time AND NEW.start_time < dts.end_time) OR
                (NEW.end_time > dts.start_time AND NEW.end_time <= dts.end_time) OR
                (NEW.start_time <= dts.start_time AND NEW.end_time >= dts.end_time)
            )
        ) THEN
            RAISE EXCEPTION 'Doctor already has a conflicting recurring time slot on this day. A doctor cannot be available in multiple clinics at the same time.';
        END IF;
    END IF;
    
    -- Case 2: NEW slot is a specific date slot
    IF NEW.specific_date IS NOT NULL THEN
        -- Check against other slots on the same specific date
        IF EXISTS (
            SELECT 1 FROM doctor_time_slots dts
            WHERE dts.doctor_id = NEW.doctor_id
            AND dts.specific_date = NEW.specific_date
            AND dts.is_active = TRUE
            AND dts.id != COALESCE(NEW.id, '00000000-0000-0000-0000-000000000000'::uuid)
            AND (
                (NEW.start_time >= dts.start_time AND NEW.start_time < dts.end_time) OR
                (NEW.end_time > dts.start_time AND NEW.end_time <= dts.end_time) OR
                (NEW.start_time <= dts.start_time AND NEW.end_time >= dts.end_time)
            )
        ) THEN
            RAISE EXCEPTION 'Doctor already has a conflicting time slot on this specific date. A doctor cannot be available in multiple clinics at the same time.';
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Recreate trigger
CREATE TRIGGER prevent_doctor_time_slot_overlap
    BEFORE INSERT OR UPDATE ON doctor_time_slots
    FOR EACH ROW
    EXECUTE FUNCTION check_doctor_time_slot_overlap();

