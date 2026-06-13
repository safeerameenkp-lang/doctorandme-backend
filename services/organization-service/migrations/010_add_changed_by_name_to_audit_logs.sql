-- 010_add_changed_by_name_to_audit_logs.sql
ALTER TABLE inventory.medicine_audit_logs 
ADD COLUMN IF NOT EXISTS changed_by_name VARCHAR(255) DEFAULT 'System';
