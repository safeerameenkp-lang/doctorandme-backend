-- Add leave_duration column to doctor_leaves
ALTER TABLE doctor_leaves ADD COLUMN IF NOT EXISTS leave_duration VARCHAR(20) DEFAULT 'full_day';
COMMENT ON COLUMN doctor_leaves.leave_duration IS 'Duration of leave: morning, afternoon, full_day';
