ALTER TABLE sales_schema.sales 
DROP COLUMN IF EXISTS total_discount,
DROP COLUMN IF EXISTS total_tax;
