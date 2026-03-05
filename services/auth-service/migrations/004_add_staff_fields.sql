-- Add department and joining_date to staff table
ALTER TABLE staff ADD COLUMN IF NOT EXISTS department VARCHAR(100);
ALTER TABLE staff ADD COLUMN IF NOT EXISTS joining_date DATE;

-- Note: user_roles is also used to map users to roles and clinics in the auth system.
-- we will sync the staff_type to user_roles
