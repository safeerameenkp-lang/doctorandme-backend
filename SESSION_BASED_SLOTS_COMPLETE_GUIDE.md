# Session-Based Time Slots - Complete Implementation Guide

## 🎯 Overview

This is a **3-tier slot management system** that auto-generates individual bookable slots:

```
doctor_time_slots (Day Level)
    ↓
doctor_slot_sessions (Session Level: Morning, Afternoon, etc.)
    ↓
doctor_individual_slots (Individual 5-min bookable slots)
```

### Key Features
- ✅ **Auto-generates** individual slots based on interval
- ✅ **Prevents overlapping** sessions
- ✅ **Tracks booking** at individual slot level
- ✅ **Auto-calculates** day_of_week from date
- ✅ **Prevents duplicate** slots for same date
- ✅ **Smart capacity** management

---

## 📊 Database Schema

### Table 1: `doctor_time_slots` (Day Level)
```sql
id                 UUID PRIMARY KEY
doctor_id          UUID → doctors(id)
clinic_id          UUID → clinics(id)
specific_date      DATE
day_of_week        INT (1=Monday to 7=Sunday, auto-calculated)
slot_type          VARCHAR(20) -- 'offline' or 'online'
slot_duration      INT -- Default duration in minutes
is_active          BOOLEAN
notes              TEXT
created_at         TIMESTAMP
updated_at         TIMESTAMP
```

### Table 2: `doctor_slot_sessions` (Session Level)
```sql
id                      UUID PRIMARY KEY
time_slot_id            UUID → doctor_time_slots(id)
session_name            VARCHAR(50) -- e.g., "Morning Session"
start_time              TIME
end_time                TIME
max_patients            INT
slot_interval_minutes   INT -- Generate slots every X minutes
notes                   TEXT
created_at              TIMESTAMP
updated_at              TIMESTAMP
```

### Table 3: `doctor_individual_slots` (Individual Slot Level)
```sql
id                      UUID PRIMARY KEY
session_id              UUID → doctor_slot_sessions(id)
slot_start              TIME
slot_end                TIME
is_booked               BOOLEAN
booked_patient_id       UUID → users(id)
booked_appointment_id   UUID → appointments(id)
status                  VARCHAR(20) -- 'available', 'booked', 'cancelled', 'blocked'
notes                   TEXT
created_at              TIMESTAMP
updated_at              TIMESTAMP
```

---

## 📝 API Endpoints

### 1. Create Session-Based Slots

**Endpoint:**
```
POST /api/organizations/doctor-session-slots
```

**Request Body:**
```json
{
  "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "slot_type": "offline",
  "slot_duration": 5,
  "date": "2025-10-15",
  "is_available": true,
  "notes": "Regular clinic hours",
  "sessions": [
    {
      "session_name": "Morning Session",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 36,
      "slot_interval_minutes": 5,
      "notes": "Morning consultations"
    },
    {
      "session_name": "Afternoon Session",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 36,
      "slot_interval_minutes": 5,
      "notes": "Afternoon consultations"
    }
  ]
}
```

**What Happens:**
1. ✅ **Validates** doctor, clinic, and link
2. ✅ **Auto-calculates** day_of_week from date (Oct 15, 2025 = Wednesday = 3)
3. ✅ **Creates** 1 doctor_time_slots record
4. ✅ **Creates** 2 doctor_slot_sessions records
5. ✅ **Auto-generates** 72 individual slots (36 morning + 36 afternoon)

**Success Response (201):**
```json
{
  "success": true,
  "message": "Doctor time slots created successfully",
  "data": {
    "id": "time-slot-uuid",
    "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "date": "2025-10-15",
    "day_of_week": 3,
    "slot_type": "offline",
    "is_available": true,
    "sessions": [
      {
        "id": "session-morning-uuid",
        "session_name": "Morning Session",
        "start_time": "09:00",
        "end_time": "12:00",
        "max_patients": 36,
        "slot_interval_minutes": 5,
        "generated_slots": 36,
        "available_slots": 36,
        "booked_slots": 0,
        "notes": "Morning consultations",
        "slots": [
          {
            "id": "slot-1-uuid",
            "slot_start": "09:00",
            "slot_end": "09:05",
            "is_booked": false,
            "status": "available"
          },
          {
            "id": "slot-2-uuid",
            "slot_start": "09:05",
            "slot_end": "09:10",
            "is_booked": false,
            "status": "available"
          },
          // ... 34 more slots
        ]
      },
      {
        "id": "session-afternoon-uuid",
        "session_name": "Afternoon Session",
        "start_time": "14:00",
        "end_time": "17:00",
        "max_patients": 36,
        "slot_interval_minutes": 5,
        "generated_slots": 36,
        "available_slots": 36,
        "booked_slots": 0,
        "notes": "Afternoon consultations",
        "slots": [
          // 36 afternoon slots
        ]
      }
    ]
  }
}
```

---

### 2. List Session-Based Slots

**Endpoint:**
```
GET /api/organizations/doctor-session-slots?doctor_id=xxx&date=2025-10-15&clinic_id=xxx
```

**Query Parameters:**
- `doctor_id` (required) - Doctor UUID
- `date` (optional) - Filter by specific date (YYYY-MM-DD)
- `clinic_id` (optional) - Filter by clinic

**Success Response (200):**
```json
{
  "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "date": "2025-10-15",
  "total": 1,
  "slots": [
    {
      "id": "time-slot-uuid",
      "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "date": "2025-10-15",
      "day_of_week": 3,
      "slot_type": "offline",
      "is_available": true,
      "sessions": [
        {
          "id": "session-uuid",
          "session_name": "Morning Session",
          "start_time": "09:00",
          "end_time": "12:00",
          "max_patients": 36,
          "slot_interval_minutes": 5,
          "generated_slots": 36,
          "available_slots": 35,
          "booked_slots": 1,
          "slots": [
            {
              "id": "slot-1-uuid",
              "slot_start": "09:00",
              "slot_end": "09:05",
              "is_booked": true,
              "booked_patient_id": "patient-uuid",
              "booked_appointment_id": "appointment-uuid",
              "status": "booked"
            },
            {
              "id": "slot-2-uuid",
              "slot_start": "09:05",
              "slot_end": "09:10",
              "is_booked": false,
              "status": "available"
            }
            // ... remaining slots
          ]
        }
      ]
    }
  ]
}
```

---

## 🧠 Auto-Generation Logic

### Example: Morning Session (09:00 - 12:00)

**Input:**
```json
{
  "session_name": "Morning Session",
  "start_time": "09:00",
  "end_time": "12:00",
  "slot_interval_minutes": 5
}
```

**Auto-Generated Slots:**
```
09:00 → 09:05 (Slot 1)
09:05 → 09:10 (Slot 2)
09:10 → 09:15 (Slot 3)
...
11:50 → 11:55 (Slot 35)
11:55 → 12:00 (Slot 36)
```

**Total:** 36 individual bookable slots

### Calculation:
```
Duration = 12:00 - 09:00 = 180 minutes
Slots = 180 ÷ 5 = 36 slots
```

---

## ⚡ Smart Validations

### 1. **Prevent Duplicate Slots**
```sql
-- Checks if slots already exist for this doctor on this date
SELECT EXISTS(
    SELECT 1 FROM doctor_time_slots
    WHERE doctor_id = 'xxx' 
    AND specific_date = '2025-10-15'
    AND is_active = true
)
```

**Error Response (409):**
```json
{
  "error": "Duplicate slots",
  "message": "Time slots already exist for this doctor on 2025-10-15"
}
```

---

### 2. **Prevent Overlapping Sessions**

**Request:**
```json
{
  "sessions": [
    {
      "session_name": "Morning",
      "start_time": "09:00",
      "end_time": "12:00"
    },
    {
      "session_name": "Late Morning",
      "start_time": "11:00",  // ❌ Overlaps with Morning!
      "end_time": "13:00"
    }
  ]
}
```

**Error Response (400):**
```json
{
  "error": "Overlapping sessions",
  "message": "Session 'Late Morning' overlaps with session 'Morning'"
}
```

---

### 3. **Auto-Calculate day_of_week**

**Input:** `"date": "2025-10-15"`

**System Automatically:**
1. Parses date → Wednesday
2. Converts to ISO 8601 → `day_of_week = 3`
3. Stores in database

**Conversion Table:**
```
Monday    → 1
Tuesday   → 2
Wednesday → 3
Thursday  → 4
Friday    → 5
Saturday  → 6
Sunday    → 7
```

---

## 🔄 Complete Flow Example

### Scenario: Dr. Smith's Wednesday Schedule

**Step 1: Create Slots**
```bash
POST /api/organizations/doctor-session-slots
{
  "doctor_id": "dr-smith-uuid",
  "clinic_id": "main-clinic-uuid",
  "slot_type": "offline",
  "slot_duration": 5,
  "date": "2025-10-20",
  "sessions": [
    {
      "session_name": "Morning",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 36,
      "slot_interval_minutes": 5
    }
  ]
}
```

**Result:**
- ✅ 1 `doctor_time_slots` record created
- ✅ 1 `doctor_slot_sessions` record created  
- ✅ 36 `doctor_individual_slots` auto-generated
- ✅ `day_of_week` = 1 (Monday) auto-calculated

---

**Step 2: Patient Books Slot**
```bash
POST /appointments
{
  "patient_id": "patient-1",
  "doctor_id": "dr-smith-uuid",
  "individual_slot_id": "slot-09-00-uuid",
  ...
}
```

**System Updates:**
```sql
UPDATE doctor_individual_slots
SET 
    is_booked = true,
    booked_patient_id = 'patient-1',
    booked_appointment_id = 'appointment-uuid',
    status = 'booked'
WHERE id = 'slot-09-00-uuid'
```

---

**Step 3: List Available Slots**
```bash
GET /api/organizations/doctor-session-slots?doctor_id=dr-smith-uuid&date=2025-10-20
```

**Response Shows:**
```json
{
  "sessions": [
    {
      "session_name": "Morning",
      "generated_slots": 36,
      "available_slots": 35,  // ✅ One slot booked
      "booked_slots": 1,
      "slots": [
        {
          "slot_start": "09:00",
          "slot_end": "09:05",
          "is_booked": true,  // ✅ This one is booked
          "status": "booked"
        },
        {
          "slot_start": "09:05",
          "slot_end": "09:10",
          "is_booked": false,
          "status": "available"
        }
        // ... 34 more available slots
      ]
    }
  ]
}
```

---

## 🎨 UI Integration Example

### Display Available Slots

```javascript
async function loadAvailableSlots(doctorId, date) {
  const response = await fetch(
    `/api/organizations/doctor-session-slots?doctor_id=${doctorId}&date=${date}`
  );
  const data = await response.json();
  
  // Display sessions
  data.slots.forEach(daySlot => {
    daySlot.sessions.forEach(session => {
      console.log(`${session.session_name}: ${session.available_slots}/${session.generated_slots} available`);
      
      // Show only available slots
      const availableSlots = session.slots.filter(slot => 
        !slot.is_booked && slot.status === 'available'
      );
      
      availableSlots.forEach(slot => {
        showSlotButton(slot.id, slot.slot_start, slot.slot_end);
      });
    });
  });
}

function showSlotButton(slotId, start, end) {
  // Display as clickable button in UI
  const button = document.createElement('button');
  button.textContent = `${start} - ${end}`;
  button.onclick = () => bookSlot(slotId);
  document.getElementById('slots-container').appendChild(button);
}
```

---

## 📋 Error Handling

### Error 1: Doctor Not Found
```json
{
  "error": "Doctor not found",
  "message": "Doctor not found or is inactive"
}
```

### Error 2: Clinic Not Found
```json
{
  "error": "Clinic not found",
  "message": "Clinic not found or is inactive"
}
```

### Error 3: Doctor Not Linked to Clinic
```json
{
  "error": "Doctor is not linked to this clinic",
  "message": "The specified doctor is not associated with this clinic"
}
```

### Error 4: Invalid Time Format
```json
{
  "error": "Invalid start_time format in session 0. Use HH:MM format"
}
```

### Error 5: End Time Before Start Time
```json
{
  "error": "Session 0: end_time must be after start_time"
}
```

---

## 💡 Best Practices

### 1. **Use Reasonable Intervals**
```json
// ✅ Good
"slot_interval_minutes": 5   // 5-minute slots
"slot_interval_minutes": 10  // 10-minute slots
"slot_interval_minutes": 15  // 15-minute slots

// ❌ Too Small
"slot_interval_minutes": 1   // Generates too many slots

// ❌ Too Large
"slot_interval_minutes": 120 // Only 1-2 slots per session
```

### 2. **Session Naming**
```json
// ✅ Clear names
"Morning Session"
"Afternoon Session"
"Evening Clinic"
"Emergency Hours"

// ❌ Vague names
"Session 1"
"Time Slot A"
```

### 3. **Max Patients = Generated Slots**
```
For 3-hour session with 5-min intervals:
Slots = 180 ÷ 5 = 36 slots
max_patients = 36 (one patient per slot)
```

---

## ✅ Migration Required

Before using this API, run:

```bash
# Apply migration 013
docker exec -i drandme-backend-postgres-1 psql -U postgres -d drandme -f migrations/013_session_based_slots.sql
```

---

## 🎯 Summary

| Feature | Status |
|---------|--------|
| Auto-generate slots | ✅ Working |
| Auto-calculate day_of_week | ✅ Working |
| Prevent duplicates | ✅ Working |
| Prevent overlaps | ✅ Working |
| Track bookings | ✅ Working |
| Session management | ✅ Working |
| Individual slot status | ✅ Working |

---

**Status:** ✅ Ready for Production  
**Last Updated:** October 15, 2025  
**Version:** 1.0

