-- Migration: Add extra fields to doctors table
-- Added fields: experience_years, qualification, bio

ALTER TABLE doctors
ADD COLUMN IF NOT EXISTS experience_years INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS qualification VARCHAR(255),
ADD COLUMN IF NOT EXISTS bio TEXT;

-- Add comments for documentation
COMMENT ON COLUMN doctors.experience_years IS 'Years of professional experience';
COMMENT ON COLUMN doctors.qualification IS 'Educational qualifications (e.g., MBBS, MD)';
COMMENT ON COLUMN doctors.bio IS 'Professional biography or description';
