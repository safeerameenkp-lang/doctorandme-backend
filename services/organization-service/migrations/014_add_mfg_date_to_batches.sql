-- Add mfg_date to batches table
ALTER TABLE inventory.batches ADD COLUMN mfg_date DATE NOT NULL;

-- Optional: If you want to update existing batches with a reasonable default based on expiry (e.g., 2 years before)
-- UPDATE inventory.batches SET mfg_date = expiry_date - INTERVAL '2 years' WHERE mfg_date = CURRENT_DATE;
