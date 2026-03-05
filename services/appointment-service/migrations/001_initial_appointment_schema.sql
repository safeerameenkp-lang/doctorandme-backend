-- Appointment Service: Initial Schema
-- This migration creates appointment, check-in, and vitals tables

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Appointments
-- Note: References patients, clinics, doctors which are created by organization-service
CREATE TABLE IF NOT EXISTS appointments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_id UUID,  -- References patients table (created by org service)
    clinic_id UUID,   -- References clinics table (created by org service)
    doctor_id UUID,   -- References doctors table (created by org service)
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
-- Note: checked_in_by references users table (created by auth-service)
CREATE TABLE IF NOT EXISTS patient_checkins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    appointment_id UUID REFERENCES appointments(id) ON DELETE CASCADE,
    checkin_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    vitals_recorded BOOLEAN DEFAULT FALSE,
    payment_collected BOOLEAN DEFAULT FALSE,
    checked_in_by UUID,  -- References users table (created by auth-service)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Patient vitals
-- Note: recorded_by references users table (created by auth-service)
CREATE TABLE IF NOT EXISTS patient_vitals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    appointment_id UUID REFERENCES appointments(id) ON DELETE CASCADE,
    systolic_bp INTEGER,
    diastolic_bp INTEGER,
    temperature DECIMAL(4,1),
    pulse_rate INTEGER,
    height_cm INTEGER,
    weight_kg DECIMAL(5,2),
    recorded_by UUID,  -- References users table (created by auth-service)
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_appointments_patient_id ON appointments(patient_id);
CREATE INDEX IF NOT EXISTS idx_appointments_clinic_id ON appointments(clinic_id);
CREATE INDEX IF NOT EXISTS idx_appointments_doctor_id ON appointments(doctor_id);
CREATE INDEX IF NOT EXISTS idx_appointments_booking_number ON appointments(booking_number);
CREATE INDEX IF NOT EXISTS idx_appointments_appointment_time ON appointments(appointment_time);
CREATE INDEX IF NOT EXISTS idx_appointments_status ON appointments(status);
CREATE INDEX IF NOT EXISTS idx_appointments_payment_status ON appointments(payment_status);
CREATE INDEX IF NOT EXISTS idx_patient_checkins_appointment_id ON patient_checkins(appointment_id);
CREATE INDEX IF NOT EXISTS idx_patient_checkins_checkin_time ON patient_checkins(checkin_time);
CREATE INDEX IF NOT EXISTS idx_patient_vitals_appointment_id ON patient_vitals(appointment_id);
CREATE INDEX IF NOT EXISTS idx_patient_vitals_recorded_at ON patient_vitals(recorded_at);

