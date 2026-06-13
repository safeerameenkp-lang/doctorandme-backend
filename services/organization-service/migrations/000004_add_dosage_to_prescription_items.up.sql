ALTER TABLE sales_schema.prescription_items
ADD COLUMN duration_days INT DEFAULT 0,
ADD COLUMN dosage_per_day INT DEFAULT 0,
ADD COLUMN morning BOOLEAN DEFAULT false,
ADD COLUMN noon BOOLEAN DEFAULT false,
ADD COLUMN night BOOLEAN DEFAULT false;
