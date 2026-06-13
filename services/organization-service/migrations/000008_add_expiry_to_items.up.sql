ALTER TABLE sales_schema.sale_items
ADD COLUMN IF NOT EXISTS expiry_date TIMESTAMP;
