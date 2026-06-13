ALTER TABLE sales_schema.sales
ADD COLUMN generated_credit NUMERIC(10, 2) DEFAULT 0.00,
ADD COLUMN generated_due NUMERIC(10, 2) DEFAULT 0.00;
