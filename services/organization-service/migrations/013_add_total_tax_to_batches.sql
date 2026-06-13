-- Migration to add total_tax_percentage to inventory.batches table
ALTER TABLE inventory.batches ADD COLUMN IF NOT EXISTS total_tax_percentage DECIMAL(5, 2) DEFAULT 0.00;
