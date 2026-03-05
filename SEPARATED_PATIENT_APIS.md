# Separated Patient Creation APIs

## Overview
Two separate APIs for patient creation based on user roles:
1. **Super Admin**: Create patients globally (no clinic assignment)
2. **Clinic Admin**: Create patients and assign to specific clinic

## API Endpoints

### 1. Super Admin - Global Patient Creation
**Endpoint**: `POST /api/patients`

**Access**: Super Admin only

**Purpose**: Create patients globally without clinic assignment

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
  "blood_group": "A+"
}
```

**Response**:
```json
{
  "message": "Patient created successfully",
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
  }
}
```

### 2. Clinic Admin - Patient Creation with Clinic Assignment
**Endpoint**: `POST /api/clinic-patients`

**Access**: Clinic Admin only

**Purpose**: Create patients and assign to specific clinic

**Request Body**:
```json
{
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
}
```

**Response**:
```json
{
  "message": "Patient created and assigned to clinic successfully",
  "patient": {
    "id": "patient-uuid",
    "user_id": "user-uuid",
    "mo_id": "MO789012",
    "first_name": "Jane",
    "last_name": "Smith",
    "phone": "9876543210",
    "email": "jane.smith@email.com",
    "date_of_birth": "1985-05-20",
    "gender": "female",
    "medical_history": "No significant medical history",
    "allergies": "None known",
    "blood_group": "O+",
    "is_active": true
  },
  "clinic": {
    "id": "clinic-uuid",
    "name": "City Medical Center"
  }
}
```

## Key Differences

### Super Admin API (`/api/patients`)
- **Role**: Super Admin only
- **Clinic Assignment**: Not required, not performed
- **Scope**: Global patient creation
- **Use Case**: Platform-wide patient management

### Clinic Admin API (`/api/clinic-patients`)
- **Role**: Clinic Admin only
- **Clinic Assignment**: Required, automatically performed
- **Scope**: Clinic-specific patient creation
- **Use Case**: Clinic-level patient registration

## Validation Rules

### Super Admin API
- `clinic_id`: **Not required** (optional field)
- All other validations same as Clinic Admin API

### Clinic Admin API
- `clinic_id`: **Required** (must be provided)
- Clinic must exist and be active
- Patient automatically assigned to clinic as primary

## Common Validations (Both APIs)
- Phone number uniqueness across all users
- Mo ID uniqueness across all patients
- Email format validation
- Required field validation
- Blood group validation

## Usage Examples

### Super Admin Creating Global Patient
```bash
curl -X POST http://localhost:8080/api/patients \
  -H "Authorization: Bearer super-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Global",
    "last_name": "Patient",
    "phone": "1111111111",
    "email": "global@email.com",
    "date_of_birth": "1980-01-01",
    "gender": "male",
    "mo_id": "MO999999",
    "medical_history": "Global patient",
    "allergies": "None",
    "blood_group": "A+"
  }'
```

### Clinic Admin Creating Clinic Patient
```bash
curl -X POST http://localhost:8080/api/clinic-patients \
  -H "Authorization: Bearer clinic-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Clinic",
    "last_name": "Patient",
    "phone": "2222222222",
    "email": "clinic@email.com",
    "date_of_birth": "1985-01-01",
    "gender": "female",
    "mo_id": "MO888888",
    "medical_history": "Clinic patient",
    "allergies": "None",
    "blood_group": "B+",
    "clinic_id": "clinic-uuid"
  }'
```

## Error Responses

### Missing Clinic ID (Clinic Admin API)
```json
{
  "error": "Missing clinic_id",
  "message": "clinic_id is required for clinic admin patient creation"
}
```

### Clinic Not Found (Clinic Admin API)
```json
{
  "error": "Clinic not found",
  "message": "Clinic not found or is inactive"
}
```

### Phone Number Exists (Both APIs)
```json
{
  "error": "Phone number exists",
  "message": "A user with this phone number already exists"
}
```

### Mo ID Exists (Both APIs)
```json
{
  "error": "Mo ID exists",
  "message": "A patient with this Mo ID already exists"
}
```

## Database Operations

### Super Admin API
1. Create user account
2. Assign patient role
3. Create patient record
4. **No clinic assignment**

### Clinic Admin API
1. Create user account
2. Assign patient role
3. Create patient record
4. **Assign to clinic as primary**

## Workflow Comparison

### Super Admin Workflow
```
Super Admin
├── Create patient globally
├── Patient exists without clinic assignment
├── Later assign to clinic(s) if needed
└── Use for platform-wide patient management
```

### Clinic Admin Workflow
```
Clinic Admin
├── Create patient for specific clinic
├── Patient automatically assigned to clinic
├── Ready for appointments immediately
└── Use for clinic-specific patient registration
```

## Integration with Appointment System

### Super Admin Created Patients
```json
POST /api/appointments
{
  "patient_id": "global-patient-uuid",
  "clinic_id": "clinic-uuid",
  "doctor_id": "doctor-uuid"
  // Patient exists globally, can book at any clinic
}
```

### Clinic Admin Created Patients
```json
POST /api/appointments
{
  "patient_id": "clinic-patient-uuid",
  "clinic_id": "clinic-uuid",
  "doctor_id": "doctor-uuid"
  // Patient already assigned to clinic
}
```

## Benefits of Separation

### 1. **Role-Based Access Control**
- Super Admin: Global patient management
- Clinic Admin: Clinic-specific patient management
- Clear separation of responsibilities

### 2. **Data Management**
- Super Admin: Platform-wide patient data
- Clinic Admin: Clinic-focused patient data
- Better data organization

### 3. **Workflow Clarity**
- Super Admin: Create patients for platform
- Clinic Admin: Create patients for clinic
- Clear use cases for each API

### 4. **Security**
- Role-based access prevents unauthorized access
- Clinic admins can only create patients for their clinic
- Super admins have global access

## Additional Endpoints

### Both APIs Support
- `GET /api/patients` - List patients (Super Admin)
- `GET /api/clinic-patients` - List patients (Clinic Admin)
- `GET /api/patients/:id` - Get patient details
- `GET /api/clinic-patients/:id` - Get patient details
- `PUT /api/patients/:id` - Update patient (Super Admin)
- `PUT /api/clinic-patients/:id` - Update patient (Clinic Admin)
- `DELETE /api/patients/:id` - Delete patient (Super Admin)
- `DELETE /api/clinic-patients/:id` - Delete patient (Clinic Admin)

### Clinic Admin Additional
- `POST /api/clinic-patients/:id/assign-clinic` - Assign to another clinic

## Testing

### Super Admin API Test
```bash
# Test global patient creation
curl -X POST http://localhost:8080/api/patients \
  -H "Authorization: Bearer super-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Test",
    "last_name": "Global",
    "phone": "3333333333",
    "clinic_id": "clinic-uuid"
  }'
```

### Clinic Admin API Test
```bash
# Test clinic patient creation
curl -X POST http://localhost:8080/api/clinic-patients \
  -H "Authorization: Bearer clinic-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Test",
    "last_name": "Clinic",
    "phone": "4444444444",
    "clinic_id": "clinic-uuid"
  }'
```

## Conclusion

The separated patient creation APIs provide:
- ✅ **Clear role separation** between Super Admin and Clinic Admin
- ✅ **Global vs clinic-specific** patient management
- ✅ **Automatic clinic assignment** for clinic admin API
- ✅ **Flexible patient creation** based on user role
- ✅ **Proper access control** with role-based security
- ✅ **Workflow clarity** for different use cases
- ✅ **Data organization** by scope and responsibility

This separation ensures that Super Admins can manage patients globally while Clinic Admins can create patients specifically for their clinics with automatic assignment.
