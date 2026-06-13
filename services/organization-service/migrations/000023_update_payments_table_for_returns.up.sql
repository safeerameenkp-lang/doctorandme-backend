-- 000023_update_payments_table_for_returns.up.sql
ALTER TABLE sales_schema.payments 
ADD COLUMN IF NOT EXISTS transaction_type VARCHAR(20) DEFAULT 'PAYMENT',
ADD COLUMN IF NOT EXISTS return_id UUID REFERENCES sales_schema.sales_returns(id);

CREATE INDEX IF NOT EXISTS idx_payments_return_id ON sales_schema.payments(return_id);
