-- Add destination_id to stock_outs table for supplier validation
ALTER TABLE inventory.stock_outs ADD COLUMN destination_id UUID;
