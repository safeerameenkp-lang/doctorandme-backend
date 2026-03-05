# Doctor Time Slots API - Enhanced Guide

## 🎯 Overview

The Doctor Time Slots API now supports **two types of slots**:

1. **Date-Specific Slots** - One-time slots for a specific date (e.g., "2024-10-20")
2. **Recurring Weekly Slots** - Repeating slots for a specific day of the week (e.g., "Every Monday")

## 📋 API Endpoint

```
POST /doctor-time-slots
```

## 🔑 Key Rules

- You must provide **EITHER** `date` **OR** `day_of_week`, but **NOT both**
- `day_of_week` values: 0=Sunday, 1=Monday, 2=Tuesday, 3=Wednesday, 4=Thursday, 5=Friday, 6=Saturday
- All other fields remain the same

---

## 📝 Usage Examples

### Example 1: Create Date-Specific Slots (One-Time)

Use this for **special dates**, **holiday schedules**, or **one-time availability**.

```json
POST /doctor-time-slots
{
  "doctor_id": "123e4567-e89b-12d3-a456-426614174000",
  "clinic_id": "123e4567-e89b-12d3-a456-426614174001",
  "slot_type": "offline",
  "date": "2024-10-20",           // ✅ Use date for specific date
  "slots": [
    {
      "start_time": "09:00",
      "end_time": "09:30",
      "max_patients": 1,
      "notes": "Special consultation day"
    },
    {
      "start_time": "09:30",
      "end_time": "10:00",
      "max_patients": 1
    }
  ]
}
```

**Response:**
```json
{
  "message": "Slot creation completed. 2 created, 0 failed",
  "created_slots": [
    {
      "id": "slot-id-1",
      "doctor_id": "123e4567-e89b-12d3-a456-426614174000",
      "clinic_id": "123e4567-e89b-12d3-a456-426614174001",
      "date": "2024-10-20",        // ✅ Specific date
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "09:30",
      "max_patients": 1,
      "booked_patients": 0,
      "available_spots": 1,
      "is_available": true,
      "status": "available",
      "notes": "Special consultation day",
      "is_active": true,
      "created_at": "2024-10-15T10:30:00Z",
      "updated_at": "2024-10-15T10:30:00Z"
    }
  ],
  "total_created": 2,
  "total_failed": 0
}
```

---

### Example 2: Create Recurring Weekly Slots

Use this for **regular weekly schedules** that repeat every week.

```json
POST /doctor-time-slots
{
  "doctor_id": "123e4567-e89b-12d3-a456-426614174000",
  "clinic_id": "123e4567-e89b-12d3-a456-426614174001",
  "slot_type": "online",
  "day_of_week": 1,               // ✅ 1 = Monday (recurring every Monday)
  "slots": [
    {
      "start_time": "14:00",
      "end_time": "14:30",
      "max_patients": 1,
      "notes": "Regular Monday online consultation"
    },
    {
      "start_time": "14:30",
      "end_time": "15:00",
      "max_patients": 1
    }
  ]
}
```

**Response:**
```json
{
  "message": "Slot creation completed. 2 created, 0 failed",
  "created_slots": [
    {
      "id": "slot-id-1",
      "doctor_id": "123e4567-e89b-12d3-a456-426614174000",
      "clinic_id": "123e4567-e89b-12d3-a456-426614174001",
      "date": "Every Monday",       // ✅ Human-readable format
      "day_of_week": 1,             // ✅ Also includes numeric value
      "slot_type": "online",
      "start_time": "14:00",
      "end_time": "14:30",
      "max_patients": 1,
      "booked_patients": 0,
      "available_spots": 1,
      "is_available": true,
      "status": "available",
      "notes": "Regular Monday online consultation",
      "is_active": true,
      "created_at": "2024-10-15T10:30:00Z",
      "updated_at": "2024-10-15T10:30:00Z"
    }
  ],
  "total_created": 2,
  "total_failed": 0
}
```

---

## 🗓️ Day of Week Reference

| Value | Day       | Use Case Example                    |
|-------|-----------|-------------------------------------|
| 0     | Sunday    | Weekend consultations               |
| 1     | Monday    | Start of week regular hours         |
| 2     | Tuesday   | Regular weekday hours               |
| 3     | Wednesday | Mid-week hours                      |
| 4     | Thursday  | Regular weekday hours               |
| 5     | Friday    | End of week hours                   |
| 6     | Saturday  | Weekend consultations               |

---

## 🚫 Error Examples

### ❌ Error: Both date AND day_of_week provided

```json
{
  "doctor_id": "123e4567-e89b-12d3-a456-426614174000",
  "clinic_id": "123e4567-e89b-12d3-a456-426614174001",
  "slot_type": "offline",
  "date": "2024-10-20",
  "day_of_week": 1,              // ❌ Can't have both!
  "slots": [...]
}
```

**Error Response (400):**
```json
{
  "error": "Provide either 'date' for specific date slots OR 'day_of_week' for recurring weekly slots, but not both"
}
```

---

### ❌ Error: Neither date NOR day_of_week provided

```json
{
  "doctor_id": "123e4567-e89b-12d3-a456-426614174000",
  "clinic_id": "123e4567-e89b-12d3-a456-426614174001",
  "slot_type": "offline",
  // ❌ Missing both date and day_of_week!
  "slots": [...]
}
```

**Error Response (400):**
```json
{
  "error": "Provide either 'date' for specific date slots OR 'day_of_week' for recurring weekly slots, but not both"
}
```

---

### ❌ Error: Invalid day_of_week value

```json
{
  "doctor_id": "123e4567-e89b-12d3-a456-426614174000",
  "clinic_id": "123e4567-e89b-12d3-a456-426614174001",
  "slot_type": "offline",
  "day_of_week": 7,              // ❌ Must be 0-6!
  "slots": [...]
}
```

**Error Response (400):**
```json
{
  "error": "Invalid day_of_week. Must be between 0 (Sunday) and 6 (Saturday)"
}
```

---

## 📊 Complete Real-World Example

### Scenario: Dr. Smith's Weekly Schedule at City Clinic

**Monday - Regular offline consultation:**
```json
POST /doctor-time-slots
{
  "doctor_id": "dr-smith-uuid",
  "clinic_id": "city-clinic-uuid",
  "slot_type": "offline",
  "day_of_week": 1,              // Every Monday
  "slots": [
    { "start_time": "09:00", "end_time": "09:30", "max_patients": 1 },
    { "start_time": "09:30", "end_time": "10:00", "max_patients": 1 },
    { "start_time": "10:00", "end_time": "10:30", "max_patients": 1 }
  ]
}
```

**Special date - October 20, 2024 (Holiday schedule):**
```json
POST /doctor-time-slots
{
  "doctor_id": "dr-smith-uuid",
  "clinic_id": "city-clinic-uuid",
  "slot_type": "online",
  "date": "2024-10-20",          // One-time special schedule
  "slots": [
    { 
      "start_time": "11:00", 
      "end_time": "11:30", 
      "max_patients": 1,
      "notes": "Holiday emergency consultation only"
    }
  ]
}
```

---

## 🔍 Querying Slots

All existing GET endpoints work the same way. The response will include both `date` and `day_of_week` fields:

```
GET /doctor-time-slots?doctor_id=xxx&clinic_id=xxx
```

**Response includes both types:**
```json
{
  "slots": [
    {
      "id": "slot-1",
      "date": "2024-10-20",        // Date-specific slot
      "slot_type": "offline",
      ...
    },
    {
      "id": "slot-2",
      "date": "Every Monday",      // Recurring slot (human-readable)
      "day_of_week": 1,            // Recurring slot (numeric)
      "slot_type": "online",
      ...
    }
  ],
  "total": 2
}
```

---

## ✅ Migration Applied

The database has been updated to support this feature:
- ✅ `day_of_week` column is now nullable
- ✅ Constraint ensures either `date` OR `day_of_week` is set (but not both)
- ✅ Migration 012 applied successfully

---

## 🎯 Use Case Summary

| Slot Type | Use `date` | Use `day_of_week` | Example |
|-----------|------------|-------------------|---------|
| One-time appointment | ✅ | ❌ | Holiday schedule, special events |
| Regular weekly schedule | ❌ | ✅ | Every Monday morning clinic |
| Emergency override | ✅ | ❌ | Specific date availability change |
| Permanent schedule | ❌ | ✅ | Doctor's weekly routine |

---

## 📝 Quick Reference

### Required Fields (All Requests)
- `doctor_id` (UUID)
- `clinic_id` (UUID)
- `slot_type` ("offline" or "online")
- **EITHER** `date` (YYYY-MM-DD) **OR** `day_of_week` (0-6)
- `slots` (array of time slot definitions)

### Optional Fields (Per Slot)
- `max_patients` (default: 1)
- `notes` (string)

---

**Status**: ✅ Fully Implemented and Tested
**Last Updated**: October 15, 2025

