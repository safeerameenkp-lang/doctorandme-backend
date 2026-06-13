ALTER TABLE sales_schema.patients 
DROP CONSTRAINT IF EXISTS patients_tenant_id_phone_name_key;

-- We ignore duplicates when down migrating, just applying safely
ALTER TABLE sales_schema.patients 
ADD CONSTRAINT patients_tenant_id_phone_key UNIQUE (tenant_id, phone);
