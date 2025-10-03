# Patient Creation During Appointment Booking

## Overview

This feature allows reception staff to create new patients and book appointments in a single operation. This is particularly useful for walk-in patients or when adding new patients during appointment booking.

## New Features Added

### 1. MO ID Field
- Added `mo_id` field to the patients table
- MO ID (Medical Officer ID) is a unique identifier assigned by medical officers
- Field is optional and can be up to 50 characters
- Added database index for better performance

### 2. New API Endpoint
- **POST** `/api/appointments/with-patient`
- Creates a new patient and books an appointment in a single transaction
- Available to `clinic_admin` and `receptionist` roles

### 3. Enhanced Patient Model
- Updated patient model to include MO ID field
- Updated all patient-related queries to include MO ID
- Updated appointment queries to return MO ID in patient information

## API Usage

### Endpoint: POST /api/appointments/with-patient

#### Request Body
```json
{
  "first_name": "John",
  "last_name": "Doe", 
  "phone": "+1234567890",
  "email": "john.doe@example.com",
  "date_of_birth": "1990-01-15",
  "gender": "Male",
  "mo_id": "MO123456",
  "medical_history": "No significant medical history",
  "allergies": "None known",
  "blood_group": "O+",
  "clinic_id": "clinic-uuid",
  "doctor_id": "doctor-uuid", 
  "appointment_time": "2024-01-20 10:00:00",
  "duration_minutes": 15,
  "consultation_type": "new",
  "is_priority": false,
  "payment_mode": "cash"
}
```

#### Required Fields
- `first_name`: Patient's first name (max 100 characters)
- `last_name`: Patient's last name (max 100 characters)
- `phone`: Patient's phone number (max 20 characters)
- `clinic_id`: UUID of the clinic
- `doctor_id`: UUID of the doctor
- `appointment_time`: Appointment time in YYYY-MM-DD HH:MM:SS format
- `consultation_type`: One of: new, followup, walkin, emergency

#### Optional Fields
- `email`: Patient's email address
- `date_of_birth`: Date of birth in YYYY-MM-DD format
- `gender`: Patient's gender (max 20 characters)
- `mo_id`: Medical Officer ID (max 50 characters)
- `medical_history`: Patient's medical history
- `allergies`: Known allergies
- `blood_group`: Blood group (max 10 characters)
- `duration_minutes`: Appointment duration (default: 12 minutes)
- `is_priority`: Whether this is a priority appointment (default: false)
- `payment_mode`: Payment method (cash, card, upi)

#### Response
```json
{
  "appointment": {
    "id": "appointment-uuid",
    "patient_id": "patient-uuid",
    "clinic_id": "clinic-uuid",
    "doctor_id": "doctor-uuid",
    "booking_number": "DOC001-20240120-001",
    "appointment_time": "2024-01-20T10:00:00Z",
    "duration_minutes": 15,
    "consultation_type": "new",
    "status": "booked",
    "fee_amount": 100.00,
    "payment_status": "paid",
    "payment_mode": "cash",
    "is_priority": false,
    "created_at": "2024-01-20T09:00:00Z"
  },
  "patient": {
    "id": "patient-uuid",
    "user_id": "user-uuid",
    "mo_id": "MO123456",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890",
    "email": "john.doe@example.com",
    "medical_history": "No significant medical history",
    "allergies": "None known",
    "blood_group": "O+"
  },
  "message": "Patient created and appointment booked successfully"
}
```

## Database Changes

### Migration: 002_add_mo_id_to_patients.sql
```sql
-- Add MO ID field to patients table
ALTER TABLE patients ADD COLUMN mo_id VARCHAR(50);

-- Create index for MO ID for better performance
CREATE INDEX idx_patients_mo_id ON patients(mo_id);

-- Add comment to explain the field
COMMENT ON COLUMN patients.mo_id IS 'Medical Officer ID - unique identifier for the patient assigned by medical officer';
```

## Business Logic

### Patient Creation Process
1. **User Validation**: Check if a user with the same phone number already exists
2. **User Creation**: If no existing user, create a new user record
3. **Patient Creation**: Create patient record linked to the user
4. **Clinic Assignment**: Automatically assign patient to the clinic as primary
5. **Appointment Booking**: Create appointment with all validations
6. **Payment Processing**: Handle immediate payment if specified
7. **Auto Check-in**: Create check-in record if payment is completed

### Transaction Safety
- All operations are wrapped in a database transaction
- If any step fails, all changes are rolled back
- Ensures data consistency across user, patient, and appointment records

### Duplicate Prevention
- Checks for existing users by phone number
- Prevents duplicate patient records for the same user
- Handles existing users who don't have patient records yet

## Error Handling

### Common Error Responses
- **400 Bad Request**: Invalid input data or validation errors
- **404 Not Found**: Doctor or clinic not found
- **409 Conflict**: Patient already exists with this phone number
- **500 Internal Server Error**: Database or system errors

### Validation Errors
- Invalid date formats
- Missing required fields
- Invalid consultation types
- Doctor not available at requested time

## Security

### Role-Based Access
- Only `clinic_admin` and `receptionist` roles can create patients with appointments
- Authentication required for all requests
- Input validation and sanitization

### Data Protection
- Sensitive information properly handled
- Transaction rollback on errors
- Proper error messages without exposing system details

## Testing

### Test Script
A test script is provided at `scripts/test-patient-creation.ps1` that demonstrates:
- How to call the new endpoint
- Expected request/response format
- Error handling scenarios

### Manual Testing Steps
1. Start the appointment service
2. Ensure database migration is applied
3. Use the test script with valid clinic and doctor UUIDs
4. Verify patient and appointment creation
5. Test error scenarios (duplicate phone, invalid data)

## Integration Notes

### Frontend Integration
- Form should collect all required patient information
- Validate phone number format
- Handle date formats properly
- Show appropriate error messages
- Display success confirmation with appointment details

### Existing API Compatibility
- All existing patient and appointment endpoints remain unchanged
- MO ID field is optional and backward compatible
- Existing appointments will have null MO ID values

## Future Enhancements

### Potential Improvements
1. **Bulk Patient Creation**: Support for creating multiple patients at once
2. **Patient Search**: Enhanced search by MO ID
3. **MO ID Validation**: Validate MO ID format against medical officer records
4. **Audit Trail**: Track who created patients and when
5. **Patient Merge**: Handle cases where duplicate patients are created

### Performance Considerations
- Database indexes are in place for MO ID lookups
- Transaction handling ensures data consistency
- Consider caching for frequently accessed patient data

