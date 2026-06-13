ALTER TABLE inventory.medicines 
ADD COLUMN created_by UUID,
ADD COLUMN updated_by UUID;

-- Optional: If we want to make it mandatory for future entries
-- we might want to fill it for existing entries first, 
-- but since this is dev, we can just add it.
-- COMMENT ON COLUMN inventory.medicines.created_by IS 'The user who created this medicine record.';
