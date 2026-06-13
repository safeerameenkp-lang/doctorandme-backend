-- Migration 055: Add audit columns to medicines and create medicine_audit_logs table

-- 1. Add audit columns to inventory.medicines
ALTER TABLE inventory.medicines 
ADD COLUMN IF NOT EXISTS created_by UUID,
ADD COLUMN IF NOT EXISTS created_by_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS updated_by UUID,
ADD COLUMN IF NOT EXISTS updated_by_name VARCHAR(255);

-- 2. Create inventory.medicine_audit_logs table if not exists
CREATE TABLE IF NOT EXISTS inventory.medicine_audit_logs (
    id UUID PRIMARY KEY,
    pharmacy_id UUID NOT NULL,
    medicine_id UUID NOT NULL REFERENCES inventory.medicines(id) ON DELETE CASCADE,
    action_type VARCHAR(50) NOT NULL,
    changed_by UUID NOT NULL,
    changed_by_name VARCHAR(255) NOT NULL,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Create index for medicine_audit_logs
CREATE INDEX IF NOT EXISTS idx_medicine_audit_logs_pharmacy_id ON inventory.medicine_audit_logs(pharmacy_id);
CREATE INDEX IF NOT EXISTS idx_medicine_audit_logs_medicine_id ON inventory.medicine_audit_logs(medicine_id);
