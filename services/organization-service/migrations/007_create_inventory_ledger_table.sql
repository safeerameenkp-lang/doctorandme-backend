-- Migration: Create inventory.stock_ledger table
-- Description: Audit log for all inventory movements (Purchase, Sale, Adjustments)

CREATE TABLE IF NOT EXISTS inventory.stock_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    medicine_id UUID NOT NULL,
    batch_id UUID NOT NULL,
    
    -- Transaction Categorization
    transaction_type VARCHAR(50) NOT NULL, -- PURCHASE, SALE, SALE_RETURN, PURCHASE_RETURN, ADJUSTMENT
    
    -- Movement Details
    quantity_change INTEGER NOT NULL, -- Positive for IN, Negative for OUT
    balance_after INTEGER NOT NULL,  -- Snapshot of stock AFTER the change
    
    -- Traceability
    reference_type VARCHAR(50),      -- e.g., 'PURCHASE', 'INVOICE'
    reference_id UUID,               -- ID of the header record
    
    -- Audit
    performed_by UUID,               -- ID of the user
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT fk_ledger_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_schema.tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_ledger_medicine FOREIGN KEY (medicine_id) REFERENCES inventory.medicines(id) ON DELETE RESTRICT,
    CONSTRAINT fk_ledger_batch FOREIGN KEY (batch_id) REFERENCES inventory.batches(id) ON DELETE CASCADE
);

-- Performance Indexes for Daily Auditing & Reporting
CREATE INDEX IF NOT EXISTS idx_ledger_tenant_medicine ON inventory.stock_ledger(tenant_id, medicine_id);
CREATE INDEX IF NOT EXISTS idx_ledger_tenant_batch ON inventory.stock_ledger(tenant_id, batch_id);
CREATE INDEX IF NOT EXISTS idx_ledger_reference ON inventory.stock_ledger(reference_id);
CREATE INDEX IF NOT EXISTS idx_ledger_created_at ON inventory.stock_ledger(created_at);

-- Add a comment explaining the table importance
COMMENT ON TABLE inventory.stock_ledger IS 'Immutable audit log for all stock movements to ensure inventory integrity and financial auditing.';
