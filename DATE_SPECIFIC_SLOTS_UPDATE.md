# Date-Specific Slots API - Updated

## ✅ Changes Made

### **1. Database Schema**
- ✅ Added `slot_id` column to `appointments` table (Migration 011)
- ✅ Table already has `specific_date` column (from earlier migration)
- ✅ Both `day_of_week` and `specific_date` columns exist

### **2. API Structure Changed**

#### **Previous (Weekly Recurring):**
```json
{
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "slot_type": "offline",
  "slots": [
    {
      "day_of_week": 1,  // ❌ Removed
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10
    }
  ]
}
```

#### **New (Date-Specific):**
```json
{
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "slot_type": "offline",
  "slots": [
    {
      "date": "2024-10-15",  // ✅ Now required - specific date
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10
    }
  ]
}
```

---

## 📌 Updated Create Slots API

### **Endpoint**
```
POST /api/organizations/doctor-time-slots
```

### **Request Body**
```json
{
  "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "slot_type": "offline",
  "slots": [
    {
      "date": "2024-10-15",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10,
      "notes": "Morning shift - Tuesday"
    },
    {
      "date": "2024-10-15",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 10,
      "notes": "Afternoon shift - Tuesday"
    },
    {
      "date": "2024-10-17",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10,
      "notes": "Morning shift - Thursday"
    },
    {
      "date": "2024-10-17",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 10,
      "notes": "Afternoon shift - Thursday"
    }
  ]
}
```

### **How It Works**
1. ✅ User selects a **specific date** in the UI (e.g., October 15, 2024)
2. ✅ System **automatically calculates** `day_of_week` from the date (Oct 15 = Tuesday = 2)
3. ✅ Saves both `specific_date` and calculated `day_of_week` to database
4. ✅ Can add **multiple slots for multiple dates** in one API call

### **Example: Creating Slots for Multiple Dates**
```json
{
  "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "slot_type": "offline",
  "slots": [
    // Tuesday, October 15
    {"date": "2024-10-15", "start_time": "09:00", "end_time": "12:00", "max_patients": 10},
    {"date": "2024-10-15", "start_time": "14:00", "end_time": "17:00", "max_patients": 10},
    
    // Wednesday, October 16
    {"date": "2024-10-16", "start_time": "09:00", "end_time": "12:00", "max_patients": 10},
    
    // Thursday, October 17
    {"date": "2024-10-17", "start_time": "09:00", "end_time": "12:00", "max_patients": 10},
    {"date": "2024-10-17", "start_time": "14:00", "end_time": "17:00", "max_patients": 10},
    
    // Friday, October 18
    {"date": "2024-10-18", "start_time": "09:00", "end_time": "17:00", "max_patients": 15}
  ]
}
```

---

## 📋 List Slots API (Updated)

### **Endpoint**
```
GET /api/organizations/doctor-time-slots/list/:doctor_id/:clinic_id/:slot_type?date=YYYY-MM-DD
```

### **Example Request**
```
GET http://localhost:8081/api/organizations/doctor-time-slots/list/3fd28e6d-7f9a-4dde-8172-d14a74a9b02d/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/offline?date=2024-10-15
```

### **Response**
```json
{
  "slots": [
    {
      "id": "uuid",
      "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "day_of_week": 2,
      "day_name": "Tuesday",
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10,
      "booked_patients": 3,
      "available_spots": 7,
      "is_available": true,
      "status": "available",
      "notes": "Morning shift - Tuesday",
      "is_active": true,
      "created_at": "2024-10-13T10:00:00Z",
      "updated_at": "2024-10-13T10:00:00Z"
    }
  ],
  "total": 1,
  "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "slot_type": "offline",
  "date": "2024-10-15"
}
```

---

## 🎯 Benefits of Date-Specific Slots

| Feature | Before (Weekly) | After (Date-Specific) |
|---------|----------------|---------------------|
| Slot Definition | Every Monday forever | October 15, 2024 only |
| Flexibility | Same schedule weekly | Different times each day |
| Holiday Management | Manual override needed | Simply don't create slots |
| Special Events | Hard to manage | Easy - just different dates |
| Patient Booking | Recurring pattern | Specific date selection |

---

## 💡 UI Workflow

### **Admin Creates Slots**
1. ✅ Select Doctor from dropdown
2. ✅ Select Clinic from dropdown
3. ✅ Select Slot Type (offline/online)
4. ✅ **Pick a specific date** (e.g., October 15, 2024)
5. ✅ Add multiple time slots for that date
6. ✅ Repeat for other dates as needed

### **Patient Books Appointment**
1. ✅ Select Doctor
2. ✅ Select Consultation Type (offline/online)
3. ✅ **Pick a date** from calendar
4. ✅ See available time slots for that specific date
5. ✅ Shows "Booking Full" if max_patients reached
6. ✅ Shows "Unavailable" if no slots exist for that date

---

## 🔧 cURL Examples

### **Create Slots for October 15, 2024**
```bash
curl -X POST http://localhost:8081/api/organizations/doctor-time-slots \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "doctor_id": "3fd28e6d-7f9a-4dde-8172-d14a74a9b02d",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "slot_type": "offline",
    "slots": [
      {
        "date": "2024-10-15",
        "start_time": "09:00",
        "end_time": "12:00",
        "max_patients": 10,
        "notes": "Morning shift"
      },
      {
        "date": "2024-10-15",
        "start_time": "14:00",
        "end_time": "17:00",
        "max_patients": 10,
        "notes": "Afternoon shift"
      }
    ]
  }'
```

### **List Slots for October 15, 2024**
```bash
curl -X GET "http://localhost:8081/api/organizations/doctor-time-slots/list/3fd28e6d-7f9a-4dde-8172-d14a74a9b02d/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/offline?date=2024-10-15" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 📊 Database Structure

The `doctor_time_slots` table now stores:
```sql
{
  id: UUID,
  doctor_id: UUID,
  clinic_id: UUID,
  day_of_week: INT,        -- Auto-calculated from specific_date
  specific_date: DATE,     -- The actual date (e.g., 2024-10-15)
  slot_type: VARCHAR,      -- offline/online
  start_time: TIME,
  end_time: TIME,
  max_patients: INT,
  notes: TEXT,
  is_active: BOOLEAN
}
```

---

## ✨ Key Improvements

1. ✅ **Date-based scheduling** - No more weekly recurring slots
2. ✅ **Flexible timing** - Different times for different dates
3. ✅ **Easy holiday management** - Just don't create slots
4. ✅ **Auto day calculation** - System calculates day_of_week from date
5. ✅ **Bulk creation** - Add multiple slots for multiple dates at once
6. ✅ **Real-time availability** - Shows booked/available status
7. ✅ **Slot_id tracking** - Appointments now linked to specific slots

---

## 🚀 Ready to Test

The service has been restarted. You can now test with:

**Create slots:**
```
POST http://localhost:8081/api/organizations/doctor-time-slots
```

**List slots:**
```
GET http://localhost:8081/api/organizations/doctor-time-slots/list/3fd28e6d-7f9a-4dde-8172-d14a74a9b02d/7a6c1211-c029-4923-a1a6-fe3dfe48bdf2/offline?date=2024-10-15
```

