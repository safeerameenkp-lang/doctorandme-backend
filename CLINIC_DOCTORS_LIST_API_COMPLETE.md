# Clinic Doctors List API - Complete Documentation

## Overview
This document describes the APIs available for listing doctors associated with clinics. The system supports **clinic-specific fees** where the same doctor can work at multiple clinics with different consultation fees for each clinic.

---

## API Endpoints

### 1. **Get Doctors by Clinic** (Recommended)
**Endpoint:** `GET /api/doctors/clinic/:clinic_id`

**Description:** Get all doctors linked to a specific clinic with their clinic-specific fees.

**Authentication:** Required (Bearer Token)

**URL Parameters:**
- `clinic_id` (UUID, required) - The ID of the clinic

**Response Format:**
```json
{
  "clinic_id": "clinic-uuid",
  "doctors": [
    {
      "link_id": "link-uuid",
      "doctor_id": "doctor-uuid",
      "user_id": "user-uuid",
      "doctor_code": "DOC001",
      "specialization": "Cardiology",
      "license_number": "LIC123456",
      "first_name": "John",
      "last_name": "Doe",
      "full_name": "John Doe",
      "email": "john.doe@example.com",
      "username": "johndoe",
      "phone": "+1234567890",
      "is_active": true,
      "clinic_specific_fees": {
        "consultation_fee_offline": 500.00,
        "consultation_fee_online": 300.00,
        "follow_up_fee": 250.00,
        "follow_up_days": 7,
        "notes": "Available on weekdays"
      },
      "default_fees": {
        "consultation_fee": 450.00,
        "follow_up_fee": 225.00,
        "follow_up_days": 5
      }
    }
  ],
  "total_doctors": 1
}
```

**Example Usage (PowerShell):**
```powershell
$clinicId = "your-clinic-uuid"
$token = "your-jwt-token"

$response = Invoke-RestMethod -Uri "http://localhost:8080/api/doctors/clinic/$clinicId" `
    -Method GET `
    -Headers @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

Write-Host "Total Doctors: $($response.total_doctors)"
$response.doctors | ForEach-Object {
    Write-Host "Doctor: $($_.full_name) - Specialization: $($_.specialization)"
    Write-Host "Offline Fee: $($_.clinic_specific_fees.consultation_fee_offline)"
}
```

**Example Usage (cURL):**
```bash
curl -X GET "http://localhost:8080/api/doctors/clinic/{clinic-uuid}" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"
```

---

### 2. **Get Clinic-Doctor Links** (Alternative)
**Endpoint:** `GET /api/clinic-doctor-links`

**Description:** List all clinic-doctor links with optional filtering by clinic or doctor.

**Authentication:** Required (Bearer Token)

**Query Parameters:**
- `clinic_id` (UUID, optional) - Filter by clinic
- `doctor_id` (UUID, optional) - Filter by doctor

**Response Format:**
```json
{
  "links": [
    {
      "link_id": "link-uuid",
      "is_active": true,
      "created_at": "2025-10-11T10:30:00Z",
      "clinic": {
        "clinic_id": "clinic-uuid",
        "name": "City Medical Center",
        "clinic_code": "CMC001"
      },
      "doctor": {
        "doctor_id": "doctor-uuid",
        "doctor_code": "DOC001",
        "specialization": "Cardiology",
        "license_number": "LIC123456",
        "first_name": "John",
        "last_name": "Doe",
        "email": "john.doe@example.com",
        "username": "johndoe",
        "phone": "+1234567890"
      },
      "fees": {
        "consultation_fee_offline": 500.00,
        "consultation_fee_online": 300.00,
        "follow_up_fee": 250.00,
        "follow_up_days": 7
      },
      "notes": "Available on weekdays"
    }
  ],
  "count": 1
}
```

**Example Usage (PowerShell):**
```powershell
$clinicId = "your-clinic-uuid"
$token = "your-jwt-token"

$response = Invoke-RestMethod -Uri "http://localhost:8080/api/clinic-doctor-links?clinic_id=$clinicId" `
    -Method GET `
    -Headers @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

Write-Host "Total Links: $($response.count)"
$response.links | ForEach-Object {
    $doctorName = "$($_.doctor.first_name) $($_.doctor.last_name)"
    Write-Host "Doctor: $doctorName - Clinic: $($_.clinic.name)"
    Write-Host "Offline Fee: $($_.fees.consultation_fee_offline)"
}
```

**Example Usage (cURL):**
```bash
# Get all doctors for a specific clinic
curl -X GET "http://localhost:8080/api/clinic-doctor-links?clinic_id={clinic-uuid}" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"

# Get all clinics for a specific doctor
curl -X GET "http://localhost:8080/api/clinic-doctor-links?doctor_id={doctor-uuid}" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"

# Get all clinic-doctor links
curl -X GET "http://localhost:8080/api/clinic-doctor-links" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"
```

---

### 3. **Get Doctors (Generic)**
**Endpoint:** `GET /api/doctors`

**Description:** List all doctors with optional filtering by clinic.

**Authentication:** Required (Bearer Token)

**Query Parameters:**
- `clinic_id` (UUID, optional) - Filter by clinic

**Response Format:**
```json
[
  {
    "doctor": {
      "id": "doctor-uuid",
      "user_id": "user-uuid",
      "clinic_id": "clinic-uuid",
      "doctor_code": "DOC001",
      "specialization": "Cardiology",
      "license_number": "LIC123456",
      "consultation_fee": 450.00,
      "follow_up_fee": 225.00,
      "follow_up_days": 5,
      "is_main_doctor": false,
      "is_active": true,
      "created_at": "2025-10-11T10:30:00Z"
    },
    "user": {
      "first_name": "John",
      "last_name": "Doe",
      "email": "john.doe@example.com",
      "username": "johndoe",
      "phone": "+1234567890"
    }
  }
]
```

**Example Usage:**
```powershell
# Get all doctors
$response = Invoke-RestMethod -Uri "http://localhost:8080/api/doctors" `
    -Method GET `
    -Headers @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

# Get doctors for a specific clinic
$response = Invoke-RestMethod -Uri "http://localhost:8080/api/doctors?clinic_id=$clinicId" `
    -Method GET `
    -Headers @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
```

---

## Comparison of APIs

| Feature | `/doctors/clinic/:clinic_id` | `/clinic-doctor-links?clinic_id=...` | `/doctors?clinic_id=...` |
|---------|------------------------------|--------------------------------------|--------------------------|
| **Clinic-specific fees** | ✅ Yes | ✅ Yes | ❌ No (only default fees) |
| **Link metadata** | ✅ Yes | ✅ Yes | ❌ No |
| **Clinic details** | ❌ No | ✅ Yes | ❌ No |
| **Default fees** | ✅ Yes | ❌ No | ✅ Yes |
| **Full name** | ✅ Yes | ❌ No | ❌ No |
| **Recommended for** | Getting doctors for a clinic | Managing links | General doctor listing |

---

## Use Cases

### Use Case 1: Display Doctors on Appointment Booking
**Scenario:** User wants to book an appointment at a specific clinic and needs to see all available doctors with their fees.

**Recommended API:** `GET /api/doctors/clinic/:clinic_id`

**Reason:** 
- Returns doctors with clinic-specific fees (important for billing)
- Includes full name for display
- Clean response format
- Only active doctors

### Use Case 2: Admin Panel - Manage Clinic-Doctor Relationships
**Scenario:** Admin wants to view all doctors linked to a clinic and their fees.

**Recommended API:** `GET /api/clinic-doctor-links?clinic_id=...`

**Reason:**
- Returns link metadata (link_id, is_active, created_at)
- Includes clinic details
- Shows clinic-specific fees
- Can filter by both clinic and doctor

### Use Case 3: Doctor Profile - Show All Clinics
**Scenario:** Show all clinics where a doctor works.

**Recommended API:** `GET /api/clinic-doctor-links?doctor_id=...`

**Reason:**
- Shows all clinics for the doctor
- Includes clinic details
- Shows fees at each clinic

---

## Understanding Clinic-Specific Fees

### Concept
A doctor can work at multiple clinics with **different fees** at each clinic:

```
Dr. John Doe:
├── City Medical Center
│   ├── Offline: $500
│   ├── Online: $300
│   └── Follow-up: $250 (7 days)
├── Downtown Clinic
│   ├── Offline: $450
│   ├── Online: $250
│   └── Follow-up: $200 (5 days)
└── Suburban Hospital
    ├── Offline: $600
    ├── Online: $350
    └── Follow-up: $300 (10 days)
```

### Default Fees vs Clinic-Specific Fees
- **Default fees** are stored in the `doctors` table
- **Clinic-specific fees** are stored in the `clinic_doctor_links` table
- When clinic-specific fees are set, they override default fees for that clinic
- If clinic-specific fees are NULL, the default fees apply

---

## Complete Test Script (PowerShell)

```powershell
# Configuration
$baseUrl = "http://localhost:8080/api"
$token = "your-jwt-token-here"

# Headers
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# Test 1: Get doctors for a specific clinic
Write-Host "`n=== Test 1: Get Doctors for Clinic ===" -ForegroundColor Cyan
$clinicId = "your-clinic-uuid"
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/doctors/clinic/$clinicId" `
        -Method GET -Headers $headers
    
    Write-Host "✓ Found $($response.total_doctors) doctors" -ForegroundColor Green
    
    $response.doctors | ForEach-Object {
        Write-Host "`nDoctor: $($_.full_name)"
        Write-Host "  Specialization: $($_.specialization)"
        Write-Host "  Clinic Fees:"
        Write-Host "    - Offline: ₹$($_.clinic_specific_fees.consultation_fee_offline)"
        Write-Host "    - Online: ₹$($_.clinic_specific_fees.consultation_fee_online)"
        Write-Host "    - Follow-up: ₹$($_.clinic_specific_fees.follow_up_fee) ($($_.clinic_specific_fees.follow_up_days) days)"
    }
} catch {
    Write-Host "✗ Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 2: Get clinic-doctor links
Write-Host "`n=== Test 2: Get Clinic-Doctor Links ===" -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/clinic-doctor-links?clinic_id=$clinicId" `
        -Method GET -Headers $headers
    
    Write-Host "✓ Found $($response.count) links" -ForegroundColor Green
    
    $response.links | ForEach-Object {
        Write-Host "`nClinic: $($_.clinic.name)"
        Write-Host "Doctor: $($_.doctor.first_name) $($_.doctor.last_name)"
        Write-Host "Link Status: $($_.is_active)"
    }
} catch {
    Write-Host "✗ Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 3: Get all doctors (generic)
Write-Host "`n=== Test 3: Get All Doctors ===" -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/doctors" `
        -Method GET -Headers $headers
    
    Write-Host "✓ Found $($response.Length) doctors" -ForegroundColor Green
    
    $response | Select-Object -First 3 | ForEach-Object {
        Write-Host "`nDoctor: $($_.user.first_name) $($_.user.last_name)"
        Write-Host "  Code: $($_.doctor.doctor_code)"
        Write-Host "  Specialization: $($_.doctor.specialization)"
    }
} catch {
    Write-Host "✗ Error: $($_.Exception.Message)" -ForegroundColor Red
}
```

---

## Error Handling

### Common Errors

1. **404 - Clinic not found or inactive**
```json
{
  "error": "Clinic not found or inactive"
}
```

2. **401 - Unauthorized**
```json
{
  "error": "Unauthorized"
}
```

3. **400 - Invalid UUID**
```json
{
  "error": "Invalid UUID format"
}
```

### Empty Results
When a clinic has no doctors:
```json
{
  "clinic_id": "clinic-uuid",
  "doctors": [],
  "total_doctors": 0
}
```

---

## Summary

### Quick Reference

**Best API for most use cases:**
```
GET /api/doctors/clinic/:clinic_id
```

**To manage links (admin):**
```
GET /api/clinic-doctor-links?clinic_id=...
```

**To see all clinics for a doctor:**
```
GET /api/clinic-doctor-links?doctor_id=...
```

### Key Features
✅ Returns clinic-specific fees (important for billing)  
✅ Returns default fees as fallback  
✅ Only returns active doctors, users, and links  
✅ Sorted by doctor name  
✅ Includes full doctor profile and user information  
✅ Role-based access control (authentication required)  

---

## Next Steps

1. **Link a doctor to a clinic:**
   - Use `POST /api/clinic-doctor-links`
   - See `DOCTOR_LINKING_WORKFLOW.md`

2. **Update clinic-specific fees:**
   - Use `PUT /api/clinic-doctor-links/:link_id`

3. **Remove doctor from clinic:**
   - Use `DELETE /api/clinic-doctor-links/:link_id`

For more details on the complete doctor-clinic workflow, refer to:
- `DOCTOR_LINKING_WORKFLOW.md`
- `DOCTOR_CLINIC_RELATIONSHIP_EXPLAINED.md`
- `CLINIC_SPECIFIC_DOCTOR_FEES.md` (migration 007)

