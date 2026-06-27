-- Migration: Create Dummy Tenant Schema to support older migrations
-- This allows table creation referencing tenant_schema.tenants(id) to succeed.
-- These references are later migrated to pharmacy_id by subsequent migrations.

CREATE SCHEMA IF NOT EXISTS tenant_schema;

CREATE TABLE IF NOT EXISTS tenant_schema.tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255)
);
