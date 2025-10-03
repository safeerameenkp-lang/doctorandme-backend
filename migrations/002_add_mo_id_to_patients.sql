-- Add MO ID field to patients table
ALTER TABLE patients ADD COLUMN mo_id VARCHAR(50);

-- Create index for MO ID for better performance
CREATE INDEX idx_patients_mo_id ON patients(mo_id);

-- Add comment to explain the field
COMMENT ON COLUMN patients.mo_id IS 'Medical Officer ID - unique identifier for the patient assigned by medical officer';
