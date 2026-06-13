-- Migration 059: Add customer_age and customer_gender to sales
ALTER TABLE sales_schema.sales ADD COLUMN IF NOT EXISTS customer_age INTEGER;
ALTER TABLE sales_schema.sales ADD COLUMN IF NOT EXISTS customer_gender TEXT;
