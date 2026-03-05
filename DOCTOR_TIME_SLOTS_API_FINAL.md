# Doctor Time Slots API - Complete Implementation

## Overview
A complete RESTful API for managing doctor time slots in the clinic system. This implementation includes all CRUD operations with proper authentication, validation, and real-time availability tracking.

---

## Implementation Files

### 1. **Model** (`services/organization-service/models/organization.model.go`)
```go
type DoctorTimeSlot struct {
    ID          string     `json:"id" db:"id"`
    DoctorID    string     `json:"doctor_id" db:"doctor_id"`
    ClinicID    string     `json:"clinic_id" db:"clinic_id"`
    Date        string     `json:"date" db:"specific_date"`
    SlotType    string     `json:"slot_type" db:"slot_type"`
    StartTime   string     `json:"start_time" db:"start_time"`
    EndTime     string     `json:"end_time" db:"end_time"`
    MaxPatients int        `json:"max_patients" db:"max_patients"`
    Notes       *string    `json:"notes,omitempty" db:"notes"`
    IsActive    bool       `json:"is_active" db:"is_active"`
    CreatedAt   time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
```

### 2. **Controller** (`services/organization-service/controllers/doctor_time_slots.controller.go`)
Implements 5 complete endpoints with proper error handling and validation.

### 3. **Routes** (`services/organization-service/routes/organization.routes.go`)
All routes are authenticated and role-based access controlled.

---

## API Endpoints

### Base URL
```
http://localhost:8081/api/doctor-time-slots
```

### Authentication
All endpoints require Bearer token authentication:
```
Authorization: Bearer <your-token>
```

---

## 1. Create Doctor Time Slots (Bulk)

**Endpoint:** `POST /doctor-time-slots`

**Access:** Doctor, Clinic Admin

**Description:** Create multiple time slots for a doctor at a clinic for a specific date.

**Request Body:**
```json
{
  "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "slot_type": "offline",
  "date": "2024-10-15",
  "slots": [
    {
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10,
      "notes": "Morning shift - Monday"
    },
    {
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 10,
      "notes": "Afternoon shift - Monday"
    }
  ]
}
```

**Response (201 Created):**
```json
{
  "message": "Slot creation completed. 2 created, 0 failed",
  "created_slots": [
    {
      "id": "slot-uuid-1",
      "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "date": "2024-10-15",
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10,
      "booked_patients": 0,
      "available_spots": 10,
      "is_available": true,
      "status": "available",
      "notes": "Morning shift - Monday",
      "is_active": true,
      "created_at": "2024-10-14T10:00:00Z",
      "updated_at": "2024-10-14T10:00:00Z"
    }
  ],
  "failed_slots": [],
  "total_created": 2,
  "total_failed": 0
}
```

**Validations:**
- Doctor must exist and be active
- Clinic must exist and be active
- Doctor must be linked to the clinic
- `slot_type` must be "offline" or "online"
- `date` must be in YYYY-MM-DD format
- `start_time` and `end_time` must be in HH:MM format
- `max_patients` defaults to 1 if not provided

---

## 2. List Doctor Time Slots

**Endpoint:** `GET /doctor-time-slots`

**Access:** All authenticated users

**Description:** Retrieve time slots for a doctor with optional filtering.

**Query Parameters:**
| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| doctor_id | Yes | UUID | Doctor's unique identifier |
| clinic_id | No | UUID | Filter by clinic |
| slot_type | No | String | Filter by type: "offline" or "online" |
| date | No | String | Filter by date (YYYY-MM-DD) |

**Example Request:**
```
GET /doctor-time-slots?doctor_id=3fd28e6d-7f9a-4dde-8172-d14a74a9b02d&clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&slot_type=offline&date=2024-10-15
```

**Response (200 OK):**
```json
{
  "slots": [
    {
      "id": "slot-uuid-1",
      "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "date": "2024-10-15",
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10,
      "booked_patients": 3,
      "available_spots": 7,
      "is_available": true,
      "status": "available",
      "notes": "Morning shift - Monday",
      "is_active": true,
      "created_at": "2024-10-14T10:00:00Z",
      "updated_at": "2024-10-14T10:00:00Z"
    }
  ],
  "total": 1,
  "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "slot_type": "offline",
  "date": "2024-10-15"
}
```

**Features:**
- Real-time availability calculation
- Automatic status determination (available/booking_full)
- Counts confirmed and completed appointments
- Only returns active slots (`is_active = true`)

---

## 3. Get Single Time Slot

**Endpoint:** `GET /doctor-time-slots/:id`

**Access:** All authenticated users

**Description:** Retrieve a specific time slot by ID with availability information.

**Example Request:**
```
GET /doctor-time-slots/slot-uuid-1
```

**Response (200 OK):**
```json
{
  "slot": {
    "id": "slot-uuid-1",
    "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "date": "2024-10-15",
    "slot_type": "offline",
    "start_time": "09:00",
    "end_time": "12:00",
    "max_patients": 10,
    "booked_patients": 3,
    "available_spots": 7,
    "is_available": true,
    "status": "available",
    "notes": "Morning shift - Monday",
    "is_active": true,
    "created_at": "2024-10-14T10:00:00Z",
    "updated_at": "2024-10-14T10:00:00Z"
  }
}
```

---

## 4. Update Time Slot

**Endpoint:** `PUT /doctor-time-slots/:id`

**Access:** Doctor, Clinic Admin

**Description:** Update an existing time slot. All fields are optional.

**Request Body:**
```json
{
  "slot_type": "online",
  "start_time": "10:00",
  "end_time": "13:00",
  "max_patients": 15,
  "notes": "Updated morning shift",
  "is_active": true
}
```

**Response (200 OK):**
```json
{
  "message": "Time slot updated successfully",
  "slot": {
    "id": "slot-uuid-1",
    "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "date": "2024-10-15",
    "slot_type": "online",
    "start_time": "10:00",
    "end_time": "13:00",
    "max_patients": 15,
    "booked_patients": 3,
    "available_spots": 12,
    "is_available": true,
    "status": "available",
    "notes": "Updated morning shift",
    "is_active": true,
    "created_at": "2024-10-14T10:00:00Z",
    "updated_at": "2024-10-14T11:30:00Z"
  }
}
```

**Updatable Fields:**
- `slot_type` - "offline" or "online"
- `start_time` - HH:MM format
- `end_time` - HH:MM format
- `max_patients` - Integer
- `notes` - String
- `is_active` - Boolean

---

## 5. Delete Time Slot (Soft Delete)

**Endpoint:** `DELETE /doctor-time-slots/:id`

**Access:** Doctor, Clinic Admin

**Description:** Soft delete a time slot by setting `is_active` to false.

**Example Request:**
```
DELETE /doctor-time-slots/slot-uuid-1
```

**Response (200 OK):**
```json
{
  "message": "Time slot deleted successfully",
  "slot_id": "slot-uuid-1"
}
```

**Note:** Soft delete is used to maintain data integrity. Deleted slots are not permanently removed from the database.

---

## Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 206 | Partial Content (some slots created, some failed) |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Invalid or missing token |
| 403 | Forbidden - Insufficient permissions or doctor not linked to clinic |
| 404 | Not Found - Resource doesn't exist |
| 500 | Internal Server Error |

---

## Validation Rules

### Slot Type
- Must be "offline" or "online"
- Case-sensitive

### Date Format
- Must be YYYY-MM-DD
- Example: 2024-10-15

### Time Format
- Must be HH:MM (24-hour)
- Valid range: 00:00 to 23:59
- Example: 09:00, 14:30

### Max Patients
- Must be positive integer
- Minimum: 1
- Default: 1 if not provided

### UUID Format
- Must be valid UUID v4
- Example: 3fd28e6d-7f9a-4dde-8172-d14a74a9b02d

---

## Key Features

### 1. **Real-time Availability**
- Automatically calculates `booked_patients` from appointments table
- Computes `available_spots` = `max_patients` - `booked_patients`
- Sets `is_available` and `status` based on availability

### 2. **Automatic Status Determination**
- `"available"` - Has available spots
- `"booking_full"` - No available spots

### 3. **Soft Delete**
- Slots are marked as inactive rather than permanently deleted
- Preserves historical data and appointment references

### 4. **Bulk Creation**
- Create multiple slots for the same date in one request
- Partial success handling (some slots may fail individually)

### 5. **Comprehensive Validation**
- Doctor existence and active status
- Clinic existence and active status
- Doctor-clinic linkage verification
- Time and date format validation
- Slot type validation

### 6. **Security**
- JWT authentication required for all endpoints
- Role-based access control (RBAC)
- Doctor and Clinic Admin can create/update/delete
- All authenticated users can view

---

## Database Schema Reference

The API uses the `doctor_time_slots` table with the following structure:

```sql
CREATE TABLE doctor_time_slots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    doctor_id UUID NOT NULL REFERENCES doctors(id),
    clinic_id UUID NOT NULL REFERENCES clinics(id),
    specific_date DATE NOT NULL,
    slot_type VARCHAR(20) NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    max_patients INTEGER NOT NULL DEFAULT 1,
    notes TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## Integration Example (Flutter/Dart)

### Create Slots
```dart
final response = await http.post(
  Uri.parse('$baseUrl/doctor-time-slots'),
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer $token',
  },
  body: jsonEncode({
    'doctor_id': doctorId,
    'clinic_id': clinicId,
    'slot_type': 'offline',
    'date': '2024-10-15',
    'slots': [
      {
        'start_time': '09:00',
        'end_time': '12:00',
        'max_patients': 10,
        'notes': 'Morning shift',
      },
    ],
  }),
);
```

### List Slots
```dart
final response = await http.get(
  Uri.parse('$baseUrl/doctor-time-slots?doctor_id=$doctorId&date=2024-10-15'),
  headers: {'Authorization': 'Bearer $token'},
);
```

### Update Slot
```dart
final response = await http.put(
  Uri.parse('$baseUrl/doctor-time-slots/$slotId'),
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer $token',
  },
  body: jsonEncode({
    'max_patients': 15,
    'notes': 'Updated shift',
  }),
);
```

### Delete Slot
```dart
final response = await http.delete(
  Uri.parse('$baseUrl/doctor-time-slots/$slotId'),
  headers: {'Authorization': 'Bearer $token'},
);
```

---

## Error Handling

### Common Error Responses

**Doctor Not Found (404):**
```json
{
  "error": "Doctor not found",
  "message": "Doctor not found or is inactive"
}
```

**Doctor Not Linked to Clinic (403):**
```json
{
  "error": "Doctor is not linked to this clinic",
  "message": "The specified doctor is not associated with this clinic"
}
```

**Invalid Input (400):**
```json
{
  "error": "Invalid input data",
  "message": "Field validation for 'doctor_id' failed on the 'uuid' tag"
}
```

**Invalid Date Format (400):**
```json
{
  "error": "Invalid date format. Use YYYY-MM-DD format"
}
```

---

## Testing

A comprehensive test script is provided: `test-time-slots.ps1`

Run the test:
```powershell
powershell -ExecutionPolicy Bypass -File test-time-slots.ps1
```

The test script covers:
1. Authentication
2. Bulk slot creation
3. Listing with filters
4. Single slot retrieval
5. Slot update
6. Soft delete

---

## Notes

1. All timestamps are in UTC format (ISO 8601)
2. Soft delete is used - slots are marked as inactive rather than permanently deleted
3. `booked_patients` and `available_spots` are calculated in real-time
4. The `status` field is automatically determined based on availability
5. Multiple slots can be created for the same date with different times
6. The API only returns active slots by default in list operations

---

## Implementation Complete ✓

All 5 endpoints have been implemented with:
- ✓ Complete model definition
- ✓ Full controller implementation
- ✓ Route registration with RBAC
- ✓ Comprehensive validation
- ✓ Error handling
- ✓ Real-time availability tracking
- ✓ Test script provided

The API is production-ready and follows RESTful best practices.

