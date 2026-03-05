# Slot Reschedule Fix - Exclude Current Appointment

## Problem
During appointment reschedule, the current appointment's slot shows as "booked" (red) instead of being selectable, because the slot list API counts the current appointment when checking if a slot is available.

## Root Cause
The `ListDoctorSessionSlots` API was counting ALL appointments for a slot, including the current appointment being rescheduled, making it appear as "booked" and unselectable.

## Solution
✅ **Updated Backend API** to accept an optional `appointment_id` parameter. When provided, it excludes the current appointment from the booking count during reschedule.

## Changes Made

### File: `doctor_session_slots.controller.go`

#### 1. Added appointment_id Parameter
```go
func ListDoctorSessionSlots(c *gin.Context) {
    doctorID := c.Query("doctor_id")
    date := c.Query("date")
    clinicID := c.Query("clinic_id")
    slotType := c.Query("slot_type")
    appointmentID := c.Query("appointment_id") // ✅ NEW: For reschedule - exclude current appointment
```

#### 2. Updated Slot Booking Count Logic
```go
// Before (Incorrect) ❌
err = config.DB.QueryRow(`
    SELECT COUNT(*) 
    FROM appointments 
    WHERE individual_slot_id = $1 
    AND status NOT IN ('cancelled', 'no_show')
`, slot.ID).Scan(&actualBookedCount)

// After (Fixed) ✅
if appointmentID != "" {
    // ✅ RESCHEDULE MODE: Exclude current appointment from booking count
    countQuery = `
        SELECT COUNT(*) 
        FROM appointments 
        WHERE individual_slot_id = $1 
        AND status NOT IN ('cancelled', 'no_show')
        AND id != $2  -- Exclude current appointment during reschedule
    `
    countArgs = []interface{}{slot.ID, appointmentID}
} else {
    // Normal mode for new appointments
    countQuery = `
        SELECT COUNT(*) 
        FROM appointments 
        WHERE individual_slot_id = $1 
        AND status NOT IN ('cancelled', 'no_show')
    `
    countArgs = []interface{}{slot.ID}
}

err = config.DB.QueryRow(countQuery, countArgs...).Scan(&actualBookedCount)
```

#### 3. Updated Response
```go
c.JSON(http.StatusOK, gin.H{
    "doctor_id":      doctorID,
    "clinic_id":      clinicID,
    "date":           date,
    "slot_type":      slotType,
    "appointment_id": appointmentID, // ✅ Include for debugging reschedule mode
    "slots":          results,
    "total":          len(results),
})
```

## API Usage

### For New Appointments (No Change)
```bash
GET /api/v1/doctor-session-slots?doctor_id=xxx&clinic_id=xxx&date=2024-10-17&slot_type=offline
```

### For Reschedule (New)
```bash
GET /api/v1/doctor-session-slots?doctor_id=xxx&clinic_id=xxx&date=2024-10-17&slot_type=offline&appointment_id=APPOINTMENT_ID
```

## Frontend Integration

### Flutter Implementation
```dart
class SlotService {
  // For new appointments
  Future<List<Slot>> getAvailableSlots({
    required String doctorId,
    required String clinicId,
    required String date,
    required String slotType,
  }) async {
    final url = 'doctor-session-slots?'
        'doctor_id=$doctorId'
        '&clinic_id=$clinicId'
        '&date=$date'
        '&slot_type=$slotType';
    
    final response = await _serviceRepo.request(url, method: 'GET', token: token);
    return parseSlots(response);
  }

  // ✅ For reschedule - exclude current appointment
  Future<List<Slot>> getAvailableSlotsForReschedule({
    required String doctorId,
    required String clinicId,
    required String date,
    required String slotType,
    required String appointmentId, // ✅ NEW parameter
  }) async {
    final url = 'doctor-session-slots?'
        'doctor_id=$doctorId'
        '&clinic_id=$clinicId'
        '&date=$date'
        '&slot_type=$slotType'
        '&appointment_id=$appointmentId'; // ✅ Excludes current appointment
    
    final response = await _serviceRepo.request(url, method: 'GET', token: token);
    return parseSlots(response);
  }
}
```

### Usage in Reschedule Modal
```dart
class RescheduleModal extends StatefulWidget {
  final Appointment appointment; // Current appointment being rescheduled
  
  @override
  Widget build(BuildContext context) {
    return FutureBuilder<List<Slot>>(
      future: SlotService().getAvailableSlotsForReschedule(
        doctorId: selectedDoctor.id,
        clinicId: appointment.clinicId,
        date: selectedDate,
        slotType: 'offline',
        appointmentId: appointment.id, // ✅ This makes current slot selectable
      ),
      builder: (context, snapshot) {
        // Now the current appointment's slot will show as "Available" (green)
        // instead of "Booked" (red)
      },
    );
  }
}
```

## How It Works

### Scenario: Rescheduling Appointment

**Before Fix:**
1. User opens reschedule modal for Appointment A (slot at 10:30 AM)
2. API counts ALL appointments for 10:30 AM slot: 1 appointment (Appointment A)
3. Slot shows as "Booked" (red) - user cannot select it
4. User confused - cannot reschedule to same slot

**After Fix:**
1. User opens reschedule modal for Appointment A (slot at 10:30 AM)
2. API counts appointments for 10:30 AM slot EXCLUDING Appointment A: 0 appointments
3. Slot shows as "Available" (green) - user can select it
4. User can reschedule to same slot or any other available slot

### Database Query Comparison

**Before (Incorrect):**
```sql
SELECT COUNT(*) 
FROM appointments 
WHERE individual_slot_id = 'slot-uuid' 
AND status NOT IN ('cancelled', 'no_show')
-- Result: 1 (includes current appointment)
```

**After (Fixed):**
```sql
SELECT COUNT(*) 
FROM appointments 
WHERE individual_slot_id = 'slot-uuid' 
AND status NOT IN ('cancelled', 'no_show')
AND id != 'current-appointment-uuid'  -- ✅ Excludes current appointment
-- Result: 0 (current appointment excluded)
```

## Response Example

### New Appointment (No appointment_id)
```json
{
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid", 
  "date": "2024-10-17",
  "slot_type": "offline",
  "appointment_id": null,
  "slots": [
    {
      "id": "slot-uuid",
      "slot_start": "10:30:00",
      "slot_end": "10:45:00",
      "is_booked": true,        // ❌ Shows as booked
      "is_bookable": false,     // ❌ Not selectable
      "display_message": "Fully Booked"
    }
  ]
}
```

### Reschedule (With appointment_id)
```json
{
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "date": "2024-10-17", 
  "slot_type": "offline",
  "appointment_id": "current-appointment-uuid", // ✅ Current appointment ID
  "slots": [
    {
      "id": "slot-uuid",
      "slot_start": "10:30:00",
      "slot_end": "10:45:00",
      "is_booked": false,       // ✅ Shows as available
      "is_bookable": true,      // ✅ Selectable
      "display_message": "Available"
    }
  ]
}
```

## Testing

### Test 1: New Appointment (Should show booked slots as red)
```bash
curl -X GET "http://localhost:8082/api/v1/doctor-session-slots?doctor_id=DOCTOR_ID&clinic_id=CLINIC_ID&date=2024-10-17&slot_type=offline"
```

### Test 2: Reschedule (Should show current slot as available)
```bash
curl -X GET "http://localhost:8082/api/v1/doctor-session-slots?doctor_id=DOCTOR_ID&clinic_id=CLINIC_ID&date=2024-10-17&slot_type=offline&appointment_id=APPOINTMENT_ID"
```

## Benefits

1. ✅ **Fix Reschedule UX**: Current appointment's slot now shows as selectable during reschedule
2. ✅ **Backward Compatible**: Existing new appointment flows work unchanged
3. ✅ **Accurate Availability**: Shows true slot availability excluding current appointment
4. ✅ **Better User Experience**: Users can reschedule to same slot or any available slot
5. ✅ **Debug Support**: Response includes appointment_id for troubleshooting

## Related APIs

- **List Slots**: `GET /api/v1/doctor-session-slots` - Get available slots
- **Reschedule**: `POST /api/v1/appointments/simple/:id/reschedule` - Reschedule appointment
- **Get Appointment**: `GET /api/v1/appointments/simple/:id` - Get appointment details

## Status
✅ **IMPLEMENTED AND TESTED**

The slot list API now properly handles reschedule scenarios by excluding the current appointment from booking counts.

---

**Problem**: Current appointment's slot showed as "booked" during reschedule  
**Solution**: Added appointment_id parameter to exclude current appointment  
**Result**: Current slot now shows as "available" and selectable during reschedule ✅
