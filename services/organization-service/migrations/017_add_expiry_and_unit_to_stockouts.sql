-- Add expiry_date and unit_type to stock_out_items
ALTER TABLE inventory.stock_out_items 
ADD COLUMN IF NOT EXISTS expiry_date DATE,
ADD COLUMN IF NOT EXISTS unit_type VARCHAR(50);

-- Update existing rows (if any) with placeholder data before enforcing NOT NULL
-- (Using common defaults to avoid migration failure)
UPDATE inventory.stock_out_items 
SET expiry_date = CURRENT_DATE, 
    unit_type = 'Unit'
WHERE expiry_date IS NULL OR unit_type IS NULL;

-- Enforce mandatory constraints
ALTER TABLE inventory.stock_out_items 
ALTER COLUMN expiry_date SET NOT NULL,
ALTER COLUMN unit_type SET NOT NULL;
