# Reschedule NULL Scan Error Fix

## Error
```
Request failed with status 500: {
  "details": "sql: Scan error on column index 2, name \"mo_id\": converting NULL to string is unsupported",
  "error": "Failed to retrieve updated appointment"
}
```

## Root Cause
The reschedule API was trying to scan nullable database columns (`mo_id`, `token_number`) into non-pointer Go string types. When these fields are NULL in the database, Go's SQL scanner cannot convert NULL to a string value.

## Database Schema

### Fields That Can Be NULL:
```sql
-- clinic_patients table
mo_id VARCHAR(50)  -- ❌ Can be NULL (no NOT NULL constraint)

-- appointments table
token_number VARCHAR(20)  -- ❌ Can be NULL (no NOT NULL constraint)
department_id UUID  -- ❌ Can be NULL (optional)
consultation_type VARCHAR(20)  -- ❌ Can be NULL (optional)
appointment_date DATE  -- ❌ Can be NULL (optional)
fee_amount DECIMAL(10,2)  -- ❌ Can be NULL (optional)
```

## The Problem

### Before (❌ Error):
```go
var (
    updatedID, updatedTokenNumber, updatedMoID, updatedPatientName, updatedDoctorName string
    // ❌ updatedTokenNumber and updatedMoID are non-pointer strings
    // ❌ Cannot handle NULL values from database
)

err = config.DB.QueryRow(`
    SELECT 
        a.id,
        a.token_number,  -- Can be NULL
        cp.mo_id,        -- Can be NULL
        ...
`).Scan(
    &updatedID,
    &updatedTokenNumber,  // ❌ Error if NULL
    &updatedMoID,         // ❌ Error if NULL
    ...
)
```

### After (✅ Fixed):
```go
var (
    updatedID, updatedPatientName, updatedDoctorName string
    updatedTokenNumber, updatedMoID *string  // ✅ Pointer types handle NULL
    updatedDepartment, updatedConsultationType *string
    updatedAppointmentDate *string
    updatedFeeAmount *float64
)

err = config.DB.QueryRow(`
    SELECT 
        a.id,
        a.token_number,  -- Can be NULL → *string ✅
        cp.mo_id,        -- Can be NULL → *string ✅
        ...
`).Scan(
    &updatedID,
    &updatedTokenNumber,  // ✅ Can handle NULL
    &updatedMoID,         // ✅ Can handle NULL
    ...
)
```

## How Go Handles NULL Values

### Non-Pointer Types (❌ Error)
```go
var name string
db.QueryRow("SELECT name FROM users WHERE id = $1", id).Scan(&name)
// If name is NULL in database → ERROR: converting NULL to string is unsupported
```

### Pointer Types (✅ Works)
```go
var name *string
db.QueryRow("SELECT name FROM users WHERE id = $1", id).Scan(&name)
// If name is NULL in database → name = nil ✅
// If name is 'John' in database → name = &"John" ✅
```

### JSON Encoding
```go
var tokenNumber *string  // nil

json.Marshal(gin.H{
    "token_number": tokenNumber,  // Outputs: "token_number": null ✅
})

var moID *string = &"MO123456"

json.Marshal(gin.H{
    "mo_id": moID,  // Outputs: "mo_id": "MO123456" ✅
})
```

## Complete Variable Mapping

| Database Column | Go Variable Type | Reason |
|----------------|------------------|--------|
| `a.id` | `string` | ✅ NOT NULL (PRIMARY KEY) |
| `a.token_number` | `*string` | ❌ Can be NULL |
| `cp.mo_id` | `*string` | ❌ Can be NULL |
| Patient name | `string` | ✅ COALESCE provides default |
| Doctor name | `string` | ✅ COALESCE provides default |
| `dept.name` | `*string` | ❌ Can be NULL (LEFT JOIN) |
| `a.consultation_type` | `*string` | ❌ Can be NULL |
| `a.appointment_date` | `*string` | ❌ Can be NULL |
| `a.appointment_time` | `time.Time` | ✅ NOT NULL |
| `a.status` | `string` | ✅ DEFAULT value |
| `a.fee_amount` | `*float64` | ❌ Can be NULL |
| `a.payment_status` | `string` | ✅ DEFAULT value |
| `a.booking_number` | `string` | ✅ NOT NULL UNIQUE |
| `a.created_at` | `time.Time` | ✅ DEFAULT CURRENT_TIMESTAMP |

## Fixed Code

### Variable Declaration
```go
// Step 9: Return updated appointment details using the same structure as GetSimpleAppointmentDetails
var (
    updatedID, updatedPatientName, updatedDoctorName                                 string
    updatedTokenNumber, updatedMoID                                                  *string  // ✅ Nullable
    updatedDepartment, updatedConsultationType                                       *string  // ✅ Nullable
    updatedAppointmentDate                                                           *string  // ✅ Nullable
    updatedAppointmentTime, updatedCreatedAt                                         time.Time
    updatedStatus, updatedBookingNumber                                              string
    updatedFeeAmount                                                                 *float64  // ✅ Nullable
    updatedPaymentStatus                                                             string
)
```

### Query and Scan
```go
err = config.DB.QueryRow(`
    SELECT 
        a.id,
        a.token_number,
        cp.mo_id,
        COALESCE(cp.first_name || ' ' || cp.last_name, cp.first_name, 'Unknown') as patient_name,
        COALESCE(u.first_name || ' ' || u.last_name, u.first_name, 'Unknown Doctor') as doctor_name,
        COALESCE(dept_appt.name, dept_doc.name) as department,
        a.consultation_type,
        a.appointment_date,
        a.appointment_time,
        a.status,
        a.fee_amount,
        a.payment_status,
        a.booking_number,
        a.created_at
    FROM appointments a
    LEFT JOIN clinic_patients cp ON cp.id = a.clinic_patient_id
    LEFT JOIN doctors d ON d.id = a.doctor_id
    LEFT JOIN users u ON u.id = d.user_id
    LEFT JOIN departments dept_appt ON dept_appt.id = a.department_id
    LEFT JOIN departments dept_doc ON dept_doc.id = d.department_id
    WHERE a.id = $1
`, appointmentID).Scan(
    &updatedID,
    &updatedTokenNumber,        // ✅ *string handles NULL
    &updatedMoID,               // ✅ *string handles NULL
    &updatedPatientName,
    &updatedDoctorName,
    &updatedDepartment,         // ✅ *string handles NULL
    &updatedConsultationType,   // ✅ *string handles NULL
    &updatedAppointmentDate,    // ✅ *string handles NULL
    &updatedAppointmentTime,
    &updatedStatus,
    &updatedFeeAmount,          // ✅ *float64 handles NULL
    &updatedPaymentStatus,
    &updatedBookingNumber,
    &updatedCreatedAt,
)
```

### JSON Response
```go
c.JSON(http.StatusOK, gin.H{
    "success": true,
    "message": "Appointment rescheduled successfully",
    "appointment": gin.H{
        "id":                    updatedID,
        "token_number":          updatedTokenNumber,          // nil → null in JSON ✅
        "mo_id":                 updatedMoID,                 // nil → null in JSON ✅
        "patient_name":          updatedPatientName,
        "doctor_name":           updatedDoctorName,
        "department":            updatedDepartment,           // nil → null in JSON ✅
        "consultation_type":     updatedConsultationType,     // nil → null in JSON ✅
        "appointment_date_time": updatedAppointmentDateTime,
        "status":                updatedStatus,
        "fee_amount":            updatedFeeAmount,            // nil → null in JSON ✅
        "payment_status":        updatedPaymentStatus,
        "fee_status":            updatedPaymentStatus,
        "booking_number":        updatedBookingNumber,
        "created_at":            updatedCreatedAt.Format("2006-01-02 15:04:05"),
    },
})
```

## Response Examples

### With NULL Values
```json
{
  "success": true,
  "message": "Appointment rescheduled successfully",
  "appointment": {
    "id": "appt-123",
    "token_number": null,        // ✅ NULL in database
    "mo_id": null,               // ✅ NULL in database
    "patient_name": "John Doe",
    "doctor_name": "Dr. Smith",
    "department": null,          // ✅ NULL in database
    "consultation_type": null,   // ✅ NULL in database
    "appointment_date_time": "2024-10-17 10:30:00",
    "status": "scheduled",
    "fee_amount": null,          // ✅ NULL in database
    "payment_status": "pending",
    "booking_number": "BK20241017001",
    "created_at": "2024-10-17 09:00:00"
  }
}
```

### With Values
```json
{
  "success": true,
  "message": "Appointment rescheduled successfully",
  "appointment": {
    "id": "appt-123",
    "token_number": "T001",      // ✅ Value from database
    "mo_id": "MO123456",         // ✅ Value from database
    "patient_name": "John Doe",
    "doctor_name": "Dr. Smith",
    "department": "Cardiology",  // ✅ Value from database
    "consultation_type": "followup",  // ✅ Value from database
    "appointment_date_time": "2024-10-17 10:30:00",
    "status": "scheduled",
    "fee_amount": 500.00,        // ✅ Value from database
    "payment_status": "paid",
    "booking_number": "BK20241017001",
    "created_at": "2024-10-17 09:00:00"
  }
}
```

## Testing

### Test 1: Appointment with NULL mo_id
```bash
POST /api/v1/appointments/simple/APPOINTMENT_ID/reschedule
{
  "doctor_id": "doctor-uuid",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2024-10-18",
  "appointment_time": "2024-10-18 11:00:00"
}

# Expected: Success (mo_id = null in response) ✅
{
  "appointment": {
    "mo_id": null,
    "token_number": null,
    ...
  }
}
```

### Test 2: Appointment with mo_id
```bash
# Same request with appointment that has mo_id

# Expected: Success (mo_id = "MO123456" in response) ✅
{
  "appointment": {
    "mo_id": "MO123456",
    "token_number": "T001",
    ...
  }
}
```

## Related Fixes

This fix is consistent with the similar fix made for `GetSimpleAppointmentDetails` API in a previous update. Both APIs now properly handle NULL values from the database.

## Files Modified
- ✅ `services/appointment-service/controllers/appointment_list_simple.controller.go`

## Summary

| Issue | Solution |
|-------|----------|
| **Error** | Cannot convert NULL to string |
| **Cause** | Non-pointer types for nullable columns |
| **Fix** | Changed to pointer types (*string, *float64) |
| **Impact** | Reschedule works with NULL values |
| **Status** | ✅ FIXED |

---

**Error**: `sql: Scan error on column index 2, name "mo_id": converting NULL to string is unsupported`  
**Fix**: Changed nullable fields to pointer types (*string, *float64)  
**Result**: Reschedule API handles NULL values correctly ✅


