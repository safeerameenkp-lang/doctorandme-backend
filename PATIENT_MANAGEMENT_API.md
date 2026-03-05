# Patient Management API - Clinic Admin

## Overview
A dedicated API for Clinic Admins to create patients and assign them to clinics in a single operation. This separates patient creation from appointment booking, providing better control and data management.

## Key Features
- **Single API Call**: Create patient and assign to clinic in one operation
- **Complete Patient Data**: Full patient information including medical history
- **Clinic Assignment**: Automatic clinic assignment during creation
- **Validation**: Phone number and Mo ID uniqueness checks
- **Role-Based Access**: Clinic Admin only for patient creation
- **Transaction Safety**: Database transactions ensure data consistency

## API Endpoints

### 1. Create Patient and Assign to Clinic
**Endpoint**: `POST /api/patients-new`

**Access**: Clinic Admin only

**Request Body**:
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "phone": "1234567890",
  "email": "john.doe@email.com",
  "date_of_birth": "1990-01-15",
  "gender": "male",
  "mo_id": "MO123456",
  "medical_history": "Diabetes, Hypertension",
  "allergies": "Penicillin",
  "blood_group": "A+",
  "clinic_id": "clinic-uuid"
}
```

**Response**:
```json
{
  "message": "Patient created and assigned to clinic successfully",
  "patient": {
    "id": "patient-uuid",
    "user_id": "user-uuid",
    "mo_id": "MO123456",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "1234567890",
    "email": "john.doe@email.com",
    "date_of_birth": "1990-01-15",
    "gender": "male",
    "medical_history": "Diabetes, Hypertension",
    "allergies": "Penicillin",
    "blood_group": "A+",
    "is_active": true
  },
  "clinic": {
    "id": "clinic-uuid",
    "name": "City Medical Center"
  }
}
```

### 2. List Patients
**Endpoint**: `GET /api/patients-new`

**Query Parameters**:
- `clinic_id` (optional): Filter by clinic
- `only_active` (optional): Show only active patients (default: true)
- `search` (optional): Search by name, phone, or Mo ID

**Response**:
```json
{
  "patients": [
    {
      "id": "patient-uuid",
      "user_id": "user-uuid",
      "mo_id": "MO123456",
      "first_name": "John",
      "last_name": "Doe",
      "phone": "1234567890",
      "email": "john.doe@email.com",
      "date_of_birth": "1990-01-15",
      "gender": "male",
      "medical_history": "Diabetes, Hypertension",
      "allergies": "Penicillin",
      "blood_group": "A+",
      "is_active": true,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total_count": 1
}
```

### 3. Get Patient Details
**Endpoint**: `GET /api/patients-new/:id`

**Response**:
```json
{
  "id": "patient-uuid",
  "user_id": "user-uuid",
  "mo_id": "MO123456",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "1234567890",
  "email": "john.doe@email.com",
  "date_of_birth": "1990-01-15",
  "gender": "male",
  "medical_history": "Diabetes, Hypertension",
  "allergies": "Penicillin",
  "blood_group": "A+",
  "is_active": true,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### 4. Update Patient
**Endpoint**: `PUT /api/patients-new/:id`

**Access**: Clinic Admin only

**Request Body** (all fields optional):
```json
{
  "first_name": "John",
  "last_name": "Smith",
  "phone": "9876543210",
  "email": "john.smith@email.com",
  "date_of_birth": "1990-01-15",
  "gender": "male",
  "mo_id": "MO123456",
  "medical_history": "Updated medical history",
  "allergies": "Updated allergies",
  "blood_group": "B+",
  "is_active": true
}
```

**Response**:
```json
{
  "message": "Patient updated successfully"
}
```

### 5. Delete Patient
**Endpoint**: `DELETE /api/patients-new/:id`

**Access**: Clinic Admin only

**Response**:
```json
{
  "message": "Patient deleted successfully"
}
```

### 6. Assign Patient to Another Clinic
**Endpoint**: `POST /api/patients-new/:id/assign-clinic`

**Access**: Clinic Admin only

**Request Body**:
```json
{
  "clinic_id": "another-clinic-uuid"
}
```

**Response**:
```json
{
  "message": "Patient assigned to clinic successfully"
}
```

## Validation Rules

### Required Fields
- `first_name`: 2-50 characters
- `last_name`: 2-50 characters
- `phone`: 10-15 characters
- `clinic_id`: Valid UUID

### Optional Fields
- `email`: Valid email format
- `date_of_birth`: Date format
- `gender`: One of: male, female, other
- `mo_id`: 3-20 characters
- `blood_group`: One of: A+, A-, B+, B-, AB+, AB-, O+, O-
- `medical_history`: Text
- `allergies`: Text

### Uniqueness Checks
- **Phone Number**: Must be unique across all users
- **Mo ID**: Must be unique across all patients
- **Email**: Must be unique across all users (if provided)

## Error Responses

### Validation Errors
```json
{
  "error": "Invalid input data",
  "message": "Field validation failed"
}
```

### Duplicate Phone Number
```json
{
  "error": "Phone number exists",
  "message": "A user with this phone number already exists"
}
```

### Duplicate Mo ID
```json
{
  "error": "Mo ID exists",
  "message": "A patient with this Mo ID already exists"
}
```

### Clinic Not Found
```json
{
  "error": "Clinic not found",
  "message": "Clinic not found or is inactive"
}
```

### Patient Not Found
```json
{
  "error": "Patient not found",
  "message": "The specified patient does not exist"
}
```

### Cannot Delete Patient
```json
{
  "error": "Cannot delete patient",
  "message": "Patient has 5 appointments. Please handle appointments before deleting."
}
```

## Usage Examples

### 1. Create New Patient
```bash
curl -X POST http://localhost:8080/api/patients-new \
  -H "Authorization: Bearer clinic-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Jane",
    "last_name": "Smith",
    "phone": "9876543210",
    "email": "jane.smith@email.com",
    "date_of_birth": "1985-05-20",
    "gender": "female",
    "mo_id": "MO789012",
    "medical_history": "No significant medical history",
    "allergies": "None known",
    "blood_group": "O+",
    "clinic_id": "clinic-uuid"
  }'
```

### 2. Search Patients
```bash
curl -X GET "http://localhost:8080/api/patients-new?clinic_id=clinic-uuid&search=Jane" \
  -H "Authorization: Bearer clinic-admin-token"
```

### 3. Update Patient
```bash
curl -X PUT http://localhost:8080/api/patients-new/patient-uuid \
  -H "Authorization: Bearer clinic-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "medical_history": "Updated: Diabetes diagnosed",
    "allergies": "Penicillin, Shellfish"
  }'
```

### 4. Assign to Another Clinic
```bash
curl -X POST http://localhost:8080/api/patients-new/patient-uuid/assign-clinic \
  -H "Authorization: Bearer clinic-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "clinic_id": "another-clinic-uuid"
  }'
```

## Database Operations

### What Happens During Patient Creation
1. **User Account Creation**: Creates user record with basic info
2. **Role Assignment**: Assigns "patient" role to user
3. **Patient Record Creation**: Creates patient record with medical info
4. **Clinic Assignment**: Links patient to clinic as primary
5. **Transaction Safety**: All operations in a single transaction

### Data Flow
```
CreatePatient API
├── Validate input data
├── Check clinic exists
├── Check phone uniqueness
├── Check Mo ID uniqueness
├── Start database transaction
│   ├── Create user record
│   ├── Assign patient role
│   ├── Create patient record
│   └── Assign to clinic
└── Commit transaction
```

## Integration with Appointment System

### Before (Old Way)
```json
POST /api/appointments
{
  "user_id": "user-uuid",
  "clinic_id": "clinic-uuid",
  "doctor_id": "doctor-uuid",
  // ... other appointment fields
}
```
*Patient was created automatically if not found*

### After (New Way)
```json
// Step 1: Create patient
POST /api/patients-new
{
  "first_name": "John",
  "last_name": "Doe",
  "phone": "1234567890",
  "clinic_id": "clinic-uuid"
  // ... other patient fields
}

// Step 2: Create appointment
POST /api/appointments
{
  "patient_id": "patient-uuid",
  "clinic_id": "clinic-uuid",
  "doctor_id": "doctor-uuid",
  // ... other appointment fields
}
```

## Benefits

### 1. **Separation of Concerns**
- Patient creation is separate from appointment booking
- Better data management and validation
- Clearer workflow for clinic admins

### 2. **Data Integrity**
- Complete patient information captured upfront
- Proper clinic assignment
- Transaction safety ensures consistency

### 3. **Better User Experience**
- Clinic admins can pre-register patients
- Patients can be created before appointments
- Search and manage patients independently

### 4. **Flexibility**
- Patients can be assigned to multiple clinics
- Easy patient data updates
- Proper patient lifecycle management

## Security Features

### Role-Based Access
- **Patient Creation**: Clinic Admin only
- **Patient Updates**: Clinic Admin only
- **Patient Deletion**: Clinic Admin only
- **Patient Viewing**: Clinic Admin, Doctor, Receptionist

### Data Validation
- Input validation for all fields
- SQL injection prevention
- Cross-clinic data isolation
- Uniqueness constraints

### Audit Trail
- Created_at and updated_at timestamps
- Soft delete for data retention
- Transaction logging

## Performance Considerations

### Database Indexes
- `idx_users_phone` - Fast phone lookup
- `idx_patients_mo_id` - Fast Mo ID lookup
- `idx_patient_clinics_patient_id` - Fast patient-clinic lookup
- `idx_patient_clinics_clinic_id` - Fast clinic-patient lookup

### Query Optimization
- Efficient patient search
- Optimized clinic filtering
- Minimal database round trips

## Migration Notes

### Backward Compatibility
- Old `/api/patients` routes still work
- New `/api/patients-new` routes provide enhanced functionality
- Gradual migration path available

### Data Migration
- Existing patients remain unchanged
- New patients use enhanced creation process
- Clinic assignments maintained

## Testing

### Test Scenarios
1. **Patient Creation**: Valid data, invalid data, duplicates
2. **Patient Search**: By name, phone, Mo ID, clinic
3. **Patient Updates**: All fields, partial updates
4. **Clinic Assignment**: Single clinic, multiple clinics
5. **Patient Deletion**: With appointments, without appointments
6. **Error Handling**: All error scenarios

### API Testing
```bash
# Test patient creation
curl -X POST http://localhost:8080/api/patients-new \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{"first_name": "Test", "last_name": "Patient", "phone": "1234567890", "clinic_id": "clinic-uuid"}'

# Test patient search
curl -X GET "http://localhost:8080/api/patients-new?search=Test" \
  -H "Authorization: Bearer token"
```

## Conclusion

The new Patient Management API provides:
- ✅ **Dedicated patient creation** separate from appointments
- ✅ **Complete patient data** including medical history
- ✅ **Clinic assignment** in single operation
- ✅ **Role-based security** with Clinic Admin access
- ✅ **Data validation** and uniqueness checks
- ✅ **Transaction safety** for data consistency
- ✅ **Flexible patient management** with search and updates
- ✅ **Integration ready** for appointment system

This API enables clinic admins to properly manage patient data and clinic assignments before creating appointments, providing better control and data integrity.
