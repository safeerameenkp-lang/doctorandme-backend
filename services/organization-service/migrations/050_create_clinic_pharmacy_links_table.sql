-- Migration 050: Create clinic_pharmacy_links table for direct linking

CREATE TABLE IF NOT EXISTS clinic_pharmacy_links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE,
    pharmacy_id UUID REFERENCES pharmacies(id) ON DELETE CASCADE,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE(clinic_id, pharmacy_id)
);

CREATE INDEX IF NOT EXISTS idx_clinic_pharmacy_links_clinic_id ON clinic_pharmacy_links(clinic_id);
CREATE INDEX IF NOT EXISTS idx_clinic_pharmacy_links_pharmacy_id ON clinic_pharmacy_links(pharmacy_id);

-- Add updated_at trigger matching existing trigger patterns
DROP TRIGGER IF EXISTS update_clinic_pharmacy_links_updated_at ON clinic_pharmacy_links;
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_clinic_pharmacy_links_updated_at
    BEFORE UPDATE ON clinic_pharmacy_links
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
