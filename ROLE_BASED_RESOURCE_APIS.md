# Role-Based Resource APIs Documentation

## Overview

This document describes the **role-based, scope-filtered resource APIs** that automatically return data based on the user's role and their associated organization/clinic. The same API endpoint returns different data depending on who calls it.

## Core Principle

**One API, Different Data Based on Role**

```
┌────────────────────────────────────────┐
│  GET /api/v1/auth/resources/patients  │
└────────────────────────────────────────┘
                  │
    ┌─────────────┴─────────────┐
    │                           │
    ▼                           ▼
Super Admin                 Clinic Admin
Returns: ALL               Returns: Only
1000+ patients             patients in
platform-wide              their clinic
                           (20 patients)
```

---

## Resource Categories

### 1. Clinics
- List all clinics with filtering
- Includes doctor, patient, and staff counts
- Auto-scoped by role

### 2. Patients
- List patients with medical history
- Shows associated clinics
- Auto-scoped by role

### 3. Doctors
- List doctors with specializations
- Shows consultation fees, schedules
- Auto-scoped by role

### 4. Staff
- List staff (receptionists, pharmacists, lab technicians, billing staff)
- Shows role and clinic assignment
- Auto-scoped by role

---

## API Endpoints by Admin Level

### Super Admin APIs
**Base Path:** `/api/v1/auth/admin/resources`

| Endpoint | Returns |
|----------|---------|
| `GET /admin/resources/clinics` | ALL clinics platform-wide |
| `GET /admin/resources/patients` | ALL patients platform-wide |
| `GET /admin/resources/doctors` | ALL doctors platform-wide |
| `GET /admin/resources/staff` | ALL staff platform-wide |

### Organization Admin APIs
**Base Path:** `/api/v1/auth/org-admin/resources`

| Endpoint | Returns |
|----------|---------|
| `GET /org-admin/resources/clinics` | Clinics in their organization |
| `GET /org-admin/resources/patients` | Patients in their organization's clinics |
| `GET /org-admin/resources/doctors` | Doctors in their organization's clinics |
| `GET /org-admin/resources/staff` | Staff in their organization's clinics |

### Clinic Admin APIs
**Base Path:** `/api/v1/auth/clinic-admin/resources`

| Endpoint | Returns |
|----------|---------|
| `GET /clinic-admin/resources/clinics` | Only their clinic(s) |
| `GET /clinic-admin/resources/patients` | Patients in their clinic |
| `GET /clinic-admin/resources/doctors` | Doctors in their clinic |
| `GET /clinic-admin/resources/staff` | Staff in their clinic |

### Staff APIs (Doctors, Receptionists, Pharmacy, Lab)
**Base Path:** `/api/v1/auth/resources`

| Endpoint | Returns |
|----------|---------|
| `GET /resources/clinics` | Their clinic(s) |
| `GET /resources/patients` | Patients in their clinic |
| `GET /resources/doctors` | Doctors in their clinic |
| `GET /resources/staff` | Staff in their clinic |

---

## Detailed API Documentation

### 1. List Clinics

**Endpoints:**
- `GET /admin/resources/clinics` (Super Admin)
- `GET /org-admin/resources/clinics` (Org Admin)
- `GET /clinic-admin/resources/clinics` (Clinic Admin)
- `GET /resources/clinics` (All Staff)

**Query Parameters:**
- `page` (integer, optional): Page number (default: 1)
- `page_size` (integer, optional): Items per page (default: 20, max: 100)
- `search` (string, optional): Search in clinic name, code, email
- `is_active` (boolean, optional): Filter by active status

**Response:**
```json
{
  "clinics": [
    {
      "id": "clinic-uuid",
      "organization_id": "org-uuid",
      "organization_name": "Hospital ABC",
      "user_id": "user-uuid",
      "clinic_code": "CLINIC001",
      "name": "Downtown Clinic",
      "email": "downtown@hospital.com",
      "phone": "+1234567890",
      "address": "123 Main St",
      "license_number": "LIC-12345",
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "doctor_count": 5,
      "patient_count": 150,
      "staff_count": 8
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_count": 45,
    "total_pages": 3
  },
  "scope": {
    "is_super_admin": false,
    "is_organization_admin": true,
    "is_clinic_admin": false
  }
}
```

**Scoping Logic:**
- **Super Admin:** Returns ALL clinics
- **Org Admin:** Returns clinics WHERE `organization_id` IN (their organizations)
- **Clinic Admin:** Returns clinics WHERE `id` IN (their clinics)
- **Staff:** Returns clinics WHERE `id` IN (clinics from their user_roles)

---

### 2. List Patients

**Endpoints:**
- `GET /admin/resources/patients` (Super Admin)
- `GET /org-admin/resources/patients` (Org Admin)
- `GET /clinic-admin/resources/patients` (Clinic Admin)
- `GET /resources/patients` (All Staff)

**Query Parameters:**
- `page` (integer, optional): Page number (default: 1)
- `page_size` (integer, optional): Items per page (default: 20, max: 100)
- `search` (string, optional): Search in patient name, email, phone
- `clinic_id` (string, optional): Filter by specific clinic

**Response:**
```json
{
  "patients": [
    {
      "id": "patient-uuid",
      "user_id": "user-uuid",
      "first_name": "John",
      "last_name": "Doe",
      "email": "john.doe@example.com",
      "phone": "+1234567890",
      "date_of_birth": "1990-01-01T00:00:00Z",
      "gender": "male",
      "blood_group": "O+",
      "medical_history": "No significant history",
      "allergies": "None",
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "clinics": ["clinic-uuid-1", "clinic-uuid-2"],
      "clinic_names": ["Downtown Clinic", "Uptown Clinic"]
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_count": 150,
    "total_pages": 8
  },
  "scope": {
    "is_super_admin": false,
    "is_organization_admin": false,
    "is_clinic_admin": true
  }
}
```

**Scoping Logic:**
- **Super Admin:** Returns ALL patients
- **Org Admin:** Returns patients registered in their organization's clinics
- **Clinic Admin:** Returns patients registered in their clinic(s)
- **Staff:** Returns patients registered in their clinic(s)

---

### 3. List Doctors

**Endpoints:**
- `GET /admin/resources/doctors` (Super Admin)
- `GET /org-admin/resources/doctors` (Org Admin)
- `GET /clinic-admin/resources/doctors` (Clinic Admin)
- `GET /resources/doctors` (All Staff)

**Query Parameters:**
- `page` (integer, optional): Page number (default: 1)
- `page_size` (integer, optional): Items per page (default: 20, max: 100)
- `search` (string, optional): Search in doctor name, code, specialization
- `clinic_id` (string, optional): Filter by specific clinic
- `specialization` (string, optional): Filter by specialization

**Response:**
```json
{
  "doctors": [
    {
      "id": "doctor-uuid",
      "user_id": "user-uuid",
      "clinic_id": "clinic-uuid",
      "clinic_name": "Downtown Clinic",
      "first_name": "Dr. Jane",
      "last_name": "Smith",
      "email": "jane.smith@hospital.com",
      "phone": "+1234567890",
      "doctor_code": "DOC001",
      "specialization": "Cardiology",
      "license_number": "MED-12345",
      "consultation_fee": 150.00,
      "follow_up_fee": 75.00,
      "follow_up_days": 7,
      "is_main_doctor": true,
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_count": 25,
    "total_pages": 2
  },
  "scope": {
    "is_super_admin": false,
    "is_organization_admin": true,
    "is_clinic_admin": false
  }
}
```

**Scoping Logic:**
- **Super Admin:** Returns ALL doctors
- **Org Admin:** Returns doctors in their organization's clinics
- **Clinic Admin:** Returns doctors in their clinic(s)
- **Staff:** Returns doctors in their clinic(s)

---

### 4. List Staff

**Endpoints:**
- `GET /admin/resources/staff` (Super Admin)
- `GET /org-admin/resources/staff` (Org Admin)
- `GET /clinic-admin/resources/staff` (Clinic Admin)
- `GET /resources/staff` (All Staff)

**Query Parameters:**
- `page` (integer, optional): Page number (default: 1)
- `page_size` (integer, optional): Items per page (default: 20, max: 100)
- `search` (string, optional): Search in staff name, email
- `clinic_id` (string, optional): Filter by specific clinic
- `role` (string, optional): Filter by role (receptionist, pharmacist, lab_technician, billing_staff)

**Response:**
```json
{
  "staff": [
    {
      "id": "user-uuid",
      "first_name": "Alice",
      "last_name": "Johnson",
      "email": "alice.johnson@hospital.com",
      "phone": "+1234567890",
      "role": "receptionist",
      "clinic_id": "clinic-uuid",
      "clinic_name": "Downtown Clinic",
      "is_active": true,
      "assigned_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_count": 8,
    "total_pages": 1
  },
  "scope": {
    "is_super_admin": false,
    "is_organization_admin": false,
    "is_clinic_admin": true
  }
}
```

**Staff Roles Included:**
- `receptionist`
- `pharmacist`
- `lab_technician`
- `billing_staff`

**Scoping Logic:**
- **Super Admin:** Returns ALL staff
- **Org Admin:** Returns staff in their organization's clinics
- **Clinic Admin:** Returns staff in their clinic(s)
- **Staff:** Returns staff in their clinic(s)

---

## Usage Examples

### Example 1: Super Admin Lists All Clinics

```bash
curl -X GET "http://localhost:8000/api/v1/auth/admin/resources/clinics?page=1&page_size=50" \
  -H "Authorization: Bearer SUPER_ADMIN_TOKEN"
```

**Result:** Returns ALL clinics across the entire platform

---

### Example 2: Organization Admin Lists Patients

```bash
curl -X GET "http://localhost:8000/api/v1/auth/org-admin/resources/patients?search=john" \
  -H "Authorization: Bearer ORG_ADMIN_TOKEN"
```

**Result:** Returns patients in their organization's clinics whose name contains "john"

---

### Example 3: Clinic Admin Lists Doctors

```bash
curl -X GET "http://localhost:8000/api/v1/auth/clinic-admin/resources/doctors?specialization=cardiology" \
  -H "Authorization: Bearer CLINIC_ADMIN_TOKEN"
```

**Result:** Returns cardiologists in their clinic only

---

### Example 4: Receptionist Lists Patients

```bash
curl -X GET "http://localhost:8000/api/v1/auth/resources/patients" \
  -H "Authorization: Bearer RECEPTIONIST_TOKEN"
```

**Result:** Returns patients in the receptionist's clinic only

---

### Example 5: Doctor Lists Staff

```bash
curl -X GET "http://localhost:8000/api/v1/auth/resources/staff?role=receptionist" \
  -H "Authorization: Bearer DOCTOR_TOKEN"
```

**Result:** Returns receptionists in the doctor's clinic only

---

## Scope Comparison Matrix

| Resource | Super Admin | Org Admin | Clinic Admin | Doctor/Staff |
|----------|-------------|-----------|--------------|--------------|
| **Clinics** | All platform | Org's clinics | Their clinic | Their clinic |
| **Patients** | All platform | Org's patients | Clinic patients | Clinic patients |
| **Doctors** | All platform | Org's doctors | Clinic doctors | Clinic doctors |
| **Staff** | All platform | Org's staff | Clinic staff | Clinic staff |

---

## Automatic Filtering Logic

### How It Works

1. **User logs in** → JWT token contains `user_id`
2. **Middleware checks role:**
   - Super Admin? → No filtering, return ALL
   - Org Admin? → Get `organization_ids` from user_roles
   - Clinic Admin? → Get `clinic_ids` from user_roles
   - Staff? → Get `clinic_ids` from user_roles
3. **Query automatically filters** data based on scope
4. **Response includes scope information** for debugging

### Database Query Example

**For Organization Admin:**
```sql
-- Automatically added to patients query
WHERE patients.id IN (
    SELECT DISTINCT patient_clinics.patient_id 
    FROM patient_clinics
    JOIN clinics ON clinics.id = patient_clinics.clinic_id
    WHERE clinics.organization_id IN ('org-admin-org-1', 'org-admin-org-2')
)
```

**For Clinic Admin:**
```sql
-- Automatically added to patients query
WHERE patients.id IN (
    SELECT DISTINCT patient_clinics.patient_id 
    FROM patient_clinics
    WHERE patient_clinics.clinic_id IN ('clinic-admin-clinic-1', 'clinic-admin-clinic-2')
)
```

---

## Security Features

### 1. Automatic Scope Enforcement
- ✅ Cannot bypass scope filtering
- ✅ Enforced at controller level
- ✅ No manual filtering required

### 2. Multi-Tenant Isolation
- ✅ Organization A cannot see Organization B data
- ✅ Clinic 1 cannot see Clinic 2 data
- ✅ Complete data isolation

### 3. Role-Based Access
- ✅ Different roles see different data
- ✅ Same API, different results
- ✅ Transparent to frontend

### 4. Query Performance
- ✅ Optimized database queries
- ✅ Proper indexing on foreign keys
- ✅ Pagination support

---

## Integration with Frontend

### Single API Call for All Roles

```javascript
// Frontend code - same for all roles!
async function fetchPatients() {
  const token = getAuthToken(); // Get user's token
  const response = await fetch('/api/v1/auth/resources/patients', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  const data = await response.json();
  
  // data.patients automatically contains only the patients
  // the user is allowed to see based on their role!
  return data.patients;
}
```

**No role checking needed in frontend** - the backend automatically returns the correct data!

---

## Benefits

### For Development
- ✅ **DRY Principle:** One controller function for all roles
- ✅ **Maintainability:** Single source of truth
- ✅ **Testability:** Easy to test scope filtering

### For Frontend
- ✅ **Simplicity:** Same API endpoint for all users
- ✅ **No role logic:** No need to check roles in frontend
- ✅ **Consistent UX:** Same components for all roles

### For Security
- ✅ **Automatic enforcement:** Cannot be bypassed
- ✅ **Centralized logic:** Scope filtering in one place
- ✅ **Audit trail:** Scope information in logs

### For Users
- ✅ **Fast responses:** Optimized queries with proper indexing
- ✅ **Relevant data:** Only see what matters to them
- ✅ **Intuitive:** Data matches their expectations

---

## Testing

### Test 1: Super Admin Sees All

```bash
# Login as Super Admin
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"superadmin","password":"pass"}' | jq -r '.accessToken')

# List all clinics
curl -X GET "http://localhost:8000/api/v1/auth/admin/resources/clinics" \
  -H "Authorization: Bearer $TOKEN"

# Expected: ALL clinics across platform
```

### Test 2: Org Admin Sees Only Their Org

```bash
# Login as Org Admin
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"orgadmin","password":"pass"}' | jq -r '.accessToken')

# List clinics
curl -X GET "http://localhost:8000/api/v1/auth/org-admin/resources/clinics" \
  -H "Authorization: Bearer $TOKEN"

# Expected: Only clinics in their organization
```

### Test 3: Clinic Admin Sees Only Their Clinic

```bash
# Login as Clinic Admin
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"clinicadmin","password":"pass"}' | jq -r '.accessToken')

# List patients
curl -X GET "http://localhost:8000/api/v1/auth/clinic-admin/resources/patients" \
  -H "Authorization: Bearer $TOKEN"

# Expected: Only patients in their clinic
```

### Test 4: Doctor Sees Only Their Clinic Patients

```bash
# Login as Doctor
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"doctor","password":"pass"}' | jq -r '.accessToken')

# List patients
curl -X GET "http://localhost:8000/api/v1/auth/resources/patients" \
  -H "Authorization: Bearer $TOKEN"

# Expected: Only patients in the doctor's clinic
```

---

## Advanced Features

### Multiple Clinic Support

Users can be assigned to multiple clinics. The API automatically returns data from ALL their clinics.

```json
// User has access to 2 clinics
{
  "user_id": "user-123",
  "clinics": ["clinic-a", "clinic-b"]
}

// API returns patients from BOTH clinics
GET /resources/patients
→ Returns patients from clinic-a AND clinic-b
```

### Clinic-Specific Filtering

Users can filter by specific clinic even if they have access to multiple:

```bash
# User has access to 3 clinics
# But wants to see only Clinic A patients
GET /resources/patients?clinic_id=clinic-a

# Returns: Only Clinic A patients
```

### Combined Filters

Filters can be combined for powerful queries:

```bash
# Org Admin wants to see:
# - Patients in Clinic A
# - Named "John"
# - Page 2
GET /org-admin/resources/patients?clinic_id=clinic-a&search=john&page=2

# Result: John patients in Clinic A, page 2
```

---

## Performance Considerations

### Database Indexing

Ensure these indexes exist for optimal performance:

```sql
CREATE INDEX idx_patient_clinics_clinic_id ON patient_clinics(clinic_id);
CREATE INDEX idx_patient_clinics_patient_id ON patient_clinics(patient_id);
CREATE INDEX idx_doctors_clinic_id ON doctors(clinic_id);
CREATE INDEX idx_clinics_organization_id ON clinics(organization_id);
CREATE INDEX idx_user_roles_clinic_id ON user_roles(clinic_id);
CREATE INDEX idx_user_roles_organization_id ON user_roles(organization_id);
```

### Query Optimization

- ✅ Use `LIMIT` and `OFFSET` for pagination
- ✅ Filter at database level, not application level
- ✅ Use `EXISTS` instead of `COUNT(*)` where possible
- ✅ Proper JOIN ordering for query planner

---

## Troubleshooting

### Problem: User sees no data

**Possible causes:**
1. User has no clinic/organization assigned
2. Clinic/organization is inactive
3. user_roles.is_active = false

**Solution:**
```sql
-- Check user's roles and assignments
SELECT ur.*, r.name as role_name 
FROM user_roles ur
JOIN roles r ON r.id = ur.role_id
WHERE ur.user_id = 'user-uuid';

-- Verify clinic assignments
SELECT DISTINCT clinic_id 
FROM user_roles 
WHERE user_id = 'user-uuid' 
AND clinic_id IS NOT NULL 
AND is_active = true;
```

### Problem: User sees data from other clinics

**This indicates a security issue!**

**Check:**
1. Verify middleware is applied to route
2. Check database query includes scope filter
3. Review logs for the specific request

---

## Summary

This role-based resource API system provides:

- ✅ **4 Resource Types:** Clinics, Patients, Doctors, Staff
- ✅ **4 Admin Levels:** Super Admin, Org Admin, Clinic Admin, Staff
- ✅ **16 Total Endpoints:** 4 resources × 4 admin levels
- ✅ **Automatic Scope Filtering:** Based on user's role and assignments
- ✅ **Multi-Tenant Isolation:** Complete data separation
- ✅ **Single Codebase:** One controller for all roles
- ✅ **Production Ready:** Optimized, secure, tested

---

**Version:** 1.0.0  
**Last Updated:** October 7, 2025  
**Maintainer:** Dr&Me Platform Team

