-- Consolidated Migration: Initialize Sales Schema
-- This file contains the final state of the sales service schema.

-- 1. Create sales schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS sales_schema;

-- 2. Patients Table
CREATE TABLE IF NOT EXISTS sales_schema.patients (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    gender TEXT,
    age INT,
    address TEXT,
    is_recurring BOOLEAN DEFAULT FALSE,
    due_amount NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    credit_amount NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT patients_tenant_id_phone_name_key UNIQUE (tenant_id, phone, name)
);

CREATE INDEX IF NOT EXISTS idx_patients_phone ON sales_schema.patients(phone);

-- 3. Prescriptions Table
CREATE TABLE IF NOT EXISTS sales_schema.prescriptions (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id UUID NOT NULL,
    token_no VARCHAR(50),
    patient_name VARCHAR(150) NOT NULL,
    patient_phone VARCHAR(20),
    doctor_name VARCHAR(150),
    date TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    bill_amount DECIMAL(12,2) DEFAULT NULL,
    payment_method VARCHAR(20) DEFAULT NULL,
    handled_by_name VARCHAR(100) DEFAULT NULL,
    latest_sale_id UUID,
    invoice_number VARCHAR(100)
);

CREATE INDEX IF NOT EXISTS idx_temp_rx_tenant ON sales_schema.prescriptions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_prescriptions_latest_sale_id ON sales_schema.prescriptions(latest_sale_id);

-- 4. Prescription Items Table
CREATE TABLE IF NOT EXISTS sales_schema.prescription_items (
    id UUID PRIMARY KEY,
    prescription_id VARCHAR(50) NOT NULL REFERENCES sales_schema.prescriptions(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    medicine_name VARCHAR(150) NOT NULL,
    medicine_brand VARCHAR(255),
    quantity INTEGER NOT NULL,
    instructions TEXT,
    duration_days INT DEFAULT 0,
    dosage_per_day DECIMAL(10,2) DEFAULT 0,
    morning DECIMAL(10,2) DEFAULT 0,
    noon DECIMAL(10,2) DEFAULT 0,
    night DECIMAL(10,2) DEFAULT 0
);

-- 5. Sales Table
CREATE TABLE IF NOT EXISTS sales_schema.sales (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    sale_type VARCHAR(20) NOT NULL,
    prescription_id VARCHAR(50) NOT NULL,
    patient_id UUID REFERENCES sales_schema.patients(id),
    customer_name TEXT,
    customer_phone TEXT,
    customer_address TEXT,
    status VARCHAR(20) NOT NULL,
    gross_amount DECIMAL(15,2) DEFAULT 0.00,
    total_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    total_discount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    total_tax DECIMAL(15, 2) NOT NULL DEFAULT 0,
    invoice_number VARCHAR(50) UNIQUE,
    is_recurring BOOLEAN DEFAULT FALSE,
    days_supply INTEGER DEFAULT 0,
    next_refill_date TIMESTAMP,
    applied_credit NUMERIC(10, 2) DEFAULT 0.00,
    applied_due NUMERIC(10, 2) DEFAULT 0.00,
    generated_credit NUMERIC(10, 2) DEFAULT 0.00,
    generated_due NUMERIC(10, 2) DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sales_tenant ON sales_schema.sales(tenant_id);

-- 6. Sale Items Table
CREATE TABLE IF NOT EXISTS sales_schema.sale_items (
    id UUID PRIMARY KEY,
    sale_id UUID NOT NULL REFERENCES sales_schema.sales(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    batch_id UUID NOT NULL,
    quantity INTEGER NOT NULL,
    price DECIMAL(15, 2) NOT NULL,
    reservation_id VARCHAR(100),
    mrp DECIMAL(15,2) DEFAULT 0,
    discount_percentage DECIMAL(5,2) DEFAULT 0,
    tax_percentage DECIMAL(5,2) DEFAULT 0,
    subtotal DECIMAL(15,2) DEFAULT 0,
    medicine_name VARCHAR(255),
    medicine_brand VARCHAR(255),
    expiry_date TIMESTAMP,
    batch_no VARCHAR(255),
    retail_disc_perc DECIMAL(5,2) DEFAULT 0,
    staff_disc_perc DECIMAL(5,2) DEFAULT 0,
    special_disc_perc DECIMAL(5,2) DEFAULT 0,
    max_disc_perc DECIMAL(5,2) DEFAULT 0,
    rack_no TEXT DEFAULT 'N/A',
    returned_quantity INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sales_items_sale ON sales_schema.sale_items(sale_id);

-- 7. Sales Returns Table
CREATE TABLE IF NOT EXISTS sales_schema.sales_returns (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    sale_id UUID NOT NULL REFERENCES sales_schema.sales(id),
    return_number VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL,
    total_refund DECIMAL(15, 2) NOT NULL DEFAULT 0,
    reason TEXT,
    handled_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sales_returns_sale_id ON sales_schema.sales_returns(sale_id);
CREATE INDEX IF NOT EXISTS idx_sales_returns_tenant_id ON sales_schema.sales_returns(tenant_id);

-- 8. Sales Return Items Table
CREATE TABLE IF NOT EXISTS sales_schema.sales_return_items (
    id UUID PRIMARY KEY,
    return_id UUID NOT NULL REFERENCES sales_schema.sales_returns(id),
    sale_item_id UUID NOT NULL REFERENCES sales_schema.sale_items(id),
    product_id UUID NOT NULL,
    medicine_name VARCHAR(255) NOT NULL,
    batch_id UUID NOT NULL,
    batch_no VARCHAR(100),
    quantity INTEGER NOT NULL,
    refund_amount DECIMAL(15, 2) NOT NULL,
    condition VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 9. Payments Table
CREATE TABLE IF NOT EXISTS sales_schema.payments (
    id UUID PRIMARY KEY,
    sale_id UUID NOT NULL REFERENCES sales_schema.sales(id) ON DELETE CASCADE,
    mode VARCHAR(20) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    transaction_type VARCHAR(20) DEFAULT 'PAYMENT',
    return_id UUID REFERENCES sales_schema.sales_returns(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payments_return_id ON sales_schema.payments(return_id);
