ALTER TABLE sales_schema.sales ADD COLUMN IF NOT EXISTS patient_id UUID REFERENCES sales_schema.patients(id);
ALTER TABLE sales_schema.sales ADD COLUMN IF NOT EXISTS customer_name TEXT;
ALTER TABLE sales_schema.sales ADD COLUMN IF NOT EXISTS customer_phone TEXT;
