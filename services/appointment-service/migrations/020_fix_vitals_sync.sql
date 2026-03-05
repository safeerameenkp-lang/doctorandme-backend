-- Migration 020: Fix Vitals Precision and Column Sync
-- This migration ensures patient_vitals matches the frontend's extended data model and prevents numeric overflows.

-- 1. Correct temperature and weight precision to avoid overflows
ALTER TABLE patient_vitals 
ALTER COLUMN temperature TYPE NUMERIC(8,2),
ALTER COLUMN weight_kg TYPE NUMERIC(8,2);

-- 2. Add any missing columns that might have been skipped in previous manual syncs
-- (These already exist in current psql view but we include them here for completeness)
DO $$ 
BEGIN 
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
END $$;
