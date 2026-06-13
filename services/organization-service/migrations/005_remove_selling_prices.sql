-- Migration: Drop selling price columns from purchase_items
ALTER TABLE inventory.purchase_items
DROP COLUMN IF EXISTS selling_price_per_mode,
DROP COLUMN IF EXISTS selling_price_per_unit;
