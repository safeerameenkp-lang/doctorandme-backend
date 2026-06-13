ALTER TABLE sales_schema.prescriptions ADD COLUMN latest_sale_id UUID;
CREATE INDEX idx_prescriptions_latest_sale_id ON sales_schema.prescriptions(latest_sale_id);
