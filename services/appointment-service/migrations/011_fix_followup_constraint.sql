-- Fix follow_ups constraint to allow future appointments
-- The constraint was too restrictive and prevented creating follow-ups for future appointments

-- Drop the restrictive constraint
ALTER TABLE follow_ups DROP CONSTRAINT IF EXISTS chk_follow_ups_no_future_dates;

-- Add a more reasonable constraint
-- Allow valid_from to be up to 30 days in the future
-- And valid_until to be up to 60 days in the future
ALTER TABLE follow_ups 
ADD CONSTRAINT chk_follow_ups_valid_dates 
CHECK (
  valid_from >= CURRENT_DATE - INTERVAL '30 days'
  AND valid_from <= CURRENT_DATE + INTERVAL '60 days'
  AND valid_until >= valid_from
  AND valid_until <= CURRENT_DATE + INTERVAL '90 days'
);

COMMENT ON CONSTRAINT chk_follow_ups_valid_dates ON follow_ups IS 'Ensures follow-up dates are reasonable but allows future appointments';

