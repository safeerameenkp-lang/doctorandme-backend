ALTER TABLE sales_schema.sale_items
ADD COLUMN IF NOT EXISTS batch_no VARCHAR(255);
