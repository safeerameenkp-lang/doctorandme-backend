# Appointment List API

## Overview
New API endpoint specifically designed for the appointment list UI, providing all the data fields required by the frontend table display.

## API Endpoint

**Endpoint**: `GET /api/appointments/list`

**Access**: Clinic Admin, Doctor, Receptionist

**Purpose**: Get appointments list formatted for UI table display

## Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `clinic_id` | string | No | Filter by clinic ID |
| `doctor_id` | string | No | Filter by doctor ID |
| `patient_id` | string | No | Filter by patient ID |
| `status` | string | No | Filter by appointment status |
| `date` | string | No | Filter by appointment date (YYYY-MM-DD) |
| `limit` | integer | No | Number of records to return (default: 50) |
| `offset` | integer | No | Number of records to skip (default: 0) |

## Response Structure

```json
{
  "appointments": [
    {
      "id": "appointment-uuid",
      "serial_number": 1,
      "mo_id": "#23455H",
      "patient_name": "Sarah Johnson (Patient)",
      "doctor_name": "Dr. Maria Garcia",
      "department": "Dermatology",
      "consultation_type": "Follow Up",
      "appointment_date_time": "12-03-2025 10:30 AM",
      "status": "Completed",
      "fee_status": "₹600.00",
      "fee_amount": 600.00,
      "payment_status": "paid",
      "booking_number": "APT-001",
      "created_at": "2025-03-12T10:30:00Z"
    }
  ],
  "total_count": 1
}
```

## Field Descriptions

### Appointment List Item Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique appointment identifier |
| `serial_number` | integer | Sequential number for table display |
| `mo_id` | string | Patient's Medical Officer ID (nullable) |
| `patient_name` | string | Patient's full name with "(Patient)" suffix |
| `doctor_name` | string | Doctor's full name with "Dr." prefix |
| `department` | string | Department name (nullable) |
| `consultation_type` | string | Type of consultation (Follow Up, Online Consultation, Clinic Visit) |
| `appointment_date_time` | string | Formatted date and time (DD-MM-YYYY HH:MM AM/PM) |
| `status` | string | Appointment status (Completed, Cancelled, Upcoming) |
| `fee_status` | string | Fee display text (₹amount or "Pay Now") |
| `fee_amount` | float | Actual fee amount (nullable) |
| `payment_status` | string | Payment status (paid, pending, etc.) |
| `booking_number` | string | Booking reference number |
| `created_at` | string | ISO timestamp of appointment creation |

## Status Values

### Appointment Status
- `completed` - Appointment finished successfully
- `cancelled` - Appointment was cancelled
- `upcoming` - Appointment scheduled for future
- `in_progress` - Appointment currently ongoing
- `no_show` - Patient didn't show up

### Payment Status
- `paid` - Payment completed
- `pending` - Payment pending
- `failed` - Payment failed
- `refunded` - Payment refunded

### Consultation Types
- `follow_up` - Follow-up consultation
- `online` - Online/video consultation
- `offline` - In-person clinic visit
- `video` - Video call consultation
- `in_person` - Physical clinic visit

## Fee Status Logic

The `fee_status` field is determined by:
- If `payment_status` is "paid" and `fee_amount` exists: Shows "₹{amount}"
- Otherwise: Shows "Pay Now"

## Usage Examples

### Get All Appointments for a Clinic
```bash
curl -X GET "http://localhost:8080/api/appointments/list?clinic_id=clinic-uuid" \
  -H "Authorization: Bearer clinic-admin-token"
```

### Get Appointments by Status
```bash
curl -X GET "http://localhost:8080/api/appointments/list?status=completed" \
  -H "Authorization: Bearer doctor-token"
```

### Get Appointments by Date
```bash
curl -X GET "http://localhost:8080/api/appointments/list?date=2025-03-12" \
  -H "Authorization: Bearer receptionist-token"
```

### Get Appointments with Pagination
```bash
curl -X GET "http://localhost:8080/api/appointments/list?limit=20&offset=0" \
  -H "Authorization: Bearer clinic-admin-token"
```

### Get Appointments by Doctor
```bash
curl -X GET "http://localhost:8080/api/appointments/list?doctor_id=doctor-uuid" \
  -H "Authorization: Bearer doctor-token"
```

## Response Examples

### Successful Response
```json
{
  "appointments": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "serial_number": 1,
      "mo_id": "#23455H",
      "patient_name": "Sarah Johnson (Patient)",
      "doctor_name": "Dr. Maria Garcia",
      "department": "Dermatology",
      "consultation_type": "Follow Up",
      "appointment_date_time": "12-03-2025 10:30 AM",
      "status": "Completed",
      "fee_status": "₹600.00",
      "fee_amount": 600.00,
      "payment_status": "paid",
      "booking_number": "APT-001",
      "created_at": "2025-03-12T10:30:00Z"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "serial_number": 2,
      "mo_id": "#23456I",
      "patient_name": "John Smith (Patient)",
      "doctor_name": "Dr. David Wilson",
      "department": "Neurology",
      "consultation_type": "Online Consultation",
      "appointment_date_time": "13-03-2025 02:00 PM",
      "status": "Cancelled",
      "fee_status": "Pay Now",
      "fee_amount": 500.00,
      "payment_status": "pending",
      "booking_number": "APT-002",
      "created_at": "2025-03-13T14:00:00Z"
    }
  ],
  "total_count": 2
}
```

### Empty Response
```json
{
  "appointments": [],
  "total_count": 0
}
```

## Error Responses

### Unauthorized Access
```json
{
  "error": "Unauthorized",
  "message": "Invalid or missing token"
}
```

### Database Error
```json
{
  "error": "Database error"
}
```

### Invalid Query Parameters
```json
{
  "error": "Invalid query parameters",
  "message": "Invalid date format"
}
```

## Database Query

The API uses an optimized query that joins multiple tables:

```sql
SELECT a.id, a.booking_number, a.appointment_time, a.consultation_type, 
       a.status, a.fee_amount, a.payment_status, a.created_at,
       p.mo_id, u.first_name as patient_first_name, u.last_name as patient_last_name,
       du.first_name as doctor_first_name, du.last_name as doctor_last_name,
       dept.name as department_name
FROM appointments a
JOIN patients p ON p.id = a.patient_id
JOIN users u ON u.id = p.user_id
JOIN doctors d ON d.id = a.doctor_id
JOIN users du ON du.id = d.user_id
LEFT JOIN departments dept ON dept.id = a.department_id
WHERE 1=1
ORDER BY a.appointment_time DESC
```

## Performance Considerations

### Indexing
Ensure proper indexes exist on:
- `appointments.clinic_id`
- `appointments.doctor_id`
- `appointments.patient_id`
- `appointments.status`
- `appointments.appointment_time`

### Pagination
- Default limit: 50 records
- Maximum recommended limit: 100 records
- Use `offset` for pagination

### Filtering
- Combine multiple filters for better performance
- Use date filters to limit result set
- Filter by clinic_id for clinic-specific views

## Integration with Frontend

### Table Display
The response structure is designed to match the UI table columns:

| UI Column | API Field | Description |
|-----------|-----------|-------------|
| # | `serial_number` | Sequential number |
| Mo ID | `mo_id` | Patient's MO ID |
| Patient Name | `patient_name` | Full patient name |
| Doctor Name | `doctor_name` | Full doctor name |
| Department | `department` | Department name |
| Consultation Type | `consultation_type` | Type of consultation |
| Appointment Date & Time | `appointment_date_time` | Formatted datetime |
| STATUS | `status` | Appointment status |
| Fee Status | `fee_status` | Fee display text |

### Status Indicators
- **Completed**: Green dot + "Completed" text
- **Cancelled**: Red dot + "Cancelled" text  
- **Upcoming**: Yellow dot + "Upcoming" text

### Fee Status Display
- **Paid**: Shows amount (e.g., "₹600.00")
- **Pending**: Shows "Pay Now" (clickable link)

## Security

### Role-Based Access
- **Clinic Admin**: Can view all appointments in their clinic
- **Doctor**: Can view their own appointments
- **Receptionist**: Can view appointments in their clinic

### Data Filtering
- Clinic admins see only their clinic's appointments
- Doctors see only their own appointments
- Receptionists see only their clinic's appointments

## Testing

### Test Cases
1. **Basic List**: Get all appointments
2. **Clinic Filter**: Filter by clinic ID
3. **Status Filter**: Filter by appointment status
4. **Date Filter**: Filter by specific date
5. **Pagination**: Test limit and offset
6. **Empty Results**: Test with no matching records
7. **Role Access**: Test different user roles

### Sample Test Data
```json
{
  "appointments": [
    {
      "id": "test-appointment-1",
      "serial_number": 1,
      "mo_id": "#TEST001",
      "patient_name": "Test Patient (Patient)",
      "doctor_name": "Dr. Test Doctor",
      "department": "Test Department",
      "consultation_type": "Follow Up",
      "appointment_date_time": "15-03-2025 11:00 AM",
      "status": "Upcoming",
      "fee_status": "Pay Now",
      "fee_amount": 500.00,
      "payment_status": "pending",
      "booking_number": "TEST-001",
      "created_at": "2025-03-15T11:00:00Z"
    }
  ],
  "total_count": 1
}
```

## Comparison with Original API

### Original GetAppointments API
- Returns full appointment details with nested objects
- Includes patient, doctor, and clinic information
- More comprehensive but heavier response

### New GetAppointmentList API
- Returns flattened structure optimized for table display
- Includes only necessary fields for UI
- Lighter response, faster loading
- Serial numbers for table display
- Formatted date/time strings
- Simplified fee status logic

## Migration Guide

### From GetAppointments to GetAppointmentList

**Before (GetAppointments)**:
```javascript
// Frontend code
const response = await fetch('/api/appointments');
const appointments = response.data;
// Process nested structure
```

**After (GetAppointmentList)**:
```javascript
// Frontend code
const response = await fetch('/api/appointments/list');
const appointments = response.appointments;
// Direct table mapping
```

### Benefits of Migration
1. **Performance**: Faster response times
2. **Simplicity**: Easier frontend integration
3. **Consistency**: Matches UI requirements exactly
4. **Maintainability**: Clear separation of concerns

## Conclusion

The new Appointment List API provides:
- ✅ **UI-optimized response** structure
- ✅ **Serial numbers** for table display
- ✅ **Formatted date/time** strings
- ✅ **Simplified fee status** logic
- ✅ **Department information** included
- ✅ **Role-based access control**
- ✅ **Comprehensive filtering** options
- ✅ **Pagination support**
- ✅ **Performance optimized** queries
- ✅ **Frontend-ready** data format

This API is specifically designed to match the appointment list UI requirements and provides all necessary data in the exact format needed for table display.
