-- 000022_create_sales_returns_table.up.sql
CREATE TABLE IF NOT EXISTS sales_schema.sales_returns (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    sale_id UUID NOT NULL REFERENCES sales_schema.sales(id),
    return_number VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL,
    total_refund DECIMAL(15, 2) NOT NULL DEFAULT 0,
    reason TEXT,
    handled_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sales_schema.sales_return_items (
    id UUID PRIMARY KEY,
    return_id UUID NOT NULL REFERENCES sales_schema.sales_returns(id),
    sale_item_id UUID NOT NULL REFERENCES sales_schema.sale_items(id),
    product_id UUID NOT NULL,
    medicine_name VARCHAR(255) NOT NULL,
    batch_id UUID NOT NULL,
    batch_no VARCHAR(100),
    quantity INTEGER NOT NULL,
    refund_amount DECIMAL(15, 2) NOT NULL,
    condition VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sales_returns_sale_id ON sales_schema.sales_returns(sale_id);
CREATE INDEX idx_sales_returns_tenant_id ON sales_schema.sales_returns(tenant_id);
