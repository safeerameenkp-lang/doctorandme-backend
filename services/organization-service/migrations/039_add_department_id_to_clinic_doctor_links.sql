-- Migration 039: Add department_id to clinic_doctor_links
-- This allows linking a doctor to a specific department when they are assigned to a clinic

ALTER TABLE clinic_doctor_links
ADD COLUMN department_id UUID REFERENCES departments(id) ON DELETE SET NULL;

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_clinic_doctor_links_department_id ON clinic_doctor_links(department_id);

-- Add comment
COMMENT ON COLUMN clinic_doctor_links.department_id IS 'Specific department assignment for the doctor within this clinic';
