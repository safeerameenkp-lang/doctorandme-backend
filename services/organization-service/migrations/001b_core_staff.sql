-- 001b_core_staff.sql
-- Staff Management Table (for clinic staff roles)
-- Links users to clinics with specialized roles

CREATE TABLE IF NOT EXISTS staff (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL, -- References users table (created by auth-service)
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE,
    staff_type VARCHAR(50) NOT NULL, -- receptionist, doctor, lab_tech, pharmacist, billing
    permissions JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, clinic_id)
);

-- Index for better performance
CREATE INDEX IF NOT EXISTS idx_staff_user_id ON staff(user_id);
CREATE INDEX IF NOT EXISTS idx_staff_clinic_id ON staff(clinic_id);
CREATE INDEX IF NOT EXISTS idx_staff_staff_type ON staff(staff_type);
