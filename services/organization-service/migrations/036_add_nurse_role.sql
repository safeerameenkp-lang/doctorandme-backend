-- Migration 036: Add nurse role and missing columns to roles table
-- Fixes: "Failed to find role: nurse" error when creating staff
-- Fixes: ListRoles query expecting is_active and description columns

-- Step 1: Add missing columns to roles table (safe, idempotent)
ALTER TABLE roles ADD COLUMN IF NOT EXISTS description  VARCHAR(255) DEFAULT '';
ALTER TABLE roles ADD COLUMN IF NOT EXISTS is_system_role BOOLEAN DEFAULT TRUE;
ALTER TABLE roles ADD COLUMN IF NOT EXISTS is_active    BOOLEAN DEFAULT TRUE;

-- Step 2: Insert the missing nurse role
INSERT INTO roles (name, description, is_system_role, is_active, permissions)
VALUES (
    'nurse',
    'Assists doctors and cares for patients',
    TRUE,
    TRUE,
    '{"patients": ["read", "update"], "appointments": ["read", "update"], "vitals": ["read", "create", "update"]}'
)
ON CONFLICT (name) DO NOTHING;

-- Step 3: Backfill description for existing roles that have none
UPDATE roles SET description = 'Full system access'                          WHERE name = 'super_admin'       AND (description IS NULL OR description = '');
UPDATE roles SET description = 'Manages organization and its clinics'        WHERE name = 'organization_admin' AND (description IS NULL OR description = '');
UPDATE roles SET description = 'Manages clinic staff and settings'           WHERE name = 'clinic_admin'       AND (description IS NULL OR description = '');
UPDATE roles SET description = 'Consults patients and manages prescriptions' WHERE name = 'doctor'             AND (description IS NULL OR description = '');
UPDATE roles SET description = 'Front desk and appointment management'       WHERE name = 'receptionist'       AND (description IS NULL OR description = '');
UPDATE roles SET description = 'Manages pharmacy dispensing and inventory'   WHERE name = 'pharmacist'         AND (description IS NULL OR description = '');
UPDATE roles SET description = 'Performs lab tests and uploads results'      WHERE name = 'lab_technician'     AND (description IS NULL OR description = '');
UPDATE roles SET description = 'Handles billing, payments and invoices'      WHERE name = 'billing_staff'      AND (description IS NULL OR description = '');
UPDATE roles SET description = 'Patient with access to own records'          WHERE name = 'patient'            AND (description IS NULL OR description = '');
