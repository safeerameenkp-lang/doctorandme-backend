-- Migration to make last_name optional in clinic_patients
ALTER TABLE clinic_patients ALTER COLUMN last_name DROP NOT NULL;
