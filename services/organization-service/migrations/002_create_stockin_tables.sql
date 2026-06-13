-- Migration: Create Stock-In Tables (Standardized with Billing Control Tiers)
-- Description: Header-Detail tables with specific focus on item-level tax and multi-category discount tracking.

-- 1. Purchases Table (Header)
CREATE TABLE IF NOT EXISTS inventory.purchases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    supplier_id UUID NOT NULL,
    
    invoice_no VARCHAR(100) NOT NULL,
    purchase_date DATE NOT NULL DEFAULT CURRENT_DATE,
    received_by VARCHAR(255) NOT NULL,
    
    -- Final Payable Amount
    grand_total NUMERIC(15,2) DEFAULT 0.00,
    
    -- Metadata
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT fk_purchases_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_schema.tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_purchases_supplier FOREIGN KEY (supplier_id) REFERENCES supplier_schema.suppliers(id) ON DELETE RESTRICT
);

-- 2. Purchase Items Table (Detail)
CREATE TABLE IF NOT EXISTS inventory.purchase_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    medicine_id UUID NOT NULL,
    
    -- Batch & Expiry
    batch_no VARCHAR(100) NOT NULL,
    mfg_date DATE,
    expiry_date DATE NOT NULL,
    
    -- Quantity Details
    unit_mode VARCHAR(50) NOT NULL,
    units_per_mode INTEGER NOT NULL,
    received_qty INTEGER NOT NULL,
    bonus_qty INTEGER DEFAULT 0,
    total_qty_units INTEGER NOT NULL,
    base_unit VARCHAR(50) NOT NULL, -- tablets, bottles, piece, etc.
    
    -- Pricing Details (Per Mode)
    purchase_price_per_mode NUMERIC(15,2) NOT NULL,
    mrp_per_mode NUMERIC(15,2) NOT NULL,
    
    -- TAX Details (Item Level - Rates Only)
    cgst_rate NUMERIC(5,2) DEFAULT 0.00,
    sgst_rate NUMERIC(5,2) DEFAULT 0.00,
    total_tax_percentage NUMERIC(5,2) DEFAULT 0.00,

    -- DISCOUNT/BILLING CONTROL TIERS (Rates Only)
    retail_discount_percentage NUMERIC(5,2) DEFAULT 0.00,
    staff_discount_percentage NUMERIC(5,2) DEFAULT 0.00,
    special_discount_percentage NUMERIC(5,2) DEFAULT 0.00,
    max_discount_percentage NUMERIC(5,2) DEFAULT 0.00, -- Hard limit for security
    
    -- Calculated Costs (Per Unit - Derived from Manual Item Total)
    cost_price_per_mode NUMERIC(15,2) NOT NULL,
    cost_price_per_unit NUMERIC(15,2) NOT NULL,
    selling_price_per_mode NUMERIC(15,2) NOT NULL,
    selling_price_per_unit NUMERIC(15,2) NOT NULL,
    item_total_amount NUMERIC(15,2) NOT NULL, -- Manual Entry from UI
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT fk_items_purchase FOREIGN KEY (purchase_id) REFERENCES inventory.purchases(id) ON DELETE CASCADE,
    CONSTRAINT fk_items_medicine FOREIGN KEY (medicine_id) REFERENCES inventory.medicines(id) ON DELETE RESTRICT,
    CONSTRAINT fk_items_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_schema.tenants(id) ON DELETE CASCADE
);

-- 3. Indexes
CREATE INDEX IF NOT EXISTS idx_purchases_tenant ON inventory.purchases(tenant_id);
CREATE INDEX IF NOT EXISTS idx_purchases_supplier ON inventory.purchases(supplier_id);
CREATE INDEX IF NOT EXISTS idx_purchases_invoice ON inventory.purchases(invoice_no);

CREATE INDEX IF NOT EXISTS idx_purchase_items_purchase ON inventory.purchase_items(purchase_id);
CREATE INDEX IF NOT EXISTS idx_purchase_items_medicine ON inventory.purchase_items(medicine_id);
CREATE INDEX IF NOT EXISTS idx_purchase_items_batch ON inventory.purchase_items(batch_no);
CREATE INDEX IF NOT EXISTS idx_purchase_items_expiry ON inventory.purchase_items(expiry_date);
