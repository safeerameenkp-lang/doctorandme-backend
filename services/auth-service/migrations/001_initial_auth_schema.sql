-- Auth Service: Initial Schema
-- This migration creates core authentication and authorization tables

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    date_of_birth DATE,
    gender VARCHAR(20),
    is_active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Roles (system + custom)
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,   -- e.g. super_admin, clinic_admin, doctor
    description VARCHAR(255) DEFAULT '',
    permissions JSONB DEFAULT '{}',
    is_system_role BOOLEAN DEFAULT TRUE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User Roles (per org/clinic/service)
-- Note: References organizations, clinics, external_services which are created by organization-service
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    organization_id UUID,  -- Will reference organizations table (created by org service)
    clinic_id UUID,        -- Will reference clinics table (created by org service)
    service_id UUID,        -- Will reference external_services table (created by org service)
    is_active BOOLEAN DEFAULT TRUE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id, organization_id, clinic_id, service_id)
);

-- Refresh Tokens (multi-device support)
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL, -- store hashed if extra security needed
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP,
    UNIQUE(user_id, token)
);

-- Insert default roles with hierarchical permissions
INSERT INTO roles (name, description, is_system_role, is_active, permissions) VALUES 
('super_admin',       'Full system access',                          TRUE, TRUE, '{"organizations": ["create", "read", "update", "delete"], "clinics": ["create", "read", "update", "delete"], "services": ["create", "read", "update", "delete"], "users": ["create", "read", "update", "delete"], "roles": ["create", "read", "update", "delete"]}'),
('organization_admin','Manages organization and its clinics',        TRUE, TRUE, '{"organizations": ["read", "update"], "clinics": ["create", "read", "update", "delete"], "users": ["create", "read", "update", "delete"], "roles": ["read"]}'),
('clinic_admin',      'Manages clinic staff and settings',           TRUE, TRUE, '{"clinics": ["read", "update"], "users": ["create", "read", "update", "delete"], "roles": ["read"], "staff": ["create", "read", "update", "delete"]}'),
('doctor',            'Consults patients and manages prescriptions', TRUE, TRUE, '{"patients": ["read", "update"], "appointments": ["read", "create", "update"], "prescriptions": ["read", "create", "update"]}'),
('receptionist',      'Front desk and appointment management',       TRUE, TRUE, '{"patients": ["read", "create", "update"], "appointments": ["read", "create", "update"], "billing": ["read", "create"]}'),
('pharmacist',        'Manages pharmacy dispensing and inventory',   TRUE, TRUE, '{"prescriptions": ["read", "update"], "medications": ["read", "create", "update"], "inventory": ["read", "update"]}'),
('lab_technician',    'Performs lab tests and uploads results',      TRUE, TRUE, '{"lab_orders": ["read", "create", "update"], "lab_results": ["read", "create", "update"], "reports": ["read", "create"]}'),
('billing_staff',     'Handles billing, payments and invoices',      TRUE, TRUE, '{"billing": ["read", "create", "update"], "payments": ["read", "create", "update"], "invoices": ["read", "create", "update"]}'),
('nurse',             'Assists doctors and cares for patients',      TRUE, TRUE, '{"patients": ["read", "update"], "appointments": ["read", "update"], "vitals": ["read", "create", "update"]}'),
('patient',           'Patient with access to own records',          TRUE, TRUE, '{"profile": ["read", "update"], "appointments": ["read", "create"], "prescriptions": ["read"], "lab_results": ["read"]}')
ON CONFLICT (name) DO NOTHING;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_organization_id ON user_roles(organization_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_clinic_id ON user_roles(clinic_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_service_id ON user_roles(service_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);

