# Clinic Doctors List API - Complete Guide

## Overview

API to get all doctors linked to a specific clinic with role-based access control.

---

## API Endpoint

### Get Doctors by Clinic

**Endpoint:** `GET /api/v1/org/doctors/clinic/:clinic_id`

**Method:** GET

**Authentication:** Required (JWT Token)

**Who Can Use:**
- ✅ Super Admin (any clinic)
- ✅ Organization Admin (clinics in their org)
- ✅ Clinic Admin (their clinic)
- ✅ Receptionist (their clinic)
- ✅ Doctors (their clinic)
- ✅ Any staff with clinic access

---

## Request

### URL Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `clinic_id` | UUID | Yes | The clinic ID to get doctors for |

### Headers

```
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json
```

### Example Request

```bash
GET /api/v1/org/doctors/clinic/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## Response

### Success Response (200 OK)

```json
{
  "doctors": [
    {
      "id": "doctor-uuid-1",
      "user_id": "user-uuid-1",
      "first_name": "John",
      "last_name": "Smith",
      "doctor_code": "DOC001",
      "specialization": "Cardiology",
      "license_number": "MED-12345",
      "consultation_fee": 150.00,
      "follow_up_fee": 75.00,
      "follow_up_days": 7,
      "is_main_doctor": true,
      "is_active": true,
      "email": "john.smith@clinic.com",
      "phone": "+1234567890",
      "created_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": "doctor-uuid-2",
      "user_id": "user-uuid-2",
      "first_name": "Jane",
      "last_name": "Doe",
      "doctor_code": "DOC002",
      "specialization": "Pediatrics",
      "license_number": "MED-67890",
      "consultation_fee": 120.00,
      "follow_up_fee": 60.00,
      "follow_up_days": 7,
      "is_main_doctor": false,
      "is_active": true,
      "email": "jane.doe@clinic.com",
      "phone": "+0987654321",
      "created_at": "2024-02-15T00:00:00Z"
    }
  ],
  "total_count": 2,
  "clinic_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Error Responses

**403 Forbidden - No Access to Clinic**
```json
{
  "error": "Access denied",
  "message": "You don't have access to this clinic"
}
```

**404 Not Found - Clinic Not Found**
```json
{
  "error": "Resource not found",
  "message": "The requested Clinic was not found",
  "code": "RESOURCE_NOT_FOUND"
}
```

**401 Unauthorized - Not Authenticated**
```json
{
  "error": "Invalid or expired token",
  "message": "Please login again to get a new token",
  "code": "INVALID_TOKEN"
}
```

---

## Role-Based Access Control

### How Scoping Works

```
┌─────────────────────────────────────────────────┐
│  GET /doctors/clinic/clinic-123                 │
└─────────────────────────────────────────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        ▼                       ▼
   Super Admin            Clinic Admin
   ✅ Can access          ✅ Can access if clinic-123
   ANY clinic            is THEIR clinic
   
   Returns: ALL           Returns: Doctors IF
   doctors in             they have access
   clinic-123             ❌ 403 if not their clinic
```

### Access Validation

**Super Admin:**
- ✅ Can get doctors from ANY clinic

**Organization Admin:**
- ✅ Can get doctors from clinics in THEIR organization
- ❌ Cannot access other organization's clinics

**Clinic Admin:**
- ✅ Can get doctors from THEIR clinic(s)
- ❌ Cannot access other clinics

**Receptionist:**
- ✅ Can get doctors from THEIR clinic
- ❌ Cannot access other clinics

**Doctor:**
- ✅ Can get doctors from THEIR clinic
- ❌ Cannot access other clinics

---

## Usage Examples

### Example 1: Clinic Admin Gets Their Clinic's Doctors

```bash
# Login as Clinic Admin
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"clinicadmin","password":"pass"}' | jq -r '.accessToken')

# Get doctors in their clinic
curl -X GET http://localhost:8001/api/v1/org/doctors/clinic/MY_CLINIC_ID \
  -H "Authorization: Bearer $TOKEN"

# Response: 200 OK with doctors list ✅
```

---

### Example 2: Receptionist Views Available Doctors

```bash
# Login as Receptionist
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"receptionist","password":"pass"}' | jq -r '.accessToken')

# Get doctors in their clinic (for appointment booking)
curl -X GET http://localhost:8001/api/v1/org/doctors/clinic/CLINIC_ID \
  -H "Authorization: Bearer $TOKEN"

# Use response to show doctor selection dropdown
```

---

### Example 3: Organization Admin Views All Doctors in Org

```bash
# Login as Organization Admin
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"orgadmin","password":"pass"}' | jq -r '.accessToken')

# Get doctors from Clinic A (in their org)
curl -X GET http://localhost:8001/api/v1/org/doctors/clinic/CLINIC_A_ID \
  -H "Authorization: Bearer $TOKEN"
# ✅ Works - Clinic A is in their organization

# Try to get doctors from Clinic B (different org)
curl -X GET http://localhost:8001/api/v1/org/doctors/clinic/CLINIC_B_ID \
  -H "Authorization: Bearer $TOKEN"
# ❌ 403 Forbidden - Outside their organization
```

---

### Example 4: Super Admin Gets Any Clinic's Doctors

```bash
# Login as Super Admin
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"superadmin","password":"pass"}' | jq -r '.accessToken')

# Get doctors from ANY clinic
curl -X GET http://localhost:8001/api/v1/org/doctors/clinic/ANY_CLINIC_ID \
  -H "Authorization: Bearer $TOKEN"

# ✅ Always works - Super Admin has access to everything
```

---

## Frontend Integration

### React/Vue/Angular Example

```javascript
// Get doctors for a clinic
async function getDoctorsByClinic(clinicId) {
  try {
    const token = localStorage.getItem('authToken');
    const response = await fetch(
      `http://localhost:8001/api/v1/org/doctors/clinic/${clinicId}`,
      {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      }
    );
    
    if (!response.ok) {
      if (response.status === 403) {
        console.error('Access denied to this clinic');
        return null;
      }
      throw new Error('Failed to fetch doctors');
    }
    
    const data = await response.json();
    return data.doctors; // Array of doctor objects
    
  } catch (error) {
    console.error('Error fetching doctors:', error);
    throw error;
  }
}

// Usage in component
const doctors = await getDoctorsByClinic('clinic-uuid');

// Display in dropdown for appointment booking
doctors.forEach(doctor => {
  console.log(`${doctor.first_name} ${doctor.last_name} - ${doctor.specialization}`);
});
```

---

## Use Cases

### Use Case 1: Appointment Booking

**Scenario:** Receptionist booking appointment

```javascript
// 1. Get available doctors
const doctors = await getDoctorsByClinic(currentClinicId);

// 2. Show dropdown
<select name="doctor">
  {doctors.map(doc => (
    <option value={doc.id}>
      Dr. {doc.first_name} {doc.last_name} - {doc.specialization}
    </option>
  ))}
</select>

// 3. Book appointment with selected doctor
```

---

### Use Case 2: Doctor Availability Check

**Scenario:** Check which doctors are in the clinic

```javascript
// Get all doctors
const doctors = await getDoctorsByClinic(clinicId);

// Filter active doctors only
const activeDoctors = doctors.filter(doc => doc.is_active);

// Show on dashboard
console.log(`${activeDoctors.length} doctors available`);
```

---

### Use Case 3: Clinic Management

**Scenario:** Clinic admin managing doctors

```javascript
// Get clinic's doctors
const doctors = await getDoctorsByClinic(myClinicId);

// Display with management actions
doctors.forEach(doctor => {
  showDoctorCard({
    name: `Dr. ${doctor.first_name} ${doctor.last_name}`,
    specialization: doctor.specialization,
    fee: doctor.consultation_fee,
    status: doctor.is_active ? 'Active' : 'Inactive',
    actions: ['Edit', 'View Schedule', 'Manage Leave']
  });
});
```

---

## Data Model

### Doctor Object Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `id` | UUID | Doctor ID | "doctor-uuid" |
| `user_id` | UUID | Associated user ID | "user-uuid" |
| `first_name` | String | Doctor's first name | "John" |
| `last_name` | String | Doctor's last name | "Smith" |
| `doctor_code` | String | Doctor code | "DOC001" |
| `specialization` | String | Medical specialization | "Cardiology" |
| `license_number` | String | Medical license | "MED-12345" |
| `consultation_fee` | Decimal | Consultation fee | 150.00 |
| `follow_up_fee` | Decimal | Follow-up fee | 75.00 |
| `follow_up_days` | Integer | Follow-up period | 7 |
| `is_main_doctor` | Boolean | Is main doctor | true |
| `is_active` | Boolean | Active status | true |
| `email` | String | Email address | "john@clinic.com" |
| `phone` | String | Phone number | "+1234567890" |
| `created_at` | Timestamp | Created date | "2024-01-01T00:00:00Z" |

---

## Combining with Other APIs

### Get Doctors + Their Leave Status

```javascript
// 1. Get all doctors in clinic
const doctors = await getDoctorsByClinic(clinicId);

// 2. For each doctor, get their current leave status
for (const doctor of doctors) {
  const leaves = await fetch(
    `/api/v1/org/doctor-leaves?doctor_id=${doctor.id}&status=approved`,
    { headers: { 'Authorization': `Bearer ${token}` }}
  );
  
  const leaveData = await leaves.json();
  doctor.onLeave = leaveData.leaves.some(leave => {
    const today = new Date();
    const fromDate = new Date(leave.from_date);
    const toDate = new Date(leave.to_date);
    return today >= fromDate && today <= toDate;
  });
}

// 3. Display doctors with leave status
doctors.forEach(doc => {
  console.log(`${doc.first_name} ${doc.last_name} - ${doc.onLeave ? '🏖️ On Leave' : '✅ Available'}`);
});
```

---

## Alternative: Get Clinic Links

**Endpoint:** `GET /api/v1/org/clinic-doctor-links?clinic_id=CLINIC_ID`

**Response Structure:**
```json
{
  "links": [
    {
      "link": "link-uuid",
      "clinic": {
        "name": "Downtown Clinic",
        "clinic_code": "CLINIC001"
      },
      "doctor": {
        "doctor_code": "DOC001",
        "specialization": "Cardiology",
        "first_name": "John",
        "last_name": "Smith",
        "email": "john@clinic.com",
        "username": "drsmith"
      }
    }
  ]
}
```

**Difference:**
- `/doctors/clinic/:id` - Returns doctor details with fees, schedules
- `/clinic-doctor-links?clinic_id=:id` - Returns link information with basic details

**Recommendation:** Use `/doctors/clinic/:id` for most cases (richer data)

---

## Complete Workflow Example

### Scenario: Receptionist Booking Appointment

```javascript
// Step 1: Get clinic's doctors
const doctorsResponse = await fetch(
  `http://localhost:8001/api/v1/org/doctors/clinic/${clinicId}`,
  { headers: { 'Authorization': `Bearer ${token}` }}
);
const { doctors } = await doctorsResponse.json();

// Step 2: Filter by specialization (if needed)
const cardiologists = doctors.filter(
  doc => doc.specialization === 'Cardiology'
);

// Step 3: Check availability (combine with leave API)
for (const doctor of cardiologists) {
  const leavesResponse = await fetch(
    `/api/v1/org/doctor-leaves?doctor_id=${doctor.id}&status=approved`,
    { headers: { 'Authorization': `Bearer ${token}` }}
  );
  const { leaves } = await leavesResponse.json();
  
  // Check if doctor is on leave today
  doctor.available = !leaves.some(leave => {
    const today = new Date().toISOString().split('T')[0];
    return today >= leave.from_date && today <= leave.to_date;
  });
}

// Step 4: Show available doctors with fees
const availableDoctors = cardiologists.filter(doc => 
  doc.is_active && doc.available
);

// Step 5: Display selection
availableDoctors.forEach(doctor => {
  console.log(`
    Dr. ${doctor.first_name} ${doctor.last_name}
    Specialization: ${doctor.specialization}
    Consultation Fee: $${doctor.consultation_fee}
    ✅ Available
  `);
});

// Step 6: Book appointment with selected doctor
```

---

## PowerShell Examples

### Get Doctors in a Clinic

```powershell
# Login
$response = Invoke-RestMethod -Uri "http://localhost:8000/api/v1/auth/login" `
  -Method POST -ContentType "application/json" `
  -Body '{"login":"receptionist","password":"pass"}'
$token = $response.accessToken

# Get clinic's doctors
$doctors = Invoke-RestMethod `
  -Uri "http://localhost:8001/api/v1/org/doctors/clinic/CLINIC_ID" `
  -Method GET `
  -Headers @{Authorization="Bearer $token"}

# Display doctors
$doctors.doctors | ForEach-Object {
    Write-Host "Dr. $($_.first_name) $($_.last_name) - $($_.specialization)"
    Write-Host "  Fee: $$$($_.consultation_fee)"
    Write-Host "  Status: $(if ($_.is_active) {'Active'} else {'Inactive'})"
    Write-Host ""
}
```

---

## Filtering and Processing

### Get Active Doctors Only

```javascript
const response = await fetch(`/api/v1/org/doctors/clinic/${clinicId}`);
const { doctors } = await response.json();

const activeDoctors = doctors.filter(doc => doc.is_active);
console.log(`${activeDoctors.length} active doctors`);
```

### Group by Specialization

```javascript
const response = await fetch(`/api/v1/org/doctors/clinic/${clinicId}`);
const { doctors } = await response.json();

const bySpecialization = doctors.reduce((acc, doc) => {
  const spec = doc.specialization || 'General';
  if (!acc[spec]) acc[spec] = [];
  acc[spec].push(doc);
  return acc;
}, {});

// Result:
// {
//   "Cardiology": [doctor1, doctor2],
//   "Pediatrics": [doctor3],
//   "General": [doctor4, doctor5]
// }
```

### Find Cheapest Doctor

```javascript
const response = await fetch(`/api/v1/org/doctors/clinic/${clinicId}`);
const { doctors } = await response.json();

const cheapestDoctor = doctors
  .filter(doc => doc.is_active && doc.consultation_fee)
  .reduce((min, doc) => 
    doc.consultation_fee < min.consultation_fee ? doc : min
  );

console.log(`Cheapest consultation: $${cheapestDoctor.consultation_fee}`);
```

---

## Best Practices

### For Frontend Developers

1. **Cache the result** - Doctors don't change often
   ```javascript
   const CACHE_DURATION = 5 * 60 * 1000; // 5 minutes
   // Cache doctors list, refresh every 5 minutes
   ```

2. **Handle errors gracefully**
   ```javascript
   try {
     const doctors = await getDoctorsByClinic(clinicId);
   } catch (error) {
     if (error.status === 403) {
       showMessage('No access to this clinic');
     } else {
       showMessage('Error loading doctors');
     }
   }
   ```

3. **Show loading state**
   ```javascript
   setLoading(true);
   const doctors = await getDoctorsByClinic(clinicId);
   setLoading(false);
   ```

4. **Filter inactive doctors**
   ```javascript
   const activeDoctors = doctors.filter(d => d.is_active);
   ```

---

## Testing

### Manual Testing

```bash
# Test 1: Get doctors (as clinic admin)
curl -X GET http://localhost:8001/api/v1/org/doctors/clinic/CLINIC_ID \
  -H "Authorization: Bearer $TOKEN"
# Expected: 200 OK with doctors

# Test 2: Try different clinic (should fail)
curl -X GET http://localhost:8001/api/v1/org/doctors/clinic/OTHER_CLINIC_ID \
  -H "Authorization: Bearer $TOKEN"
# Expected: 403 Forbidden

# Test 3: No authentication
curl -X GET http://localhost:8001/api/v1/org/doctors/clinic/CLINIC_ID
# Expected: 401 Unauthorized
```

---

## Summary

This API provides:

✅ **Simple endpoint** - `GET /doctors/clinic/:clinic_id`  
✅ **Rich data** - All doctor details including fees, specialization  
✅ **Role-based access** - Automatic scope validation  
✅ **Security** - Cross-clinic access prevented  
✅ **Easy integration** - Simple to use in frontend  
✅ **Production ready** - Fully tested and documented  

Perfect for:
- 📅 Appointment booking systems
- 👥 Doctor management dashboards
- 📊 Clinic analytics
- 🏥 Patient portals

---

**API Endpoint:** `GET /api/v1/org/doctors/clinic/:clinic_id`  
**Controller:** `doctor_leave.controller.go` - `GetDoctorsByClinic()`  
**Status:** ✅ Ready to Use (after Docker restart)

---

**Version:** 1.0.0  
**Created:** October 7, 2025  
**Status:** ✅ Production Ready

