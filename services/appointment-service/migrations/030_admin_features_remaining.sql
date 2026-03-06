-- 030_admin_features_remaining.sql
-- Queue, Pharmacy, Lab, Billing, and Analytics tables
-- This migration runs after appointments and staff table are created

-- Queue Management Tables
CREATE TABLE IF NOT EXISTS queues (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID NOT NULL, -- References clinics table
    queue_type VARCHAR(50) NOT NULL, -- doctor, lab, pharmacy
    doctor_id UUID, -- References doctors table
    is_active BOOLEAN DEFAULT TRUE,
    is_paused BOOLEAN DEFAULT FALSE,
    current_token INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS queue_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    queue_id UUID REFERENCES queues(id) ON DELETE CASCADE,
    patient_id UUID NOT NULL, -- References patients table
    appointment_id UUID REFERENCES appointments(id) ON DELETE CASCADE,
    token_number INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'waiting', -- waiting, called, completed, cancelled
    priority BOOLEAN DEFAULT FALSE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    called_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Pharmacy Management Tables
CREATE TABLE IF NOT EXISTS pharmacy_inventory (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID NOT NULL,
    medicine_name VARCHAR(255) NOT NULL,
    generic_name VARCHAR(255),
    medicine_code VARCHAR(50) UNIQUE NOT NULL,
    category VARCHAR(100),
    unit VARCHAR(20) NOT NULL, -- tablet, ml, mg, etc.
    current_stock INTEGER DEFAULT 0,
    min_stock_level INTEGER DEFAULT 0,
    max_stock_level INTEGER DEFAULT 0,
    unit_price DECIMAL(10,2) NOT NULL,
    expiry_date DATE,
    supplier_name VARCHAR(255),
    batch_number VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pharmacy_suppliers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID NOT NULL,
    supplier_name VARCHAR(255) NOT NULL,
    contact_person VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(20),
    address TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pharmacy_discounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID NOT NULL,
    discount_name VARCHAR(255) NOT NULL,
    discount_type VARCHAR(20) NOT NULL, -- percentage, fixed_amount
    discount_value DECIMAL(10,2) NOT NULL,
    min_purchase_amount DECIMAL(10,2) DEFAULT 0,
    max_discount_amount DECIMAL(10,2),
    valid_from DATE NOT NULL,
    valid_to DATE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pharmacy_billing (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID NOT NULL,
    patient_id UUID NOT NULL,
    appointment_id UUID REFERENCES appointments(id) ON DELETE CASCADE,
    prescription_id UUID,
    total_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    final_amount DECIMAL(10,2) NOT NULL,
    payment_status VARCHAR(20) DEFAULT 'pending', -- pending, paid, partial, refunded
    payment_mode VARCHAR(20), -- cash, card, insurance
    billing_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pharmacy_billing_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    billing_id UUID REFERENCES pharmacy_billing(id) ON DELETE CASCADE,
    medicine_id UUID REFERENCES pharmacy_inventory(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    final_price DECIMAL(10,2) NOT NULL
);

-- Lab Management Tables
CREATE TABLE IF NOT EXISTS lab_tests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID NOT NULL,
    test_code VARCHAR(50) UNIQUE NOT NULL,
    test_name VARCHAR(255) NOT NULL,
    test_category VARCHAR(100),
    description TEXT,
    sample_type VARCHAR(100), -- blood, urine, stool, etc.
    preparation_instructions TEXT,
    normal_range TEXT,
    unit VARCHAR(20),
    price DECIMAL(10,2) NOT NULL,
    turnaround_time_hours INTEGER DEFAULT 24,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS lab_sample_collectors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    clinic_id UUID NOT NULL,
    collector_code VARCHAR(20),
    specialization VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS lab_orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID NOT NULL,
    patient_id UUID NOT NULL,
    appointment_id UUID REFERENCES appointments(id) ON DELETE CASCADE,
    doctor_id UUID NOT NULL,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'ordered', -- ordered, collected, processing, completed, cancelled
    total_amount DECIMAL(10,2) NOT NULL,
    payment_status VARCHAR(20) DEFAULT 'pending',
    collector_id UUID REFERENCES lab_sample_collectors(id),
    collection_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS lab_order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID REFERENCES lab_orders(id) ON DELETE CASCADE,
    test_id UUID REFERENCES lab_tests(id) ON DELETE CASCADE,
    quantity INTEGER DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL
);

CREATE TABLE IF NOT EXISTS lab_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID REFERENCES lab_orders(id) ON DELETE CASCADE,
    test_id UUID REFERENCES lab_tests(id) ON DELETE CASCADE,
    result_value VARCHAR(255),
    result_unit VARCHAR(20),
    normal_range VARCHAR(100),
    status VARCHAR(20) DEFAULT 'normal', -- normal, abnormal, critical
    notes TEXT,
    uploaded_by UUID NOT NULL,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_visible_to_patient BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Billing & Payment Control Tables
CREATE TABLE IF NOT EXISTS fee_structures (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID NOT NULL,
    service_type VARCHAR(50) NOT NULL, -- consultation, lab, pharmacy, procedure
    service_name VARCHAR(255) NOT NULL,
    base_fee DECIMAL(10,2) NOT NULL,
    follow_up_fee DECIMAL(10,2),
    follow_up_days INTEGER DEFAULT 7,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS billing_discounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID NOT NULL,
    discount_name VARCHAR(255) NOT NULL,
    discount_type VARCHAR(20) NOT NULL, -- percentage, fixed_amount
    discount_value DECIMAL(10,2) NOT NULL,
    applicable_services JSONB DEFAULT '[]', -- array of service types
    min_amount DECIMAL(10,2) DEFAULT 0,
    max_discount_amount DECIMAL(10,2),
    valid_from DATE NOT NULL,
    valid_to DATE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS daily_collections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID NOT NULL,
    collection_date DATE NOT NULL,
    consultation_amount DECIMAL(10,2) DEFAULT 0,
    lab_amount DECIMAL(10,2) DEFAULT 0,
    pharmacy_amount DECIMAL(10,2) DEFAULT 0,
    procedure_amount DECIMAL(10,2) DEFAULT 0,
    total_amount DECIMAL(10,2) DEFAULT 0,
    cash_amount DECIMAL(10,2) DEFAULT 0,
    card_amount DECIMAL(10,2) DEFAULT 0,
    insurance_amount DECIMAL(10,2) DEFAULT 0,
    outstanding_amount DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(clinic_id, collection_date)
);

-- Insurance Provider Master Tables
CREATE TABLE IF NOT EXISTS insurance_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID NOT NULL,
    provider_name VARCHAR(255) NOT NULL,
    provider_code VARCHAR(50),
    contact_details JSONB DEFAULT '{}',
    consultation_covered BOOLEAN DEFAULT FALSE,
    medicines_covered BOOLEAN DEFAULT FALSE,
    lab_tests_covered BOOLEAN DEFAULT FALSE,
    coverage_percentage DECIMAL(5,2) DEFAULT 0,
    max_coverage_amount DECIMAL(10,2),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS patient_insurance (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_id UUID NOT NULL,
    provider_id UUID REFERENCES insurance_providers(id) ON DELETE CASCADE,
    policy_number VARCHAR(100) NOT NULL,
    policy_holder_name VARCHAR(255),
    relationship_to_patient VARCHAR(50), -- self, spouse, child, parent
    coverage_start_date DATE,
    coverage_end_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS insurance_claims (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_id UUID NOT NULL,
    provider_id UUID REFERENCES insurance_providers(id) ON DELETE CASCADE,
    appointment_id UUID REFERENCES appointments(id) ON DELETE CASCADE,
    claim_number VARCHAR(100) UNIQUE NOT NULL,
    claim_amount DECIMAL(10,2) NOT NULL,
    covered_amount DECIMAL(10,2) DEFAULT 0,
    patient_payable DECIMAL(10,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending', -- pending, submitted, approved, rejected, paid
    submission_date DATE,
    approval_date DATE,
    rejection_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Reports & Analytics Tables
CREATE TABLE IF NOT EXISTS analytics_daily_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID NOT NULL,
    stat_date DATE NOT NULL,
    total_patients INTEGER DEFAULT 0,
    new_patients INTEGER DEFAULT 0,
    total_appointments INTEGER DEFAULT 0,
    completed_appointments INTEGER DEFAULT 0,
    cancelled_appointments INTEGER DEFAULT 0,
    total_revenue DECIMAL(10,2) DEFAULT 0,
    consultation_revenue DECIMAL(10,2) DEFAULT 0,
    lab_revenue DECIMAL(10,2) DEFAULT 0,
    pharmacy_revenue DECIMAL(10,2) DEFAULT 0,
    avg_wait_time_minutes INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(clinic_id, stat_date)
);

CREATE TABLE IF NOT EXISTS analytics_doctor_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID NOT NULL,
    doctor_id UUID NOT NULL,
    stat_date DATE NOT NULL,
    total_appointments INTEGER DEFAULT 0,
    completed_appointments INTEGER DEFAULT 0,
    avg_consultation_time_minutes INTEGER DEFAULT 0,
    total_revenue DECIMAL(10,2) DEFAULT 0,
    patient_satisfaction_score DECIMAL(3,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(clinic_id, doctor_id, stat_date)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_queues_clinic_id ON queues(clinic_id);
CREATE INDEX IF NOT EXISTS idx_queues_queue_type ON queues(queue_type);
CREATE INDEX IF NOT EXISTS idx_queue_tokens_queue_id ON queue_tokens(queue_id);
CREATE INDEX IF NOT EXISTS idx_queue_tokens_status ON queue_tokens(status);
CREATE INDEX IF NOT EXISTS idx_pharmacy_inventory_clinic_id ON pharmacy_inventory(clinic_id);
CREATE INDEX IF NOT EXISTS idx_pharmacy_billing_clinic_id ON pharmacy_billing(clinic_id);
CREATE INDEX IF NOT EXISTS idx_lab_tests_clinic_id ON lab_tests(clinic_id);
CREATE INDEX IF NOT EXISTS idx_lab_orders_clinic_id ON lab_orders(clinic_id);
CREATE INDEX IF NOT EXISTS idx_lab_results_order_id ON lab_results(order_id);
CREATE INDEX IF NOT EXISTS idx_insurance_claims_status ON insurance_claims(status);
CREATE INDEX IF NOT EXISTS idx_analytics_daily_stats_stat_date ON analytics_daily_stats(stat_date);
