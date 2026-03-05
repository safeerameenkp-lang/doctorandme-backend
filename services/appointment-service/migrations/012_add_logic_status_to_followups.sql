-- Add follow_up_logic_status and logic_notes to follow_ups table for better tracking

ALTER TABLE follow_ups
ADD COLUMN IF NOT EXISTS follow_up_logic_status VARCHAR(20) DEFAULT 'new' 
CHECK (follow_up_logic_status IN ('new', 'expired', 'used', 'renewed')),
ADD COLUMN IF NOT EXISTS logic_notes TEXT;

-- Update existing records
UPDATE follow_ups
SET follow_up_logic_status = CASE
    WHEN status = 'used' THEN 'used'
    WHEN status = 'expired' THEN 'expired'
    WHEN status = 'renewed' THEN 'renewed'
    WHEN status = 'active' THEN 'new'
    ELSE 'new'
END;

-- Add index for performance
CREATE INDEX IF NOT EXISTS idx_follow_ups_logic_status ON follow_ups(follow_up_logic_status);

COMMENT ON COLUMN follow_ups.follow_up_logic_status IS 'Logic status: new (active), used (consumed), expired (5 days passed), renewed (replaced by new follow-up)';
COMMENT ON COLUMN follow_ups.logic_notes IS 'Optional notes for debugging or production logs';

