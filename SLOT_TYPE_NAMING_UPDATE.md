# Slot Type Naming Convention Update ✅

## 🎯 Overview

Updated slot type naming from technical terms to user-friendly business terms.

---

## 📋 Changes Made

### Old Naming (Technical):
- `offline` - Unclear what this means
- `online` - Ambiguous term

### New Naming (Business-Friendly): ✅
- `clinic_visit` - Clear: patient visits clinic
- `video_consultation` - Clear: remote video appointment

---

## 🔄 What Changed

### API Values

| Old Value | New Value | Meaning |
|-----------|-----------|---------|
| `offline` | `clinic_visit` | In-person clinic visit |
| `online` | `video_consultation` | Remote video appointment |

### Follow-Up Types

| Old Value | New Value | Maps To |
|-----------|-----------|---------|
| `follow-up-via-offline` | `follow-up-via-clinic` | `clinic_visit` |
| `follow-up-via-online` | `follow-up-via-video` | `video_consultation` |

---

## 📝 Files Updated

### 1. Database Migration

**File:** `migrations/023_rename_slot_types.sql`

```sql
-- Update existing records
UPDATE doctor_time_slots 
SET slot_type = 'clinic_visit' 
WHERE slot_type = 'offline';

UPDATE doctor_time_slots 
SET slot_type = 'video_consultation' 
WHERE slot_type = 'online';
```

---

### 2. API Controllers

#### `services/organization-service/controllers/doctor_session_slots.controller.go`

**Changes:**
- Updated validation: `oneof=clinic_visit video_consultation`
- Updated mapping logic for follow-up types
- Updated error messages

**Code:**
```go
// Create slots
type CreateDoctorSessionSlotsInput struct {
    SlotType string `json:"slot_type" binding:"required,oneof=clinic_visit video_consultation"`
    // ...
}

// List slots - mapping logic
switch slotType {
case "clinic_visit":
    actualSlotType = "clinic_visit"
case "video_consultation":
    actualSlotType = "video_consultation"
case "follow-up-via-clinic":
    actualSlotType = "clinic_visit"
case "follow-up-via-video":
    actualSlotType = "video_consultation"
}
```

---

#### `services/organization-service/controllers/doctor_time_slots.controller.go`

**Changes:**
- Updated all validation maps
- Updated all error messages

**Code:**
```go
validSlotTypes := map[string]bool{
    "clinic_visit":        true,
    "video_consultation": true,
}
```

---

#### `services/appointment-service/controllers/appointment.controller.go`

**Changes:**
- Updated slot type mapping
- Updated default value

**Code:**
```go
// Map consultation type to slot type
if input.ConsultationType == "video" || input.ConsultationType == "online" {
    slotType = "video_consultation"
} else {
    slotType = "clinic_visit"
}
```

---

### 3. Routes Documentation

**File:** `services/organization-service/routes/organization.routes.go`

**Updated comment:**
```go
// List session-based slots - Query params: 
// doctor_id (required), clinic_id, date, 
// slot_type (clinic_visit/video_consultation/follow-up-via-clinic/follow-up-via-video)
```

---

## 🌐 API Usage

### Create Session Slots

**Before:**
```json
POST /api/organizations/doctor-session-slots
{
  "slot_type": "offline"  // ❌ Old
}
```

**After:** ✅
```json
POST /api/organizations/doctor-session-slots
{
  "slot_type": "clinic_visit"  // ✅ New
}
```

---

### List Session Slots

**Before:**
```bash
GET /doctor-session-slots?
  slot_type=offline  # ❌ Old
```

**After:** ✅
```bash
GET /doctor-session-slots?
  slot_type=clinic_visit  # ✅ New
```

---

### Follow-Up Appointments

**Before:**
```bash
GET /doctor-session-slots?
  slot_type=follow-up-via-offline  # ❌ Old
```

**After:** ✅
```bash
GET /doctor-session-slots?
  slot_type=follow-up-via-clinic  # ✅ New
```

---

## 💻 Flutter Integration Updates

### Dropdown Values

**Before:**
```dart
DropdownMenuItem(value: 'offline', child: Text('Offline')),
DropdownMenuItem(value: 'online', child: Text('Online')),
```

**After:** ✅
```dart
DropdownMenuItem(
  value: 'clinic_visit',
  child: Text('🏥 Clinic Visit'),
),
DropdownMenuItem(
  value: 'video_consultation',
  child: Text('💻 Video Consultation'),
),
```

---

### Full Dropdown with Follow-Up

```dart
final List<Map<String, String>> slotTypes = [
  {
    'value': 'clinic_visit',
    'label': '🏥 Clinic Visit',
  },
  {
    'value': 'video_consultation',
    'label': '💻 Video Consultation',
  },
  {
    'value': 'follow-up-via-clinic',
    'label': '🔄 Follow-Up (Clinic Visit)',
  },
  {
    'value': 'follow-up-via-video',
    'label': '🔄 Follow-Up (Video Consultation)',
  },
];

DropdownButtonFormField<String>(
  decoration: InputDecoration(
    labelText: 'Appointment Type',
    border: OutlineInputBorder(),
  ),
  items: slotTypes.map((type) {
    return DropdownMenuItem<String>(
      value: type['value'],
      child: Text(type['label']!),
    );
  }).toList(),
  onChanged: (value) {
    setState(() {
      selectedSlotType = value;
      _loadSlots(); // Fetch slots with new filter
    });
  },
);
```

---

## 📊 Complete Mapping Table

| User Sees | API Value | Database Value | Meaning |
|-----------|-----------|----------------|---------|
| 🏥 Clinic Visit | `clinic_visit` | `clinic_visit` | In-person appointment |
| 💻 Video Consultation | `video_consultation` | `video_consultation` | Remote video call |
| 🔄 Follow-Up (Clinic) | `follow-up-via-clinic` | `clinic_visit` | Return visit in person |
| 🔄 Follow-Up (Video) | `follow-up-via-video` | `video_consultation` | Return visit via video |

---

## 🧪 Testing

### Test 1: Create Clinic Visit Slots

```bash
POST /api/organizations/doctor-session-slots
Authorization: Bearer {token}
Content-Type: application/json

{
  "doctor_id": "xxx",
  "clinic_id": "xxx",
  "slot_type": "clinic_visit",  // ✅ New value
  "date": "2025-10-20",
  "slot_duration": 5,
  "sessions": [
    {
      "session_name": "Morning",
      "start_time": "09:00:00",
      "end_time": "12:00:00",
      "max_patients": 10,
      "slot_interval_minutes": 5
    }
  ]
}
```

**Expected:** ✅ Success - slots created with `slot_type = 'clinic_visit'`

---

### Test 2: List Video Consultation Slots

```bash
GET /api/organizations/doctor-session-slots?
  doctor_id=xxx&
  clinic_id=xxx&
  date=2025-10-20&
  slot_type=video_consultation  # ✅ New value

Authorization: Bearer {token}
```

**Expected:** ✅ Returns only video consultation slots

---

### Test 3: Follow-Up Clinic Visit

```bash
GET /api/organizations/doctor-session-slots?
  doctor_id=xxx&
  slot_type=follow-up-via-clinic  # ✅ New value
```

**Expected:** ✅ Returns clinic visit slots (mapped from follow-up-via-clinic)

---

### Test 4: Old Values Should Fail

```bash
GET /api/organizations/doctor-session-slots?
  slot_type=offline  # ❌ Old value
```

**Expected:** ❌ Error 400 - Invalid slot_type

---

## ❌ Error Handling

### Invalid Slot Type (Old Value)

**Request:**
```bash
POST /api/organizations/doctor-session-slots
{
  "slot_type": "offline"  // ❌ Old value
}
```

**Response:**
```json
{
  "error": "Invalid slot_type. Must be one of: clinic_visit, video_consultation"
}
```

---

### Invalid Follow-Up Type

**Request:**
```bash
GET /doctor-session-slots?slot_type=follow-up-via-offline  // ❌ Old
```

**Response:**
```json
{
  "error": "Invalid slot_type. Must be 'clinic_visit', 'video_consultation', 'follow-up-via-clinic', or 'follow-up-via-video'"
}
```

---

## 🚀 Deployment Steps

### 1. Run Database Migration

```bash
# Apply migration to update existing data
psql -U postgres -d drandme -f migrations/023_rename_slot_types.sql
```

**Verifies:**
- Existing `offline` → `clinic_visit`
- Existing `online` → `video_consultation`

---

### 2. Deploy Backend Services

```bash
# Rebuild and restart services
docker-compose build organization-service appointment-service
docker-compose up -d organization-service appointment-service
```

---

### 3. Update Flutter App

Update all hardcoded `"offline"` and `"online"` values to:
- `"clinic_visit"`
- `"video_consultation"`

---

### 4. Verify

```bash
# Test new values work
curl "http://localhost:8083/api/organizations/doctor-session-slots?doctor_id=xxx&slot_type=clinic_visit"

# Test old values fail appropriately
curl "http://localhost:8083/api/organizations/doctor-session-slots?doctor_id=xxx&slot_type=offline"
# Should return 400 error
```

---

## 📋 Checklist

| Task | Status | Notes |
|------|--------|-------|
| Database migration created | ✅ Done | `023_rename_slot_types.sql` |
| Update CreateDoctorSessionSlots | ✅ Done | Validation updated |
| Update ListDoctorSessionSlots | ✅ Done | Mapping updated |
| Update doctor_time_slots controller | ✅ Done | All validations updated |
| Update appointment controller | ✅ Done | Slot type mapping updated |
| Update routes documentation | ✅ Done | Comments updated |
| No linter errors | ✅ Done | All services clean |
| Documentation created | ✅ Done | This guide |

---

## 🎯 Benefits of New Naming

✅ **User-Friendly:** "Clinic Visit" is clearer than "offline"  
✅ **Business Terms:** Aligns with medical domain language  
✅ **Self-Documenting:** Code is more readable  
✅ **Less Ambiguous:** "Video Consultation" vs "online"  
✅ **Professional:** Matches industry standards  

---

## 🔄 Migration Impact

### What Gets Updated:
- ✅ All existing `doctor_time_slots` records
- ✅ API validation rules
- ✅ Error messages
- ✅ Documentation

### What Stays Same:
- ✅ Database schema structure (no column changes)
- ✅ Appointment consultation_type (separate field)
- ✅ Table relationships
- ✅ Foreign keys

---

## ✅ Summary

| Aspect | Status |
|--------|--------|
| **Database:** | ✅ Migration ready |
| **Backend:** | ✅ Code updated |
| **API:** | ✅ Values changed |
| **Validation:** | ✅ Updated |
| **Error Messages:** | ✅ Updated |
| **Documentation:** | ✅ Complete |
| **Testing:** | ✅ Ready |
| **Deployment:** | ✅ Ready |

---

**Status:** ✅ **Complete and ready for deployment!**

**Next Steps:**
1. Run migration: `023_rename_slot_types.sql`
2. Deploy backend services
3. Update Flutter app constants
4. Test thoroughly

---

**Done!** 🏥💻✅

