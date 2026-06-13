-- 012_create_stockin_audit_logs_table.sql
CREATE TABLE IF NOT EXISTS inventory.stock_in_audit_logs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    stock_in_id UUID NOT NULL,
    action_type VARCHAR(20) NOT NULL, -- CREATE / UPDATE
    changed_by UUID NOT NULL,
    changed_by_name VARCHAR(255) DEFAULT 'System',
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_stock_in
        FOREIGN KEY(stock_in_id) 
        REFERENCES inventory.purchases(id)
        ON DELETE CASCADE
);

-- Index for efficient lookups by stock_in_id
CREATE INDEX idx_stock_in_audit_logs_stock_in_id ON inventory.stock_in_audit_logs(stock_in_id);
CREATE INDEX idx_stock_in_audit_logs_tenant_id ON inventory.stock_in_audit_logs(tenant_id);
