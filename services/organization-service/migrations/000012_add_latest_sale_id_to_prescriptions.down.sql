DROP INDEX IF EXISTS sales_schema.idx_prescriptions_latest_sale_id;
ALTER TABLE sales_schema.prescriptions DROP COLUMN IF EXISTS latest_sale_id;
