ALTER TABLE sales_schema.sales
ADD COLUMN applied_credit NUMERIC(10, 2) DEFAULT 0.00,
ADD COLUMN applied_due NUMERIC(10, 2) DEFAULT 0.00;
