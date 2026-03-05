-- Migration 008: Doctor Time Slots Management
-- Allows doctors to set different time slots for each clinic
-- Prevents overlapping time slots across all clinics for the same doctor

-- Create doctor time slots table
CREATE TABLE IF NOT EXISTS doctor_time_slots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    doctor_id UUID REFERENCES doctors(id) ON DELETE CASCADE NOT NULL,
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE NOT NULL,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6), -- 0=Sunday, 1=Monday, etc.
    slot_type VARCHAR(20) NOT NULL CHECK (slot_type IN ('offline', 'online')), -- offline or online
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    max_patients INTEGER DEFAULT 1, -- Maximum patients per slot
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_time_range CHECK (end_time > start_time),
    CONSTRAINT valid_max_patients CHECK (max_patients > 0)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_doctor_time_slots_doctor_id ON doctor_time_slots(doctor_id);
CREATE INDEX IF NOT EXISTS idx_doctor_time_slots_clinic_id ON doctor_time_slots(clinic_id);
CREATE INDEX IF NOT EXISTS idx_doctor_time_slots_day_of_week ON doctor_time_slots(day_of_week);
CREATE INDEX IF NOT EXISTS idx_doctor_time_slots_slot_type ON doctor_time_slots(slot_type);
CREATE INDEX IF NOT EXISTS idx_doctor_time_slots_active ON doctor_time_slots(is_active);
CREATE INDEX IF NOT EXISTS idx_doctor_time_slots_time_range ON doctor_time_slots(start_time, end_time);

-- Create composite index for overlap checking
CREATE INDEX IF NOT EXISTS idx_doctor_time_slots_overlap_check ON doctor_time_slots(doctor_id, day_of_week, start_time, end_time, is_active);

-- Create trigger for updated_at
DROP TRIGGER IF EXISTS update_doctor_time_slots_updated_at ON doctor_time_slots;
CREATE TRIGGER update_doctor_time_slots_updated_at 
    BEFORE UPDATE ON doctor_time_slots
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE doctor_time_slots IS 'Stores doctor time slots for each clinic with offline/online separation';
COMMENT ON COLUMN doctor_time_slots.day_of_week IS 'Day of week: 0=Sunday, 1=Monday, 2=Tuesday, 3=Wednesday, 4=Thursday, 5=Friday, 6=Saturday';
COMMENT ON COLUMN doctor_time_slots.slot_type IS 'Type of consultation: offline (in-person) or online (telemedicine)';
COMMENT ON COLUMN doctor_time_slots.start_time IS 'Start time of the slot (HH:MM format)';
COMMENT ON COLUMN doctor_time_slots.end_time IS 'End time of the slot (HH:MM format)';
COMMENT ON COLUMN doctor_time_slots.max_patients IS 'Maximum number of patients that can book this slot';
COMMENT ON COLUMN doctor_time_slots.notes IS 'Additional notes about this time slot';

-- Create function to check for overlapping time slots
CREATE OR REPLACE FUNCTION check_doctor_time_slot_overlap()
RETURNS TRIGGER AS $$
BEGIN
    -- Check if there's an overlapping time slot for the same doctor on the same day
    -- (regardless of clinic - doctor can't be in two places at once)
    IF EXISTS (
        SELECT 1 FROM doctor_time_slots dts
        WHERE dts.doctor_id = NEW.doctor_id
        AND dts.day_of_week = NEW.day_of_week
        AND dts.is_active = TRUE
        AND dts.id != COALESCE(NEW.id, '00000000-0000-0000-0000-000000000000'::uuid)
        AND (
            -- New slot starts during existing slot
            (NEW.start_time >= dts.start_time AND NEW.start_time < dts.end_time) OR
            -- New slot ends during existing slot
            (NEW.end_time > dts.start_time AND NEW.end_time <= dts.end_time) OR
            -- New slot completely contains existing slot
            (NEW.start_time <= dts.start_time AND NEW.end_time >= dts.end_time)
        )
    ) THEN
        RAISE EXCEPTION 'Doctor already has a conflicting time slot on this day. A doctor cannot be available in multiple clinics at the same time.';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to prevent overlapping time slots
DROP TRIGGER IF EXISTS prevent_doctor_time_slot_overlap ON doctor_time_slots;
CREATE TRIGGER prevent_doctor_time_slot_overlap
    BEFORE INSERT OR UPDATE ON doctor_time_slots
    FOR EACH ROW
    EXECUTE FUNCTION check_doctor_time_slot_overlap();
