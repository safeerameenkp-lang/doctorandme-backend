CREATE TABLE IF NOT EXISTS sales_schema.prescriptions (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id UUID NOT NULL,
    patient_name VARCHAR(150) NOT NULL,
    patient_phone VARCHAR(20),
    doctor_name VARCHAR(150),
    date TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING'
);

CREATE TABLE IF NOT EXISTS sales_schema.prescription_items (
    id UUID PRIMARY KEY,
    prescription_id VARCHAR(50) NOT NULL REFERENCES sales_schema.prescriptions(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    medicine_name VARCHAR(150) NOT NULL,
    quantity INTEGER NOT NULL,
    instructions TEXT
);

CREATE INDEX IF NOT EXISTS idx_temp_rx_tenant ON sales_schema.prescriptions(tenant_id);
