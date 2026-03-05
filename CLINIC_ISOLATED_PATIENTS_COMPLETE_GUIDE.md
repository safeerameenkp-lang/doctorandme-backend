# Clinic-Isolated Patients - Complete Guide

## 🎯 Problem Solved

### ❌ Old Problem (Global Users System)
```
Clinic A creates patient: phone +971501234567
  → Creates GLOBAL user
  → Creates GLOBAL patient
  ↓
Clinic B tries same phone: +971501234567
  → ❌ Error: "Phone already exists"
  → OR worse: Links to Clinic A's patient (data leak!)
```

### ✅ New Solution (Clinic-Isolated Patients)
```
Clinic A creates patient: phone +971501234567
  → Creates CLINIC A patient (isolated)
  ↓
Clinic B creates same phone: +971501234567
  → Creates CLINIC B patient (isolated)
  → ✅ NO CONFLICT! Different clinics, different records
```

---

## 📊 Database Structure

### New Table: `clinic_patients`
```sql
CREATE TABLE clinic_patients (
    id              UUID PRIMARY KEY,
    clinic_id       UUID NOT NULL,  -- Belongs to THIS clinic only
    
    -- Personal info
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    phone           VARCHAR(20) NOT NULL,
    email           VARCHAR(100),
    date_of_birth   DATE,
    gender          VARCHAR(20),
    
    -- Medical info
    mo_id           VARCHAR(50),
    medical_history TEXT,
    allergies       TEXT,
    blood_group     VARCHAR(10),
    
    -- Status
    is_active       BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at      TIMESTAMP,
    updated_at      TIMESTAMP,
    
    -- ✅ Key constraint: Phone unique PER CLINIC (not globally)
    UNIQUE (clinic_id, phone),
    UNIQUE (clinic_id, mo_id)
);
```

**Key Feature:** Same phone can exist in different clinics! ✅

---

## 📝 API Endpoints

### 1. Create Clinic-Specific Patient

**Endpoint:**
```
POST /api/organizations/clinic-specific-patients
```

**Request:**
```json
POST /api/organizations/clinic-specific-patients
Content-Type: application/json
Authorization: Bearer {token}

{
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "first_name": "Ahmed",
  "last_name": "Khan",
  "phone": "+971501234567",
  "email": "ahmed@email.com",
  "date_of_birth": "1985-03-20",
  "gender": "male",
  "mo_id": "MO123456",
  "medical_history": "Diabetes",
  "allergies": "Penicillin",
  "blood_group": "O+"
}
```

**Response (201 Created):**
```json
{
  "message": "Patient created successfully for this clinic",
  "patient": {
    "id": "patient-clinic-a-uuid",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "first_name": "Ahmed",
    "last_name": "Khan",
    "phone": "+971501234567",
    "email": "ahmed@email.com",
    "date_of_birth": "1985-03-20",
    "gender": "male",
    "mo_id": "MO123456",
    "medical_history": "Diabetes",
    "allergies": "Penicillin",
    "blood_group": "O+",
    "is_active": true,
    "created_at": "2024-10-15T10:30:00Z",
    "updated_at": "2024-10-15T10:30:00Z"
  }
}
```

**What's Different:**
- ✅ NO global user created
- ✅ NO global patient created
- ✅ Patient belongs ONLY to this clinic
- ✅ Other clinics can't see this patient

---

### 2. List Clinic Patients

**Endpoint:**
```
GET /api/organizations/clinic-specific-patients?clinic_id={uuid}
```

**Request:**
```
GET /clinic-specific-patients?clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&only_active=true
```

**Response (200 OK):**
```json
{
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "total": 2,
  "patients": [
    {
      "id": "patient-1-uuid",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "first_name": "Ahmed",
      "last_name": "Khan",
      "phone": "+971501234567",
      "mo_id": "MO123456",
      "is_active": true,
      "created_at": "2024-10-15T10:30:00Z"
    },
    {
      "id": "patient-2-uuid",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "first_name": "Sara",
      "last_name": "Ali",
      "phone": "+971507654321",
      "mo_id": "MO789012",
      "is_active": true,
      "created_at": "2024-10-14T14:20:00Z"
    }
  ]
}
```

---

### 3. Search Patients (Clinic-Isolated)

**Endpoint:**
```
GET /clinic-specific-patients?clinic_id={uuid}&search={query}
```

**Examples:**
```
GET /clinic-specific-patients?clinic_id=xxx&search=Ahmed
GET /clinic-specific-patients?clinic_id=xxx&search=+971501234567
GET /clinic-specific-patients?clinic_id=xxx&search=MO123456
```

---

### 4. Get Single Patient

**Endpoint:**
```
GET /clinic-specific-patients/:id
```

---

### 5. Update Patient

**Endpoint:**
```
PUT /clinic-specific-patients/:id
```

**Request:**
```json
PUT /clinic-specific-patients/patient-uuid
{
  "medical_history": "Diabetes, Updated with new diagnosis",
  "blood_group": "O+",
  "allergies": "Penicillin, Aspirin"
}
```

---

### 6. Delete Patient (Soft Delete)

**Endpoint:**
```
DELETE /clinic-specific-patients/:id
```

---

## 🔍 How Clinic Isolation Works

### Scenario: Same Person, Different Clinics

**Clinic A (Main Hospital):**
```json
POST /clinic-specific-patients
{
  "clinic_id": "clinic-a-uuid",
  "first_name": "Ahmed",
  "last_name": "Khan",
  "phone": "+971501234567",
  "mo_id": "A-MO-001"
}
```

**Result:**
```
clinic_patients table:
  id: patient-a-uuid
  clinic_id: clinic-a-uuid
  phone: +971501234567
  mo_id: A-MO-001
```

---

**Clinic B (Branch Clinic):**
```json
POST /clinic-specific-patients
{
  "clinic_id": "clinic-b-uuid",
  "first_name": "Ahmed",
  "last_name": "Khan",
  "phone": "+971501234567",
  "mo_id": "B-MO-100"
}
```

**Result:**
```
clinic_patients table:
  id: patient-b-uuid
  clinic_id: clinic-b-uuid
  phone: +971501234567  ✅ Same phone, different clinic OK!
  mo_id: B-MO-100
```

**✅ NO CONFLICT!** Both clinics have their own isolated patient record.

---

## 📋 List Patients (Isolated View)

**Clinic A Query:**
```
GET /clinic-specific-patients?clinic_id=clinic-a-uuid
```

**Clinic A Sees:**
```json
{
  "patients": [
    {
      "id": "patient-a-uuid",
      "clinic_id": "clinic-a-uuid",
      "first_name": "Ahmed",
      "mo_id": "A-MO-001"
    }
  ]
}
```

**Clinic B Query:**
```
GET /clinic-specific-patients?clinic_id=clinic-b-uuid
```

**Clinic B Sees:**
```json
{
  "patients": [
    {
      "id": "patient-b-uuid",
      "clinic_id": "clinic-b-uuid",
      "first_name": "Ahmed",
      "mo_id": "B-MO-100"
    }
  ]
}
```

**✅ Complete Isolation!** Each clinic only sees their own patients.

---

## 🔄 Complete Workflow

### Step 1: Receptionist Creates Patient

**Request:**
```json
POST /clinic-specific-patients
{
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "first_name": "Mohammed",
  "last_name": "Hassan",
  "phone": "+971509876543",
  "mo_id": "MO999888",
  "blood_group": "A+"
}
```

**Response:**
```json
{
  "message": "Patient created successfully for this clinic",
  "patient": {
    "id": "clinic-patient-uuid",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "first_name": "Mohammed",
    "last_name": "Hassan",
    "phone": "+971509876543",
    "mo_id": "MO999888",
    "is_active": true
  }
}
```

---

### Step 2: Book Appointment (Update appointment.controller.go needed)

**Request:**
```json
POST /appointments
{
  "clinic_patient_id": "clinic-patient-uuid",  // NEW field
  "doctor_id": "doctor-uuid",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "consultation_type": "offline"
}
```

---

## ✅ Advantages of Clinic-Specific System

### 1. Complete Isolation ✅
```
Clinic A patients: Only Clinic A can see
Clinic B patients: Only Clinic B can see
NO cross-clinic data access
```

### 2. No Phone Conflicts ✅
```
Same phone can exist in multiple clinics
Constraint: UNIQUE (clinic_id, phone)
```

### 3. No MO ID Conflicts ✅
```
Each clinic has its own MO ID sequence
MO123 at Clinic A ≠ MO123 at Clinic B
```

### 4. Simple Queries ✅
```sql
-- Get clinic patients (fast, simple)
SELECT * FROM clinic_patients
WHERE clinic_id = 'your-clinic-uuid';
-- No JOINs needed!
```

---

## 🔗 Optional: Link to Global Patient Later

**If patient wants multi-clinic access:**

```json
POST /clinic-specific-patients/:id/link-global
{
  "global_patient_id": "global-patient-uuid"
}
```

**Database:**
```sql
UPDATE clinic_patients
SET global_patient_id = 'global-patient-uuid'
WHERE id = 'clinic-patient-id';
```

**Now:**
- ✅ Clinic patient record exists
- ✅ Optionally linked to global patient
- ✅ Can access global medical history if patient consents

---

## 📊 Comparison

| Feature | Global Patients | Clinic-Specific Patients |
|---------|----------------|-------------------------|
| Phone conflicts | ❌ Global conflict | ✅ Per-clinic unique |
| Data isolation | ❌ Shared | ✅ Complete |
| Multi-clinic support | ✅ Built-in | ✅ Optional (via global_patient_id) |
| Query simplicity | ❌ Requires JOINs | ✅ Direct SELECT |
| Privacy | ❌ Shared data | ✅ Clinic-owned |

---

## ✅ Your New API

**Create Patient:**
```
POST /api/organizations/clinic-specific-patients
```

**List Patients:**
```
GET /api/organizations/clinic-specific-patients?clinic_id=your-clinic-uuid
```

**Get Patient:**
```
GET /api/organizations/clinic-specific-patients/:id
```

**Update Patient:**
```
PUT /api/organizations/clinic-specific-patients/:id
```

**Delete Patient:**
```
DELETE /api/organizations/clinic-specific-patients/:id
```

---

## 🚀 Ready to Use

| Component | Status |
|-----------|--------|
| Database table created | ✅ Applied |
| API controller created | ✅ Done |
| Routes registered | ✅ Done |
| Clinic isolation | ✅ Working |
| Phone per-clinic unique | ✅ Working |
| MO ID per-clinic unique | ✅ Working |

---

**Status:** ✅ **Clinic-Isolated Patients Ready!**  
**Endpoint:** `POST /api/organizations/clinic-specific-patients`

Now each clinic has their OWN patients with NO global conflicts! 🎉

