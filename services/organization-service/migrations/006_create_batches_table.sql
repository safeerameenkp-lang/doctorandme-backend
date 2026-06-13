-- Migration: Create inventory.batches schema
-- This table tracks the live, on-hand inventory state (quantity, expiry, location)
-- and pricing rules (tax rates, discount tiers) for each batch.

CREATE TABLE IF NOT EXISTS inventory.batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    medicine_id UUID NOT NULL,
    
    -- Core Tracking
    batch_no VARCHAR(100) NOT NULL,
    expiry_date DATE NOT NULL,
    rack_no VARCHAR(50),
    
    -- Quantity (Live State)
    quantity_available INTEGER NOT NULL DEFAULT 0 CHECK (quantity_available >= 0),
    
    -- Pricing
    cost_price NUMERIC(15,2) NOT NULL DEFAULT 0.00, -- Per Unit
    mrp NUMERIC(15,2) NOT NULL DEFAULT 0.00,        -- Per Unit
    unit_price NUMERIC(15,2) NOT NULL DEFAULT 0.00, -- Usually same as MRP or Selling Price
    
    -- Tax Rates (Inherited from Purchase/Master)
    cgst_rate NUMERIC(5,2) DEFAULT 0.00,
    sgst_rate NUMERIC(5,2) DEFAULT 0.00,
    
    -- Discount Tiers (Inherited from Purchase)
    retail_disc_perc NUMERIC(5,2) DEFAULT 0.00 CHECK (retail_disc_perc >= 0 AND retail_disc_perc <= 100),
    staff_disc_perc NUMERIC(5,2) DEFAULT 0.00 CHECK (staff_disc_perc >= 0 AND staff_disc_perc <= 100),
    special_disc_perc NUMERIC(5,2) DEFAULT 0.00 CHECK (special_disc_perc >= 0 AND special_disc_perc <= 100),
    max_disc_perc NUMERIC(5,2) DEFAULT 0.00 CHECK (max_disc_perc >= 0 AND max_disc_perc <= 100),
    
    -- Audit
    supplier_id UUID, -- Optional: Links to the LATEST supplier (or first)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT fk_batches_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_schema.tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_batches_medicine FOREIGN KEY (medicine_id) REFERENCES inventory.medicines(id) ON DELETE RESTRICT,
    -- Ensure unique batch per medicine per tenant (so we upsert instead of duplicating)
    CONSTRAINT uq_batches_medicine_batch UNIQUE (tenant_id, medicine_id, batch_no)
);

-- Performance Indexes
CREATE INDEX IF NOT EXISTS idx_batches_expiry ON inventory.batches(expiry_date);
CREATE INDEX IF NOT EXISTS idx_batches_medicine ON inventory.batches(medicine_id);
CREATE INDEX IF NOT EXISTS idx_batches_tenant_medicine ON inventory.batches(tenant_id, medicine_id);
