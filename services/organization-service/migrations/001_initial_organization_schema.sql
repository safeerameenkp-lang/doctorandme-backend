-- Organization Service: Initial Schema
-- This migration creates core organization, clinic, doctor, and patient tables

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Organizations (e.g., ABC Company with multiple branches)
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20),
    address TEXT,
    license_number VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Clinics (linked to organizations)
-- Note: user_id references users table (created by auth-service)
CREATE TABLE IF NOT EXISTS clinics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID,  -- References users table (created by auth-service)
    clinic_code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20),
    address TEXT,
    license_number VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- External Services (Labs/Pharmacies)
CREATE TABLE IF NOT EXISTS external_services (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    service_type VARCHAR(20) NOT NULL,
    email VARCHAR(255), 
    phone VARCHAR(20),
    address TEXT,
    license_number VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Clinic-Service link
CREATE TABLE IF NOT EXISTS clinic_service_links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE,
    service_id UUID REFERENCES external_services(id) ON DELETE CASCADE,
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(clinic_id, service_id)
);

-- Doctors (can be main doctors or regular clinic doctors)
-- Note: user_id references users table (created by auth-service)
CREATE TABLE IF NOT EXISTS doctors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID,  -- References users table (created by auth-service)
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE,
    doctor_code VARCHAR(20),
    specialization VARCHAR(100),
    license_number VARCHAR(100),
    consultation_fee DECIMAL(10,2),
    follow_up_fee DECIMAL(10,2),
    follow_up_days INTEGER DEFAULT 7,
    is_main_doctor BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Clinic Doctor Links (links main doctors to clinics)
CREATE TABLE IF NOT EXISTS clinic_doctor_links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE,
    doctor_id UUID REFERENCES doctors(id) ON DELETE CASCADE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(clinic_id, doctor_id)
);

-- Doctor schedules
CREATE TABLE IF NOT EXISTS doctor_schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    doctor_id UUID REFERENCES doctors(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    slot_duration_minutes INTEGER DEFAULT 12,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Patients (linked to users)
-- Note: user_id references users table (created by auth-service)
CREATE TABLE IF NOT EXISTS patients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID,  -- References users table (created by auth-service)
    medical_history TEXT,
    allergies TEXT,
    blood_group VARCHAR(10),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Patient - Clinic assignment (multi-clinic support)
CREATE TABLE IF NOT EXISTS patient_clinics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_id UUID REFERENCES patients(id) ON DELETE CASCADE,
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(patient_id, clinic_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_clinics_organization_id ON clinics(organization_id);
CREATE INDEX IF NOT EXISTS idx_clinics_user_id ON clinics(user_id);
CREATE INDEX IF NOT EXISTS idx_clinic_service_links_clinic_id ON clinic_service_links(clinic_id);
CREATE INDEX IF NOT EXISTS idx_clinic_service_links_service_id ON clinic_service_links(service_id);
CREATE INDEX IF NOT EXISTS idx_doctors_user_id ON doctors(user_id);
CREATE INDEX IF NOT EXISTS idx_doctors_clinic_id ON doctors(clinic_id);
CREATE INDEX IF NOT EXISTS idx_doctors_doctor_code ON doctors(doctor_code);
CREATE INDEX IF NOT EXISTS idx_doctors_is_main_doctor ON doctors(is_main_doctor);
CREATE INDEX IF NOT EXISTS idx_clinic_doctor_links_clinic_id ON clinic_doctor_links(clinic_id);
CREATE INDEX IF NOT EXISTS idx_clinic_doctor_links_doctor_id ON clinic_doctor_links(doctor_id);
CREATE INDEX IF NOT EXISTS idx_doctor_schedules_doctor_id ON doctor_schedules(doctor_id);
CREATE INDEX IF NOT EXISTS idx_doctor_schedules_day_of_week ON doctor_schedules(day_of_week);
CREATE INDEX IF NOT EXISTS idx_patients_user_id ON patients(user_id);
CREATE INDEX IF NOT EXISTS idx_patients_is_active ON patients(is_active);
CREATE INDEX IF NOT EXISTS idx_patient_clinics_patient_id ON patient_clinics(patient_id);
CREATE INDEX IF NOT EXISTS idx_patient_clinics_clinic_id ON patient_clinics(clinic_id);
CREATE INDEX IF NOT EXISTS idx_patient_clinics_is_primary ON patient_clinics(is_primary);

