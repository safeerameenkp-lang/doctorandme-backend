-- Migration to change token_number from INTEGER to VARCHAR
-- To support alphanumeric tokens like "T01"

ALTER TABLE appointments 
ALTER COLUMN token_number TYPE VARCHAR(20);

COMMENT ON COLUMN appointments.token_number IS 'Alphanumeric token number for queue management (per doctor per clinic per date)';
