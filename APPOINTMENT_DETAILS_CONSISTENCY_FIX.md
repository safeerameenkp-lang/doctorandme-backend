# Appointment Details API - Consistency Fix

## Issue
The `GetSimpleAppointmentDetails` API was using different fields than `GetSimpleAppointmentList`, causing SQL errors when trying to access non-existent columns.

### Error
```
Request failed with status 404: 
{"details":"pq: column [column_name] does not exist","error":"Appointment not found"}
```

## Root Cause
The two APIs were using different SQL queries:
- **GetSimpleAppointmentList**: Simple query with basic fields
- **GetSimpleAppointmentDetails**: Complex query with extra fields (phone, email, age, payment_mode, clinic details, slot details, etc.)

This mismatch caused:
1. SQL errors when columns didn't exist in the database
2. Inconsistent data structure between list and details APIs
3. Confusion for frontend developers

## Solution
✅ **Made both APIs use the EXACT SAME data structure**

### Fields Now Used by Both APIs

| Field                  | Type   | Source                                    |
|------------------------|--------|-------------------------------------------|
| `id`                   | string | `appointments.id`                         |
| `token_number`         | string | `appointments.token_number`               |
| `mo_id`                | string | `clinic_patients.mo_id`                   |
| `patient_name`         | string | `clinic_patients.first_name + last_name`  |
| `doctor_name`          | string | `users.first_name + last_name`            |
| `department`           | string | `departments.name`                        |
| `consultation_type`    | string | `appointments.consultation_type`          |
| `appointment_date`     | string | `appointments.appointment_date`           |
| `appointment_time`     | time   | `appointments.appointment_time`           |
| `status`               | string | `appointments.status`                     |
| `fee_amount`           | float  | `appointments.fee_amount`                 |
| `payment_status`       | string | `appointments.payment_status`             |
| `booking_number`       | string | `appointments.booking_number`             |
| `created_at`           | time   | `appointments.created_at`                 |

### Response Structure (Both APIs)

**List API Response:**
```json
{
  "success": true,
  "clinic_id": "clinic-uuid",
  "date": "2024-10-17",
  "total": 1,
  "appointments": [
    {
      "id": "appointment-uuid",
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
  ]
}
```

**Details API Response:**
```json
{
  "success": true,
  "appointment": {
    "id": "appointment-uuid",
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

**Notice**: The appointment object in the details API is **identical** to each item in the list API's appointments array!

## Changes Made

### File: `appointment_list_simple.controller.go`

#### GetSimpleAppointmentDetails Function
- ✅ Simplified SQL query to match GetSimpleAppointmentList
- ✅ Removed extra fields: patient phone, email, age, gender
- ✅ Removed extra fields: payment_mode, clinic details, slot details
- ✅ Removed department_id, doctor_id (only return names)
- ✅ Used same variable structure as list
- ✅ Used same response format as list items

### SQL Query (Both Functions)
```sql
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
```

## Benefits

### 1. ✅ Consistency
Both APIs now return the same data structure, making frontend development easier.

### 2. ✅ No SQL Errors
Only uses columns that exist in the database.

### 3. ✅ Simplified Code
Less complex queries, easier to maintain.

### 4. ✅ Better Performance
Fewer joins and fields to retrieve.

### 5. ✅ Easy to Use
Frontend developers can use the same data model for both list and detail views.

## Frontend Integration

### Flutter Example
```dart
class AppointmentItem {
  final String id;
  final String tokenNumber;
  final String moId;
  final String patientName;
  final String doctorName;
  final String? department;
  final String? consultationType;
  final String appointmentDateTime;
  final String status;
  final double? feeAmount;
  final String paymentStatus;
  final String feeStatus;
  final String bookingNumber;
  final String createdAt;

  // This SAME model works for BOTH list items AND detail view!
  factory AppointmentItem.fromJson(Map<String, dynamic> json) {
    return AppointmentItem(
      id: json['id'],
      tokenNumber: json['token_number'],
      moId: json['mo_id'],
      patientName: json['patient_name'],
      doctorName: json['doctor_name'],
      department: json['department'],
      consultationType: json['consultation_type'],
      appointmentDateTime: json['appointment_date_time'],
      status: json['status'],
      feeAmount: json['fee_amount']?.toDouble(),
      paymentStatus: json['payment_status'],
      feeStatus: json['fee_status'],
      bookingNumber: json['booking_number'],
      createdAt: json['created_at'],
    );
  }
}

// Usage in list
Future<List<AppointmentItem>> fetchAppointments() async {
  final response = await http.get(Uri.parse('$baseUrl/appointments/simple-list?clinic_id=$clinicId'));
  final data = json.decode(response.body);
  return (data['appointments'] as List)
      .map((item) => AppointmentItem.fromJson(item))
      .toList();
}

// Usage in details
Future<AppointmentItem> fetchAppointmentDetails(String id) async {
  final response = await http.get(Uri.parse('$baseUrl/appointments/simple/$id'));
  final data = json.decode(response.body);
  return AppointmentItem.fromJson(data['appointment']);
}
```

## Migration Notes

### Breaking Changes
⚠️ If you were using the details API and expecting these fields, they are now removed:
- `patient.phone`
- `patient.email`
- `patient.age`
- `patient.gender`
- `doctor.id`
- `department.id`
- `clinic` object
- `payment_method`
- `slot_details` object
- `notes`
- `cancellation_reason`
- `duration_minutes`
- `session_type`
- `updated_at`

### Solution
If you need these additional fields, use the original detailed API:
```
GET /api/v1/appointments/:id
```

The **simple** APIs are meant for lightweight list/details views with consistent structure.

## Testing

### Test the List API
```bash
curl -X GET "http://localhost:8082/api/v1/appointments/simple-list?clinic_id=YOUR_CLINIC_ID" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Test the Details API
```bash
curl -X GET "http://localhost:8082/api/v1/appointments/simple/APPOINTMENT_ID" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Expected Result
The appointment object in the details response should match each item in the list response.

## Files Modified
1. ✅ `services/appointment-service/controllers/appointment_list_simple.controller.go`
2. ✅ `GET_APPOINTMENT_DETAILS_API.md` (documentation)

## Status
✅ **FIXED AND TESTED**

Both APIs now use identical data structures and work without SQL errors.

---

**Date**: October 17, 2024  
**Issue**: SQL column errors and API inconsistency  
**Solution**: Aligned both APIs to use the same simple structure  
**Result**: Consistent, working APIs ✅

