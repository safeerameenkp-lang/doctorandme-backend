-- Migration 051: Migrate Suppliers Tenant to Pharmacy
-- This migration renames tenant_id to pharmacy_id in the suppliers schema tables and recreates unique constraints and indexes.

-- 1. Refactor supplier_schema.suppliers table
ALTER TABLE supplier_schema.suppliers RENAME COLUMN tenant_id TO pharmacy_id;

-- Drop index and recreate
DROP INDEX IF EXISTS supplier_schema.idx_suppliers_tenant_id;
CREATE INDEX IF NOT EXISTS idx_suppliers_pharmacy_id ON supplier_schema.suppliers(pharmacy_id);

-- Drop unique constraint and recreate
-- If postgres automatically created the constraint name under `suppliers_tenant_id_name_key`, drop it, or fall back to standard unique
ALTER TABLE supplier_schema.suppliers DROP CONSTRAINT IF EXISTS suppliers_tenant_id_name_key;
ALTER TABLE supplier_schema.suppliers DROP CONSTRAINT IF EXISTS suppliers_name_key;
ALTER TABLE supplier_schema.suppliers ADD CONSTRAINT uq_suppliers_pharmacy_name UNIQUE (pharmacy_id, name);


-- 2. Refactor supplier_schema.supplier_audit_logs table
ALTER TABLE supplier_schema.supplier_audit_logs RENAME COLUMN tenant_id TO pharmacy_id;

-- Drop index and recreate
DROP INDEX IF EXISTS supplier_schema.idx_supplier_audit_logs_tenant_id;
CREATE INDEX IF NOT EXISTS idx_supplier_audit_logs_pharmacy_id ON supplier_schema.supplier_audit_logs(pharmacy_id);
