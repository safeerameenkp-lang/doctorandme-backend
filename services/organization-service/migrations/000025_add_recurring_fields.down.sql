ALTER TABLE sales_schema.sales 
DROP COLUMN IF EXISTS is_recurring,
DROP COLUMN IF EXISTS days_supply,
DROP COLUMN IF EXISTS next_refill_date;

ALTER TABLE sales_schema.patients 
DROP COLUMN IF EXISTS is_recurring;
