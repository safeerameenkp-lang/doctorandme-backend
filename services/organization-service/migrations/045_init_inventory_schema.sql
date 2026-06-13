-- Consolidated Migration: Initialize Inventory Schema
-- This file contains the final state of the inventory service schema.

-- 1. Create inventory schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS inventory;

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

-- Foreign Key Constraint (Depends on supplier-service schema)
CREATE SCHEMA IF NOT EXISTS supplier_schema;
ALTER TABLE inventory.medicines
ADD CONSTRAINT fk_medicines_supplier 
FOREIGN KEY (supplier_id) REFERENCES supplier_schema.suppliers(id) ON DELETE RESTRICT;

-- Performance Indexes
CREATE INDEX IF NOT EXISTS idx_medicines_tenant_id ON inventory.medicines(tenant_id);
CREATE INDEX IF NOT EXISTS idx_medicines_tenant_name ON inventory.medicines(tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_medicines_tenant_category ON inventory.medicines(tenant_id, category);
CREATE INDEX IF NOT EXISTS idx_medicines_tenant_active ON inventory.medicines(tenant_id, is_active);
CREATE INDEX IF NOT EXISTS idx_medicines_barcode ON inventory.medicines(barcode) WHERE barcode IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_medicines_supplier_id ON inventory.medicines(supplier_id);

-- Case-Insensitive Unique Constraint for Duplicate Prevention
CREATE UNIQUE INDEX idx_medicines_unique_case_insensitive 
ON inventory.medicines (
    tenant_id, 
    LOWER(name), 
    LOWER(COALESCE(brand_name, '')), 
    LOWER(dosage_form), 
    unit_type,
    COALESCE(supplier_id::text, '')
);

COMMENT ON COLUMN inventory.medicines.supplier_id IS 'Required foreign key to supplier. Every medicine must have a supplier.';


-- 3. Purchases Table (Header)
CREATE TABLE IF NOT EXISTS inventory.purchases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    supplier_id UUID NOT NULL,
    invoice_no VARCHAR(100) NOT NULL,
    purchase_date DATE NOT NULL DEFAULT CURRENT_DATE,
    received_by VARCHAR(255) NOT NULL,
    grand_total NUMERIC(15,2) DEFAULT 0.00,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_purchases_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_schema.tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_purchases_supplier FOREIGN KEY (supplier_id) REFERENCES supplier_schema.suppliers(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_purchases_tenant ON inventory.purchases(tenant_id);
CREATE INDEX IF NOT EXISTS idx_purchases_supplier ON inventory.purchases(supplier_id);
CREATE INDEX IF NOT EXISTS idx_purchases_invoice ON inventory.purchases(invoice_no);


-- 4. Purchase Items Table (Detail)
CREATE TABLE IF NOT EXISTS inventory.purchase_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    medicine_id UUID NOT NULL,
    batch_no VARCHAR(100) NOT NULL,
    mfg_date DATE,
    expiry_date DATE NOT NULL,
    unit_mode VARCHAR(50) NOT NULL,
    units_per_mode INTEGER NOT NULL,
    received_qty INTEGER NOT NULL,
    bonus_qty INTEGER DEFAULT 0,
    total_qty_units INTEGER NOT NULL,
    base_unit VARCHAR(50) NOT NULL,
    purchase_price_per_mode NUMERIC(15,2) NOT NULL,
    mrp_per_mode NUMERIC(15,2) NOT NULL,
    cgst_rate NUMERIC(5,2) DEFAULT 0.00,
    sgst_rate NUMERIC(5,2) DEFAULT 0.00,
    total_tax_percentage NUMERIC(5,2) DEFAULT 0.00,
    retail_discount_percentage NUMERIC(5,2) DEFAULT 0.00,
    staff_discount_percentage NUMERIC(5,2) DEFAULT 0.00,
    special_discount_percentage NUMERIC(5,2) DEFAULT 0.00,
    max_discount_percentage NUMERIC(5,2) DEFAULT 0.00,
    cost_price_per_mode NUMERIC(15,2) NOT NULL,
    cost_price_per_unit NUMERIC(15,2) NOT NULL,
    selling_price_per_mode NUMERIC(15,2) NOT NULL,
    selling_price_per_unit NUMERIC(15,2) NOT NULL,
    item_total_amount NUMERIC(15,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_items_purchase FOREIGN KEY (purchase_id) REFERENCES inventory.purchases(id) ON DELETE CASCADE,
    CONSTRAINT fk_items_medicine FOREIGN KEY (medicine_id) REFERENCES inventory.medicines(id) ON DELETE RESTRICT,
    CONSTRAINT fk_items_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_schema.tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_purchase_items_purchase ON inventory.purchase_items(purchase_id);
CREATE INDEX IF NOT EXISTS idx_purchase_items_medicine ON inventory.purchase_items(medicine_id);
CREATE INDEX IF NOT EXISTS idx_purchase_items_batch ON inventory.purchase_items(batch_no);
CREATE INDEX IF NOT EXISTS idx_purchase_items_expiry ON inventory.purchase_items(expiry_date);


-- 5. Batches Table
CREATE TABLE IF NOT EXISTS inventory.batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    medicine_id UUID NOT NULL,
    batch_no VARCHAR(100) NOT NULL,
    expiry_date DATE NOT NULL,
    rack_no VARCHAR(50),
    quantity_available INTEGER NOT NULL DEFAULT 0 CHECK (quantity_available >= 0),
    cost_price NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    mrp NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    unit_price NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    cgst_rate NUMERIC(5,2) DEFAULT 0.00,
    sgst_rate NUMERIC(5,2) DEFAULT 0.00,
    retail_disc_perc NUMERIC(5,2) DEFAULT 0.00 CHECK (retail_disc_perc >= 0 AND retail_disc_perc <= 100),
    staff_disc_perc NUMERIC(5,2) DEFAULT 0.00 CHECK (staff_disc_perc >= 0 AND staff_disc_perc <= 100),
    special_disc_perc NUMERIC(5,2) DEFAULT 0.00 CHECK (special_disc_perc >= 0 AND special_disc_perc <= 100),
    max_disc_perc NUMERIC(5,2) DEFAULT 0.00 CHECK (max_disc_perc >= 0 AND max_disc_perc <= 100),
    supplier_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_batches_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_schema.tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_batches_medicine FOREIGN KEY (medicine_id) REFERENCES inventory.medicines(id) ON DELETE RESTRICT,
    CONSTRAINT uq_batches_medicine_batch UNIQUE (tenant_id, medicine_id, batch_no)
);

CREATE INDEX IF NOT EXISTS idx_batches_expiry ON inventory.batches(expiry_date);
CREATE INDEX IF NOT EXISTS idx_batches_medicine ON inventory.batches(medicine_id);
CREATE INDEX IF NOT EXISTS idx_batches_tenant_medicine ON inventory.batches(tenant_id, medicine_id);


-- 6. Stock Ledger Table
CREATE TABLE IF NOT EXISTS inventory.stock_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    medicine_id UUID NOT NULL,
    batch_id UUID NOT NULL,
    transaction_type VARCHAR(50) NOT NULL, -- PURCHASE, SALE, SALE_RETURN, PURCHASE_RETURN, ADJUSTMENT
    quantity_change INTEGER NOT NULL,
    balance_after INTEGER NOT NULL,
    reference_type VARCHAR(50),
    reference_id UUID,
    performed_by UUID,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_ledger_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_schema.tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_ledger_medicine FOREIGN KEY (medicine_id) REFERENCES inventory.medicines(id) ON DELETE RESTRICT,
    CONSTRAINT fk_ledger_batch FOREIGN KEY (batch_id) REFERENCES inventory.batches(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_ledger_tenant_medicine ON inventory.stock_ledger(tenant_id, medicine_id);
CREATE INDEX IF NOT EXISTS idx_ledger_tenant_batch ON inventory.stock_ledger(tenant_id, batch_id);
CREATE INDEX IF NOT EXISTS idx_ledger_reference ON inventory.stock_ledger(reference_id);
CREATE INDEX IF NOT EXISTS idx_ledger_created_at ON inventory.stock_ledger(created_at);

COMMENT ON TABLE inventory.stock_ledger IS 'Immutable audit log for all stock movements to ensure inventory integrity and financial auditing.';


-- 7. Stock Outs Table
CREATE TABLE IF NOT EXISTS inventory.stock_outs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    type VARCHAR(50) NOT NULL,
    reason TEXT,
    total_loss_value DECIMAL(12, 2) NOT NULL DEFAULT 0.0,
    created_by_id UUID NOT NULL,
    created_by_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_stock_outs_tenant ON inventory.stock_outs(tenant_id);


-- 8. Stock Out Items Table
CREATE TABLE IF NOT EXISTS inventory.stock_out_items (
    id UUID PRIMARY KEY,
    stock_out_id UUID NOT NULL REFERENCES inventory.stock_outs(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    medicine_id UUID NOT NULL,
    medicine_name VARCHAR(255) NOT NULL,
    batch_id UUID NOT NULL,
    batch_no VARCHAR(100) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_cost_price DECIMAL(10, 2) NOT NULL DEFAULT 0.0,
    total_loss DECIMAL(12, 2) NOT NULL DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_stock_out_items_stock_out ON inventory.stock_out_items(stock_out_id);
CREATE INDEX IF NOT EXISTS idx_stock_out_items_batch ON inventory.stock_out_items(batch_id);


-- 9. Stock Out Audit Logs Table
CREATE TABLE IF NOT EXISTS inventory.stock_out_audit_logs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    stock_out_id UUID NOT NULL REFERENCES inventory.stock_outs(id) ON DELETE CASCADE,
    action_type VARCHAR(50) NOT NULL,
    changed_by UUID NOT NULL,
    changed_by_name VARCHAR(255) NOT NULL,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_stock_out_audit_logs_stock_out ON inventory.stock_out_audit_logs(stock_out_id);


-- 10. Reservations Table
CREATE TABLE IF NOT EXISTS inventory.reservations (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    product_id UUID NOT NULL REFERENCES inventory.medicines(id),
    batch_id UUID NOT NULL REFERENCES inventory.batches(id),
    quantity INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_reservations_batch ON inventory.reservations(batch_id, tenant_id);
CREATE INDEX IF NOT EXISTS idx_reservations_expires ON inventory.reservations(expires_at);
