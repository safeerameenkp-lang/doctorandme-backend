-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    date_of_birth DATE,
    gender VARCHAR(20),
    is_active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Roles (system + custom)
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,   -- e.g. super_admin, clinic_admin, doctor
    permissions JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Organizations (e.g., ABC Company with multiple branches)
CREATE TABLE organizations (
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
CREATE TABLE clinics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
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
CREATE TABLE external_services (
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

-- User Roles (per org/clinic/service)
CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE,
    service_id UUID REFERENCES external_services(id) ON DELETE CASCADE,
    is_active BOOLEAN DEFAULT TRUE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id, organization_id, clinic_id, service_id)
);

-- Clinic-Service link
CREATE TABLE clinic_service_links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE,
    service_id UUID REFERENCES external_services(id) ON DELETE CASCADE,
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(clinic_id, service_id)
);

-- Refresh Tokens (multi-device support)
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL, -- store hashed if extra security needed
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP,
    UNIQUE(user_id, token)
);

-- Doctors (can be main doctors or regular clinic doctors)
CREATE TABLE doctors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
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
CREATE TABLE clinic_doctor_links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE,
    doctor_id UUID REFERENCES doctors(id) ON DELETE CASCADE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(clinic_id, doctor_id)
);

-- Doctor schedules
CREATE TABLE doctor_schedules (
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
CREATE TABLE patients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    medical_history TEXT,
    allergies TEXT,
    blood_group VARCHAR(10),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Patient - Clinic assignment (multi-clinic support)
CREATE TABLE patient_clinics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_id UUID REFERENCES patients(id) ON DELETE CASCADE,
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(patient_id, clinic_id)
);

-- Appointments
CREATE TABLE appointments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_id UUID REFERENCES patients(id) ON DELETE CASCADE,
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE,
    doctor_id UUID REFERENCES doctors(id) ON DELETE CASCADE,
    booking_number VARCHAR(50) UNIQUE NOT NULL,
    appointment_time TIMESTAMP NOT NULL,
    duration_minutes INTEGER DEFAULT 12,
    consultation_type VARCHAR(20) DEFAULT 'new',
    status VARCHAR(20) DEFAULT 'booked',
    fee_amount DECIMAL(10,2),
    payment_status VARCHAR(20) DEFAULT 'pending',
    payment_mode VARCHAR(20),
    is_priority BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Patient check-ins
CREATE TABLE patient_checkins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    appointment_id UUID REFERENCES appointments(id) ON DELETE CASCADE,
    checkin_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    vitals_recorded BOOLEAN DEFAULT FALSE,
    payment_collected BOOLEAN DEFAULT FALSE,
    checked_in_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Patient vitals
CREATE TABLE patient_vitals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    appointment_id UUID REFERENCES appointments(id) ON DELETE CASCADE,
    systolic_bp INTEGER,
    diastolic_bp INTEGER,
    temperature DECIMAL(4,1),
    pulse_rate INTEGER,
    height_cm INTEGER,
    weight_kg DECIMAL(5,2),
    recorded_by UUID REFERENCES users(id),
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default roles with hierarchical permissions
INSERT INTO roles (name, permissions) VALUES 
('super_admin', '{"organizations": ["create", "read", "update", "delete"], "clinics": ["create", "read", "update", "delete"], "services": ["create", "read", "update", "delete"], "users": ["create", "read", "update", "delete"], "roles": ["create", "read", "update", "delete"]}'),
('organization_admin', '{"organizations": ["read", "update"], "clinics": ["create", "read", "update", "delete"], "users": ["create", "read", "update", "delete"], "roles": ["read"]}'),
('clinic_admin', '{"clinics": ["read", "update"], "users": ["create", "read", "update", "delete"], "roles": ["read"], "staff": ["create", "read", "update", "delete"]}'),
('doctor', '{"patients": ["read", "update"], "appointments": ["read", "create", "update"], "prescriptions": ["read", "create", "update"]}'),
('receptionist', '{"patients": ["read", "create", "update"], "appointments": ["read", "create", "update"], "billing": ["read", "create"]}'),
('pharmacist', '{"prescriptions": ["read", "update"], "medications": ["read", "create", "update"], "inventory": ["read", "update"]}'),
('lab_technician', '{"lab_orders": ["read", "create", "update"], "lab_results": ["read", "create", "update"], "reports": ["read", "create"]}'),
('billing_staff', '{"billing": ["read", "create", "update"], "payments": ["read", "create", "update"], "invoices": ["read", "create", "update"]}'),
('patient', '{"profile": ["read", "update"], "appointments": ["read", "create"], "prescriptions": ["read"], "lab_results": ["read"]}');

-- Create indexes for better performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_organization_id ON user_roles(organization_id);
CREATE INDEX idx_user_roles_clinic_id ON user_roles(clinic_id);
CREATE INDEX idx_user_roles_service_id ON user_roles(service_id);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_clinics_organization_id ON clinics(organization_id);
CREATE INDEX idx_clinics_user_id ON clinics(user_id);
CREATE INDEX idx_clinic_service_links_clinic_id ON clinic_service_links(clinic_id);
CREATE INDEX idx_clinic_service_links_service_id ON clinic_service_links(service_id);
CREATE INDEX idx_doctors_user_id ON doctors(user_id);
CREATE INDEX idx_doctors_clinic_id ON doctors(clinic_id);
CREATE INDEX idx_doctors_doctor_code ON doctors(doctor_code);
CREATE INDEX idx_doctors_is_main_doctor ON doctors(is_main_doctor);
CREATE INDEX idx_clinic_doctor_links_clinic_id ON clinic_doctor_links(clinic_id);
CREATE INDEX idx_clinic_doctor_links_doctor_id ON clinic_doctor_links(doctor_id);
CREATE INDEX idx_doctor_schedules_doctor_id ON doctor_schedules(doctor_id);
CREATE INDEX idx_doctor_schedules_day_of_week ON doctor_schedules(day_of_week);
CREATE INDEX idx_patients_user_id ON patients(user_id);
CREATE INDEX idx_patients_is_active ON patients(is_active);
CREATE INDEX idx_patient_clinics_patient_id ON patient_clinics(patient_id);
CREATE INDEX idx_patient_clinics_clinic_id ON patient_clinics(clinic_id);
CREATE INDEX idx_patient_clinics_is_primary ON patient_clinics(is_primary);
CREATE INDEX idx_appointments_patient_id ON appointments(patient_id);
CREATE INDEX idx_appointments_clinic_id ON appointments(clinic_id);
CREATE INDEX idx_appointments_doctor_id ON appointments(doctor_id);
CREATE INDEX idx_appointments_booking_number ON appointments(booking_number);
CREATE INDEX idx_appointments_appointment_time ON appointments(appointment_time);
CREATE INDEX idx_appointments_status ON appointments(status);
CREATE INDEX idx_appointments_payment_status ON appointments(payment_status);
CREATE INDEX idx_patient_checkins_appointment_id ON patient_checkins(appointment_id);
CREATE INDEX idx_patient_checkins_checkin_time ON patient_checkins(checkin_time);
CREATE INDEX idx_patient_vitals_appointment_id ON patient_vitals(appointment_id);
CREATE INDEX idx_patient_vitals_recorded_at ON patient_vitals(recorded_at);
