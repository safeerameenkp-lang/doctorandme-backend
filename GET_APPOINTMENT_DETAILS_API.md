# Get Single Appointment Details API

## Overview
This API endpoint retrieves detailed information about a single appointment by its ID.

## Endpoint
```
GET /api/v1/appointments/simple/:id
```

## Authentication
- **Required**: Yes
- **Roles**: `clinic_admin`, `receptionist`, `doctor`
- **Header**: `Authorization: Bearer <token>`

## Path Parameters

| Parameter | Type   | Required | Description           |
|-----------|--------|----------|-----------------------|
| `id`      | string | Yes      | Appointment ID (UUID) |

## Response

### Success Response (200 OK)
```json
{
  "success": true,
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

**Note**: This API returns the **same data structure** as the list API (`GetSimpleAppointmentList`), but for a single appointment.

### Error Responses

#### 400 Bad Request
```json
{
  "error": "appointment_id is required"
}
```

#### 404 Not Found
```json
{
  "error": "Appointment not found",
  "details": "sql: no rows in result set"
}
```

#### 401 Unauthorized
```json
{
  "error": "Unauthorized"
}
```

#### 403 Forbidden
```json
{
  "error": "Forbidden - insufficient permissions"
}
```

## Response Fields

### Appointment Object

| Field                    | Type   | Description                                      |
|--------------------------|--------|--------------------------------------------------|
| `id`                     | string | Unique appointment identifier                    |
| `token_number`           | string | Token number for the appointment                 |
| `mo_id`                  | string | Patient's Medical Office ID                      |
| `patient_name`           | string | Patient's full name                              |
| `doctor_name`            | string | Doctor's full name                               |
| `department`             | string | Department name                                  |
| `consultation_type`      | string | Type: `new`, `followup`, `emergency`             |
| `appointment_date_time`  | string | Full appointment date and time                   |
| `status`                 | string | Status: `scheduled`, `completed`, `cancelled`, etc. |
| `fee_amount`             | float  | Consultation fee                                 |
| `payment_status`         | string | Payment: `paid`, `pending`, `refunded`           |
| `fee_status`             | string | Same as payment_status (for compatibility)       |
| `booking_number`         | string | Unique booking reference number                  |
| `created_at`             | string | Timestamp when appointment was created           |

**Note**: This API uses a **simplified structure** matching the appointment list API. All essential appointment information is returned in a flat, easy-to-use format.

## Usage Examples

### cURL
```bash
curl -X GET "http://localhost:8082/api/v1/appointments/simple/550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json"
```

### JavaScript (Fetch)
```javascript
const appointmentId = '550e8400-e29b-41d4-a716-446655440000';

fetch(`http://localhost:8082/api/v1/appointments/simple/${appointmentId}`, {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer YOUR_TOKEN_HERE',
    'Content-Type': 'application/json'
  }
})
.then(response => response.json())
.then(data => {
  console.log('Appointment Details:', data.appointment);
})
.catch(error => console.error('Error:', error));
```

### Flutter (Dart)
```dart
import 'package:http/http.dart' as http;
import 'dart:convert';

Future<Map<String, dynamic>> getAppointmentDetails(String appointmentId) async {
  final response = await http.get(
    Uri.parse('http://localhost:8082/api/v1/appointments/simple/$appointmentId'),
    headers: {
      'Authorization': 'Bearer YOUR_TOKEN_HERE',
      'Content-Type': 'application/json',
    },
  );

  if (response.statusCode == 200) {
    return json.decode(response.body);
  } else {
    throw Exception('Failed to load appointment details');
  }
}

// Usage
void main() async {
  try {
    final result = await getAppointmentDetails('550e8400-e29b-41d4-a716-446655440000');
    print('Appointment: ${result['appointment']}');
  } catch (e) {
    print('Error: $e');
  }
}
```

### Python (Requests)
```python
import requests

appointment_id = "550e8400-e29b-41d4-a716-446655440000"
url = f"http://localhost:8082/api/v1/appointments/simple/{appointment_id}"

headers = {
    "Authorization": "Bearer YOUR_TOKEN_HERE",
    "Content-Type": "application/json"
}

response = requests.get(url, headers=headers)

if response.status_code == 200:
    data = response.json()
    print("Appointment Details:", data['appointment'])
else:
    print("Error:", response.json())
```

### PowerShell
```powershell
$appointmentId = "550e8400-e29b-41d4-a716-446655440000"
$url = "http://localhost:8082/api/v1/appointments/simple/$appointmentId"

$headers = @{
    "Authorization" = "Bearer YOUR_TOKEN_HERE"
    "Content-Type" = "application/json"
}

$response = Invoke-RestMethod -Uri $url -Method GET -Headers $headers
Write-Host "Appointment Details:"
$response.appointment | ConvertTo-Json -Depth 5
```

## Use Cases

1. **View Appointment Details**: Display complete appointment information on a details page
2. **Print Appointment**: Get full details for printing appointment slips
3. **Verify Appointment**: Check appointment information before check-in
4. **Update References**: Retrieve current appointment data before updates
5. **Audit Trail**: Get appointment details for record-keeping

## Notes

- This API uses the **exact same data structure** as `GetSimpleAppointmentList` for consistency
- Returns a single appointment instead of a list
- All fields that can be null are properly handled
- The appointment date and time are combined into a single formatted string
- The API uses LEFT JOINs to ensure data is returned even if some related records are missing
- Department information is retrieved from either the appointment's department_id or the doctor's department_id
- Simple, flat structure makes it easy to display appointment details in your UI
- `fee_status` is provided for backward compatibility (same value as `payment_status`)

## Related APIs

- **List Appointments**: `GET /api/v1/appointments/simple-list` - Get list of appointments
- **Create Appointment**: `POST /api/v1/appointments/simple` - Create new appointment
- **Reschedule Appointment**: `POST /api/v1/appointments/:id/reschedule-simple` - Reschedule appointment

## Version
- **Added**: v1.0
- **Last Updated**: October 17, 2024

