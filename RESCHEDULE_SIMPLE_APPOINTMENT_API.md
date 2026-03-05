# Reschedule Simple Appointment API Documentation

## Overview
The Reschedule Simple Appointment API allows clinic administrators and receptionists to reschedule existing appointments to new time slots while automatically managing slot availability.

## Endpoint

```
POST /appointments/:id/reschedule-simple
```

**Authentication Required:** Yes  
**Roles Allowed:** `clinic_admin`, `receptionist`

---

## Request Format

### URL Parameters
- `id` (string, UUID, required): The ID of the appointment to reschedule

### Request Body

```json
{
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "individual_slot_id": "uuid",
  "appointment_date": "YYYY-MM-DD",
  "appointment_time": "YYYY-MM-DD HH:MM:SS",
  "department_id": "uuid",
  "reason": "string",
  "notes": "string"
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `doctor_id` | UUID | Yes | ID of the new doctor (can be same or different) |
| `clinic_id` | UUID | Yes | ID of the clinic |
| `individual_slot_id` | UUID | Yes | ID of the new individual time slot |
| `appointment_date` | String | Yes | New appointment date in `YYYY-MM-DD` format |
| `appointment_time` | String | Yes | New appointment time in `YYYY-MM-DD HH:MM:SS` format |
| `department_id` | UUID | No | Department ID (optional) |
| `reason` | String | No | Reason for the appointment |
| `notes` | String | No | Additional notes |

---

## Response Format

### Success Response (200 OK)

```json
{
  "message": "Appointment rescheduled successfully",
  "appointment": {
    "id": "uuid",
    "clinic_patient_id": "uuid",
    "clinic_id": "uuid",
    "doctor_id": "uuid",
    "department_id": "uuid",
    "booking_number": "string",
    "token_number": 1,
    "appointment_date": "2024-01-15",
    "appointment_time": "2024-01-15T10:30:00Z",
    "duration_minutes": 5,
    "consultation_type": "offline",
    "reason": "Follow-up checkup",
    "notes": "Patient requested morning slot",
    "status": "confirmed",
    "fee_amount": 500.00,
    "payment_status": "paid",
    "payment_mode": "cash",
    "created_at": "2024-01-10T09:00:00Z"
  },
  "slot_re_enabled": {
    "old_slot_id": "uuid",
    "message": "Previous slot has been made available again"
  }
}
```

---

## Error Responses

### 400 Bad Request - Invalid Input
```json
{
  "error": "Invalid input",
  "details": "Key: 'RescheduleSimpleAppointmentInput.DoctorID' Error:Field validation for 'DoctorID' failed on the 'required' tag"
}
```

### 400 Bad Request - Invalid Date Format
```json
{
  "error": "Invalid date format. Use YYYY-MM-DD"
}
```

### 400 Bad Request - Invalid Time Format
```json
{
  "error": "Invalid time format. Use YYYY-MM-DD HH:MM:SS"
}
```

### 400 Bad Request - Patient Belongs to Different Clinic
```json
{
  "error": "Patient belongs to different clinic"
}
```

### 400 Bad Request - Slot Belongs to Different Clinic
```json
{
  "error": "Slot belongs to different clinic"
}
```

### 404 Not Found - Appointment Not Found
```json
{
  "error": "Appointment not found or cannot be rescheduled"
}
```

### 404 Not Found - Patient Not Found
```json
{
  "error": "Patient not found"
}
```

### 404 Not Found - Doctor Not Found
```json
{
  "error": "Doctor not found"
}
```

### 404 Not Found - Slot Not Found
```json
{
  "error": "Slot not found"
}
```

### 409 Conflict - Slot Not Available
```json
{
  "error": "Slot not available",
  "message": "This slot is fully booked. Please select another slot.",
  "details": {
    "max_patients": 3,
    "available_count": 0,
    "booked_count": 3
  }
}
```

### 409 Conflict - Slot Just Got Booked (Race Condition)
```json
{
  "error": "Slot just got booked",
  "message": "This slot was just booked by another patient. Please select another slot."
}
```

### 500 Internal Server Error
```json
{
  "error": "Failed to update appointment",
  "details": "Database error details"
}
```

---

## How It Works

### Step-by-Step Process

1. **Validate Existing Appointment**
   - Fetches the current appointment details
   - Only allows rescheduling for appointments with status `confirmed` or `pending`
   - Retrieves current slot information

2. **Validate Patient**
   - Verifies the clinic patient exists and is active
   - Ensures patient belongs to the specified clinic

3. **Validate New Slot**
   - Checks if the new slot exists and belongs to the correct clinic
   - Verifies slot has available capacity (`available_count > 0`)
   - Ensures slot status is `available`

4. **Doctor Change Handling**
   - If doctor changes, recalculates consultation fee based on new doctor's rates
   - Generates new booking number and token number
   - If doctor remains same, keeps existing booking number and token

5. **Transaction-based Slot Management**
   - **Frees Old Slot:** Increases `available_count` of the previous slot
   - **Books New Slot:** Decreases `available_count` of the new slot
   - Updates appointment with new details
   - All operations in a database transaction (atomic)

6. **Slot Re-enabling Logic**
   ```
   New Available Count = Old Available Count + 1
   
   If (New Available Count >= Max Patients):
     - Set is_booked = false
     - Set status = 'available'
     - Clear booked_appointment_id
   ```

7. **New Slot Booking Logic**
   ```
   New Available Count = Old Available Count - 1
   
   If (New Available Count <= 0):
     - Set is_booked = true
     - Set status = 'booked'
     - Set booked_appointment_id = appointment_id
   ```

---

## Key Features

### 🔄 Automatic Slot Management
- Automatically frees up the old slot when rescheduling
- Automatically books the new slot
- Maintains slot capacity tracking

### 🔒 Race Condition Protection
- Uses database transactions to prevent double-booking
- Validates slot availability before committing changes
- Returns clear error if slot gets booked during the process

### 💰 Dynamic Fee Calculation
- Recalculates fee if doctor changes
- Maintains existing fee if same doctor
- Respects clinic-specific doctor fees

### 🎫 Smart Token Management
- Generates new token number if doctor changes
- Preserves token number if same doctor and date
- Ensures proper queue ordering

### ✅ Comprehensive Validations
- Patient must belong to the clinic
- Slot must belong to the clinic
- Appointment must be in reschedulable status
- New slot must have available capacity

---

## Example Usage

### Example 1: Reschedule to Different Doctor

**Request:**
```bash
curl -X POST http://localhost:8082/appointments/550e8400-e29b-41d4-a716-446655440000/reschedule-simple \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "doctor_id": "660e8400-e29b-41d4-a716-446655440000",
    "clinic_id": "770e8400-e29b-41d4-a716-446655440000",
    "individual_slot_id": "880e8400-e29b-41d4-a716-446655440000",
    "appointment_date": "2024-01-20",
    "appointment_time": "2024-01-20 14:00:00",
    "department_id": "990e8400-e29b-41d4-a716-446655440000",
    "reason": "Specialist consultation",
    "notes": "Patient requested specialist"
  }'
```

**Response:**
```json
{
  "message": "Appointment rescheduled successfully",
  "appointment": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "clinic_patient_id": "111e8400-e29b-41d4-a716-446655440000",
    "clinic_id": "770e8400-e29b-41d4-a716-446655440000",
    "doctor_id": "660e8400-e29b-41d4-a716-446655440000",
    "booking_number": "DR002-20240120-001",
    "token_number": 1,
    "appointment_date": "2024-01-20",
    "appointment_time": "2024-01-20T14:00:00Z",
    "consultation_type": "offline",
    "fee_amount": 700.00,
    "payment_status": "paid",
    "status": "confirmed"
  },
  "slot_re_enabled": {
    "old_slot_id": "abc12345-e29b-41d4-a716-446655440000",
    "message": "Previous slot has been made available again"
  }
}
```

### Example 2: Reschedule to Different Time (Same Doctor)

**Request:**
```bash
curl -X POST http://localhost:8082/appointments/550e8400-e29b-41d4-a716-446655440000/reschedule-simple \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "doctor_id": "660e8400-e29b-41d4-a716-446655440000",
    "clinic_id": "770e8400-e29b-41d4-a716-446655440000",
    "individual_slot_id": "new-slot-uuid",
    "appointment_date": "2024-01-20",
    "appointment_time": "2024-01-20 16:00:00",
    "reason": "Patient prefers evening slot"
  }'
```

**Response:**
```json
{
  "message": "Appointment rescheduled successfully",
  "appointment": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "booking_number": "DR001-20240115-001",
    "token_number": 5,
    "appointment_date": "2024-01-20",
    "appointment_time": "2024-01-20T16:00:00Z",
    "fee_amount": 500.00,
    "status": "confirmed"
  },
  "slot_re_enabled": {
    "old_slot_id": "old-slot-uuid",
    "message": "Previous slot has been made available again"
  }
}
```

---

## Important Notes

### ⚠️ Payment Handling
- Payment status and mode are **preserved** during rescheduling
- If fee changes due to doctor change, payment status remains unchanged
- Frontend should handle payment adjustment logic if needed

### ⚠️ Status Restrictions
- Only appointments with status `confirmed` or `pending` can be rescheduled
- Completed, cancelled, or no-show appointments cannot be rescheduled

### ⚠️ Slot Capacity
- Respects `max_patients` configuration for multi-patient slots
- Prevents overbooking by checking `available_count`
- Uses database-level constraints to prevent race conditions

### ⚠️ Token Number Changes
- Token number changes when:
  - Doctor changes
  - Date changes (even for same doctor)
- Token number preserved when:
  - Only time changes on same day for same doctor

---

## Differences from Original Reschedule API

| Feature | Original `/reschedule` | New `/reschedule-simple` |
|---------|------------------------|--------------------------|
| **Input Structure** | Simpler (only time) | Comprehensive (all fields) |
| **Slot Management** | Manual | Automatic |
| **Old Slot Handling** | Not freed automatically | Automatically re-enabled |
| **Fee Recalculation** | No | Yes (on doctor change) |
| **Token Generation** | No | Yes (on doctor change) |
| **Clinic Validation** | Limited | Comprehensive |
| **Race Condition Protection** | Basic | Advanced |

---

## Database Changes

The reschedule operation affects these tables:

1. **`appointments` table**
   - Updates: `doctor_id`, `department_id`, `individual_slot_id`, `appointment_date`, `appointment_time`, `booking_number`, `token_number`, `fee_amount`, `reason`, `notes`

2. **`doctor_individual_slots` table**
   - Old Slot: Increases `available_count`, may update `is_booked` and `status`
   - New Slot: Decreases `available_count`, may update `is_booked` and `status`

---

## Testing Checklist

- [ ] Reschedule to different doctor
- [ ] Reschedule to same doctor, different time
- [ ] Reschedule to same doctor, different date
- [ ] Verify old slot becomes available
- [ ] Verify new slot capacity decreases
- [ ] Test with fully booked slot (should fail)
- [ ] Test with invalid appointment ID
- [ ] Test with cancelled appointment (should fail)
- [ ] Test with patient from different clinic (should fail)
- [ ] Test concurrent reschedule requests (race condition)
- [ ] Verify fee recalculation on doctor change
- [ ] Verify token number generation
- [ ] Test with invalid date/time formats

---

## Related APIs

- **Create Simple Appointment:** `POST /appointments/simple`
- **Get Appointment List:** `GET /appointments/simple-list`
- **Get Doctor Slots:** `GET /organization/doctors/:doctor_id/clinics/:clinic_id/session-slots`
- **Cancel Appointment:** `POST /appointments/:id/cancel`

---

## Migration Notes

### From Old Reschedule API
If migrating from the original `POST /appointments/:id/reschedule` endpoint:

**Old Request Format:**
```json
{
  "new_appointment_time": "2024-01-20 14:00:00",
  "reason": "Patient request"
}
```

**New Request Format:**
```json
{
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "individual_slot_id": "uuid",
  "appointment_date": "2024-01-20",
  "appointment_time": "2024-01-20 14:00:00",
  "reason": "Patient request"
}
```

**Migration Steps:**
1. Update frontend to include all required fields
2. Change endpoint from `/reschedule` to `/reschedule-simple`
3. Handle the enhanced response format
4. Remove manual slot re-enabling logic (now automatic)

---

## Support

For issues or questions:
- Check error responses for detailed messages
- Verify JWT token is valid
- Ensure user has correct role (`clinic_admin` or `receptionist`)
- Check database logs for transaction failures
- Verify slot capacity settings

---

*Last Updated: October 17, 2024*

