-- Migration 049: Create pharmacies table, link to user_roles, and add pharmacy_admin role

-- Step 1: Create pharmacies table
CREATE TABLE IF NOT EXISTS pharmacies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    clinic_id UUID REFERENCES clinics(id) ON DELETE SET NULL, -- Nullable, for linking to a clinic later
    user_id UUID, -- References users(id) in public schema
    pharmacy_code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    pharmacy_type VARCHAR(50) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20),
    address TEXT,
    license_number VARCHAR(100),
    logo VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_pharmacies_organization_id ON pharmacies(organization_id);
CREATE INDEX IF NOT EXISTS idx_pharmacies_clinic_id ON pharmacies(clinic_id);
CREATE INDEX IF NOT EXISTS idx_pharmacies_user_id ON pharmacies(user_id);
CREATE INDEX IF NOT EXISTS idx_pharmacies_pharmacy_code ON pharmacies(pharmacy_code);

-- Step 2: Add pharmacy_id column to user_roles
ALTER TABLE user_roles ADD COLUMN IF NOT EXISTS pharmacy_id UUID REFERENCES pharmacies(id) ON DELETE CASCADE;

-- Add index on user_roles(pharmacy_id)
CREATE INDEX IF NOT EXISTS idx_user_roles_pharmacy_id ON user_roles(pharmacy_id);

-- Drop existing unique constraint if present and add the new one
-- We do a soft drop/add by checking if it exists or dropping it directly
ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS uq_user_roles_pharmacy;
ALTER TABLE user_roles ADD CONSTRAINT uq_user_roles_pharmacy UNIQUE (user_id, role_id, pharmacy_id);

-- Step 3: Insert pharmacy_admin role
INSERT INTO roles (name, description, is_system_role, is_active, permissions)
VALUES (
    'pharmacy_admin',
    'Manages pharmacy staff, settings and inventory',
    TRUE,
    TRUE,
    '{"pharmacy": ["read", "update"], "users": ["create", "read", "update", "delete"], "roles": ["read"], "inventory": ["read", "create", "update", "delete"]}'
)
ON CONFLICT (name) DO NOTHING;
