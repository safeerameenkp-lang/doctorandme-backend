-- 011_add_user_names_to_medicines.sql
ALTER TABLE inventory.medicines 
ADD COLUMN IF NOT EXISTS created_by_name VARCHAR(255) DEFAULT 'System',
ADD COLUMN IF NOT EXISTS updated_by_name VARCHAR(255) DEFAULT 'System';
