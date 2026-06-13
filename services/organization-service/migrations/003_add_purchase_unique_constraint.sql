-- Migration: Add unique constraint to prevent duplicate purchases
-- A purchase is considered duplicate if the same tenant records the same invoice number from the same supplier.

ALTER TABLE inventory.purchases
ADD CONSTRAINT uq_purchases_invoice_supplier_tenant 
UNIQUE (tenant_id, supplier_id, invoice_no);

-- Optional: Create an index to support faster lookups for this constraint is implicit with the UNIQUE constraint
