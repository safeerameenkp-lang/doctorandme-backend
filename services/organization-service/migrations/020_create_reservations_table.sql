CREATE TABLE IF NOT EXISTS inventory.reservations (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    product_id UUID NOT NULL REFERENCES inventory.medicines(id),
    batch_id UUID NOT NULL REFERENCES inventory.batches(id),
    quantity INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_reservations_batch ON inventory.reservations(batch_id, tenant_id);
CREATE INDEX idx_reservations_expires ON inventory.reservations(expires_at);
