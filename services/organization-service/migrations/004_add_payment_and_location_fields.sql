-- Migration: Add payment tracking fields to purchases and location tracking to items
-- 1. Add payment tracking columns to Purchase Header
ALTER TABLE inventory.purchases
ADD COLUMN paid_amount NUMERIC(15,2) DEFAULT 0.00,
ADD COLUMN due_amount NUMERIC(15,2) GENERATED ALWAYS AS (grand_total - paid_amount) STORED,
ADD COLUMN payment_status VARCHAR(20) DEFAULT 'PENDING';

-- 2. Add Location Tracking to Purchase Items
ALTER TABLE inventory.purchase_items
ADD COLUMN rack_no VARCHAR(50);

-- 3. Create Index for Payment Status for quick filtering of outstanding bills
CREATE INDEX idx_purchases_payment_status ON inventory.purchases(payment_status);
