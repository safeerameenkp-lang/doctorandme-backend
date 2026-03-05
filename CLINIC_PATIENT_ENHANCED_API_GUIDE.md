# Enhanced Clinic Patient API Guide

## Overview
The clinic patient API has been enhanced to include all fields from the patient form. This includes personal information, address details, lifestyle information, and physical measurements.

**New Feature**: Automatic MO ID Generation
- When creating a patient, if `mo_id` is not provided, the system automatically generates a unique MO ID
- Format: `{clinic_code}{sequential_number}` (e.g., MO0001, MO0002, ABC0001, ABC0002)
- Each clinic maintains its own sequence
- See [Auto MO ID Guide](./CLINIC_PATIENT_AUTO_MO_ID_GUIDE.md) for details

## New Fields Added

### Personal Information
- `age` (integer): Patient age in years (0-150)
- `address1` (string): Primary address line (max 200 chars)
- `address2` (string): Secondary address line (max 200 chars)
- `district` (string): Patient district (max 100 chars)
- `state` (string): Patient state (max 100 chars)

### Lifestyle Information
- `smoking_status` (string): Smoking status (Yes/No, max 20 chars)
- `alcohol_use` (string): Alcohol use status (Yes/No, max 20 chars)

### Physical Measurements
- `height_cm` (integer): Patient height in centimeters (0-300)
- `weight_kg` (integer): Patient weight in kilograms (0-500)

## API Endpoints

### 1. Create Clinic Patient
**POST** `/clinic-specific-patients`

#### Request Body

**Option 1: Auto-Generate MO ID** (Recommended)
```json
{
  "clinic_id": "uuid-required",
  "first_name": "Safeer",
  "last_name": "Ameen",
  "phone": "+91 476738737",
  "email": "safeer@example.com",
  "date_of_birth": "1988-01-01",
  "age": 36,
  "gender": "Male",
  "address1": "123 Main Street",
  "address2": "Apt 4B",
  "district": "Malappuram",
  "state": "Kerala",
  "medical_history": "No significant history",
  "allergies": "None known",
  "blood_group": "O+",
  "smoking_status": "No",
  "alcohol_use": "No",
  "height_cm": 165,
  "weight_kg": 69
  // mo_id is NOT provided - will be auto-generated (e.g., MO0001)
}
```

**Option 2: Custom MO ID**
```json
{
  "clinic_id": "uuid-required",
  "first_name": "Safeer",
  "last_name": "Ameen",
  "phone": "+91 476738737",
  "email": "safeer@example.com",
  "date_of_birth": "1988-01-01",
  "age": 36,
  "gender": "Male",
  "address1": "123 Main Street",
  "address2": "Apt 4B",
  "district": "Malappuram",
  "state": "Kerala",
  "mo_id": "CUSTOM001",  // Custom MO ID provided
  "medical_history": "No significant history",
  "allergies": "None known",
  "blood_group": "O+",
  "smoking_status": "No",
  "alcohol_use": "No",
  "height_cm": 165,
  "weight_kg": 69
}
```

#### Response
```json
{
  "message": "Patient created successfully for this clinic",
  "patient": {
    "id": "patient-uuid",
    "clinic_id": "clinic-uuid",
    "first_name": "Safeer",
    "last_name": "Ameen",
    "phone": "+91 476738737",
    "email": "safeer@example.com",
    "date_of_birth": "1988-01-01",
    "age": 36,
    "gender": "Male",
    "address1": "123 Main Street",
    "address2": "Apt 4B",
    "district": "Malappuram",
    "state": "Kerala",
    "mo_id": "MO0001",  // ✅ Auto-generated based on clinic_code
    "medical_history": "No significant history",
    "allergies": "None known",
    "blood_group": "O+",
    "smoking_status": "No",
    "alcohol_use": "No",
    "height_cm": 165,
    "weight_kg": 69,
    "is_active": true,
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z"
  }
}
```

### 2. List Clinic Patients
**GET** `/clinic-specific-patients?clinic_id=uuid&search=term&only_active=true`

#### Query Parameters
- `clinic_id` (required): Clinic UUID
- `search` (optional): Search term (searches name, phone, mo_id, address, district, state)
- `only_active` (optional): Filter by active status (default: true)

#### Response
```json
{
  "clinic_id": "clinic-uuid",
  "total": 1,
  "patients": [
    {
      "id": "patient-uuid",
      "clinic_id": "clinic-uuid",
      "first_name": "Safeer",
      "last_name": "Ameen",
      "phone": "+91 476738737",
      "email": "safeer@example.com",
      "date_of_birth": "1988-01-01",
      "age": 36,
      "gender": "Male",
      "address1": "123 Main Street",
      "address2": "Apt 4B",
      "district": "Malappuram",
      "state": "Kerala",
      "mo_id": "CLINIC001",
      "medical_history": "No significant history",
      "allergies": "None known",
      "blood_group": "O+",
      "smoking_status": "No",
      "alcohol_use": "No",
      "height_cm": 165,
      "weight_kg": 69,
      "is_active": true,
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T10:00:00Z"
    }
  ]
}
```

### 3. Get Single Clinic Patient
**GET** `/clinic-specific-patients/:id`

#### Response
```json
{
  "patient": {
    "id": "patient-uuid",
    "clinic_id": "clinic-uuid",
    "first_name": "Safeer",
    "last_name": "Ameen",
    "phone": "+91 476738737",
    "email": "safeer@example.com",
    "date_of_birth": "1988-01-01",
    "age": 36,
    "gender": "Male",
    "address1": "123 Main Street",
    "address2": "Apt 4B",
    "district": "Malappuram",
    "state": "Kerala",
    "mo_id": "CLINIC001",
    "medical_history": "No significant history",
    "allergies": "None known",
    "blood_group": "O+",
    "smoking_status": "No",
    "alcohol_use": "No",
    "height_cm": 165,
    "weight_kg": 69,
    "is_active": true,
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z"
  }
}
```

### 4. Update Clinic Patient
**PUT** `/clinic-specific-patients/:id`

#### Request Body (All fields optional)
```json
{
  "first_name": "Updated Name",
  "age": 37,
  "address1": "456 New Street",
  "district": "Updated District",
  "smoking_status": "Yes",
  "height_cm": 166,
  "weight_kg": 70
}
```

#### Response
```json
{
  "message": "Patient updated successfully",
  "patient": {
    "id": "patient-uuid",
    "clinic_id": "clinic-uuid",
    "first_name": "Updated Name",
    "last_name": "Ameen",
    "phone": "+91 476738737",
    "email": "safeer@example.com",
    "date_of_birth": "1988-01-01",
    "age": 37,
    "gender": "Male",
    "address1": "456 New Street",
    "address2": "Apt 4B",
    "district": "Updated District",
    "state": "Kerala",
    "mo_id": "CLINIC001",
    "medical_history": "No significant history",
    "allergies": "None known",
    "blood_group": "O+",
    "smoking_status": "Yes",
    "alcohol_use": "No",
    "height_cm": 166,
    "weight_kg": 70,
    "is_active": true,
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T11:00:00Z"
  }
}
```

### 5. Delete Clinic Patient (Soft Delete)
**DELETE** `/clinic-specific-patients/:id`

#### Response
```json
{
  "message": "Patient deleted successfully"
}
```

## Field Validation Rules

### Required Fields (Create Only)
- `clinic_id`: Must be valid UUID
- `first_name`: Max 100 characters
- `last_name`: Max 100 characters
- `phone`: Max 20 characters

### Optional Fields with Validation
- `email`: Valid email format
- `age`: Integer between 0-150
- `gender`: Max 20 characters
- `address1`: Max 200 characters
- `address2`: Max 200 characters
- `district`: Max 100 characters
- `state`: Max 100 characters
- `mo_id`: Max 50 characters (auto-generated if not provided)
- `blood_group`: Max 10 characters
- `smoking_status`: Max 20 characters
- `alcohol_use`: Max 20 characters
- `height_cm`: Integer between 0-300
- `weight_kg`: Integer between 0-500

## Database Migration

Run the migration to add the new fields:
```sql
-- File: migrations/019_add_missing_patient_fields.sql
```

## Search Functionality

The search parameter in the list endpoint now searches across:
- First name
- Last name
- Phone number
- MO ID
- Address 1
- District
- State

## Error Responses

### Validation Errors
```json
{
  "error": "Invalid input data",
  "message": "Field validation failed: age must be between 0 and 150"
}
```

### Not Found Errors
```json
{
  "error": "Patient not found"
}
```

### Conflict Errors
```json
{
  "error": "Phone number exists in this clinic",
  "message": "A patient with this phone number already exists in your clinic"
}
```

## Usage Examples

### Create Patient with All Fields
```bash
curl -X POST http://localhost:8080/clinic-specific-patients \
  -H "Content-Type: application/json" \
  -d '{
    "clinic_id": "clinic-uuid",
    "first_name": "Safeer",
    "last_name": "Ameen",
    "phone": "+91 476738737",
    "email": "safeer@example.com",
    "age": 36,
    "gender": "Male",
    "address1": "123 Main Street",
    "address2": "Apt 4B",
    "district": "Malappuram",
    "state": "Kerala",
    "blood_group": "O+",
    "smoking_status": "No",
    "alcohol_use": "No",
    "height_cm": 165,
    "weight_kg": 69
  }'
```

### Update Patient Address and Measurements
```bash
curl -X PUT http://localhost:8080/clinic-specific-patients/patient-uuid \
  -H "Content-Type: application/json" \
  -d '{
    "address1": "456 New Street",
    "district": "Updated District",
    "height_cm": 166,
    "weight_kg": 70
  }'
```

### Search Patients by District
```bash
curl "http://localhost:8080/clinic-specific-patients?clinic_id=clinic-uuid&search=Malappuram"
```

## Notes

1. All new fields are optional and can be updated independently
2. The API maintains backward compatibility with existing clients
3. Search functionality has been enhanced to include address fields
4. Age and physical measurements have reasonable validation ranges
5. The system supports both date_of_birth and age fields for flexibility
6. **MO ID Auto-Generation**: If `mo_id` is not provided during patient creation, the system automatically generates a unique MO ID using the format `{clinic_code}{sequential_number}` (e.g., MO0001, MO0002). See [Auto MO ID Guide](./CLINIC_PATIENT_AUTO_MO_ID_GUIDE.md) for details.
