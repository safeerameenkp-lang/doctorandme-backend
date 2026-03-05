-- Migration 017: Clinic-Specific Patients (No Global Users)
-- Creates isolated patient records for each clinic

-- Create clinic_patients table (replaces patient_clinics link)
CREATE TABLE IF NOT EXISTS clinic_patients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE NOT NULL,
    
    -- Personal Information (clinic-specific)
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(100),
    date_of_birth DATE,
    gender VARCHAR(20),
    
    -- Patient-specific fields
    mo_id VARCHAR(50),  -- Clinic's internal patient ID
    medical_history TEXT,
    allergies TEXT,
    blood_group VARCHAR(10),
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Optional: Link to global patient if they opt-in for multi-clinic
    global_patient_id UUID REFERENCES patients(id) ON DELETE SET NULL,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint: phone + clinic_id (same person can have different records at different clinics)
    CONSTRAINT unique_phone_per_clinic UNIQUE (clinic_id, phone),
    -- Unique constraint: mo_id per clinic
    CONSTRAINT unique_mo_id_per_clinic UNIQUE (clinic_id, mo_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_clinic_patients_clinic_id ON clinic_patients(clinic_id);
CREATE INDEX IF NOT EXISTS idx_clinic_patients_phone ON clinic_patients(phone);
CREATE INDEX IF NOT EXISTS idx_clinic_patients_mo_id ON clinic_patients(mo_id);
CREATE INDEX IF NOT EXISTS idx_clinic_patients_global_patient ON clinic_patients(global_patient_id);
CREATE INDEX IF NOT EXISTS idx_clinic_patients_active ON clinic_patients(clinic_id, is_active);

-- Composite index for common queries
CREATE INDEX IF NOT EXISTS idx_clinic_patients_search ON clinic_patients(clinic_id, phone, mo_id);

-- Trigger for updated_at
CREATE TRIGGER update_clinic_patients_updated_at 
    BEFORE UPDATE ON clinic_patients
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments
COMMENT ON TABLE clinic_patients IS 'Clinic-specific patient records with optional global patient linking';
COMMENT ON COLUMN clinic_patients.global_patient_id IS 'Optional link to global patient record if patient opts-in for multi-clinic access';
COMMENT ON COLUMN clinic_patients.mo_id IS 'Clinic internal patient ID (unique per clinic)';
COMMENT ON CONSTRAINT unique_phone_per_clinic ON clinic_patients IS 'Same phone can exist in different clinics';

