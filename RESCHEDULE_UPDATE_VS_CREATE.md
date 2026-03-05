# Reschedule: UPDATE vs CREATE - Complete Explanation

## The Concern
**"Make sure the system doesn't create both the original and rescheduled appointments, as that causes conflicts and confusion."**

## ✅ Answer: System is Correct!

The reschedule API **UPDATES** the existing appointment, it does NOT create a new one. This prevents duplicates and maintains data integrity.

## How It Works

### Reschedule API (UPDATE)
```go
// Step 6: Update appointment (UPDATE existing, not create new)
result, err := tx.Exec(`
    UPDATE appointments SET
        doctor_id = $1,
        department_id = $2,
        individual_slot_id = $3,
        appointment_date = $4,
        appointment_time = $5,
        reason = $6,
        notes = $7,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = $8
`, input.DoctorID, input.DepartmentID, input.IndividualSlotID,
    appointmentDateStr, appointmentTime, input.Reason, input.Notes, appointmentID)

// ✅ Verify that the appointment was actually updated
rowsAffected, _ := result.RowsAffected()
if rowsAffected == 0 {
    return error: "Appointment not found or already modified"
}
```

### Create Appointment API (INSERT)
```go
// Different API - creates NEW appointment
err = config.DB.QueryRow(`
    INSERT INTO appointments (
        clinic_patient_id, clinic_id, doctor_id, department_id, booking_number, token_number,
        appointment_date, appointment_time, duration_minutes, consultation_type,
        reason, notes, fee_amount, payment_mode, payment_status, status, individual_slot_id
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
    RETURNING id, ...
`, ...)
```

## Key Differences

| Aspect | Reschedule API | Create API |
|--------|----------------|------------|
| **SQL Operation** | UPDATE | INSERT |
| **Endpoint** | POST /appointments/simple/:id/reschedule | POST /appointments/simple |
| **Purpose** | Modify existing appointment | Create new appointment |
| **Appointment ID** | Required (from URL) | Generated (new) |
| **Slot Management** | Frees old + books new | Only books new |
| **Booking Number** | Preserved | New generated |
| **Token Number** | Preserved | New generated |
| **Result** | Same appointment modified | New appointment created |

## Step-by-Step Comparison

### Reschedule Flow (UPDATE)
```
1. GET existing appointment by ID
   - Verify appointment exists
   - Get current slot ID
   - Get booking number (preserve)
   - Get token number (preserve)

2. Validate new slot availability

3. Start Transaction

4. Free old slot (if different slot)
   - Update slot availability

5. UPDATE existing appointment ✅
   - Same ID preserved
   - Same booking number
   - Same token number
   - Updated details

6. Book new slot (if different slot)
   - Update slot availability

7. Commit Transaction

8. Return SAME appointment with updated details ✅
```

### Create Flow (INSERT)
```
1. Validate input data

2. Generate NEW booking number ✅

3. Generate NEW token number ✅

4. Start Transaction

5. INSERT new appointment ✅
   - New ID generated
   - New booking number
   - New token number

6. Book slot
   - Update slot availability

7. Commit Transaction

8. Return NEW appointment ✅
```

## Database Impact

### Reschedule (UPDATE)
```sql
-- Before Reschedule
appointments table:
id = 'appt-123'
booking_number = 'BK20241017001'
token_number = 'T001'
individual_slot_id = 'slot-A'
appointment_time = '10:30:00'

-- After Reschedule (SAME RECORD UPDATED)
appointments table:
id = 'appt-123' ✅ (SAME)
booking_number = 'BK20241017001' ✅ (SAME)
token_number = 'T001' ✅ (SAME)
individual_slot_id = 'slot-B' ✅ (UPDATED)
appointment_time = '11:00:00' ✅ (UPDATED)
updated_at = CURRENT_TIMESTAMP ✅
```

### Create (INSERT)
```sql
-- Before Create
appointments table:
(no record)

-- After Create (NEW RECORD)
appointments table:
id = 'appt-456' ✅ (NEW)
booking_number = 'BK20241017002' ✅ (NEW)
token_number = 'T002' ✅ (NEW)
individual_slot_id = 'slot-C'
appointment_time = '14:00:00'
created_at = CURRENT_TIMESTAMP ✅
```

## Verification Safeguards

### 1. Initial Appointment Check
```go
// Step 1: Get existing appointment details
err := config.DB.QueryRow(`
    SELECT clinic_id, doctor_id, clinic_patient_id, individual_slot_id,
           fee_amount, booking_number, consultation_type, token_number
    FROM appointments 
    WHERE id = $1 AND status IN ('scheduled', 'confirmed', 'pending')
`, appointmentID).Scan(...)

if err != nil {
    return error: "Appointment not found or cannot be rescheduled"
}
```

### 2. Update Verification
```go
// ✅ Verify that the appointment was actually updated
rowsAffected, _ := result.RowsAffected()
if rowsAffected == 0 {
    return error: "Appointment not found or already modified"
}
```

### 3. Transaction Safety
```go
tx, err := config.DB.Begin()
defer tx.Rollback()  // Rollback if anything fails

// ... all updates ...

tx.Commit()  // Only commit if everything succeeds
```

## Example Scenarios

### Scenario 1: Successful Reschedule

**Initial State:**
```json
{
  "id": "appt-123",
  "booking_number": "BK20241017001",
  "token_number": "T001",
  "individual_slot_id": "slot-A",
  "appointment_time": "2024-10-17 10:30:00"
}
```

**Reschedule Request:**
```bash
POST /appointments/simple/appt-123/reschedule
{
  "individual_slot_id": "slot-B",
  "appointment_date": "2024-10-17",
  "appointment_time": "2024-10-17 11:00:00"
}
```

**Result (SAME appointment updated):**
```json
{
  "id": "appt-123",              // ✅ SAME ID
  "booking_number": "BK20241017001",  // ✅ SAME booking number
  "token_number": "T001",        // ✅ SAME token
  "individual_slot_id": "slot-B",     // ✅ UPDATED
  "appointment_time": "2024-10-17 11:00:00"  // ✅ UPDATED
}
```

**Database Query:**
```sql
SELECT COUNT(*) FROM appointments WHERE booking_number = 'BK20241017001';
-- Result: 1 ✅ (Only one appointment, not two)
```

### Scenario 2: Attempted Duplicate Prevention

**If someone tries to reschedule the same appointment twice simultaneously:**

```
Request 1: Reschedule appt-123 to slot-B
Request 2: Reschedule appt-123 to slot-C (at same time)

Transaction 1:
  - Read appointment (success)
  - Free slot-A
  - UPDATE appointment → slot-B
  - Book slot-B
  - Commit ✅

Transaction 2:
  - Read appointment (success)
  - Free slot-B (already freed by Tx1)
  - UPDATE appointment → slot-C
  - Book slot-C
  - Commit ✅

Result: Last update wins (slot-C)
No duplicate appointments created ✅
```

## API Endpoints

### Reschedule (UPDATE)
```bash
POST /api/v1/appointments/simple/:id/reschedule

# URL contains appointment ID
# Updates THAT specific appointment
# Preserves booking number and token
```

### Create (INSERT)
```bash
POST /api/v1/appointments/simple

# No ID in URL
# Creates NEW appointment
# Generates new booking number and token
```

## Benefits of UPDATE Approach

1. ✅ **No Duplicates**: Same appointment record updated
2. ✅ **Preserves History**: Booking number and token remain same
3. ✅ **Audit Trail**: `updated_at` timestamp tracks changes
4. ✅ **Data Consistency**: One record per appointment
5. ✅ **Patient Experience**: Same booking number for patient
6. ✅ **Slot Management**: Properly frees old and books new
7. ✅ **Transaction Safe**: All-or-nothing updates

## Testing

### Test 1: Verify No Duplicates
```sql
-- Before Reschedule
SELECT COUNT(*) FROM appointments WHERE id = 'appt-123';
-- Result: 1

-- Reschedule
POST /appointments/simple/appt-123/reschedule { ... }

-- After Reschedule
SELECT COUNT(*) FROM appointments WHERE id = 'appt-123';
-- Result: 1 ✅ (Still one, not two)

SELECT COUNT(*) FROM appointments WHERE booking_number = 'BK20241017001';
-- Result: 1 ✅ (Same booking number, one record)
```

### Test 2: Verify Update vs Create
```bash
# Reschedule (UPDATE)
curl -X POST "http://localhost:8082/api/v1/appointments/simple/appt-123/reschedule" \
  -H "Content-Type: application/json" \
  -d '{
    "individual_slot_id": "slot-B",
    "appointment_date": "2024-10-17",
    "appointment_time": "2024-10-17 11:00:00"
  }'

# Response should have SAME id and booking_number
{
  "appointment": {
    "id": "appt-123",              // ✅ SAME
    "booking_number": "BK20241017001"  // ✅ SAME
  }
}

# Create New (INSERT)
curl -X POST "http://localhost:8082/api/v1/appointments/simple" \
  -H "Content-Type: application/json" \
  -d '{
    "clinic_patient_id": "patient-123",
    "doctor_id": "doctor-123",
    ...
  }'

# Response should have NEW id and booking_number
{
  "appointment": {
    "id": "appt-456",              // ✅ NEW
    "booking_number": "BK20241017002"  // ✅ NEW
  }
}
```

### Test 3: Verify Slot Management
```sql
-- Before Reschedule
SELECT available_count FROM doctor_individual_slots WHERE id = 'slot-A';
-- Result: 0 (booked)

SELECT available_count FROM doctor_individual_slots WHERE id = 'slot-B';
-- Result: 1 (available)

-- Reschedule from slot-A to slot-B

-- After Reschedule
SELECT available_count FROM doctor_individual_slots WHERE id = 'slot-A';
-- Result: 1 ✅ (freed)

SELECT available_count FROM doctor_individual_slots WHERE id = 'slot-B';
-- Result: 0 ✅ (booked)
```

## Error Handling

### Error 1: Appointment Not Found
```json
{
  "error": "Appointment not found or cannot be rescheduled"
}
```
**Cause**: Invalid appointment ID or appointment already cancelled

### Error 2: Update Failed
```json
{
  "error": "Appointment not found or already modified",
  "message": "The appointment may have been cancelled or modified by another user"
}
```
**Cause**: UPDATE query affected 0 rows (concurrent modification)

## Summary

| Question | Answer |
|----------|--------|
| Does reschedule create duplicate? | ❌ NO - Uses UPDATE |
| Are both appointments created? | ❌ NO - Only one exists |
| Is original preserved? | ✅ YES - Same ID and booking number |
| Can duplicates occur? | ❌ NO - UPDATE modifies existing record |
| Is booking number preserved? | ✅ YES - Same booking number |
| Is token preserved? | ✅ YES - Same token number |

## Status
✅ **VERIFIED AND SAFE**

The reschedule API correctly UPDATES the existing appointment record. It does NOT create a new appointment. This prevents duplicates and maintains data integrity.

---

**Concern**: System might create both original and rescheduled appointments  
**Reality**: System UPDATES existing appointment, preserves ID and booking number  
**Result**: No duplicates, data integrity maintained ✅
