-- Migration 011: Add slot_id to appointments table
-- This links appointments to specific doctor time slots

-- Add slot_id column to appointments
ALTER TABLE appointments 
ADD COLUMN slot_id UUID REFERENCES doctor_time_slots(id) ON DELETE SET NULL;

-- Create index for slot_id for better query performance
CREATE INDEX idx_appointments_slot_id ON appointments(slot_id);

-- Add a composite index for slot_id and appointment_date for availability queries
CREATE INDEX idx_appointments_slot_date ON appointments(slot_id, appointment_date);

-- Add a composite index for slot_id and status for counting booked appointments
CREATE INDEX idx_appointments_slot_status ON appointments(slot_id, status);

-- Optional: Add a comment to document the relationship
COMMENT ON COLUMN appointments.slot_id IS 'Links appointment to a specific doctor time slot. NULL for appointments not using the slot system.';

