CREATE SCHEMA IF NOT EXISTS supplier_schema;

CREATE TABLE IF NOT EXISTS supplier_schema.suppliers (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    supplier_type VARCHAR(100), -- Example: Wholesaler, Manufacturer, etc.
    contact_person VARCHAR(255),
    contact_number VARCHAR(50),
    website VARCHAR(255),
    email VARCHAR(255),
    address TEXT,
    state VARCHAR(100),
    country_code VARCHAR(10) NOT NULL,
    pincode VARCHAR(20),
    gst_number VARCHAR(50),
    pan_number VARCHAR(50),
    license_number VARCHAR(100),
    
    -- Bank Details
    bank_name VARCHAR(255),
    account_name VARCHAR(255),
    account_number VARCHAR(100),
    ifsc_code VARCHAR(50), -- Or SWIFT based on country
    
    -- Credit Terms
    credit_period_days INT DEFAULT 0,
    credit_limit DECIMAL(15, 2) DEFAULT 0.00,
    
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE (tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_suppliers_tenant_id ON supplier_schema.suppliers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_suppliers_country_code ON supplier_schema.suppliers(country_code);
