# Appointment Details API with Slot Information

## Overview
The appointment details API has been enhanced to include detailed information about the booked time slot when an appointment was created using the individual slot booking system.

## Implementation Date
**October 17, 2024**

## What's New

### Enhanced Response Structure
The appointment details API now returns slot information when available, including:
- Slot ID and status
- Exact slot start and end times
- Session name (e.g., "Morning Session", "Afternoon Session")
- Formatted slot date and time range
- Slot-specific date

### Conditional Slot Details
The `slot_details` object is only included in the response when:
1. The appointment was booked through the individual slot booking system
2. The appointment has a valid `individual_slot_id` linking it to a specific slot

If an appointment was created without using the slot system, the `slot_details` field will not be present in the response.

## Technical Implementation

### Database Schema
The API joins four tables to retrieve slot information:

```sql
appointments (a)
  └─ individual_slot_id → doctor_individual_slots (dis)
                            └─ session_id → doctor_slot_sessions (dss)
                                              └─ time_slot_id → doctor_time_slots (dts)
```

### SQL Query Structure
```sql
LEFT JOIN doctor_individual_slots dis ON dis.id = a.individual_slot_id
LEFT JOIN doctor_slot_sessions dss ON dss.id = dis.session_id
LEFT JOIN doctor_time_slots dts ON dts.id = dss.time_slot_id
```

Using LEFT JOINs ensures the query returns results even when slot information is not available.

## API Endpoint

```
GET /api/v1/appointments/simple/:id
```

### Authentication Required
- Roles: `clinic_admin`, `receptionist`, `doctor`
- Header: `Authorization: Bearer <token>`

## Response Example

### With Slot Details
```json
{
  "success": true,
  "appointment": {
    "id": "appointment-uuid",
    "token_number": "T001",
    "mo_id": "MO123456",
    "booking_number": "BK20241017001",
    "patient": {
      "name": "John Doe",
      "phone": "+1234567890",
      "email": "john.doe@example.com",
      "age": 35,
      "gender": "male"
    },
    "doctor": {
      "id": "doctor-uuid",
      "name": "Dr. Smith"
    },
    "department": {
      "id": "dept-uuid",
      "name": "Cardiology"
    },
    "clinic": {
      "id": "clinic-uuid",
      "name": "Main Clinic"
    },
    "consultation_type": "followup",
    "appointment_date_time": "2024-10-17 10:30:00",
    "duration_minutes": 15,
    "session_type": "morning",
    "status": "scheduled",
    "fee_amount": 500.00,
    "payment_status": "paid",
    "payment_method": "cash",
    "slot_details": {
      "slot_id": "slot-uuid",
      "slot_status": "booked",
      "slot_start_time": "10:30:00",
      "slot_end_time": "10:45:00",
      "slot_date": "2024-10-17",
      "slot_full_time": "2024-10-17 10:30:00 - 10:45:00",
      "session_name": "Morning Session"
    },
    "created_at": "2024-10-17 09:00:00",
    "updated_at": "2024-10-17 09:00:00"
  }
}
```

### Without Slot Details (Legacy Appointments)
```json
{
  "success": true,
  "appointment": {
    "id": "appointment-uuid",
    "token_number": "T002",
    "mo_id": "MO123457",
    "booking_number": "BK20241017002",
    "patient": { ... },
    "doctor": { ... },
    "department": { ... },
    "clinic": { ... },
    "consultation_type": "new",
    "appointment_date_time": "2024-10-17 11:00:00",
    "duration_minutes": 15,
    "session_type": "morning",
    "status": "scheduled",
    "fee_amount": 1000.00,
    "payment_status": "pending",
    "payment_method": null,
    "created_at": "2024-10-17 10:30:00",
    "updated_at": "2024-10-17 10:30:00"
  }
}
```

Note: No `slot_details` field is present for appointments not booked via the slot system.

## Slot Details Object Fields

| Field              | Type   | Description                                      |
|--------------------|--------|--------------------------------------------------|
| `slot_id`          | string | UUID of the individual slot                      |
| `slot_status`      | string | Current status: `available`, `booked`, `cancelled`, `blocked` |
| `slot_start_time`  | string | Start time in HH:MM:SS format                    |
| `slot_end_time`    | string | End time in HH:MM:SS format                      |
| `slot_date`        | string | Date of the slot (YYYY-MM-DD)                    |
| `slot_full_time`   | string | Human-readable format: "YYYY-MM-DD HH:MM:SS - HH:MM:SS" |
| `session_name`     | string | Name of the session (e.g., "Morning Session")    |

## Use Cases

### 1. Display Appointment Details
Show complete appointment information including the exact time slot booked:
```
Appointment: BK20241017001
Patient: John Doe (MO123456)
Doctor: Dr. Smith
Time Slot: 2024-10-17 10:30:00 - 10:45:00 (Morning Session)
Status: Scheduled
```

### 2. Verify Slot Booking
Check if an appointment has an associated slot and its status:
```javascript
if (appointment.slot_details) {
  console.log(`Slot Status: ${appointment.slot_details.slot_status}`);
  console.log(`Session: ${appointment.slot_details.session_name}`);
} else {
  console.log('Legacy appointment without slot booking');
}
```

### 3. Print Appointment Slip
Include slot details on printed appointment slips:
```
╔═══════════════════════════════════════════╗
║        APPOINTMENT CONFIRMATION           ║
╠═══════════════════════════════════════════╣
║ Booking Number: BK20241017001             ║
║ Token: T001                               ║
║ Patient: John Doe (MO123456)              ║
║ Doctor: Dr. Smith                         ║
║ Department: Cardiology                    ║
║ Slot: Morning Session                     ║
║ Time: 2024-10-17 10:30:00 - 10:45:00      ║
║ Fee: ₹500.00                              ║
║ Payment: Paid                             ║
╚═══════════════════════════════════════════╝
```

### 4. Slot Management Dashboard
Track which slots are being used for appointments:
```javascript
appointments.forEach(apt => {
  if (apt.slot_details) {
    console.log(`Slot ${apt.slot_details.slot_id}: ${apt.slot_details.slot_status}`);
  }
});
```

### 5. Rescheduling Assistance
Show original slot information when rescheduling:
```
Current Appointment:
  Date/Time: 2024-10-17 10:30:00 - 10:45:00
  Session: Morning Session
  
Select new slot:
  [ Available slots for rescheduling... ]
```

## Code Example (Flutter)

```dart
class AppointmentDetailScreen extends StatelessWidget {
  final String appointmentId;

  Future<Map<String, dynamic>> fetchAppointmentDetails() async {
    final response = await http.get(
      Uri.parse('http://api.example.com/api/v1/appointments/simple/$appointmentId'),
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
    );

    if (response.statusCode == 200) {
      return json.decode(response.body);
    } else {
      throw Exception('Failed to load appointment');
    }
  }

  Widget buildSlotInfo(Map<String, dynamic>? slotDetails) {
    if (slotDetails == null) {
      return Card(
        child: Padding(
          padding: EdgeInsets.all(16),
          child: Text('No slot information available'),
        ),
      );
    }

    return Card(
      child: Padding(
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Slot Details', style: TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.bold,
            )),
            SizedBox(height: 8),
            Text('Session: ${slotDetails['session_name'] ?? 'N/A'}'),
            Text('Time: ${slotDetails['slot_full_time'] ?? 'N/A'}'),
            Text('Status: ${slotDetails['slot_status'] ?? 'N/A'}'),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<Map<String, dynamic>>(
      future: fetchAppointmentDetails(),
      builder: (context, snapshot) {
        if (snapshot.hasData) {
          final appointment = snapshot.data!['appointment'];
          final slotDetails = appointment['slot_details'];
          
          return Scaffold(
            appBar: AppBar(title: Text('Appointment Details')),
            body: SingleChildScrollView(
              padding: EdgeInsets.all(16),
              child: Column(
                children: [
                  // Patient info
                  Card(child: /* patient details */),
                  
                  // Doctor info
                  Card(child: /* doctor details */),
                  
                  // Slot info (conditional)
                  buildSlotInfo(slotDetails),
                  
                  // Other details...
                ],
              ),
            ),
          );
        }
        
        return CircularProgressIndicator();
      },
    );
  }
}
```

## Migration Notes

### Backward Compatibility
- ✅ Existing appointments without slot information continue to work
- ✅ Response structure remains valid JSON
- ✅ No breaking changes to existing fields
- ✅ New field (`slot_details`) is optional

### Frontend Considerations
Always check if `slot_details` exists before accessing it:

```javascript
// ✅ Correct
if (appointment.slot_details) {
  displaySlotInfo(appointment.slot_details);
}

// ❌ Incorrect
displaySlotInfo(appointment.slot_details); // May cause error if undefined
```

## Testing

### Test Script
A PowerShell test script is available: `test-appointment-details.ps1`

```powershell
# Update configuration
$BASE_URL = "http://localhost:8082/api/v1"
$TOKEN = "your-auth-token"
$appointmentId = "actual-appointment-id"

# Run the test
.\test-appointment-details.ps1
```

### Expected Output
```
=== Test 1: Get Appointment Details ===
✓ Success!

Appointment Details:
ID: appointment-uuid
Booking Number: BK20241017001
Token Number: T001
Patient: John Doe (MO123456)
Doctor: Dr. Smith
Department: Cardiology
Date/Time: 2024-10-17 10:30:00
Status: scheduled
Payment: paid

Slot Details:
Slot ID: slot-uuid
Slot Time: 2024-10-17 10:30:00 - 10:45:00
Session: Morning Session
Slot Status: booked
```

## Files Modified

1. **Controller**: `services/appointment-service/controllers/appointment_list_simple.controller.go`
   - Enhanced `GetSimpleAppointmentDetails()` function
   - Added SQL joins for slot tables
   - Added slot data formatting logic

2. **Route**: `services/appointment-service/routes/appointment.routes.go`
   - Added: `GET /appointments/simple/:id`

3. **Documentation**: 
   - `GET_APPOINTMENT_DETAILS_API.md` - Complete API documentation
   - `APPOINTMENT_DETAILS_WITH_SLOT_INFO.md` - This file
   - `test-appointment-details.ps1` - Test script

## Related Documentation

- [Get Appointment Details API](./GET_APPOINTMENT_DETAILS_API.md)
- [Session-Based Slots Guide](./SESSION_SLOTS_COMPLETE_API_DOCUMENTATION.md)
- [Individual Slot Booking Guide](./APPOINTMENT_INDIVIDUAL_SLOT_BOOKING_COMPLETE.md)
- [Appointment Creation API](./SIMPLE_APPOINTMENT_API_GUIDE.md)

## Benefits

1. **Complete Information**: Users get full details about their booked slot
2. **Better UX**: Display exact slot timings on appointment details
3. **Slot Tracking**: Easy to track which slots are being used
4. **Print Support**: Slot details available for printing appointment slips
5. **Backward Compatible**: Works with both slot-based and legacy appointments
6. **Flexible**: Optional field doesn't break existing integrations

## Version History

| Version | Date       | Changes                                        |
|---------|------------|------------------------------------------------|
| 1.0     | 2024-10-17 | Initial implementation with slot details       |

---

**Last Updated**: October 17, 2024  
**Author**: Dr&Me Backend Team  
**Status**: ✅ Implemented and Tested

