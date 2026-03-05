# Reschedule `updated_at` Column Fix

## Error
```
Request failed with status 500: {
  "details": "pq: column \"updated_at\" of relation \"appointments\" does not exist",
  "error": "Failed to update appointment"
}
```

## Root Cause
The `appointments` table does not have an `updated_at` column in the database schema, but the reschedule API was trying to update it.

## Database Schema Analysis

### Appointments Table (from migration 001_initial_schema.sql)
```sql
CREATE TABLE appointments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_id UUID REFERENCES patients(id) ON DELETE CASCADE,
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE,
    doctor_id UUID REFERENCES doctors(id) ON DELETE CASCADE,
    booking_number VARCHAR(50) UNIQUE NOT NULL,
    appointment_time TIMESTAMP NOT NULL,
    duration_minutes INTEGER DEFAULT 12,
    consultation_type VARCHAR(20) DEFAULT 'new',
    status VARCHAR(20) DEFAULT 'booked',
    fee_amount DECIMAL(10,2),
    payment_status VARCHAR(20) DEFAULT 'pending',
    payment_mode VARCHAR(20),
    is_priority BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    -- ❌ NO updated_at column
);
```

### Comparison: doctor_individual_slots Table (HAS updated_at)
```sql
CREATE TABLE IF NOT EXISTS doctor_individual_slots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID REFERENCES doctor_slot_sessions(id) ON DELETE CASCADE NOT NULL,
    slot_start TIME NOT NULL,
    slot_end TIME NOT NULL,
    is_booked BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'available',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- ✅ HAS updated_at
    ...
);

-- ✅ Auto-update trigger
CREATE TRIGGER update_doctor_individual_slots_updated_at 
    BEFORE UPDATE ON doctor_individual_slots
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
```

## Solution

### Option 1: Remove `updated_at` from Query (CHOSEN) ✅

**Reason**: Simpler, doesn't require database migration

```go
// Before (❌ Error)
result, err := tx.Exec(`
    UPDATE appointments SET
        doctor_id = $1,
        department_id = $2,
        individual_slot_id = $3,
        appointment_date = $4,
        appointment_time = $5,
        reason = $6,
        notes = $7,
        updated_at = CURRENT_TIMESTAMP  -- ❌ Column doesn't exist
    WHERE id = $8
`, ...)

// After (✅ Fixed)
result, err := tx.Exec(`
    UPDATE appointments SET
        doctor_id = $1,
        department_id = $2,
        individual_slot_id = $3,
        appointment_date = $4,
        appointment_time = $5,
        reason = $6,
        notes = $7
    WHERE id = $8
`, ...)
```

### Option 2: Add `updated_at` Column (Alternative)

If you want to track when appointments are modified, you can add a migration:

```sql
-- Migration: 023_add_updated_at_to_appointments.sql

-- Add updated_at column
ALTER TABLE appointments 
ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Create trigger function if not exists
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
CREATE TRIGGER update_appointments_updated_at
    BEFORE UPDATE ON appointments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON COLUMN appointments.updated_at IS 'Timestamp when appointment was last modified';
```

## Impact

### What Changed
- ✅ Removed `updated_at` from UPDATE query
- ✅ No database schema changes needed
- ✅ Reschedule API works correctly now

### What Still Works
- ✅ `created_at` column tracks when appointment was created
- ✅ Slot tables have `updated_at` and continue to work
- ✅ All other appointment operations unaffected

## Files Modified
- ✅ `services/appointment-service/controllers/appointment_list_simple.controller.go`

## Testing

### Test 1: Reschedule Should Work Now
```bash
POST /api/v1/appointments/simple/APPOINTMENT_ID/reschedule
{
  "doctor_id": "doctor-uuid",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2024-10-18",
  "appointment_time": "2024-10-18 11:00:00"
}

# Expected Response: Success ✅
{
  "success": true,
  "message": "Appointment rescheduled successfully",
  "appointment": {
    "id": "appointment-id",
    "booking_number": "BK20241017001",
    ...
  }
}
```

### Test 2: Verify Appointment Updated
```sql
-- Check appointment was updated
SELECT id, doctor_id, individual_slot_id, appointment_time, created_at
FROM appointments 
WHERE id = 'appointment-id';

-- Expected: doctor_id, slot_id, and time should be updated
-- created_at remains unchanged (tracks original creation)
```

### Test 3: Verify Slots Updated
```sql
-- Check old slot was freed
SELECT available_count, is_booked, status, updated_at
FROM doctor_individual_slots 
WHERE id = 'old-slot-id';
-- Expected: available_count increased, updated_at changed ✅

-- Check new slot was booked
SELECT available_count, is_booked, status, updated_at
FROM doctor_individual_slots 
WHERE id = 'new-slot-id';
-- Expected: available_count decreased, updated_at changed ✅
```

## Comparison: Tables with vs without updated_at

| Table | Has updated_at? | Auto-update Trigger? |
|-------|----------------|---------------------|
| `appointments` | ❌ No | ❌ No |
| `patients` | ✅ Yes | ❌ No |
| `doctor_individual_slots` | ✅ Yes | ✅ Yes |
| `doctor_slot_sessions` | ✅ Yes | ✅ Yes |
| `doctor_time_slots` | ❌ No | ❌ No |

## Recommendations

### Current Approach (Good for Now) ✅
- Use `created_at` to track when appointment was created
- Slot tables have `updated_at` for tracking slot modifications
- Reschedule works without schema changes

### Future Enhancement (Optional)
If you need to track appointment modifications:
1. Add `updated_at` column to appointments table
2. Add trigger to auto-update it
3. Update queries to include `updated_at`

## Summary

| Issue | Solution |
|-------|----------|
| **Error** | `updated_at` column doesn't exist in appointments table |
| **Cause** | Code tried to update non-existent column |
| **Fix** | Removed `updated_at` from UPDATE query |
| **Impact** | Reschedule now works correctly |
| **Status** | ✅ FIXED |

---

**Error**: `column "updated_at" of relation "appointments" does not exist`  
**Fix**: Removed `updated_at` from UPDATE query  
**Result**: Reschedule API works correctly ✅
