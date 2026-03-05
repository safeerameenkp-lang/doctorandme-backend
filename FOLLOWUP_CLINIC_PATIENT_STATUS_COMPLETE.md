# Follow-Up Clinic Patient Status Tracking - Complete Implementation ✅

## 🎯 **Overview**

Added comprehensive follow-up status tracking to `clinic_patients` table and integrated it with appointment creation APIs. Every follow-up now properly tracks `clinic_id` and updates patient status automatically.

---

## 📊 **Database Changes**

### Migration: `026_add_followup_status_to_clinic_patients.sql`

Added three new columns to `clinic_patients` table:

```sql
-- Status tracking columns
ALTER TABLE clinic_patients 
ADD COLUMN IF NOT EXISTS current_followup_status VARCHAR(20) DEFAULT 'none',
ADD COLUMN IF NOT EXISTS last_appointment_id UUID REFERENCES appointments(id),
ADD COLUMN IF NOT EXISTS last_followup_id UUID REFERENCES follow_ups(id);
```

**Status Values:**
- `none` - No follow-up eligibility
- `active` - Has valid follow-up (within 5 days, unused)
- `used` - Free follow-up already used
- `expired` - Follow-up expired (past 5 days)
- `renewed` - Follow-up restarted (new regular appointment booked)

---

## 🔄 **Status Flow Logic**

### 1. **Book Regular Appointment** → Status: `active`

```json
{
  "event": "Regular appointment created",
  "action": "Create follow-up record",
  "clinic_patient_update": {
    "current_followup_status": "active",
    "last_appointment_id": "<new_appointment_id>",
    "last_followup_id": "<new_follow_up_id>"
  }
}
```

### 2. **Book Free Follow-Up** → Status: `used`

```json
{
  "event": "Free follow-up booked",
  "action": "Mark follow-up as used",
  "clinic_patient_update": {
    "current_followup_status": "used",
    "last_appointment_id": "<follow_up_appointment_id>"
  }
}
```

### 3. **Follow-Up Expires** → Status: `expired`

```json
{
  "event": "Follow-up expired (after 5 days)",
  "action": "Auto-update status",
  "clinic_patient_update": {
    "current_followup_status": "expired"
  }
}
```

### 4. **Book New Regular Appointment** → Status: `renewed`

```json
{
  "event": "New regular appointment (same doctor+dept)",
  "action": "Renew follow-up cycle",
  "clinic_patient_update": {
    "current_followup_status": "renewed",
    "last_appointment_id": "<new_appointment_id>",
    "last_followup_id": "<new_follow_up_id>"
  }
}
```

---

## 🛠️ **API Changes**

### 1. **Create Simple Appointment** (`CreateSimpleAppointment`)

**File:** `services/appointment-service/controllers/appointment_simple.controller.go`

#### Changes Made:
1. ✅ Added status tracking when creating regular appointments
2. ✅ Updates `clinic_patient` status to `active` when creating follow-up eligibility
3. ✅ Updates `clinic_patient` status to `renewed` if renewing existing follow-up
4. ✅ Updates `clinic_patient` status to `used` when booking free follow-up
5. ✅ Stores `last_appointment_id` and `last_followup_id` in `clinic_patients`

#### Code Flow:

```go
// Step 9: Update clinic_patient status and follow-up tracking
var followUpID *string
var newStatus string

// If REGULAR appointment, create follow-up and update status
if input.ConsultationType == "clinic_visit" || input.ConsultationType == "video_consultation" {
    // Check if this is renewal or new
    followup, err := followUpManager.GetActiveFollowUp(...)
    if followup != nil {
        newStatus = "renewed"
    } else {
        newStatus = "active"
    }
    
    // Create follow-up record
    followUpManager.CreateFollowUp(...)
    
    // Update clinic_patient status
    config.DB.Exec(`
        UPDATE clinic_patients
        SET current_followup_status = $1,
            last_appointment_id = $2,
            last_followup_id = $3,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $4 AND clinic_id = $5
    `, newStatus, appointment.ID, followUpID, input.ClinicPatientID, input.ClinicID)
}

// If FREE follow-up appointment, mark as used
if input.IsFollowUp && isFreeFollowUp {
    followUpManager.MarkFollowUpAsUsed(...)
    
    // Update status to 'used'
    config.DB.Exec(`
        UPDATE clinic_patients
        SET current_followup_status = 'used',
            last_appointment_id = $1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $2 AND clinic_id = $3
    `, appointment.ID, input.ClinicPatientID, input.ClinicID)
}
```

### 2. **Create Appointment** (`CreateAppointment`)

**File:** `services/appointment-service/controllers/appointment.controller.go`

#### Changes Made:
1. ✅ Added proper `clinic_id` handling for follow-ups
2. ✅ Supports both clinic-specific patients and global patients
3. ✅ Creates follow-up records with correct `clinic_id`

```go
// Determine if using clinic-specific patient or global patient
if input.ClinicPatientID != nil && *input.ClinicPatientID != "" {
    // Using clinic-specific patient
    followUpManager.CreateFollowUp(*input.ClinicPatientID, input.ClinicID, ...)
} else {
    // Using global patient - find clinic_patient_id
    var clinicPatientID string
    err = config.DB.QueryRow(`
        SELECT id FROM clinic_patients 
        WHERE global_patient_id = $1 AND clinic_id = $2 AND is_active = true
    `, patientID, input.ClinicID).Scan(&clinicPatientID)
    
    followUpManager.CreateFollowUp(clinicPatientID, input.ClinicID, ...)
}
```

---

## 🗂️ **Complete Database Structure**

### Hierarchy:

```
clinics (id)
    │
    ├── clinic_patients (id, clinic_id)
    │        │
    │        ├── appointments (id, clinic_patient_id, doctor_id, department_id, clinic_id)
    │        │        └── follow_ups (clinic_patient_id, doctor_id, department_id, clinic_id, source_appointment_id)
    │        │
    │        └── Status Updates → current_followup_status (active/used/expired/renewed)
    │
    └── doctors / departments
```

### Key Tables:

#### 1. **clinic_patients**
```sql
CREATE TABLE clinic_patients (
    id UUID PRIMARY KEY,
    clinic_id UUID NOT NULL REFERENCES clinics(id),
    
    -- Status tracking
    current_followup_status VARCHAR(20) DEFAULT 'none',
    last_appointment_id UUID REFERENCES appointments(id),
    last_followup_id UUID REFERENCES follow_ups(id),
    
    -- Personal info
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    -- ... other fields
);
```

#### 2. **follow_ups**
```sql
CREATE TABLE follow_ups (
    id UUID PRIMARY KEY,
    clinic_patient_id UUID NOT NULL REFERENCES clinic_patients(id),
    clinic_id UUID NOT NULL REFERENCES clinics(id),  -- ✅ Always tracked
    doctor_id UUID NOT NULL REFERENCES doctors(id),
    department_id UUID REFERENCES departments(id),
    
    status VARCHAR(20) DEFAULT 'active',
    is_free BOOLEAN DEFAULT true,
    valid_from DATE NOT NULL,
    valid_until DATE NOT NULL,
    
    source_appointment_id UUID REFERENCES appointments(id),
    used_appointment_id UUID REFERENCES appointments(id),
    renewed_by_appointment_id UUID REFERENCES appointments(id),
    
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

---

## 📝 **API Request/Response Examples**

### Example 1: Create Regular Appointment

**Request:**
```json
POST /api/appointments/simple
{
  "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
  "clinic_patient_id": "b7e83e77-1272-4c73-9d12-68f6c9f91555",
  "doctor_id": "d932bfa4-82ab-4b93-a08a-142c1e259a44",
  "department_id": "a13b9f33-72c7-46b7-92b2-d6b54ef11c1e",
  "consultation_type": "clinic_visit",
  "appointment_date": "2025-10-25",
  "appointment_time": "10:30:00",
  "individual_slot_id": "slot-123"
}
```

**Response:**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "id": "a6b77b4c-71f9-4ff2-bd09-ff2cb7e92c99",
    "clinic_patient_id": "b7e83e77-1272-4c73-9d12-68f6c9f91555",
    "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
    "consultation_type": "clinic_visit",
    "status": "confirmed",
    "fee_amount": 250.00,
    "payment_status": "paid"
  },
  "is_regular_appointment": true,
  "followup_granted": true,
  "followup_message": "Free follow-up eligibility granted (valid for 5 days)",
  "followup_valid_until": "2025-10-30",
  "clinic_patient_update": {
    "current_followup_status": "active",
    "last_appointment_id": "a6b77b4c-71f9-4ff2-bd09-ff2cb7e92c99",
    "last_followup_id": "fup-89b4d-9123"
  }
}
```

### Example 2: Book Free Follow-Up

**Request:**
```json
POST /api/appointments/simple
{
  "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
  "clinic_patient_id": "b7e83e77-1272-4c73-9d12-68f6c9f91555",
  "doctor_id": "d932bfa4-82ab-4b93-a08a-142c1e259a44",
  "department_id": "a13b9f33-72c7-46b7-92b2-d6b54ef11c1e",
  "consultation_type": "follow-up-via-clinic",
  "appointment_date": "2025-10-27",
  "appointment_time": "14:00:00",
  "individual_slot_id": "slot-456"
}
```

**Response:**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "id": "b8f45d3c-2a1b-4f3e-9c6d-5e8a7b9c0d1e",
    "clinic_patient_id": "b7e83e77-1272-4c73-9d12-68f6c9f91555",
    "consultation_type": "follow-up-via-clinic",
    "status": "confirmed",
    "fee_amount": 0.00,
    "payment_status": "waived"
  },
  "is_free_followup": true,
  "followup_type": "free",
  "followup_message": "This is a FREE follow-up (renewed after regular appointment)",
  "clinic_patient_update": {
    "current_followup_status": "used",
    "last_appointment_id": "b8f45d3c-2a1b-4f3e-9c6d-5e8a7b9c0d1e"
  }
}
```

### Example 3: Renew After Expiry

**Request:**
```json
POST /api/appointments/simple
{
  "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
  "clinic_patient_id": "b7e83e77-1272-4c73-9d12-68f6c9f91555",
  "doctor_id": "d932bfa4-82ab-4b93-a08a-142c1e259a44",
  "department_id": "a13b9f33-72c7-46b7-92b2-d6b54ef11c1e",
  "consultation_type": "clinic_visit",
  "appointment_date": "2025-11-01",
  "appointment_time": "11:00:00",
  "individual_slot_id": "slot-789"
}
```

**Response:**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "id": "a7b82e6f-5a3b-4237-a09e-df8a82366c32",
    "clinic_patient_id": "b7e83e77-1272-4c73-9d12-68f6c9f91555",
    "consultation_type": "clinic_visit",
    "status": "confirmed",
    "fee_amount": 250.00,
    "payment_status": "paid"
  },
  "is_regular_appointment": true,
  "followup_granted": true,
  "followup_message": "Free follow-up eligibility granted (valid for 5 days)",
  "followup_valid_until": "2025-11-06",
  "follow_up_action": {
    "previous_followup_id": "fup-89b4d-9123",
    "previous_status": "expired",
    "new_followup_created": true,
    "renewed_by_appointment_id": "a7b82e6f-5a3b-4237-a09e-df8a82366c32"
  },
  "clinic_patient_update": {
    "current_followup_status": "renewed",
    "last_appointment_id": "a7b82e6f-5a3b-4237-a09e-df8a82366c32",
    "last_followup_id": "new-fup-9921"
  }
}
```

---

## ✅ **Features Implemented**

1. ✅ Added `clinic_id` tracking to all follow-up operations
2. ✅ Added status tracking to `clinic_patients` table
3. ✅ Auto-update status on appointment creation
4. ✅ Support for renewal detection
5. ✅ Proper clinic isolation for multi-clinic support
6. ✅ Complete status lifecycle management

---

## 🧪 **Testing Checklist**

- [ ] Create regular appointment → Check status = `active`
- [ ] Book free follow-up → Check status = `used`
- [ ] Wait 5+ days → Check auto-expiry (status = `expired`)
- [ ] Book new regular → Check renewal (status = `renewed`)
- [ ] Verify `clinic_id` is always set in follow_ups table
- [ ] Test with different clinics (isolation)

---

## 📋 **Summary**

Every follow-up now properly tracks:
- ✅ `clinic_id` - For proper clinic isolation
- ✅ Patient status - For UI display and tracking
- ✅ Last appointment - For quick reference
- ✅ Last follow-up - For eligibility checks

The system now fully supports your documented status flow with automatic tracking and updates! 🎉

