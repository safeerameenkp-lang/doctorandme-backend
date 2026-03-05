-- Truncate all tables and delete all data
-- This will keep table structure but remove all records

-- Disable triggers and constraints temporarily
SET session_replication_role = 'replica';

-- Truncate all tables in order (respecting dependencies)
TRUNCATE TABLE refresh_tokens CASCADE;
TRUNCATE TABLE user_roles CASCADE;
TRUNCATE TABLE appointments CASCADE;
TRUNCATE TABLE patient_vitals CASCADE;
TRUNCATE TABLE doctor_leaves CASCADE;
TRUNCATE TABLE doctor_time_slots CASCADE;
TRUNCATE TABLE clinic_doctor_links CASCADE;
TRUNCATE TABLE doctor_clinic_fees CASCADE;
TRUNCATE TABLE patients CASCADE;
TRUNCATE TABLE doctors CASCADE;
TRUNCATE TABLE clinics CASCADE;
TRUNCATE TABLE departments CASCADE;
TRUNCATE TABLE organizations CASCADE;
TRUNCATE TABLE users CASCADE;
TRUNCATE TABLE roles CASCADE;
TRUNCATE TABLE audit_logs CASCADE;

-- Re-enable triggers and constraints
SET session_replication_role = 'origin';

-- Reset sequences (auto-increment counters) if any
-- (UUIDs don't need this, but just in case)

SELECT 'All tables truncated successfully!' as message;

