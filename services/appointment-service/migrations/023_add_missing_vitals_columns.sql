-- Migration 023: Add clinic_patient_id to patient_vitals
-- This column was missing in the patient_vitals table despite being used in the code.

DO $$ 
BEGIN 
    -- 1. Add clinic_patient_id column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='patient_vitals' AND column_name='clinic_patient_id') THEN
        ALTER TABLE patient_vitals ADD COLUMN clinic_patient_id UUID;
        
        -- Add foreign key constraint if clinic_patients table exists (cross-service visibility)
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='clinic_patients') THEN
            ALTER TABLE patient_vitals 
            ADD CONSTRAINT fk_vitals_clinic_patient 
            FOREIGN KEY (clinic_patient_id) REFERENCES clinic_patients(id) ON DELETE SET NULL;
        END IF;
    END IF;

    -- 2. Add any other missing columns from current data model (defensive sync)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='patient_vitals' AND column_name='blood_pressure') THEN
        ALTER TABLE patient_vitals ADD COLUMN blood_pressure VARCHAR(20);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='patient_vitals' AND column_name='resp_bpm') THEN
        ALTER TABLE patient_vitals ADD COLUMN resp_bpm INTEGER;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='patient_vitals' AND column_name='spo2_percent') THEN
        ALTER TABLE patient_vitals ADD COLUMN spo2_percent INTEGER;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='patient_vitals' AND column_name='sugar_mgdl') THEN
        ALTER TABLE patient_vitals ADD COLUMN sugar_mgdl NUMERIC(10,2);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='patient_vitals' AND column_name='bmi') THEN
        ALTER TABLE patient_vitals ADD COLUMN bmi NUMERIC(6,2);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='patient_vitals' AND column_name='smoking_status') THEN
        ALTER TABLE patient_vitals ADD COLUMN smoking_status VARCHAR(50);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='patient_vitals' AND column_name='alcohol_use') THEN
        ALTER TABLE patient_vitals ADD COLUMN alcohol_use VARCHAR(50);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='patient_vitals' AND column_name='notes') THEN
        ALTER TABLE patient_vitals ADD COLUMN notes TEXT;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='patient_vitals' AND column_name='updated_at') THEN
        ALTER TABLE patient_vitals ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
    END IF;
END $$;

-- 3. Ensure indexes exist
CREATE INDEX IF NOT EXISTS idx_patient_vitals_clinic_patient_id ON patient_vitals (clinic_patient_id);
CREATE INDEX IF NOT EXISTS idx_patient_vitals_appointment_id ON patient_vitals (appointment_id);

COMMENT ON COLUMN patient_vitals.clinic_patient_id IS 'Reference to clinic-specific patient (alternative to global patient_id tracked via appointment)';
