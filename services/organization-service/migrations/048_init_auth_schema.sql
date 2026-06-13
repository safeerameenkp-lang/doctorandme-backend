-- Migration: Initialize Auth Schema

-- 1. Create auth schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS auth_schema;

-- 2. Create users table
CREATE TABLE IF NOT EXISTS auth_schema.users (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (tenant_id, email)
);

-- 3. Indexes for efficient lookup
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON auth_schema.users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_tenant_role ON auth_schema.users(tenant_id, role);
