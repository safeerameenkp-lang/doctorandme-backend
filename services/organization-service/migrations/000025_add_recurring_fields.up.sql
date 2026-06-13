-- Add recurring fields to sales table
ALTER TABLE sales_schema.sales 
ADD COLUMN is_recurring BOOLEAN DEFAULT FALSE,
ADD COLUMN days_supply INTEGER DEFAULT 0,
ADD COLUMN next_refill_date TIMESTAMP;

-- Add recurring status to patients table
ALTER TABLE sales_schema.patients 
ADD COLUMN is_recurring BOOLEAN DEFAULT FALSE;
