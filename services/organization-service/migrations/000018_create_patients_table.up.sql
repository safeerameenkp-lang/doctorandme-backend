CREATE TABLE IF NOT EXISTS sales_schema.patients (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    gender TEXT,
    age INT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, phone)
);

CREATE INDEX idx_patients_phone ON sales_schema.patients(phone);
