-- Create Batch Audit Logs table
CREATE TABLE IF NOT EXISTS inventory.batch_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    batch_id UUID NOT NULL,
    action_type VARCHAR(50) NOT NULL, -- CREATE, UPDATE, STOCK_IN, STOCK_ADJUSTMENT
    changed_by UUID,
    changed_by_name VARCHAR(255),
    notes TEXT,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_batch_audit_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_schema.tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_batch_audit_batch FOREIGN KEY (batch_id) REFERENCES inventory.batches(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_batch_audit_batch ON inventory.batch_audit_logs(batch_id);
CREATE INDEX IF NOT EXISTS idx_batch_audit_tenant ON inventory.batch_audit_logs(tenant_id);
