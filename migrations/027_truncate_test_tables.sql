-- Truncate test tables for fresh start
-- This clears all data from appointments, follow_ups, and clinic_patients tables

-- Truncate appointments first (due to foreign key constraints)
TRUNCATE TABLE appointments CASCADE;

-- Truncate follow_ups table
TRUNCATE TABLE follow_ups CASCADE;

-- Truncate clinic_patients table
TRUNCATE TABLE clinic_patients CASCADE;

-- Reset sequences if needed
ALTER SEQUENCE IF EXISTS appointments_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS follow_ups_id_seq RESTART WITH 1;

-- Note: clinic_patients uses UUID (no sequence to reset)

COMMENT ON EXTENSION "unaccent" IS 'Extension for unaccent text search';

