-- Migration 060: Fix inventory.batch_audit_logs schema
-- This migration ensures batch_audit_logs has pharmacy_id instead of tenant_id

DO $$
BEGIN
    -- Check if tenant_id exists in batch_audit_logs and rename it to pharmacy_id
    IF EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_schema = 'inventory' 
          AND table_name = 'batch_audit_logs' 
          AND column_name = 'tenant_id'
    ) THEN
        -- Drop the old constraint
        ALTER TABLE inventory.batch_audit_logs DROP CONSTRAINT IF EXISTS fk_batch_audit_tenant;
        
        -- Rename column
        ALTER TABLE inventory.batch_audit_logs RENAME COLUMN tenant_id TO pharmacy_id;
        
        -- Add new constraint
        ALTER TABLE inventory.batch_audit_logs ADD CONSTRAINT fk_batch_audit_pharmacy FOREIGN KEY (pharmacy_id) REFERENCES public.pharmacies(id) ON DELETE CASCADE;
        
        -- Drop old index and create new one
        DROP INDEX IF EXISTS inventory.idx_batch_audit_tenant;
        CREATE INDEX IF NOT EXISTS idx_batch_audit_pharmacy ON inventory.batch_audit_logs(pharmacy_id);
    END IF;
END $$;
