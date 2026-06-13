ALTER TABLE sales_schema.sale_items ADD COLUMN IF NOT EXISTS retail_disc_perc DECIMAL(5,2) DEFAULT 0;
ALTER TABLE sales_schema.sale_items ADD COLUMN IF NOT EXISTS staff_disc_perc DECIMAL(5,2) DEFAULT 0;
ALTER TABLE sales_schema.sale_items ADD COLUMN IF NOT EXISTS special_disc_perc DECIMAL(5,2) DEFAULT 0;
ALTER TABLE sales_schema.sale_items ADD COLUMN IF NOT EXISTS max_disc_perc DECIMAL(5,2) DEFAULT 0;
