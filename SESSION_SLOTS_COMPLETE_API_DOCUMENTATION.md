# Session-Based Time Slots - Complete API Documentation

## 📖 Table of Contents
1. [Overview](#overview)
2. [Database Structure](#database-structure)
3. [API Endpoints](#api-endpoints)
4. [Create Session Slots - Full Examples](#create-session-slots)
5. [List Session Slots - Full Examples](#list-session-slots)
6. [Complete Workflow Examples](#complete-workflow)
7. [Error Handling](#error-handling)
8. [Integration Guide](#integration-guide)

---

## 🎯 Overview

### What is Session-Based Slot System?

A **3-tier slot management system** that automatically generates individual bookable time slots:

```
Level 1: doctor_time_slots (Day)
    ↓
Level 2: doctor_slot_sessions (Sessions: Morning, Afternoon)
    ↓
Level 3: doctor_individual_slots (Auto-generated 5-min bookable slots)
```

### Key Features
- ✅ **Auto-generates** individual slots from sessions
- ✅ **Auto-calculates** day_of_week from date
- ✅ **Multi-clinic** support with clinic_id
- ✅ **Flexible filtering** by clinic, date, slot_type
- ✅ **Real-time** booking status tracking

---

## 📊 Database Structure

### Table 1: doctor_time_slots
```sql
CREATE TABLE doctor_time_slots (
    id              UUID PRIMARY KEY,
    doctor_id       UUID REFERENCES doctors(id),
    clinic_id       UUID REFERENCES clinics(id),
    specific_date   DATE,
    day_of_week     INT,  -- Auto-calculated: 1=Monday to 7=Sunday
    slot_type       VARCHAR(20),  -- 'offline' or 'online'
    slot_duration   INT DEFAULT 5,
    is_active       BOOLEAN DEFAULT TRUE,
    notes           TEXT,
    created_at      TIMESTAMP,
    updated_at      TIMESTAMP
);
```

### Table 2: doctor_slot_sessions
```sql
CREATE TABLE doctor_slot_sessions (
    id                      UUID PRIMARY KEY,
    time_slot_id            UUID REFERENCES doctor_time_slots(id),
    clinic_id               UUID REFERENCES clinics(id),
    session_name            VARCHAR(50),
    start_time              TIME,
    end_time                TIME,
    max_patients            INT DEFAULT 10,
    slot_interval_minutes   INT DEFAULT 5,
    notes                   TEXT,
    created_at              TIMESTAMP,
    updated_at              TIMESTAMP
);
```

### Table 3: doctor_individual_slots
```sql
CREATE TABLE doctor_individual_slots (
    id                      UUID PRIMARY KEY,
    session_id              UUID REFERENCES doctor_slot_sessions(id),
    clinic_id               UUID REFERENCES clinics(id),
    slot_start              TIME,
    slot_end                TIME,
    is_booked               BOOLEAN DEFAULT FALSE,
    booked_patient_id       UUID REFERENCES users(id),
    booked_appointment_id   UUID REFERENCES appointments(id),
    status                  VARCHAR(20) DEFAULT 'available',
    notes                   TEXT,
    created_at              TIMESTAMP,
    updated_at              TIMESTAMP
);
```

---

## 🔌 API Endpoints

### Endpoint 1: Create Session-Based Slots
```
POST /api/organizations/doctor-session-slots
Content-Type: application/json
Authorization: Bearer {token}
```

### Endpoint 2: List Session-Based Slots
```
GET /api/organizations/doctor-session-slots?doctor_id={uuid}&clinic_id={uuid}&date={YYYY-MM-DD}&slot_type={offline|online}
Authorization: Bearer {token}
```

---

## 📝 Create Session Slots - Full Examples

### Example 1: Basic Single Session

**Request:**
```json
POST /api/organizations/doctor-session-slots
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "slot_type": "offline",
  "slot_duration": 5,
  "date": "2025-10-20",
  "is_available": true,
  "notes": "Regular Monday clinic",
  "sessions": [
    {
      "session_name": "Morning Session",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 36,
      "slot_interval_minutes": 5,
      "notes": "General consultations"
    }
  ]
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Doctor time slots created successfully",
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "date": "2025-10-20",
    "day_of_week": 1,
    "slot_type": "offline",
    "is_available": true,
    "sessions": [
      {
        "id": "session-uuid-1",
        "session_name": "Morning Session",
        "start_time": "09:00",
        "end_time": "12:00",
        "max_patients": 36,
        "slot_interval_minutes": 5,
        "generated_slots": 36,
        "available_slots": 36,
        "booked_slots": 0,
        "notes": "General consultations",
        "slots": [
          {
            "id": "slot-uuid-1",
            "slot_start": "09:00",
            "slot_end": "09:05",
            "is_booked": false,
            "status": "available"
          },
          {
            "id": "slot-uuid-2",
            "slot_start": "09:05",
            "slot_end": "09:10",
            "is_booked": false,
            "status": "available"
          },
          {
            "id": "slot-uuid-3",
            "slot_start": "09:10",
            "slot_end": "09:15",
            "is_booked": false,
            "status": "available"
          },
          {
            "id": "slot-uuid-4",
            "slot_start": "09:15",
            "slot_end": "09:20",
            "is_booked": false,
            "status": "available"
          },
          {
            "id": "slot-uuid-5",
            "slot_start": "09:20",
            "slot_end": "09:25",
            "is_booked": false,
            "status": "available"
          }
          // ... 31 more slots until 12:00
        ]
      }
    ]
  }
}
```

**What Happened:**
1. ✅ Created 1 `doctor_time_slots` record for Oct 20, 2025
2. ✅ Auto-calculated `day_of_week = 1` (Monday)
3. ✅ Created 1 `doctor_slot_sessions` record (Morning Session)
4. ✅ Auto-generated 36 `doctor_individual_slots` (09:00 to 12:00, every 5 minutes)

---

### Example 2: Multiple Sessions (Full Day)

**Request:**
```json
POST /api/organizations/doctor-session-slots
Content-Type: application/json
Authorization: Bearer {token}

{
  "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "slot_type": "offline",
  "slot_duration": 5,
  "date": "2025-10-22",
  "is_available": true,
  "notes": "Wednesday full day clinic",
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
    },
    {
      "session_name": "Evening Session",
      "start_time": "18:00",
      "end_time": "20:00",
      "max_patients": 24,
      "slot_interval_minutes": 5,
      "notes": "Evening walk-ins"
    }
  ]
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Doctor time slots created successfully",
  "data": {
    "id": "timeslot-uuid-wed",
    "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "date": "2025-10-22",
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
            "id": "morning-slot-1",
            "slot_start": "09:00",
            "slot_end": "09:05",
            "is_booked": false,
            "status": "available"
          }
          // ... 35 more morning slots
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
          {
            "id": "afternoon-slot-1",
            "slot_start": "14:00",
            "slot_end": "14:05",
            "is_booked": false,
            "status": "available"
          }
          // ... 35 more afternoon slots
        ]
      },
      {
        "id": "session-evening-uuid",
        "session_name": "Evening Session",
        "start_time": "18:00",
        "end_time": "20:00",
        "max_patients": 24,
        "slot_interval_minutes": 5,
        "generated_slots": 24,
        "available_slots": 24,
        "booked_slots": 0,
        "notes": "Evening walk-ins",
        "slots": [
          {
            "id": "evening-slot-1",
            "slot_start": "18:00",
            "slot_end": "18:05",
            "is_booked": false,
            "status": "available"
          }
          // ... 23 more evening slots
        ]
      }
    ]
  }
}
```

**Summary:**
- 📊 Total Slots Generated: 96 (36 + 36 + 24)
- ⏰ Total Hours: 8 hours (3h + 3h + 2h)
- 🏥 All at same clinic, same day, different sessions

---

### Example 3: Online Consultation Session

**Request:**
```json
POST /api/organizations/doctor-session-slots
Content-Type: application/json
Authorization: Bearer {token}

{
  "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "slot_type": "online",
  "slot_duration": 10,
  "date": "2025-10-23",
  "is_available": true,
  "notes": "Telemedicine Thursday",
  "sessions": [
    {
      "session_name": "Online Morning",
      "start_time": "10:00",
      "end_time": "13:00",
      "max_patients": 18,
      "slot_interval_minutes": 10,
      "notes": "Video consultations"
    },
    {
      "session_name": "Online Afternoon",
      "start_time": "15:00",
      "end_time": "18:00",
      "max_patients": 18,
      "slot_interval_minutes": 10,
      "notes": "Video consultations"
    }
  ]
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Doctor time slots created successfully",
  "data": {
    "id": "timeslot-online-thu",
    "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "date": "2025-10-23",
    "day_of_week": 4,
    "slot_type": "online",
    "is_available": true,
    "sessions": [
      {
        "id": "session-online-morning",
        "session_name": "Online Morning",
        "start_time": "10:00",
        "end_time": "13:00",
        "max_patients": 18,
        "slot_interval_minutes": 10,
        "generated_slots": 18,
        "available_slots": 18,
        "booked_slots": 0,
        "notes": "Video consultations",
        "slots": [
          {
            "id": "online-slot-1",
            "slot_start": "10:00",
            "slot_end": "10:10",
            "is_booked": false,
            "status": "available"
          },
          {
            "id": "online-slot-2",
            "slot_start": "10:10",
            "slot_end": "10:20",
            "is_booked": false,
            "status": "available"
          },
          {
            "id": "online-slot-3",
            "slot_start": "10:20",
            "slot_end": "10:30",
            "is_booked": false,
            "status": "available"
          }
          // ... 15 more 10-minute slots
        ]
      },
      {
        "id": "session-online-afternoon",
        "session_name": "Online Afternoon",
        "start_time": "15:00",
        "end_time": "18:00",
        "max_patients": 18,
        "slot_interval_minutes": 10,
        "generated_slots": 18,
        "available_slots": 18,
        "booked_slots": 0,
        "notes": "Video consultations",
        "slots": [
          {
            "id": "online-afternoon-slot-1",
            "slot_start": "15:00",
            "slot_end": "15:10",
            "is_booked": false,
            "status": "available"
          }
          // ... 17 more 10-minute slots
        ]
      }
    ]
  }
}
```

**Note:** 10-minute intervals for online consultations (longer than 5-min offline)

---

### Example 4: Multi-Clinic Doctor (Same Day, Different Clinics)

**Clinic A - Morning:**
```json
POST /api/organizations/doctor-session-slots

{
  "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
  "clinic_id": "clinic-a-uuid",
  "slot_type": "offline",
  "slot_duration": 5,
  "date": "2025-10-24",
  "sessions": [
    {
      "session_name": "Clinic A Morning",
      "start_time": "08:00",
      "end_time": "12:00",
      "max_patients": 48,
      "slot_interval_minutes": 5
    }
  ]
}
```

**Clinic B - Afternoon:**
```json
POST /api/organizations/doctor-session-slots

{
  "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
  "clinic_id": "clinic-b-uuid",
  "slot_type": "offline",
  "slot_duration": 5,
  "date": "2025-10-24",
  "sessions": [
    {
      "session_name": "Clinic B Afternoon",
      "start_time": "14:00",
      "end_time": "18:00",
      "max_patients": 48,
      "slot_interval_minutes": 5
    }
  ]
}
```

**Result:** Doctor has slots at 2 clinics on same day! ✅

---

## 🔍 List Session Slots - Full Examples

### Example 1: Get All Slots for Doctor

**Request:**
```
GET /api/organizations/doctor-session-slots?doctor_id=3fd28e6d-7f9a-4dde-8172-d14a74a9b02d
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
  "clinic_id": "",
  "date": "",
  "slot_type": "",
  "total": 3,
  "slots": [
    {
      "id": "timeslot-1",
      "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "date": "2025-10-20",
      "day_of_week": 1,
      "slot_type": "offline",
      "is_available": true,
      "sessions": [
        {
          "id": "session-1",
          "session_name": "Morning Session",
          "start_time": "09:00",
          "end_time": "12:00",
          "max_patients": 36,
          "slot_interval_minutes": 5,
          "generated_slots": 36,
          "available_slots": 34,
          "booked_slots": 2,
          "slots": [
            {
              "id": "slot-1",
              "slot_start": "09:00",
              "slot_end": "09:05",
              "is_booked": true,
              "booked_patient_id": "patient-uuid-1",
              "booked_appointment_id": "appointment-uuid-1",
              "status": "booked"
            },
            {
              "id": "slot-2",
              "slot_start": "09:05",
              "slot_end": "09:10",
              "is_booked": false,
              "status": "available"
            }
            // ... 34 more slots
          ]
        }
      ]
    },
    {
      "id": "timeslot-2",
      "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "date": "2025-10-22",
      "day_of_week": 3,
      "slot_type": "offline",
      "is_available": true,
      "sessions": [
        // Wednesday sessions
      ]
    },
    {
      "id": "timeslot-3",
      "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "date": "2025-10-23",
      "day_of_week": 4,
      "slot_type": "online",
      "is_available": true,
      "sessions": [
        // Thursday online sessions
      ]
    }
  ]
}
```

---

### Example 2: Filter by Specific Date

**Request:**
```
GET /api/organizations/doctor-session-slots?doctor_id=3fd28e6d-7f9a-4dde-8172-d14a74a9b02d&date=2025-10-20
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
  "clinic_id": "",
  "date": "2025-10-20",
  "slot_type": "",
  "total": 1,
  "slots": [
    {
      "id": "timeslot-uuid-mon",
      "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "date": "2025-10-20",
      "day_of_week": 1,
      "slot_type": "offline",
      "is_available": true,
      "sessions": [
        {
          "id": "session-morning",
          "session_name": "Morning Session",
          "start_time": "09:00",
          "end_time": "12:00",
          "max_patients": 36,
          "slot_interval_minutes": 5,
          "generated_slots": 36,
          "available_slots": 36,
          "booked_slots": 0,
          "slots": [
            /* All 36 individual slots for this session */
          ]
        }
      ]
    }
  ]
}
```

---

### Example 3: Filter by Clinic and Slot Type

**Request:**
```
GET /api/organizations/doctor-session-slots?doctor_id=3fd28e6d-7f9a-4dde-8172-d14a74a9b02d&clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&slot_type=online
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "date": "",
  "slot_type": "online",
  "total": 1,
  "slots": [
    {
      "id": "timeslot-online-thu",
      "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "date": "2025-10-23",
      "day_of_week": 4,
      "slot_type": "online",
      "is_available": true,
      "sessions": [
        {
          "id": "session-online-morning",
          "session_name": "Online Morning",
          "start_time": "10:00",
          "end_time": "13:00",
          "max_patients": 18,
          "slot_interval_minutes": 10,
          "generated_slots": 18,
          "available_slots": 15,
          "booked_slots": 3,
          "slots": [
            {
              "id": "online-slot-1",
              "slot_start": "10:00",
              "slot_end": "10:10",
              "is_booked": true,
              "booked_patient_id": "patient-uuid-1",
              "booked_appointment_id": "appt-uuid-1",
              "status": "booked"
            },
            {
              "id": "online-slot-2",
              "slot_start": "10:10",
              "slot_end": "10:20",
              "is_booked": false,
              "status": "available"
            }
            // ... 16 more slots
          ]
        },
        {
          "id": "session-online-afternoon",
          "session_name": "Online Afternoon",
          "start_time": "15:00",
          "end_time": "18:00",
          "max_patients": 18,
          "slot_interval_minutes": 10,
          "generated_slots": 18,
          "available_slots": 18,
          "booked_slots": 0,
          "slots": [
            /* All 18 afternoon slots */
          ]
        }
      ]
    }
  ]
}
```

---

### Example 4: Combine All Filters

**Request:**
```
GET /api/organizations/doctor-session-slots?doctor_id=3fd28e6d-7f9a-4dde-8172-d14a74a9b02d&clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&date=2025-10-20&slot_type=offline
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "date": "2025-10-20",
  "slot_type": "offline",
  "total": 1,
  "slots": [
    {
      "id": "timeslot-specific",
      "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "date": "2025-10-20",
      "day_of_week": 1,
      "slot_type": "offline",
      "is_available": true,
      "sessions": [
        {
          "id": "session-morning-specific",
          "session_name": "Morning Session",
          "start_time": "09:00",
          "end_time": "12:00",
          "max_patients": 36,
          "slot_interval_minutes": 5,
          "generated_slots": 36,
          "available_slots": 30,
          "booked_slots": 6,
          "slots": [
            {
              "id": "slot-09-00",
              "slot_start": "09:00",
              "slot_end": "09:05",
              "is_booked": true,
              "booked_patient_id": "patient-1-uuid",
              "booked_appointment_id": "appt-1-uuid",
              "status": "booked"
            },
            {
              "id": "slot-09-05",
              "slot_start": "09:05",
              "slot_end": "09:10",
              "is_booked": true,
              "booked_patient_id": "patient-2-uuid",
              "booked_appointment_id": "appt-2-uuid",
              "status": "booked"
            },
            {
              "id": "slot-09-10",
              "slot_start": "09:10",
              "slot_end": "09:15",
              "is_booked": false,
              "status": "available"
            },
            {
              "id": "slot-09-15",
              "slot_start": "09:15",
              "slot_end": "09:20",
              "is_booked": false,
              "status": "available"
            }
            // ... remaining 32 slots
          ]
        }
      ]
    }
  ]
}
```

---

## 🔄 Complete Workflow Examples

### Workflow 1: Doctor Creates Weekly Schedule

**Step 1: Create Monday Slots**
```json
POST /doctor-session-slots
{
  "doctor_id": "dr-smith-uuid",
  "clinic_id": "main-hospital-uuid",
  "slot_type": "offline",
  "slot_duration": 5,
  "date": "2025-10-20",
  "sessions": [
    {
      "session_name": "Monday Morning",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 36,
      "slot_interval_minutes": 5
    }
  ]
}
```

**Step 2: Create Wednesday Slots**
```json
POST /doctor-session-slots
{
  "doctor_id": "dr-smith-uuid",
  "clinic_id": "main-hospital-uuid",
  "slot_type": "offline",
  "slot_duration": 5,
  "date": "2025-10-22",
  "sessions": [
    {
      "session_name": "Wednesday Full Day",
      "start_time": "09:00",
      "end_time": "17:00",
      "max_patients": 96,
      "slot_interval_minutes": 5
    }
  ]
}
```

**Step 3: Create Friday Online Slots**
```json
POST /doctor-session-slots
{
  "doctor_id": "dr-smith-uuid",
  "clinic_id": "main-hospital-uuid",
  "slot_type": "online",
  "slot_duration": 10,
  "date": "2025-10-24",
  "sessions": [
    {
      "session_name": "Friday Telemedicine",
      "start_time": "10:00",
      "end_time": "16:00",
      "max_patients": 36,
      "slot_interval_minutes": 10
    }
  ]
}
```

**Result:** Doctor has 3 days scheduled with 168 total individual slots!

---

### Workflow 2: Patient Books Appointment

**Step 1: Patient Searches for Available Slots**
```
GET /doctor-session-slots?doctor_id=dr-smith-uuid&date=2025-10-20&slot_type=offline
```

**Response:**
```json
{
  "slots": [
    {
      "date": "2025-10-20",
      "sessions": [
        {
          "session_name": "Monday Morning",
          "available_slots": 36,
          "slots": [
            {
              "id": "slot-09-00-uuid",
              "slot_start": "09:00",
              "slot_end": "09:05",
              "is_booked": false,
              "status": "available"
            }
            // ... more available slots
          ]
        }
      ]
    }
  ]
}
```

**Step 2: Patient Books Specific Slot**
```json
POST /appointments
{
  "patient_id": "patient-uuid",
  "doctor_id": "dr-smith-uuid",
  "clinic_id": "main-hospital-uuid",
  "individual_slot_id": "slot-09-00-uuid",
  "appointment_date": "2025-10-20",
  "appointment_time": "2025-10-20 09:00:00",
  "consultation_type": "offline"
}
```

**Step 3: System Updates Slot**
```sql
UPDATE doctor_individual_slots
SET 
    is_booked = true,
    booked_patient_id = 'patient-uuid',
    booked_appointment_id = 'new-appointment-uuid',
    status = 'booked'
WHERE id = 'slot-09-00-uuid';
```

**Step 4: Verify Booking**
```
GET /doctor-session-slots?doctor_id=dr-smith-uuid&date=2025-10-20
```

**Response Shows Updated Status:**
```json
{
  "slots": [
    {
      "sessions": [
        {
          "available_slots": 35,
          "booked_slots": 1,
          "slots": [
            {
              "id": "slot-09-00-uuid",
              "is_booked": true,
              "booked_patient_id": "patient-uuid",
              "status": "booked"
            }
          ]
        }
      ]
    }
  ]
}
```

---

### Workflow 3: Multi-Clinic Doctor Schedule

**Dr. Johnson works at 3 clinics:**

**Clinic A - Monday/Wednesday:**
```json
POST /doctor-session-slots
{
  "doctor_id": "dr-johnson-uuid",
  "clinic_id": "clinic-a-uuid",
  "slot_type": "offline",
  "date": "2025-10-20",
  "sessions": [
    {
      "session_name": "Clinic A Monday",
      "start_time": "09:00",
      "end_time": "13:00",
      "max_patients": 48,
      "slot_interval_minutes": 5
    }
  ]
}
```

**Clinic B - Tuesday/Thursday:**
```json
POST /doctor-session-slots
{
  "doctor_id": "dr-johnson-uuid",
  "clinic_id": "clinic-b-uuid",
  "slot_type": "offline",
  "date": "2025-10-21",
  "sessions": [
    {
      "session_name": "Clinic B Tuesday",
      "start_time": "14:00",
      "end_time": "18:00",
      "max_patients": 48,
      "slot_interval_minutes": 5
    }
  ]
}
```

**Clinic C - Friday:**
```json
POST /doctor-session-slots
{
  "doctor_id": "dr-johnson-uuid",
  "clinic_id": "clinic-c-uuid",
  "slot_type": "online",
  "date": "2025-10-24",
  "sessions": [
    {
      "session_name": "Clinic C Online",
      "start_time": "10:00",
      "end_time": "14:00",
      "max_patients": 24,
      "slot_interval_minutes": 10
    }
  ]
}
```

**View All Clinics:**
```
GET /doctor-session-slots?doctor_id=dr-johnson-uuid
```

**Response:**
```json
{
  "total": 3,
  "slots": [
    {
      "clinic_id": "clinic-a-uuid",
      "date": "2025-10-20",
      "slot_type": "offline"
      // Clinic A details
    },
    {
      "clinic_id": "clinic-b-uuid",
      "date": "2025-10-21",
      "slot_type": "offline"
      // Clinic B details
    },
    {
      "clinic_id": "clinic-c-uuid",
      "date": "2025-10-24",
      "slot_type": "online"
      // Clinic C details
    }
  ]
}
```

---

## ❌ Error Handling

### Error 1: Missing Required Field

**Request:**
```json
POST /doctor-session-slots
{
  "doctor_id": "dr-uuid",
  "clinic_id": "clinic-uuid",
  "slot_type": "offline"
  // Missing: date, sessions
}
```

**Response (400 Bad Request):**
```json
{
  "error": "Invalid input data",
  "details": "Key: 'CreateDoctorSessionSlotsInput.Date' Error:Field validation for 'Date' failed on the 'required' tag"
}
```

---

### Error 2: Invalid Slot Type

**Request:**
```json
POST /doctor-session-slots
{
  "doctor_id": "dr-uuid",
  "clinic_id": "clinic-uuid",
  "slot_type": "hybrid",
  "date": "2025-10-20",
  "sessions": [...]
}
```

**Response (400 Bad Request):**
```json
{
  "error": "Invalid input data",
  "details": "Key: 'CreateDoctorSessionSlotsInput.SlotType' Error:Field validation for 'SlotType' failed on the 'oneof' tag"
}
```

---

### Error 3: Overlapping Sessions

**Request:**
```json
POST /doctor-session-slots
{
  "doctor_id": "dr-uuid",
  "clinic_id": "clinic-uuid",
  "slot_type": "offline",
  "date": "2025-10-20",
  "sessions": [
    {
      "session_name": "Morning",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 36,
      "slot_interval_minutes": 5
    },
    {
      "session_name": "Late Morning",
      "start_time": "11:00",
      "end_time": "13:00",
      "max_patients": 24,
      "slot_interval_minutes": 5
    }
  ]
}
```

**Response (400 Bad Request):**
```json
{
  "error": "Overlapping sessions",
  "message": "Session 'Late Morning' overlaps with session 'Morning'"
}
```

---

### Error 4: Duplicate Slots for Date

**Request:**
```json
POST /doctor-session-slots
{
  "doctor_id": "dr-uuid",
  "clinic_id": "clinic-uuid",
  "slot_type": "offline",
  "date": "2025-10-20",
  "sessions": [...]
}
```

**Response (409 Conflict):**
```json
{
  "error": "Duplicate slots",
  "message": "Time slots already exist for this doctor on 2025-10-20"
}
```

---

### Error 5: Doctor Not Linked to Clinic

**Request:**
```json
POST /doctor-session-slots
{
  "doctor_id": "dr-uuid",
  "clinic_id": "unlinked-clinic-uuid",
  "slot_type": "offline",
  "date": "2025-10-20",
  "sessions": [...]
}
```

**Response (403 Forbidden):**
```json
{
  "error": "Doctor is not linked to this clinic",
  "message": "The specified doctor is not associated with this clinic"
}
```

---

### Error 6: Invalid Filter (List API)

**Request:**
```
GET /doctor-session-slots?doctor_id=dr-uuid&slot_type=hybrid
```

**Response (400 Bad Request):**
```json
{
  "error": "Invalid slot_type. Must be 'offline' or 'online'"
}
```

---

## 🛠️ Integration Guide

### JavaScript/TypeScript Integration

```typescript
// API Client
class DoctorSessionSlotsAPI {
  private baseUrl = 'http://localhost:8081/organizations';
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  // Create session-based slots
  async createSlots(data: CreateSlotsRequest): Promise<CreateSlotsResponse> {
    const response = await fetch(`${this.baseUrl}/doctor-session-slots`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to create slots');
    }

    return response.json();
  }

  // List slots with filters
  async listSlots(filters: {
    doctor_id: string;
    clinic_id?: string;
    date?: string;
    slot_type?: 'offline' | 'online';
  }): Promise<ListSlotsResponse> {
    const params = new URLSearchParams();
    params.append('doctor_id', filters.doctor_id);
    if (filters.clinic_id) params.append('clinic_id', filters.clinic_id);
    if (filters.date) params.append('date', filters.date);
    if (filters.slot_type) params.append('slot_type', filters.slot_type);

    const response = await fetch(
      `${this.baseUrl}/doctor-session-slots?${params}`,
      {
        headers: {
          'Authorization': `Bearer ${this.token}`
        }
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to list slots');
    }

    return response.json();
  }

  // Get available slots only
  async getAvailableSlots(
    doctor_id: string,
    date: string,
    slot_type?: 'offline' | 'online'
  ) {
    const data = await this.listSlots({ doctor_id, date, slot_type });
    
    // Filter for available slots
    return data.slots.map(daySlot => ({
      ...daySlot,
      sessions: daySlot.sessions.map(session => ({
        ...session,
        slots: session.slots.filter(slot => !slot.is_booked && slot.status === 'available')
      }))
    }));
  }
}

// Usage Example
const api = new DoctorSessionSlotsAPI('your-jwt-token');

// Create slots
const createResult = await api.createSlots({
  doctor_id: 'dr-uuid',
  clinic_id: 'clinic-uuid',
  slot_type: 'offline',
  slot_duration: 5,
  date: '2025-10-20',
  sessions: [
    {
      session_name: 'Morning Session',
      start_time: '09:00',
      end_time: '12:00',
      max_patients: 36,
      slot_interval_minutes: 5
    }
  ]
});

// List available slots
const availableSlots = await api.getAvailableSlots(
  'dr-uuid',
  '2025-10-20',
  'offline'
);
```

---

### React Component Example

```tsx
import React, { useState, useEffect } from 'react';

const SlotBooking: React.FC<{ doctorId: string }> = ({ doctorId }) => {
  const [selectedDate, setSelectedDate] = useState('2025-10-20');
  const [slots, setSlots] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadSlots();
  }, [selectedDate]);

  const loadSlots = async () => {
    setLoading(true);
    try {
      const response = await fetch(
        `/api/organizations/doctor-session-slots?doctor_id=${doctorId}&date=${selectedDate}&slot_type=offline`
      );
      const data = await response.json();
      setSlots(data.slots);
    } catch (error) {
      console.error('Failed to load slots:', error);
    } finally {
      setLoading(false);
    }
  };

  const bookSlot = async (slotId: string) => {
    try {
      await fetch('/api/appointments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          patient_id: 'current-patient-id',
          doctor_id: doctorId,
          individual_slot_id: slotId,
          appointment_date: selectedDate,
          consultation_type: 'offline'
        })
      });
      alert('Slot booked successfully!');
      loadSlots(); // Refresh
    } catch (error) {
      console.error('Booking failed:', error);
    }
  };

  return (
    <div>
      <h2>Book Appointment</h2>
      
      <input
        type="date"
        value={selectedDate}
        onChange={(e) => setSelectedDate(e.target.value)}
      />

      {loading ? (
        <p>Loading...</p>
      ) : (
        slots.map(daySlot =>
          daySlot.sessions.map(session => (
            <div key={session.id} className="session">
              <h3>{session.session_name}</h3>
              <p>
                Available: {session.available_slots}/{session.generated_slots}
              </p>
              
              <div className="slots-grid">
                {session.slots
                  .filter(slot => !slot.is_booked)
                  .map(slot => (
                    <button
                      key={slot.id}
                      onClick={() => bookSlot(slot.id)}
                      className="slot-button"
                    >
                      {slot.slot_start}
                    </button>
                  ))}
              </div>
            </div>
          ))
        )
      )}
    </div>
  );
};
```

---

## 📋 Quick Reference

### Create Slots - Required Fields
```json
{
  "doctor_id": "UUID (required)",
  "clinic_id": "UUID (required)",
  "slot_type": "offline|online (required)",
  "slot_duration": "integer (required)",
  "date": "YYYY-MM-DD (required)",
  "sessions": [
    {
      "session_name": "string (required)",
      "start_time": "HH:MM (required)",
      "end_time": "HH:MM (required)",
      "max_patients": "integer (required)",
      "slot_interval_minutes": "integer (required)"
    }
  ]
}
```

### List Slots - Query Parameters
```
?doctor_id=UUID (required)
&clinic_id=UUID (optional)
&date=YYYY-MM-DD (optional)
&slot_type=offline|online (optional)
```

### Response Status Codes
- `201` - Created successfully
- `200` - Retrieved successfully
- `400` - Bad request (validation error)
- `403` - Forbidden (not linked to clinic)
- `404` - Not found (doctor/clinic doesn't exist)
- `409` - Conflict (duplicate slots)
- `500` - Internal server error

---

## ✅ Summary

| Feature | Status | Description |
|---------|--------|-------------|
| Auto-generate slots | ✅ Working | Creates individual slots from intervals |
| Auto-calculate day_of_week | ✅ Working | ISO 8601 format (1-7) |
| Multi-clinic support | ✅ Working | clinic_id in all tables |
| Multiple sessions | ✅ Working | Morning, afternoon, evening, etc. |
| Online/Offline slots | ✅ Working | Separate consultation types |
| Booking tracking | ✅ Working | Real-time status updates |
| Filter by clinic | ✅ Working | Query parameter |
| Filter by date | ✅ Working | Query parameter |
| Filter by slot_type | ✅ Working | Query parameter |
| Overlap prevention | ✅ Working | Validates session times |
| Duplicate prevention | ✅ Working | Prevents same doctor/date |

---

**Documentation Version:** 1.0  
**Last Updated:** October 15, 2025  
**Status:** ✅ Complete & Production Ready

