# Dummy Data Insertion Summary

## Overview
Successfully inserted dummy patients and appointments into the database for testing the appointment list API.

## Clinic Information
- **Clinic ID**: `699ee0fc-db11-4f50-9f5f-c4f0191b344e`
- **Clinic Name**: `abc clinic`

## Dummy Patients Created

### 1. Sarah Johnson
- **Patient ID**: `660e8400-e29b-41d4-a716-446655440001`
- **User ID**: `550e8400-e29b-41d4-a716-446655440001`
- **Mo ID**: `#23455H`
- **Phone**: `1234567890`
- **Email**: `sarah.johnson@email.com`
- **Medical History**: Diabetes, Hypertension
- **Allergies**: Penicillin
- **Blood Group**: A+

### 2. John Smith
- **Patient ID**: `660e8400-e29b-41d4-a716-446655440002`
- **User ID**: `550e8400-e29b-41d4-a716-446655440002`
- **Mo ID**: `#23456I`
- **Phone**: `2345678901`
- **Email**: `john.smith@email.com`
- **Medical History**: No significant medical history
- **Allergies**: None known
- **Blood Group**: O+

### 3. Emily Davis
- **Patient ID**: `660e8400-e29b-41d4-a716-446655440003`
- **User ID**: `550e8400-e29b-41d4-a716-446655440003`
- **Mo ID**: `#23457J`
- **Phone**: `3456789012`
- **Email**: `emily.davis@email.com`
- **Medical History**: Asthma
- **Allergies**: Shellfish
- **Blood Group**: B+

### 4. Michael Chen
- **Patient ID**: `660e8400-e29b-41d4-a716-446655440004`
- **User ID**: `550e8400-e29b-41d4-a716-446655440004`
- **Mo ID**: `#23458K`
- **Phone**: `4567890123`
- **Email**: `michael.chen@email.com`
- **Medical History**: Migraine
- **Allergies**: Aspirin
- **Blood Group**: AB+

### 5. Anna Rodriguez
- **Patient ID**: `660e8400-e29b-41d4-a716-446655440005`
- **User ID**: `550e8400-e29b-41d4-a716-446655440005`
- **Mo ID**: `#23459L`
- **Phone**: `5678901234`
- **Email**: `anna.rodriguez@email.com`
- **Medical History**: No significant medical history
- **Allergies**: None known
- **Blood Group**: O-

## Dummy Appointments Created

### 1. APT-001 - Sarah Johnson
- **Appointment ID**: `770e8400-e29b-41d4-a716-446655440001`
- **Doctor**: amar aw
- **Date**: 2025-03-12 10:30:00
- **Type**: Follow Up
- **Status**: Completed
- **Fee**: ₹600.00
- **Payment**: Paid

### 2. APT-002 - John Smith
- **Appointment ID**: `770e8400-e29b-41d4-a716-446655440002`
- **Doctor**: dsfsdf sdfsdfsdfsdfsdfs
- **Date**: 2025-03-13 14:00:00
- **Type**: Online Consultation
- **Status**: Cancelled
- **Fee**: ₹500.00
- **Payment**: Pending

### 3. APT-003 - Emily Davis
- **Appointment ID**: `770e8400-e29b-41d4-a716-446655440003`
- **Doctor**: ameeeeena ameeeeen
- **Date**: 2025-03-14 11:15:00
- **Type**: Clinic Visit
- **Status**: Cancelled
- **Fee**: ₹700.00
- **Payment**: Pending

### 4. APT-004 - Michael Chen
- **Appointment ID**: `770e8400-e29b-41d4-a716-446655440004`
- **Doctor**: sabiiiikkkkkkkkkk dddd
- **Date**: 2025-03-15 15:30:00
- **Type**: Follow Up
- **Status**: Cancelled
- **Fee**: ₹600.00
- **Payment**: Pending

### 5. APT-005 - Anna Rodriguez
- **Appointment ID**: `770e8400-e29b-41d4-a716-446655440005`
- **Doctor**: monuuu saaa
- **Date**: 2025-03-16 09:45:00
- **Type**: Clinic Visit
- **Status**: Upcoming
- **Fee**: ₹600.00
- **Payment**: Pending

### 6. APT-006 - Sarah Johnson
- **Appointment ID**: `770e8400-e29b-41d4-a716-446655440006`
- **Doctor**: amar aw
- **Date**: 2025-03-17 13:00:00
- **Type**: Online Consultation
- **Status**: Completed
- **Fee**: ₹600.00
- **Payment**: Paid

### 7. APT-007 - John Smith
- **Appointment ID**: `770e8400-e29b-41d4-a716-446655440007`
- **Doctor**: dsfsdf sdfsdfsdfsdfsdfs
- **Date**: 2025-03-18 16:15:00
- **Type**: Clinic Visit
- **Status**: Cancelled
- **Fee**: ₹700.00
- **Payment**: Pending

### 8. APT-008 - Emily Davis
- **Appointment ID**: `770e8400-e29b-41d4-a716-446655440008`
- **Doctor**: ameeeeena ameeeeen
- **Date**: 2025-03-19 10:00:00
- **Type**: Follow Up
- **Status**: Upcoming
- **Fee**: ₹600.00
- **Payment**: Pending

### 9. APT-009 - Michael Chen
- **Appointment ID**: `770e8400-e29b-41d4-a716-446655440009`
- **Doctor**: sabiiiikkkkkkkkkk dddd
- **Date**: 2025-03-20 14:30:00
- **Type**: Online Consultation
- **Status**: Completed
- **Fee**: ₹600.00
- **Payment**: Paid

## Database Tables Updated

### 1. Users Table
- Inserted 5 user records with patient information
- Assigned patient roles to all users

### 2. Patients Table
- Created 5 patient records linked to users
- Added medical history, allergies, and blood group information

### 3. Patient Clinics Table
- Linked all patients to the clinic `699ee0fc-db11-4f50-9f5f-c4f0191b344e`
- Set all as primary clinic assignments

### 4. Appointments Table
- Created 9 appointment records
- Linked to existing doctors and patients
- Included various consultation types and statuses

## API Testing

### Endpoint
```
GET http://localhost:8082/api/v1/appointments/list?clinic_id=699ee0fc-db11-4f50-9f5f-c4f0191b344e
```

### Expected Response Format
```json
{
  "appointments": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440001",
      "serial_number": 1,
      "mo_id": "#23455H",
      "patient_name": "Sarah Johnson (Patient)",
      "doctor_name": "Dr. amar aw",
      "department": null,
      "consultation_type": "Follow Up",
      "appointment_date_time": "12-03-2025 10:30 AM",
      "status": "completed",
      "fee_status": "₹600.00",
      "fee_amount": 600.00,
      "payment_status": "paid",
      "booking_number": "APT-001",
      "created_at": "2025-03-12T10:30:00Z"
    }
    // ... more appointments
  ],
  "total_count": 9
}
```

## Data Verification

### SQL Query to Verify Data
```sql
-- Check appointments for the clinic
SELECT 
    a.id,
    a.booking_number,
    u.first_name || ' ' || u.last_name as patient_name,
    du.first_name || ' ' || du.last_name as doctor_name,
    a.consultation_type,
    a.appointment_time,
    a.status,
    a.fee_amount,
    a.payment_status
FROM appointments a
JOIN patients p ON p.id = a.patient_id
JOIN users u ON u.id = p.user_id
JOIN doctors d ON d.id = a.doctor_id
JOIN users du ON du.id = d.user_id
WHERE a.clinic_id = '699ee0fc-db11-4f50-9f5f-c4f0191b344e'
ORDER BY a.appointment_time;
```

### Results
- ✅ 9 appointments created successfully
- ✅ All appointments linked to the correct clinic
- ✅ Patients properly assigned to the clinic
- ✅ Various consultation types included (Follow Up, Online Consultation, Clinic Visit)
- ✅ Different statuses included (Completed, Cancelled, Upcoming)
- ✅ Different payment statuses included (Paid, Pending)

## Next Steps

1. **Authentication**: Obtain a valid JWT token for API testing
2. **API Testing**: Test the appointment list API with authentication
3. **UI Integration**: Use this data to test the frontend appointment list UI
4. **Data Validation**: Verify that the API returns data in the expected format

## Notes

- All dummy data uses realistic values
- Patient Mo IDs follow the pattern `#23455H`, `#23456I`, etc.
- Appointment booking numbers follow the pattern `APT-001`, `APT-002`, etc.
- Consultation types match the UI requirements (Follow Up, Online Consultation, Clinic Visit)
- Fee amounts vary between ₹500-700
- Payment statuses include both paid and pending for testing different UI states
