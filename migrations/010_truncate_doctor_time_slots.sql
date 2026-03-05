-- Migration 010: Truncate Doctor Time Slots
-- This migration will clear all data from the doctor_time_slots table

-- Disable the trigger temporarily to avoid conflicts during truncation
DROP TRIGGER IF EXISTS prevent_doctor_time_slot_overlap ON doctor_time_slots;
DROP TRIGGER IF EXISTS update_doctor_time_slots_updated_at ON doctor_time_slots;

-- Truncate the table (this removes all data but keeps the table structure)
TRUNCATE TABLE doctor_time_slots RESTART IDENTITY CASCADE;

-- Re-enable the triggers
CREATE TRIGGER prevent_doctor_time_slot_overlap
    BEFORE INSERT OR UPDATE ON doctor_time_slots
    FOR EACH ROW
    EXECUTE FUNCTION check_doctor_time_slot_overlap();

CREATE TRIGGER update_doctor_time_slots_updated_at 
    BEFORE UPDATE ON doctor_time_slots
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comment
COMMENT ON TABLE doctor_time_slots IS 'Stores doctor time slots for each clinic with offline/online separation - TRUNCATED';
