-- Add logo column to clinics table
ALTER TABLE clinics ADD COLUMN IF NOT EXISTS logo VARCHAR(500);
