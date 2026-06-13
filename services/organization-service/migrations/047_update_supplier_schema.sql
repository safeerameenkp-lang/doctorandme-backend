-- Migration: Update Supplier Schema (Add audit fields and audit logs)

-- 1. Add audit columns to suppliers table
ALTER TABLE supplier_schema.suppliers 
ADD COLUMN IF NOT EXISTS created_by UUID,
ADD COLUMN IF NOT EXISTS updated_by UUID;

-- 2. Create supplier audit logs table
CREATE TABLE IF NOT EXISTS supplier_schema.supplier_audit_logs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    supplier_id UUID NOT NULL,
    action_type VARCHAR(20) NOT NULL, -- CREATE / UPDATE / STATUS_CHANGE
    changed_by UUID NOT NULL,
    changed_by_name VARCHAR(255) DEFAULT 'System',
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_supplier
        FOREIGN KEY(supplier_id) 
        REFERENCES supplier_schema.suppliers(id)
        ON DELETE CASCADE
);

-- 3. Indexes for efficient query performance
CREATE INDEX IF NOT EXISTS idx_supplier_audit_logs_supplier_id ON supplier_schema.supplier_audit_logs(supplier_id);
CREATE INDEX IF NOT EXISTS idx_supplier_audit_logs_tenant_id ON supplier_schema.supplier_audit_logs(tenant_id);
