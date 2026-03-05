-- Migration 009: Departments Management
-- Add departments table for clinic-specific departments
-- This allows clinics to organize doctors by departments (Orthology, Cardiology, etc.)

-- Create departments table
CREATE TABLE IF NOT EXISTS departments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(clinic_id, name) -- Each clinic can have unique department names
);

-- Add department_id to doctors table
ALTER TABLE doctors
ADD COLUMN department_id UUID REFERENCES departments(id) ON DELETE SET NULL;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_departments_clinic_id ON departments(clinic_id);
CREATE INDEX IF NOT EXISTS idx_departments_active ON departments(is_active);
CREATE INDEX IF NOT EXISTS idx_doctors_department_id ON doctors(department_id);

-- Create trigger for updated_at
DROP TRIGGER IF EXISTS update_departments_updated_at ON departments;
CREATE TRIGGER update_departments_updated_at 
    BEFORE UPDATE ON departments
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE departments IS 'Clinic-specific departments for organizing doctors';
COMMENT ON COLUMN departments.name IS 'Department name (e.g., Orthology, Cardiology, Pediatrics)';
COMMENT ON COLUMN departments.description IS 'Optional description of the department';
COMMENT ON COLUMN doctors.department_id IS 'Department assignment for the doctor';

-- Insert some default departments for existing clinics
INSERT INTO departments (clinic_id, name, description)
SELECT 
    c.id as clinic_id,
    'General Medicine' as name,
    'General medical consultations and primary care' as description
FROM clinics c
WHERE NOT EXISTS (
    SELECT 1 FROM departments d WHERE d.clinic_id = c.id AND d.name = 'General Medicine'
);

INSERT INTO departments (clinic_id, name, description)
SELECT 
    c.id as clinic_id,
    'Orthology' as name,
    'Orthopedic and bone-related treatments' as description
FROM clinics c
WHERE NOT EXISTS (
    SELECT 1 FROM departments d WHERE d.clinic_id = c.id AND d.name = 'Orthology'
);

INSERT INTO departments (clinic_id, name, description)
SELECT 
    c.id as clinic_id,
    'Cardiology' as name,
    'Heart and cardiovascular system treatments' as description
FROM clinics c
WHERE NOT EXISTS (
    SELECT 1 FROM departments d WHERE d.clinic_id = c.id AND d.name = 'Cardiology'
);
