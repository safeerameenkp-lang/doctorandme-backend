-- Create stock_outs table
CREATE TABLE IF NOT EXISTS inventory.stock_outs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    type VARCHAR(50) NOT NULL,
    reason TEXT,
    total_loss_value DECIMAL(12, 2) NOT NULL DEFAULT 0.0,
    created_by_id UUID NOT NULL,
    created_by_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create stock_out_items table
CREATE TABLE IF NOT EXISTS inventory.stock_out_items (
    id UUID PRIMARY KEY,
    stock_out_id UUID NOT NULL REFERENCES inventory.stock_outs(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    medicine_id UUID NOT NULL,
    medicine_name VARCHAR(255) NOT NULL,
    batch_id UUID NOT NULL,
    batch_no VARCHAR(100) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_cost_price DECIMAL(10, 2) NOT NULL DEFAULT 0.0,
    total_loss DECIMAL(12, 2) NOT NULL DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for performance
CREATE INDEX IF NOT EXISTS idx_stock_outs_tenant ON inventory.stock_outs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_stock_out_items_stock_out ON inventory.stock_out_items(stock_out_id);
CREATE INDEX IF NOT EXISTS idx_stock_out_items_batch ON inventory.stock_out_items(batch_id);

-- Create stock_out_audit_logs table for traceability
CREATE TABLE IF NOT EXISTS inventory.stock_out_audit_logs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    stock_out_id UUID NOT NULL REFERENCES inventory.stock_outs(id) ON DELETE CASCADE,
    action_type VARCHAR(50) NOT NULL,
    changed_by UUID NOT NULL,
    changed_by_name VARCHAR(255) NOT NULL,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index audit logs
CREATE INDEX IF NOT EXISTS idx_stock_out_audit_logs_stock_out ON inventory.stock_out_audit_logs(stock_out_id);
