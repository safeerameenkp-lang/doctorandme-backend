-- =====================================================
-- FOLLOW-UPS TABLE - Explicit tracking of follow-up eligibility
-- =====================================================
-- This table tracks follow-up eligibility explicitly instead of calculating on-the-fly
-- Benefits:
-- 1. Better performance (no complex queries)
-- 2. Clear status tracking (active, used, expired, renewed)
-- 3. Easy renewal management
-- 4. Historical tracking

CREATE TABLE IF NOT EXISTS follow_ups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Patient & Clinic
    clinic_patient_id UUID NOT NULL REFERENCES clinic_patients(id) ON DELETE CASCADE,
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    
    -- Doctor & Department (follow-up is per doctor+department combination)
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    
    -- Source appointment that granted this follow-up
    source_appointment_id UUID NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    
    -- Follow-up details
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'used', 'expired', 'renewed')),
    is_free BOOLEAN NOT NULL DEFAULT true,  -- First follow-up is free
    
    -- Validity period (5 days from source appointment)
    valid_from DATE NOT NULL,  -- Source appointment date
    valid_until DATE NOT NULL, -- Source appointment date + 5 days
    
    -- If used, track which appointment used it
    used_at TIMESTAMP,
    used_appointment_id UUID REFERENCES appointments(id) ON DELETE SET NULL,
    
    -- Renewal tracking
    renewed_at TIMESTAMP,
    renewed_by_appointment_id UUID REFERENCES appointments(id) ON DELETE SET NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_follow_ups_patient ON follow_ups(clinic_patient_id);
CREATE INDEX IF NOT EXISTS idx_follow_ups_clinic ON follow_ups(clinic_id);
CREATE INDEX IF NOT EXISTS idx_follow_ups_doctor_dept ON follow_ups(doctor_id, department_id);
CREATE INDEX IF NOT EXISTS idx_follow_ups_status ON follow_ups(status);
CREATE INDEX IF NOT EXISTS idx_follow_ups_validity ON follow_ups(valid_until);

-- Composite index for common query pattern (patient + doctor + dept + status)
CREATE INDEX IF NOT EXISTS idx_follow_ups_eligibility ON follow_ups(clinic_patient_id, doctor_id, department_id, status);

-- Comments
COMMENT ON TABLE follow_ups IS 'Tracks follow-up eligibility per patient-doctor-department combination';
COMMENT ON COLUMN follow_ups.status IS 'active: Available for use, used: Already consumed, expired: Past validity period, renewed: Replaced by newer follow-up';
COMMENT ON COLUMN follow_ups.is_free IS 'First follow-up per appointment is free, subsequent ones are paid';
COMMENT ON COLUMN follow_ups.valid_from IS 'Start date of follow-up eligibility (source appointment date)';
COMMENT ON COLUMN follow_ups.valid_until IS 'End date of follow-up eligibility (source appointment date + 5 days)';

