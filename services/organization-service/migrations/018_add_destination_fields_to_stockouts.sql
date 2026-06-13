-- Add destination fields to stock_outs master table
ALTER TABLE inventory.stock_outs 
ADD COLUMN IF NOT EXISTS destination_type VARCHAR(100),
ADD COLUMN IF NOT EXISTS destination_name VARCHAR(255);

-- Update existing rows (if any) with generic placeholders
UPDATE inventory.stock_outs 
SET destination_type = 'INTERNAL', 
    destination_name = 'PHARMACY'
WHERE destination_type IS NULL OR destination_name IS NULL;

-- Enforce mandatory constraints
ALTER TABLE inventory.stock_outs 
ALTER COLUMN destination_type SET NOT NULL,
ALTER COLUMN destination_name SET NOT NULL;
