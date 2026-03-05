# Updated Appointment System - UI Integration

## Overview
The appointment system has been updated to match the UI requirements shown in the image, including department management, enhanced patient search, time slot validation, and clinic-specific doctor fees.

## New Features

### 1. Department Management
- **Create Department**: `POST /api/departments`
- **List Departments**: `GET /api/departments?clinic_id=xxx`
- **Get Department**: `GET /api/departments/:id`
- **Update Department**: `PUT /api/departments/:id`
- **Delete Department**: `DELETE /api/departments/:id`
- **Get Doctors by Department**: `GET /api/departments/:department_id/doctors`

### 2. Enhanced Patient Search
The `CreateAppointment` API now supports multiple patient identification methods:
- `user_id` - Direct user ID
- `patient_id` - Direct patient ID
- `mobile_no` - Search by mobile number
- `mo_id` - Search by Mo ID

### 3. Time Slot Management
- **Get Available Time Slots**: `GET /api/appointments/slots/available?doctor_id=xxx&clinic_id=xxx&date=yyyy-mm-dd&slot_type=offline|online|both`
- Validates doctor availability, leave status, and existing appointments
- Returns slot availability with booking counts

### 4. Updated Appointment Creation
Enhanced `CreateAppointment` API with:
- Department validation
- Clinic-specific doctor fee calculation
- Time slot validation
- Leave status checking
- Appointment conflict prevention

## API Endpoints

### Department Management
```bash
# Create Department
POST /api/departments
{
  "clinic_id": "uuid",
  "name": "Orthology",
  "description": "Orthopedic treatments"
}

# List Departments
GET /api/departments?clinic_id=uuid&only_active=true

# Get Department Details
GET /api/departments/:id

# Update Department
PUT /api/departments/:id
{
  "name": "Updated Name",
  "description": "Updated description",
  "is_active": true
}

# Delete Department
DELETE /api/departments/:id

# Get Doctors in Department
GET /api/departments/:department_id/doctors?only_active=true
```

### Enhanced Appointment Creation
```bash
# Create Appointment (Multiple Patient Search Methods)
POST /api/appointments
{
  "clinic_id": "uuid",
  "doctor_id": "uuid",
  "department_id": "uuid",  // Optional
  "appointment_date": "2024-01-15",
  "appointment_time": "2024-01-15 09:00:00",
  "consultation_type": "video|in_person|offline|online",
  "duration_minutes": 30,
  "reason": "Regular checkup",
  "notes": "Patient notes",
  "payment_mode": "pay_later|pay_now|way_off|cash|card|upi",
  
  // Patient identification (choose one)
  "user_id": "uuid",        // Direct user ID
  "patient_id": "uuid",     // Direct patient ID
  "mobile_no": "1234567890", // Search by mobile
  "mo_id": "MO123456"       // Search by Mo ID
}
```

### Time Slot Availability
```bash
# Get Available Time Slots
GET /api/appointments/slots/available?doctor_id=uuid&clinic_id=uuid&date=2024-01-15&slot_type=offline

# Response
{
  "date": "2024-01-15",
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "slot_type": "offline",
  "time_slots": [
    {
      "id": "slot-uuid",
      "slot_type": "offline",
      "start_time": "09:00:00",
      "end_time": "09:30:00",
      "max_patients": 1,
      "notes": "",
      "available": true,
      "booked_count": 0
    }
  ],
  "total_count": 8
}
```

## Database Changes

### New Tables
1. **departments** - Clinic-specific departments
2. **doctor_time_slots** - Doctor availability slots per clinic

### Updated Tables
1. **doctors** - Added `department_id` field
2. **clinic_doctor_links** - Added clinic-specific fees

## UI Integration Features

### 1. Consultation Type Support
- **Video/Online**: Uses `consultation_fee_online` from clinic-doctor link
- **In-person/Offline**: Uses `consultation_fee_offline` from clinic-doctor link

### 2. Department Selection
- Departments are clinic-specific
- Doctors can be assigned to departments
- Department validation in appointment creation

### 3. Patient Search
- Search by mobile number or Mo ID
- Automatic patient creation if not found
- Clear error messages for not found patients

### 4. Time Slot Display
- Separate online/offline slots
- Real-time availability checking
- Leave status integration
- Booking count tracking

### 5. Payment Methods
- Pay Later, Pay Now, Way Off options
- Traditional payment modes (cash, card, upi)

## Validation Rules

### Appointment Creation
1. **Patient Identification**: Exactly one method required
2. **Department**: Must exist and be active in the clinic
3. **Doctor**: Must be linked to the clinic and active
4. **Time Slot**: Must exist and be available
5. **Leave Status**: Doctor cannot be on approved leave
6. **Conflicts**: No overlapping appointments across clinics

### Time Slot Validation
1. **Day of Week**: Matches doctor's schedule
2. **Slot Type**: Matches consultation type
3. **Availability**: Within doctor's time slots
4. **Leave Status**: Doctor not on leave
5. **Booking Limit**: Within max_patients limit

## Error Handling

### Patient Search Errors
```json
{
  "error": "Patient not found",
  "message": "No patient found with this mobile number. Please add new patient."
}
```

### Time Slot Errors
```json
{
  "error": "Time slot not available",
  "message": "Doctor does not have an offline time slot at this time on Monday"
}
```

### Department Errors
```json
{
  "error": "Department mismatch",
  "message": "Doctor is not assigned to the selected department"
}
```

## Usage Examples

### 1. Create Department
```bash
curl -X POST http://localhost:8080/api/departments \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{
    "clinic_id": "clinic-uuid",
    "name": "Cardiology",
    "description": "Heart and cardiovascular treatments"
  }'
```

### 2. Search Patient by Mobile
```bash
curl -X POST http://localhost:8080/api/appointments \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{
    "clinic_id": "clinic-uuid",
    "doctor_id": "doctor-uuid",
    "department_id": "department-uuid",
    "appointment_date": "2024-01-15",
    "appointment_time": "2024-01-15 09:00:00",
    "consultation_type": "video",
    "mobile_no": "1234567890",
    "reason": "Regular checkup",
    "payment_mode": "pay_now"
  }'
```

### 3. Get Available Slots
```bash
curl -X GET "http://localhost:8080/api/appointments/slots/available?doctor_id=doctor-uuid&clinic_id=clinic-uuid&date=2024-01-15&slot_type=offline" \
  -H "Authorization: Bearer token"
```

## Security Features

### Role-Based Access
- **Department Management**: Clinic Admin only
- **Appointment Creation**: Clinic Admin, Receptionist
- **Time Slot Viewing**: Clinic Admin, Receptionist, Doctor

### Data Validation
- Input validation for all fields
- SQL injection prevention
- Cross-clinic data isolation
- Doctor-clinic relationship validation

## Performance Optimizations

### Database Indexes
- `idx_departments_clinic_id` - Fast clinic department lookup
- `idx_doctor_time_slots_doctor_id` - Fast doctor slot lookup
- `idx_doctor_time_slots_clinic_id` - Fast clinic slot lookup
- `idx_doctor_time_slots_day_of_week` - Fast day-based lookup

### Query Optimization
- Efficient time slot availability checking
- Optimized patient search queries
- Minimal database round trips

## Migration Status

### Completed Migrations
1. **005_user_management_features.sql** - User management
2. **006_doctor_leave_management.sql** - Leave management
3. **007_clinic_specific_doctor_fees.sql** - Clinic-specific fees
4. **008_doctor_time_slots.sql** - Time slot management
5. **009_departments.sql** - Department management

### Database Schema
- All new tables created
- Indexes optimized
- Triggers for updated_at timestamps
- Constraints for data integrity

## Testing

### Test Scenarios
1. **Department Creation**: Create, list, update, delete departments
2. **Patient Search**: Search by mobile, Mo ID, user ID, patient ID
3. **Time Slot Validation**: Check availability, leave status, conflicts
4. **Appointment Creation**: Full workflow with all validations
5. **Error Handling**: Test all error scenarios

### API Testing
```bash
# Test department creation
curl -X POST http://localhost:8080/api/departments \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{"clinic_id": "clinic-uuid", "name": "Test Dept"}'

# Test appointment creation
curl -X POST http://localhost:8080/api/appointments \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{"clinic_id": "clinic-uuid", "doctor_id": "doctor-uuid", "appointment_date": "2024-01-15", "appointment_time": "2024-01-15 09:00:00", "consultation_type": "video", "mobile_no": "1234567890"}'
```

## Next Steps

### Potential Enhancements
1. **Bulk Appointment Creation** - Multiple appointments at once
2. **Appointment Templates** - Predefined appointment types
3. **Waitlist Management** - Queue for popular time slots
4. **Notification System** - SMS/Email reminders
5. **Analytics Dashboard** - Appointment statistics and trends

### Integration Points
1. **Payment Gateway** - Online payment processing
2. **SMS Service** - Appointment reminders
3. **Email Service** - Confirmation emails
4. **Calendar Integration** - Sync with external calendars
5. **Reporting System** - Advanced analytics and reports

## Conclusion

The updated appointment system now fully supports the UI requirements with:
- ✅ Department management
- ✅ Enhanced patient search
- ✅ Time slot validation
- ✅ Clinic-specific fees
- ✅ Leave management integration
- ✅ Comprehensive error handling
- ✅ Role-based security
- ✅ Performance optimization

The system is ready for production use and can handle the complete appointment booking workflow as shown in the UI.
