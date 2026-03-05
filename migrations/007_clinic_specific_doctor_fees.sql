-- Migration 007: Clinic-Specific Doctor Fees
-- Move fees from doctors table to clinic_doctor_links table
-- This allows each clinic to set different fees (offline/online) for the same doctor

-- Step 1: Add fee columns to clinic_doctor_links
ALTER TABLE clinic_doctor_links
ADD COLUMN consultation_fee_offline DECIMAL(10,2),
ADD COLUMN consultation_fee_online DECIMAL(10,2),
ADD COLUMN follow_up_fee DECIMAL(10,2),
ADD COLUMN follow_up_days INTEGER DEFAULT 7,
ADD COLUMN notes TEXT,
ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Step 2: Make fees in doctors table nullable (for backward compatibility)
-- These will be used as default/base fees if clinic doesn't override
ALTER TABLE doctors
ALTER COLUMN consultation_fee DROP NOT NULL,
ALTER COLUMN follow_up_fee DROP NOT NULL;

-- Step 3: Migrate existing fees from doctors to clinic_doctor_links
-- For doctors already linked to clinics, copy their fees
UPDATE clinic_doctor_links cdl
SET 
    consultation_fee_offline = d.consultation_fee,
    consultation_fee_online = d.consultation_fee,
    follow_up_fee = d.follow_up_fee,
    follow_up_days = d.follow_up_days
FROM doctors d
WHERE cdl.doctor_id = d.id
AND cdl.consultation_fee_offline IS NULL;

-- Step 4: Add trigger for updated_at
DROP TRIGGER IF EXISTS update_clinic_doctor_links_updated_at ON clinic_doctor_links;
CREATE TRIGGER update_clinic_doctor_links_updated_at 
    BEFORE UPDATE ON clinic_doctor_links
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Step 5: Create index for better performance
CREATE INDEX IF NOT EXISTS idx_clinic_doctor_links_clinic_id ON clinic_doctor_links(clinic_id);
CREATE INDEX IF NOT EXISTS idx_clinic_doctor_links_doctor_id ON clinic_doctor_links(doctor_id);

-- Add comments for documentation
COMMENT ON COLUMN clinic_doctor_links.consultation_fee_offline IS 'Fee for offline (in-person) consultation at this clinic';
COMMENT ON COLUMN clinic_doctor_links.consultation_fee_online IS 'Fee for online (telemedicine) consultation at this clinic';
COMMENT ON COLUMN clinic_doctor_links.follow_up_fee IS 'Follow-up consultation fee at this clinic';
COMMENT ON COLUMN clinic_doctor_links.follow_up_days IS 'Number of days for follow-up validity at this clinic';
COMMENT ON COLUMN clinic_doctor_links.notes IS 'Clinic-specific notes about this doctor';

