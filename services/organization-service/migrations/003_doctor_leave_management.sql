-- Create doctor leave management table
CREATE TABLE IF NOT EXISTS doctor_leaves (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    doctor_id UUID REFERENCES doctors(id) ON DELETE CASCADE NOT NULL,
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE NOT NULL,
    leave_type VARCHAR(50) NOT NULL, -- sick_leave, vacation, emergency, other
    from_date DATE NOT NULL,
    to_date DATE NOT NULL,
    total_days INTEGER NOT NULL,
    reason TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending, approved, rejected, cancelled
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reviewed_at TIMESTAMP,
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    review_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_date_range CHECK (to_date >= from_date),
    CONSTRAINT valid_total_days CHECK (total_days > 0)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_doctor_leaves_doctor_id ON doctor_leaves(doctor_id);
CREATE INDEX IF NOT EXISTS idx_doctor_leaves_clinic_id ON doctor_leaves(clinic_id);
CREATE INDEX IF NOT EXISTS idx_doctor_leaves_status ON doctor_leaves(status);
CREATE INDEX IF NOT EXISTS idx_doctor_leaves_from_date ON doctor_leaves(from_date);
CREATE INDEX IF NOT EXISTS idx_doctor_leaves_to_date ON doctor_leaves(to_date);
CREATE INDEX IF NOT EXISTS idx_doctor_leaves_applied_at ON doctor_leaves(applied_at);

-- Create trigger for updated_at
DROP TRIGGER IF EXISTS update_doctor_leaves_updated_at ON doctor_leaves;
CREATE TRIGGER update_doctor_leaves_updated_at BEFORE UPDATE ON doctor_leaves
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE doctor_leaves IS 'Stores doctor leave applications and approvals';
COMMENT ON COLUMN doctor_leaves.leave_type IS 'Type of leave: sick_leave, vacation, emergency, other';
COMMENT ON COLUMN doctor_leaves.status IS 'Leave status: pending, approved, rejected, cancelled';
COMMENT ON COLUMN doctor_leaves.total_days IS 'Total number of leave days (calculated from from_date to to_date)';

