-- Consolidated Migration: Initialize Inventory Schema
-- This file contains the final state of the inventory service schema.

-- 1. Create inventory schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS inventory;

-- Drop existing table command removed to preserve data
-- DROP TABLE IF EXISTS inventory.medicines CASCADE;

-- 2. Medicines Table (Catalog/Master Data)
CREATE TABLE IF NOT EXISTS inventory.medicines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    
    -- Mandatory Fields
    name VARCHAR(255) NOT NULL,
    dosage_form VARCHAR(100) NOT NULL,
    category VARCHAR(100) NOT NULL,
    manufacturer VARCHAR(255) NOT NULL,
    hsn_code VARCHAR(20) NOT NULL,
    unit_type VARCHAR(50) NOT NULL,
    supplier_id UUID NOT NULL, -- Required foreign key to supplier
    
    -- Optional Fields
    brand_name VARCHAR(255),
    mfg_license VARCHAR(255),
    schedule_type VARCHAR(20),
    is_rx_required BOOLEAN DEFAULT FALSE,
    barcode VARCHAR(100),
    storage_condition VARCHAR(100),
    cgst_rate NUMERIC(5,2) DEFAULT 0.00,
    sgst_rate NUMERIC(5,2) DEFAULT 0.00,
    
    -- Control Fields
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Foreign Key Constraint (Depends on supplier-service schema)
-- Note: Ensure supplier_schema exists if this migration is run on an empty DB
CREATE SCHEMA IF NOT EXISTS supplier_schema;
ALTER TABLE inventory.medicines
ADD CONSTRAINT fk_medicines_supplier 
FOREIGN KEY (supplier_id) REFERENCES supplier_schema.suppliers(id) ON DELETE RESTRICT;

-- 4. Performance Indexes
CREATE INDEX IF NOT EXISTS idx_medicines_tenant_id ON inventory.medicines(tenant_id);
CREATE INDEX IF NOT EXISTS idx_medicines_tenant_name ON inventory.medicines(tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_medicines_tenant_category ON inventory.medicines(tenant_id, category);
CREATE INDEX IF NOT EXISTS idx_medicines_tenant_active ON inventory.medicines(tenant_id, is_active);
CREATE INDEX IF NOT EXISTS idx_medicines_barcode ON inventory.medicines(barcode) WHERE barcode IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_medicines_supplier_id ON inventory.medicines(supplier_id);

-- 5. Case-Insensitive Unique Constraint for Duplicate Prevention
CREATE UNIQUE INDEX idx_medicines_unique_case_insensitive 
ON inventory.medicines (
    tenant_id, 
    LOWER(name), 
    LOWER(COALESCE(brand_name, '')), 
    LOWER(dosage_form), 
    unit_type,
    COALESCE(supplier_id::text, '')
);

-- 6. Column Comment
COMMENT ON COLUMN inventory.medicines.supplier_id IS 'Required foreign key to supplier. Every medicine must have a supplier.';
