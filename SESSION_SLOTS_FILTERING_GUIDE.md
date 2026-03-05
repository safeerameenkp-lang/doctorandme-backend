# Session-Based Slots - Filtering Guide

## 🎯 Overview

The `ListDoctorSessionSlots` API supports **multiple filters** to help you get exactly the slots you need.

---

## 📋 Available Filters

### ✅ Required Filter
- **`doctor_id`** (UUID) - Doctor's unique ID

### ✅ Optional Filters
- **`clinic_id`** (UUID) - Filter by specific clinic
- **`date`** (YYYY-MM-DD) - Filter by specific date
- **`slot_type`** (offline/online) - Filter by consultation type

---

## 🔍 Filter Examples

### Example 1: All Slots for a Doctor
```
GET /doctor-session-slots?doctor_id=dr-uuid
```

**Response:**
```json
{
  "doctor_id": "dr-uuid",
  "clinic_id": "",
  "date": "",
  "slot_type": "",
  "total": 15,
  "slots": [
    /* All slots across all clinics, all dates, all types */
  ]
}
```

---

### Example 2: Filter by Clinic
```
GET /doctor-session-slots?doctor_id=dr-uuid&clinic_id=clinic-a-uuid
```

**Use Case:** Show only slots at Clinic A

**Response:**
```json
{
  "doctor_id": "dr-uuid",
  "clinic_id": "clinic-a-uuid",
  "date": "",
  "slot_type": "",
  "total": 5,
  "slots": [
    /* Only Clinic A slots */
  ]
}
```

---

### Example 3: Filter by Date
```
GET /doctor-session-slots?doctor_id=dr-uuid&date=2025-10-20
```

**Use Case:** Show slots for a specific day

**Response:**
```json
{
  "doctor_id": "dr-uuid",
  "clinic_id": "",
  "date": "2025-10-20",
  "slot_type": "",
  "total": 2,
  "slots": [
    {
      "date": "2025-10-20",
      "sessions": [/* Morning & Afternoon */]
    }
  ]
}
```

---

### Example 4: Filter by Slot Type (NEW! ✨)
```
GET /doctor-session-slots?doctor_id=dr-uuid&slot_type=offline
```

**Use Case:** Show only in-person consultation slots

**Response:**
```json
{
  "doctor_id": "dr-uuid",
  "clinic_id": "",
  "date": "",
  "slot_type": "offline",
  "total": 8,
  "slots": [
    {
      "slot_type": "offline",
      "sessions": [/* Only offline sessions */]
    }
  ]
}
```

---

### Example 5: Combine Multiple Filters
```
GET /doctor-session-slots?doctor_id=dr-uuid&clinic_id=clinic-a-uuid&date=2025-10-20&slot_type=offline
```

**Use Case:** Show only offline slots at Clinic A on October 20, 2025

**Response:**
```json
{
  "doctor_id": "dr-uuid",
  "clinic_id": "clinic-a-uuid",
  "date": "2025-10-20",
  "slot_type": "offline",
  "total": 1,
  "slots": [
    {
      "clinic_id": "clinic-a-uuid",
      "date": "2025-10-20",
      "slot_type": "offline",
      "sessions": [
        {
          "session_name": "Morning Clinic A",
          "slots": [/* Offline slots only */]
        }
      ]
    }
  ]
}
```

---

## 🎨 UI Integration Examples

### Scenario 1: Patient Booking Page

**User selects:**
- Clinic: "Main Hospital" (clinic-a-uuid)
- Date: October 20, 2025
- Type: In-person (offline)

**API Call:**
```javascript
const response = await fetch(
  `/doctor-session-slots?` +
  `doctor_id=${doctorId}&` +
  `clinic_id=${clinicAId}&` +
  `date=2025-10-20&` +
  `slot_type=offline`
);

// Shows only offline slots at Main Hospital on Oct 20
```

---

### Scenario 2: Doctor's Schedule View

**Show all clinics for today:**
```javascript
const today = new Date().toISOString().split('T')[0];
const response = await fetch(
  `/doctor-session-slots?` +
  `doctor_id=${doctorId}&` +
  `date=${today}`
);

// Shows all clinics where doctor has slots today
```

---

### Scenario 3: Online Consultation Filter

**Show only online slots:**
```javascript
const response = await fetch(
  `/doctor-session-slots?` +
  `doctor_id=${doctorId}&` +
  `slot_type=online`
);

// Shows only telemedicine slots across all clinics
```

---

## 📊 Real-World Use Cases

### Use Case 1: Multi-Clinic Doctor

**Dr. Smith works at:**
- Clinic A (Offline): Monday, Wednesday
- Clinic B (Online): Tuesday, Thursday
- Clinic C (Offline): Friday

**Query for Friday offline slots:**
```
GET /doctor-session-slots?
    doctor_id=dr-smith&
    date=2025-10-24&
    slot_type=offline
```

**Result:** Shows Clinic C slots only ✅

---

### Use Case 2: Patient Wants Telemedicine

**Patient prefers online consultation:**

```javascript
// Filter for online slots
GET /doctor-session-slots?
    doctor_id=dr-jones&
    slot_type=online&
    date=2025-10-20

// UI shows only online/video consultation slots
```

---

### Use Case 3: Clinic Admin Dashboard

**Clinic A admin wants today's schedule:**

```javascript
const today = '2025-10-15';
GET /doctor-session-slots?
    doctor_id=dr-smith&
    clinic_id=clinic-a&
    date=${today}

// Shows only Clinic A's schedule for today
```

---

## 🚫 Validation & Errors

### Error 1: Missing doctor_id
```
GET /doctor-session-slots?clinic_id=xxx
```

**Response (400):**
```json
{
  "error": "doctor_id is required"
}
```

---

### Error 2: Invalid slot_type
```
GET /doctor-session-slots?doctor_id=xxx&slot_type=hybrid
```

**Response (400):**
```json
{
  "error": "Invalid slot_type. Must be 'offline' or 'online'"
}
```

---

### Error 3: Invalid doctor_id format
```
GET /doctor-session-slots?doctor_id=invalid-uuid
```

**Response (400):**
```json
{
  "error": "Invalid doctor_id format"
}
```

---

## 💡 Best Practices

### 1. Always Include doctor_id
```javascript
// ✅ Good
GET /doctor-session-slots?doctor_id=xxx&clinic_id=yyy

// ❌ Bad - Will fail!
GET /doctor-session-slots?clinic_id=yyy
```

---

### 2. Filter by slot_type for Patient UX
```javascript
// Let patients choose consultation type
if (userWantsOnline) {
  queryParams.slot_type = 'online';
} else {
  queryParams.slot_type = 'offline';
}
```

---

### 3. Combine Filters for Better Performance
```javascript
// ✅ Specific query = Faster response
GET /doctor-session-slots?
    doctor_id=xxx&
    clinic_id=yyy&
    date=2025-10-20&
    slot_type=offline

// ❌ Too broad = More data to process
GET /doctor-session-slots?doctor_id=xxx
```

---

## 📋 Complete Filter Reference

| Filter | Type | Required | Values | Example |
|--------|------|----------|--------|---------|
| `doctor_id` | UUID | ✅ Yes | Valid UUID | `3fd28e6d-...` |
| `clinic_id` | UUID | ❌ No | Valid UUID | `7a6c1211-...` |
| `date` | String | ❌ No | YYYY-MM-DD | `2025-10-20` |
| `slot_type` | String | ❌ No | `offline` \| `online` | `offline` |

---

## 🎯 Summary

| Filter Combination | Use Case |
|-------------------|----------|
| `doctor_id` only | All slots for doctor |
| `+ clinic_id` | Specific clinic |
| `+ date` | Specific date |
| `+ slot_type` | Specific type (offline/online) |
| All 4 filters | Most specific query |

---

## ✅ Status

| Feature | Status |
|---------|--------|
| doctor_id filter | ✅ Required |
| clinic_id filter | ✅ Optional |
| date filter | ✅ Optional |
| slot_type filter | ✅ **NEW!** Optional |
| Validation | ✅ Complete |
| Error messages | ✅ Clear |

---

**Last Updated:** October 15, 2025  
**Status:** ✅ All Filters Working

