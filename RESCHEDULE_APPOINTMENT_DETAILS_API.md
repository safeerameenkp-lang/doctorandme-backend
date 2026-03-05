# Reschedule Appointment Details API

## Overview
This API allows rescheduling an existing appointment by its ID, with support for slot selection system. It's designed to work with the reschedule modal UI shown in your image.

## Endpoint
```
POST /api/v1/appointments/simple/:id/reschedule
```

## Authentication
- **Required**: Yes
- **Roles**: `clinic_admin`, `receptionist`
- **Header**: `Authorization: Bearer <token>`

## Path Parameters

| Parameter | Type   | Required | Description           |
|-----------|--------|----------|-----------------------|
| `id`      | string | Yes      | Appointment ID (UUID) |

## Request Body

```json
{
  "doctor_id": "doctor-uuid",
  "department_id": "department-uuid",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2024-10-17",
  "appointment_time": "2024-10-17 10:30:00",
  "reason": "Patient requested time change",
  "notes": "Additional notes for reschedule"
}
```

### Request Fields

| Field              | Type   | Required | Description                                      |
|--------------------|--------|----------|--------------------------------------------------|
| `doctor_id`        | string | Yes      | UUID of the new doctor                           |
| `department_id`    | string | No       | UUID of the department (optional)                |
| `individual_slot_id` | string | Yes      | UUID of the selected individual slot             |
| `appointment_date` | string | Yes      | New appointment date (YYYY-MM-DD)                |
| `appointment_time` | string | Yes      | New appointment time (YYYY-MM-DD HH:MM:SS)       |
| `reason`           | string | No       | Reason for rescheduling                          |
| `notes`            | string | No       | Additional notes                                 |

## Response

### Success Response (200 OK)
```json
{
  "success": true,
  "message": "Appointment rescheduled successfully",
  "appointment": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "token_number": "T001",
    "mo_id": "MO123456",
    "patient_name": "John Doe",
    "doctor_name": "Dr. Smith",
    "department": "Cardiology",
    "consultation_type": "followup",
    "appointment_date_time": "2024-10-17 10:30:00",
    "status": "scheduled",
    "fee_amount": 500.00,
    "payment_status": "paid",
    "fee_status": "paid",
    "booking_number": "BK20241017001",
    "created_at": "2024-10-17 09:00:00"
  }
}
```

### Error Responses

#### 400 Bad Request
```json
{
  "error": "Invalid input",
  "details": "Key: 'RescheduleAppointmentDetailsInput.DoctorID' Error:Field validation for 'DoctorID' failed on the 'required' tag"
}
```

#### 404 Not Found
```json
{
  "error": "Appointment not found or cannot be rescheduled"
}
```

#### 409 Conflict
```json
{
  "error": "Slot not available",
  "message": "This slot is fully booked. Please select another slot."
}
```

#### 409 Conflict (Race Condition)
```json
{
  "error": "Slot just got booked",
  "message": "This slot was just booked by another patient. Please select another slot."
}
```

## How It Works

### 1. **Slot Management**
- ✅ **Frees Old Slot**: Automatically makes the previously booked slot available again
- ✅ **Books New Slot**: Reserves the newly selected slot
- ✅ **Capacity Tracking**: Properly manages `available_count` and `max_patients`
- ✅ **Status Updates**: Updates slot status (`available`, `booked`, `partially_booked`)

### 2. **Transaction Safety**
- ✅ **Atomic Operations**: All changes happen in a single transaction
- ✅ **Rollback on Error**: If any step fails, all changes are rolled back
- ✅ **Race Condition Protection**: Prevents double-booking of slots

### 3. **Data Consistency**
- ✅ **Same Response Format**: Returns appointment in the same format as `GetSimpleAppointmentDetails`
- ✅ **Updated Information**: Returns the latest appointment details after rescheduling

## UI Integration

### Reschedule Modal Flow

Based on your reschedule modal image, here's how the API integrates:

1. **Select Department**: User selects department (optional)
2. **Select Doctor**: User selects doctor (required)
3. **Add Notes**: User adds reason/notes (optional)
4. **Choose Date**: User selects new date
5. **Select Slot**: User selects available slot from morning/afternoon sessions
6. **Save**: API call to reschedule

### Flutter Integration Example

```dart
class RescheduleController {
  Future<Map<String, dynamic>> rescheduleAppointment({
    required String appointmentId,
    required String doctorId,
    String? departmentId,
    required String individualSlotId,
    required String appointmentDate,
    required String appointmentTime,
    String? reason,
    String? notes,
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl/appointments/simple/$appointmentId/reschedule'),
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
      body: json.encode({
        'doctor_id': doctorId,
        'department_id': departmentId,
        'individual_slot_id': individualSlotId,
        'appointment_date': appointmentDate,
        'appointment_time': appointmentTime,
        'reason': reason,
        'notes': notes,
      }),
    );

    if (response.statusCode == 200) {
      return json.decode(response.body);
    } else {
      throw Exception('Failed to reschedule appointment');
    }
  }
}

// Usage in UI
void _handleReschedule() async {
  try {
    final result = await RescheduleController().rescheduleAppointment(
      appointmentId: appointment.id,
      doctorId: selectedDoctor.id,
      departmentId: selectedDepartment?.id,
      individualSlotId: selectedSlot.id,
      appointmentDate: selectedDate.toString(),
      appointmentTime: '${selectedDate.toString()} ${selectedSlot.time}',
      reason: reasonController.text,
      notes: notesController.text,
    );
    
    // Show success message
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('Appointment rescheduled successfully!')),
    );
    
    // Update UI with new appointment details
    setState(() {
      appointment = Appointment.fromJson(result['appointment']);
    });
    
    // Close modal
    Navigator.pop(context);
  } catch (e) {
    // Show error message
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('Failed to reschedule: $e')),
    );
  }
}
```

## Use Cases

### 1. **Patient Requests Time Change**
```json
{
  "doctor_id": "same-doctor-id",
  "individual_slot_id": "new-slot-id",
  "appointment_date": "2024-10-18",
  "appointment_time": "2024-10-18 14:30:00",
  "reason": "Patient requested later time"
}
```

### 2. **Change Doctor**
```json
{
  "doctor_id": "different-doctor-id",
  "department_id": "department-id",
  "individual_slot_id": "available-slot-id",
  "appointment_date": "2024-10-17",
  "appointment_time": "2024-10-17 11:00:00",
  "reason": "Original doctor unavailable"
}
```

### 3. **Change Department and Doctor**
```json
{
  "doctor_id": "cardiology-doctor-id",
  "department_id": "cardiology-department-id",
  "individual_slot_id": "cardiology-slot-id",
  "appointment_date": "2024-10-19",
  "appointment_time": "2024-10-19 09:30:00",
  "reason": "Patient needs cardiology consultation",
  "notes": "Referred by general physician"
}
```

## Slot Selection Integration

### Getting Available Slots

Before rescheduling, get available slots:

```bash
GET /api/v1/doctor-session-slots/list?clinic_id=xxx&doctor_id=xxx&date=2024-10-17
```

### Slot Response Format
```json
{
  "success": true,
  "slots": [
    {
      "id": "slot-uuid",
      "slot_start": "10:30:00",
      "slot_end": "10:45:00",
      "status": "available",
      "available_count": 1,
      "max_patients": 1,
      "session_name": "Morning Session"
    }
  ]
}
```

### Reschedule Request
```json
{
  "doctor_id": "doctor-uuid",
  "individual_slot_id": "slot-uuid",  // From available slots response
  "appointment_date": "2024-10-17",
  "appointment_time": "2024-10-17 10:30:00",
  "reason": "Patient preference"
}
```

## Error Handling

### Common Error Scenarios

1. **Slot Already Booked**
   - **Cause**: Another user booked the slot between selection and submission
   - **Solution**: Refresh available slots and let user select another

2. **Invalid Slot**
   - **Cause**: Slot doesn't exist or belongs to different clinic
   - **Solution**: Validate slot before submission

3. **Appointment Not Found**
   - **Cause**: Appointment ID invalid or appointment already completed
   - **Solution**: Check appointment status before allowing reschedule

4. **Date/Time Format Error**
   - **Cause**: Invalid date or time format
   - **Solution**: Use proper formats (YYYY-MM-DD and YYYY-MM-DD HH:MM:SS)

## Testing

### Test Script
```bash
# Reschedule appointment
curl -X POST "http://localhost:8082/api/v1/appointments/simple/APPOINTMENT_ID/reschedule" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "doctor_id": "DOCTOR_ID",
    "individual_slot_id": "SLOT_ID",
    "appointment_date": "2024-10-17",
    "appointment_time": "2024-10-17 10:30:00",
    "reason": "Patient requested time change"
  }'
```

### Expected Response
```json
{
  "success": true,
  "message": "Appointment rescheduled successfully",
  "appointment": {
    "id": "appointment-id",
    "patient_name": "John Doe",
    "doctor_name": "Dr. Smith",
    "appointment_date_time": "2024-10-17 10:30:00",
    "status": "scheduled"
  }
}
```

## Related APIs

- **Get Appointment Details**: `GET /api/v1/appointments/simple/:id` - Get current appointment details
- **List Available Slots**: `GET /api/v1/doctor-session-slots/list` - Get available slots for selection
- **List Appointments**: `GET /api/v1/appointments/simple-list` - View all appointments

## Benefits

1. ✅ **Slot Management**: Properly handles slot availability and capacity
2. ✅ **Transaction Safety**: All-or-nothing updates prevent data inconsistency
3. ✅ **UI Integration**: Designed to work seamlessly with your reschedule modal
4. ✅ **Consistent Response**: Returns appointment in same format as details API
5. ✅ **Error Handling**: Comprehensive error messages for different scenarios
6. ✅ **Flexible**: Supports changing doctor, department, time, or all three

---

**Endpoint**: `POST /api/v1/appointments/simple/:id/reschedule`  
**Purpose**: Reschedule appointment with slot selection  
**UI Integration**: Works with reschedule modal  
**Status**: ✅ Ready for use

