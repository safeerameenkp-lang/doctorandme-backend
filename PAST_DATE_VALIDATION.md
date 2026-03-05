# Past Date Validation for Reschedule

## Overview
Added validation to prevent rescheduling appointments to past dates or times. This ensures data integrity and prevents user errors.

## Problem
Without validation:
- ❌ Users could reschedule appointments to yesterday
- ❌ Users could select past times for today
- ❌ System would accept invalid past dates
- ❌ No clear error messages

## Solution
✅ **Past Date Validation**: Added checks to prevent rescheduling to past dates and times in both the reschedule API and slot list API.

## Implementation

### 1. Reschedule API Validation

#### File: `appointment_list_simple.controller.go`

```go
// ✅ Step 3.1: Validate appointment is not in the past
now := time.Now()
today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

if appointmentDate.Before(today) {
    c.JSON(http.StatusBadRequest, gin.H{
        "error":   "Cannot reschedule to past date",
        "message": "Please select a date from today onwards",
    })
    return
}

// If appointment is today, check time is not in the past
if appointmentDate.Equal(today) && appointmentTime.Before(now) {
    c.JSON(http.StatusBadRequest, gin.H{
        "error":   "Cannot reschedule to past time",
        "message": "Please select a time in the future",
    })
    return
}
```

### 2. Slot List API Validation

#### File: `doctor_session_slots.controller.go`

```go
// ✅ Validate date is not in the past (optional validation)
if date != "" {
    requestedDate, err := time.Parse("2006-01-02", date)
    if err == nil { // Only validate if date is valid format
        now := time.Now()
        today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
        
        if requestedDate.Before(today) {
            c.JSON(http.StatusBadRequest, gin.H{
                "error":   "Cannot fetch slots for past dates",
                "message": "Please select a date from today onwards",
            })
            return
        }
    }
}
```

## Validation Rules

### Date Validation
1. **Past Date**: Date must be today or in the future
2. **Today's Date**: Allowed (with time validation)
3. **Future Date**: Allowed

### Time Validation
1. **Past Time Today**: Not allowed if appointment is today
2. **Future Time Today**: Allowed
3. **Any Time Future Date**: Allowed

## Error Responses

### Error 1: Past Date
```json
{
  "error": "Cannot reschedule to past date",
  "message": "Please select a date from today onwards"
}
```

**Scenario**: User tries to reschedule to yesterday
```bash
POST /appointments/simple/123/reschedule
{
  "appointment_date": "2024-10-15",  // Yesterday
  "appointment_time": "2024-10-15 10:30:00"
}
```

### Error 2: Past Time Today
```json
{
  "error": "Cannot reschedule to past time",
  "message": "Please select a time in the future"
}
```

**Scenario**: User tries to reschedule to 9:00 AM when it's already 2:00 PM today
```bash
POST /appointments/simple/123/reschedule
{
  "appointment_date": "2024-10-17",  // Today
  "appointment_time": "2024-10-17 09:00:00"  // 9 AM (already passed)
}
```

### Error 3: Slot List Past Date
```json
{
  "error": "Cannot fetch slots for past dates",
  "message": "Please select a date from today onwards"
}
```

**Scenario**: User tries to fetch slots for yesterday
```bash
GET /doctor-session-slots?date=2024-10-15  // Yesterday
```

## Use Cases

### Use Case 1: Reschedule to Tomorrow ✅
```bash
POST /appointments/simple/123/reschedule
{
  "doctor_id": "doctor-uuid",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2024-10-18",  // Tomorrow ✅
  "appointment_time": "2024-10-18 10:30:00"
}
```
**Result**: Success - Future date allowed

### Use Case 2: Reschedule to Today (Future Time) ✅
```bash
# Current time: 2024-10-17 10:00:00

POST /appointments/simple/123/reschedule
{
  "doctor_id": "doctor-uuid",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2024-10-17",  // Today ✅
  "appointment_time": "2024-10-17 15:30:00"  // 3:30 PM (future) ✅
}
```
**Result**: Success - Today with future time allowed

### Use Case 3: Reschedule to Yesterday ❌
```bash
POST /appointments/simple/123/reschedule
{
  "doctor_id": "doctor-uuid",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2024-10-15",  // Yesterday ❌
  "appointment_time": "2024-10-15 10:30:00"
}
```
**Result**: Error - Past date not allowed

### Use Case 4: Reschedule to Today (Past Time) ❌
```bash
# Current time: 2024-10-17 14:00:00 (2 PM)

POST /appointments/simple/123/reschedule
{
  "doctor_id": "doctor-uuid",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2024-10-17",  // Today ✅
  "appointment_time": "2024-10-17 10:30:00"  // 10:30 AM (past) ❌
}
```
**Result**: Error - Past time today not allowed

## Frontend Integration

### Flutter Implementation

```dart
class RescheduleValidator {
  // ✅ Validate date before making API call
  static String? validateRescheduleDate(DateTime selectedDate, TimeOfDay selectedTime) {
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final selectedDateTime = DateTime(
      selectedDate.year,
      selectedDate.month,
      selectedDate.day,
      selectedTime.hour,
      selectedTime.minute,
    );
    
    // Check if date is in the past
    if (selectedDate.isBefore(today)) {
      return "Cannot reschedule to past date. Please select a date from today onwards.";
    }
    
    // Check if time is in the past (for today)
    if (selectedDate.isAtSameMomentAs(today) && selectedDateTime.isBefore(now)) {
      return "Cannot reschedule to past time. Please select a time in the future.";
    }
    
    return null; // Valid
  }
}

// Usage in reschedule modal
void _handleReschedule() async {
  // ✅ Validate before API call
  final error = RescheduleValidator.validateRescheduleDate(selectedDate, selectedTime);
  
  if (error != null) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(error), backgroundColor: Colors.red),
    );
    return;
  }
  
  // Proceed with API call
  try {
    await rescheduleAppointment(...);
    // Success
  } catch (e) {
    // Handle error
  }
}
```

### Date Picker Configuration

```dart
// ✅ Configure date picker to disable past dates
Future<DateTime?> _selectDate(BuildContext context) async {
  final now = DateTime.now();
  
  return showDatePicker(
    context: context,
    initialDate: now,
    firstDate: now, // ✅ Only allow today and future dates
    lastDate: DateTime.now().add(Duration(days: 365)),
    selectableDayPredicate: (DateTime date) {
      // ✅ Disable past dates
      return date.isAfter(DateTime.now().subtract(Duration(days: 1)));
    },
  );
}
```

### Time Picker Validation

```dart
// ✅ Validate time if date is today
Future<TimeOfDay?> _selectTime(BuildContext context, DateTime selectedDate) async {
  final now = DateTime.now();
  final today = DateTime(now.year, now.month, now.day);
  
  final TimeOfDay? picked = await showTimePicker(
    context: context,
    initialTime: TimeOfDay.now(),
  );
  
  if (picked != null && selectedDate.isAtSameMomentAs(today)) {
    final selectedDateTime = DateTime(
      selectedDate.year,
      selectedDate.month,
      selectedDate.day,
      picked.hour,
      picked.minute,
    );
    
    // ✅ Check if time is in the past
    if (selectedDateTime.isBefore(now)) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Cannot select past time. Please choose a time in the future.'),
          backgroundColor: Colors.red,
        ),
      );
      return null;
    }
  }
  
  return picked;
}
```

## Benefits

1. ✅ **Data Integrity**: Prevents invalid past appointments
2. ✅ **User Guidance**: Clear error messages guide users
3. ✅ **Consistent Validation**: Both frontend and backend validate
4. ✅ **Better UX**: Prevents user errors before submission
5. ✅ **Flexible**: Allows same-day reschedules with future times

## Testing

### Test 1: Past Date Validation
```bash
# Should fail
curl -X POST "http://localhost:8082/api/v1/appointments/simple/APPOINTMENT_ID/reschedule" \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "doctor_id": "DOCTOR_ID",
    "individual_slot_id": "SLOT_ID",
    "appointment_date": "2024-10-15",
    "appointment_time": "2024-10-15 10:30:00"
  }'

# Expected Response:
{
  "error": "Cannot reschedule to past date",
  "message": "Please select a date from today onwards"
}
```

### Test 2: Past Time Today
```bash
# Should fail (if current time is after 10:30 AM)
curl -X POST "http://localhost:8082/api/v1/appointments/simple/APPOINTMENT_ID/reschedule" \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "doctor_id": "DOCTOR_ID",
    "individual_slot_id": "SLOT_ID",
    "appointment_date": "2024-10-17",
    "appointment_time": "2024-10-17 10:30:00"
  }'

# Expected Response:
{
  "error": "Cannot reschedule to past time",
  "message": "Please select a time in the future"
}
```

### Test 3: Future Date (Valid)
```bash
# Should succeed
curl -X POST "http://localhost:8082/api/v1/appointments/simple/APPOINTMENT_ID/reschedule" \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "doctor_id": "DOCTOR_ID",
    "individual_slot_id": "SLOT_ID",
    "appointment_date": "2024-10-20",
    "appointment_time": "2024-10-20 10:30:00"
  }'

# Expected Response:
{
  "success": true,
  "message": "Appointment rescheduled successfully",
  "appointment": { ... }
}
```

### Test 4: Slot List Past Date
```bash
# Should fail
curl -X GET "http://localhost:8082/api/v1/doctor-session-slots?doctor_id=DOCTOR_ID&clinic_id=CLINIC_ID&date=2024-10-15"

# Expected Response:
{
  "error": "Cannot fetch slots for past dates",
  "message": "Please select a date from today onwards"
}
```

## Status
✅ **IMPLEMENTED**

Both reschedule API and slot list API now validate against past dates and times.

---

**Problem**: Users could reschedule to past dates/times  
**Solution**: Added validation for date and time constraints  
**Result**: Only future dates and times are allowed ✅
