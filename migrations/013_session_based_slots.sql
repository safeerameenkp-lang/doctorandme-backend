-- Migration 013: Session-Based Time Slot System
-- Three-level structure: doctor_time_slots → sessions → individual_slots

-- Create doctor_slot_sessions table
CREATE TABLE IF NOT EXISTS doctor_slot_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    time_slot_id UUID REFERENCES doctor_time_slots(id) ON DELETE CASCADE NOT NULL,
    session_name VARCHAR(50) NOT NULL,  -- e.g., "Morning Session", "Afternoon Session"
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    max_patients INT NOT NULL DEFAULT 10,
    slot_interval_minutes INT NOT NULL DEFAULT 5,  -- Generate slots every X minutes
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_session_time_range CHECK (end_time > start_time),
    CONSTRAINT valid_max_patients_session CHECK (max_patients > 0),
    CONSTRAINT valid_slot_interval CHECK (slot_interval_minutes > 0 AND slot_interval_minutes <= 60)
);

-- Create doctor_individual_slots table (auto-generated bookable slots)
CREATE TABLE IF NOT EXISTS doctor_individual_slots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID REFERENCES doctor_slot_sessions(id) ON DELETE CASCADE NOT NULL,
    slot_start TIME NOT NULL,
    slot_end TIME NOT NULL,
    is_booked BOOLEAN DEFAULT FALSE,
    booked_patient_id UUID REFERENCES users(id) ON DELETE SET NULL,
    booked_appointment_id UUID REFERENCES appointments(id) ON DELETE SET NULL,
    status VARCHAR(20) DEFAULT 'available' CHECK (status IN ('available', 'booked', 'cancelled', 'blocked')),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_individual_slot_time CHECK (slot_end > slot_start)
);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_doctor_slot_sessions_time_slot_id ON doctor_slot_sessions(time_slot_id);
CREATE INDEX IF NOT EXISTS idx_doctor_slot_sessions_times ON doctor_slot_sessions(start_time, end_time);

CREATE INDEX IF NOT EXISTS idx_doctor_individual_slots_session_id ON doctor_individual_slots(session_id);
CREATE INDEX IF NOT EXISTS idx_doctor_individual_slots_status ON doctor_individual_slots(status);
CREATE INDEX IF NOT EXISTS idx_doctor_individual_slots_booked_patient ON doctor_individual_slots(booked_patient_id);
CREATE INDEX IF NOT EXISTS idx_doctor_individual_slots_appointment ON doctor_individual_slots(booked_appointment_id);
CREATE INDEX IF NOT EXISTS idx_doctor_individual_slots_times ON doctor_individual_slots(slot_start, slot_end);

-- Create triggers for updated_at
CREATE TRIGGER update_doctor_slot_sessions_updated_at 
    BEFORE UPDATE ON doctor_slot_sessions
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_doctor_individual_slots_updated_at 
    BEFORE UPDATE ON doctor_individual_slots
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments
COMMENT ON TABLE doctor_slot_sessions IS 'Sessions within a time slot (e.g., Morning Session, Afternoon Session)';
COMMENT ON TABLE doctor_individual_slots IS 'Auto-generated individual bookable slots within each session';

COMMENT ON COLUMN doctor_slot_sessions.slot_interval_minutes IS 'Generate individual slots every X minutes (e.g., 5 minutes)';
COMMENT ON COLUMN doctor_slot_sessions.max_patients IS 'Maximum patients for entire session';

COMMENT ON COLUMN doctor_individual_slots.is_booked IS 'Quick flag to check if slot is booked';
COMMENT ON COLUMN doctor_individual_slots.status IS 'available, booked, cancelled, or blocked';

-- Add slot_duration to doctor_time_slots if not exists
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'doctor_time_slots' AND column_name = 'slot_duration') THEN
        ALTER TABLE doctor_time_slots ADD COLUMN slot_duration INT DEFAULT 5;
    END IF;
END $$;

COMMENT ON COLUMN doctor_time_slots.slot_duration IS 'Default duration for slots in minutes';

