CREATE SCHEMA IF NOT EXISTS sales_schema;

CREATE TABLE IF NOT EXISTS sales_schema.sales (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    sale_type VARCHAR(20) NOT NULL,
    prescription_id VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    total_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    invoice_number VARCHAR(50) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sales_schema.sale_items (
    id UUID PRIMARY KEY,
    sale_id UUID NOT NULL REFERENCES sales_schema.sales(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    batch_id UUID NOT NULL,
    quantity INTEGER NOT NULL,
    price DECIMAL(15, 2) NOT NULL,
    reservation_id VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sales_schema.payments (
    id UUID PRIMARY KEY,
    sale_id UUID NOT NULL REFERENCES sales_schema.sales(id) ON DELETE CASCADE,
    mode VARCHAR(20) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for tenant-based lookups
CREATE INDEX IF NOT EXISTS idx_sales_tenant ON sales_schema.sales(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sales_items_sale ON sales_schema.sale_items(sale_id);
