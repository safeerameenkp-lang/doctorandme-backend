-- Migration 053: Migrate Inventory Tenant to Pharmacy
-- Renames tenant_id to pharmacy_id in the inventory schema tables and updates indexes/constraints.

-- 1. inventory.medicines
ALTER TABLE inventory.medicines RENAME COLUMN tenant_id TO pharmacy_id;

DROP INDEX IF EXISTS inventory.idx_medicines_tenant_id;
CREATE INDEX IF NOT EXISTS idx_medicines_pharmacy_id ON inventory.medicines(pharmacy_id);

DROP INDEX IF EXISTS inventory.idx_medicines_tenant_name;
CREATE INDEX IF NOT EXISTS idx_medicines_pharmacy_name ON inventory.medicines(pharmacy_id, name);

DROP INDEX IF EXISTS inventory.idx_medicines_tenant_category;
CREATE INDEX IF NOT EXISTS idx_medicines_pharmacy_category ON inventory.medicines(pharmacy_id, category);

DROP INDEX IF EXISTS inventory.idx_medicines_tenant_active;
CREATE INDEX IF NOT EXISTS idx_medicines_pharmacy_active ON inventory.medicines(pharmacy_id, is_active);

DROP INDEX IF EXISTS inventory.idx_medicines_unique_case_insensitive;
CREATE UNIQUE INDEX idx_medicines_unique_case_insensitive ON inventory.medicines (
    pharmacy_id, 
    LOWER(name), 
    LOWER(COALESCE(brand_name, '')), 
    LOWER(dosage_form), 
    unit_type,
    COALESCE(supplier_id::text, '')
);

-- 2. inventory.purchases
ALTER TABLE inventory.purchases DROP CONSTRAINT IF EXISTS fk_purchases_tenant;
ALTER TABLE inventory.purchases RENAME COLUMN tenant_id TO pharmacy_id;
ALTER TABLE inventory.purchases ADD CONSTRAINT fk_purchases_pharmacy FOREIGN KEY (pharmacy_id) REFERENCES public.pharmacies(id) ON DELETE CASCADE;

DROP INDEX IF EXISTS inventory.idx_purchases_tenant;
CREATE INDEX IF NOT EXISTS idx_purchases_pharmacy ON inventory.purchases(pharmacy_id);

-- 3. inventory.purchase_items
ALTER TABLE inventory.purchase_items DROP CONSTRAINT IF EXISTS fk_items_tenant;
ALTER TABLE inventory.purchase_items RENAME COLUMN tenant_id TO pharmacy_id;
ALTER TABLE inventory.purchase_items ADD CONSTRAINT fk_items_pharmacy FOREIGN KEY (pharmacy_id) REFERENCES public.pharmacies(id) ON DELETE CASCADE;

-- 4. inventory.batches
ALTER TABLE inventory.batches DROP CONSTRAINT IF EXISTS fk_batches_tenant;
ALTER TABLE inventory.batches DROP CONSTRAINT IF EXISTS uq_batches_medicine_batch;
ALTER TABLE inventory.batches RENAME COLUMN tenant_id TO pharmacy_id;
ALTER TABLE inventory.batches ADD CONSTRAINT uq_batches_medicine_batch UNIQUE (pharmacy_id, medicine_id, batch_no);
ALTER TABLE inventory.batches ADD CONSTRAINT fk_batches_pharmacy FOREIGN KEY (pharmacy_id) REFERENCES public.pharmacies(id) ON DELETE CASCADE;

DROP INDEX IF EXISTS inventory.idx_batches_tenant_medicine;
CREATE INDEX IF NOT EXISTS idx_batches_pharmacy_medicine ON inventory.batches(pharmacy_id, medicine_id);

-- 5. inventory.stock_ledger
ALTER TABLE inventory.stock_ledger DROP CONSTRAINT IF EXISTS fk_ledger_tenant;
ALTER TABLE inventory.stock_ledger RENAME COLUMN tenant_id TO pharmacy_id;
ALTER TABLE inventory.stock_ledger ADD CONSTRAINT fk_ledger_pharmacy FOREIGN KEY (pharmacy_id) REFERENCES public.pharmacies(id) ON DELETE CASCADE;

DROP INDEX IF EXISTS inventory.idx_ledger_tenant_medicine;
CREATE INDEX IF NOT EXISTS idx_ledger_pharmacy_medicine ON inventory.stock_ledger(pharmacy_id, medicine_id);

DROP INDEX IF EXISTS inventory.idx_ledger_tenant_batch;
CREATE INDEX IF NOT EXISTS idx_ledger_pharmacy_batch ON inventory.stock_ledger(pharmacy_id, batch_id);

-- 6. inventory.stock_outs
ALTER TABLE inventory.stock_outs RENAME COLUMN tenant_id TO pharmacy_id;

DROP INDEX IF EXISTS inventory.idx_stock_outs_tenant;
CREATE INDEX IF NOT EXISTS idx_stock_outs_pharmacy ON inventory.stock_outs(pharmacy_id);

-- 7. inventory.stock_out_items
ALTER TABLE inventory.stock_out_items RENAME COLUMN tenant_id TO pharmacy_id;

-- 8. inventory.stock_out_audit_logs
ALTER TABLE inventory.stock_out_audit_logs RENAME COLUMN tenant_id TO pharmacy_id;

-- 9. inventory.reservations
ALTER TABLE inventory.reservations RENAME COLUMN tenant_id TO pharmacy_id;

DROP INDEX IF EXISTS inventory.idx_reservations_batch;
CREATE INDEX IF NOT EXISTS idx_reservations_batch ON inventory.reservations(batch_id, pharmacy_id);

-- 10. inventory.medicine_audit_logs (Conditional if exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'inventory' AND table_name = 'medicine_audit_logs') THEN
        ALTER TABLE inventory.medicine_audit_logs RENAME COLUMN tenant_id TO pharmacy_id;
        DROP INDEX IF EXISTS inventory.idx_medicine_audit_logs_tenant_id;
        CREATE INDEX IF NOT EXISTS idx_medicine_audit_logs_pharmacy_id ON inventory.medicine_audit_logs(pharmacy_id);
    END IF;
END $$;
