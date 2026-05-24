-- Migration 033: Add payment columns for Clinic Appointment Management
-- This migration adds payment_method and paid_amount columns to the appointments table

ALTER TABLE appointments ADD COLUMN IF NOT EXISTS payment_method VARCHAR(20);
ALTER TABLE appointments ADD COLUMN IF NOT EXISTS paid_amount DECIMAL(10,2) DEFAULT 0.00;

-- Ensure payment_status is VARCHAR(20) and has default 'pending'
-- It was created in initial schema, but this ensures compatibility
ALTER TABLE appointments ALTER COLUMN payment_status SET DEFAULT 'pending';

-- Add indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_appointments_payment_method ON appointments(payment_method);
CREATE INDEX IF NOT EXISTS idx_appointments_paid_amount ON appointments(paid_amount);
