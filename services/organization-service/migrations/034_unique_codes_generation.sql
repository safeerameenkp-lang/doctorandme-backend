-- Make clinic_code unique and indexed across the clinics table
ALTER TABLE clinics ADD CONSTRAINT unique_clinic_code UNIQUE (clinic_code);
CREATE INDEX IF NOT EXISTS idx_clinics_clinic_code ON clinics(clinic_code);

-- Make doctor_code unique and indexed across the doctors table
ALTER TABLE doctors ADD CONSTRAINT unique_doctor_code UNIQUE (doctor_code);
CREATE INDEX IF NOT EXISTS idx_doctors_doctor_code ON doctors(doctor_code);
