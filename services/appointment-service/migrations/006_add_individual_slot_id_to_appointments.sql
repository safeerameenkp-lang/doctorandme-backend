-- Migration 020: Add individual_slot_id to appointments table
-- This links appointments to specific individual slots for better tracking

-- Add individual_slot_id column
ALTER TABLE appointments 
ADD COLUMN individual_slot_id UUID;

-- Add foreign key constraint
ALTER TABLE appointments
ADD CONSTRAINT appointments_individual_slot_id_fkey 
FOREIGN KEY (individual_slot_id) 
REFERENCES doctor_individual_slots(id) 
ON DELETE SET NULL;

-- Create index for fast lookups
CREATE INDEX IF NOT EXISTS idx_appointments_individual_slot_id 
ON appointments(individual_slot_id);

-- Add comment
COMMENT ON COLUMN appointments.individual_slot_id IS 'Links to the specific individual slot used for this appointment';

