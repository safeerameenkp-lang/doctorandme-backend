-- Migration 054: Remove country_code from suppliers
DROP INDEX IF EXISTS supplier_schema.idx_suppliers_country_code;
ALTER TABLE supplier_schema.suppliers DROP COLUMN IF EXISTS country_code;
