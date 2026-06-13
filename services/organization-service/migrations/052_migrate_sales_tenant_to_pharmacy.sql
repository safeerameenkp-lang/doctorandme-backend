-- Migration 052: Migrate Sales & Prescriptions Tenant to Pharmacy
-- Renames tenant_id to pharmacy_id in the sales schema tables and updates indexes/constraints.

-- 1. sales_schema.patients
ALTER TABLE sales_schema.patients RENAME COLUMN tenant_id TO pharmacy_id;
ALTER TABLE sales_schema.patients DROP CONSTRAINT IF EXISTS patients_tenant_id_phone_name_key;
ALTER TABLE sales_schema.patients ADD CONSTRAINT uq_patients_pharmacy_phone_name UNIQUE (pharmacy_id, phone, name);

-- 2. sales_schema.prescriptions
ALTER TABLE sales_schema.prescriptions RENAME COLUMN tenant_id TO pharmacy_id;
DROP INDEX IF EXISTS sales_schema.idx_temp_rx_tenant;
CREATE INDEX IF NOT EXISTS idx_temp_rx_pharmacy ON sales_schema.prescriptions(pharmacy_id);

-- 3. sales_schema.sales
ALTER TABLE sales_schema.sales RENAME COLUMN tenant_id TO pharmacy_id;
DROP INDEX IF EXISTS sales_schema.idx_sales_tenant;
CREATE INDEX IF NOT EXISTS idx_sales_pharmacy ON sales_schema.sales(pharmacy_id);

-- 4. sales_schema.sales_returns
ALTER TABLE sales_schema.sales_returns RENAME COLUMN tenant_id TO pharmacy_id;
DROP INDEX IF EXISTS sales_schema.idx_sales_returns_tenant_id;
CREATE INDEX IF NOT EXISTS idx_sales_returns_pharmacy_id ON sales_schema.sales_returns(pharmacy_id);
