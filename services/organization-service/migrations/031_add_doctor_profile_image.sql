-- Add profile_image column to doctors table
ALTER TABLE doctors ADD COLUMN IF NOT EXISTS profile_image VARCHAR(500);
