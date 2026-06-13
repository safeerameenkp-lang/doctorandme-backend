-- Migration 058: Create batch_audit_logs table

CREATE TABLE IF NOT EXISTS inventory.batch_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pharmacy_id UUID NOT NULL REFERENCES public.pharmacies(id) ON DELETE CASCADE,
    batch_id UUID NOT NULL REFERENCES inventory.batches(id) ON DELETE CASCADE,
    action_type VARCHAR(50) NOT NULL, -- CREATE, UPDATE, STOCK_IN, STOCK_ADJUSTMENT
    changed_by UUID,
    changed_by_name VARCHAR(255),
    notes TEXT,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_batch_audit_batch ON inventory.batch_audit_logs(batch_id);
CREATE INDEX IF NOT EXISTS idx_batch_audit_pharmacy ON inventory.batch_audit_logs(pharmacy_id);
