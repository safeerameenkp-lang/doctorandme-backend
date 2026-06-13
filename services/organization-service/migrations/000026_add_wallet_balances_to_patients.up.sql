ALTER TABLE sales_schema.patients
ADD COLUMN IF NOT EXISTS due_amount numeric(10,2) NOT NULL DEFAULT 0.00,
ADD COLUMN IF NOT EXISTS credit_amount numeric(10,2) NOT NULL DEFAULT 0.00;
