-- Add new fields to appointments table

-- Add department_id column
ALTER TABLE appointments ADD COLUMN department_id UUID;

-- Add appointment_date column (separate from appointment_time for easier date filtering)
ALTER TABLE appointments ADD COLUMN appointment_date DATE;

-- Add reason column
ALTER TABLE appointments ADD COLUMN reason TEXT;

-- Add notes column
ALTER TABLE appointments ADD COLUMN notes TEXT;

-- Create index for department_id
CREATE INDEX idx_appointments_department_id ON appointments(department_id);

-- Create index for appointment_date
CREATE INDEX idx_appointments_appointment_date ON appointments(appointment_date);

-- Backfill appointment_date from appointment_time for existing records
UPDATE appointments SET appointment_date = DATE(appointment_time) WHERE appointment_date IS NULL;

