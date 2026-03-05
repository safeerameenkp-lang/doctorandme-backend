# Doctor Time Slots - Quick Reference Guide

## 🚀 What's New?

The Doctor Time Slots API now supports **TWO types of slots**:

### 1️⃣ **Recurring Weekly Slots** (day_of_week)
- Repeat every week on specific days
- Example: Every Monday 9 AM - 5 PM

### 2️⃣ **Specific Date Slots** (specific_date)  
- One-time slots for particular dates
- Example: December 25, 2024 from 2 PM - 5 PM

---

## 📝 Quick Examples

### Create Recurring Slot (Weekly)
```json
POST /api/v1/time-slots
{
  "doctor_id": "xxx",
  "clinic_id": "yyy",
  "day_of_week": 1,        // Monday
  "slot_type": "offline",
  "start_time": "09:00",
  "end_time": "17:00"
}
```

### Create Specific Date Slot (One-time)
```json
POST /api/v1/time-slots
{
  "doctor_id": "xxx",
  "clinic_id": "yyy",
  "specific_date": "2024-12-25",  // Christmas Day
  "slot_type": "online",
  "start_time": "14:00",
  "end_time": "17:00"
}
```

---

## 🎯 UI Implementation for Date Picker

### User Flow:
1. **User selects a date** from date picker
2. **Fetch existing slots** for that date:
   ```
   GET /api/v1/time-slots?doctor_id=xxx&specific_date=2024-12-25
   ```
3. **Show "Add Slot" button** if no slots exist
4. **User adds slot** with start time, end time, type, etc.
5. **Save the slot**:
   ```json
   POST /api/v1/time-slots
   {
     "specific_date": "2024-12-25",
     "start_time": "09:00",
     "end_time": "12:00",
     "slot_type": "offline"
   }
   ```

---

## 📋 API Endpoints Summary

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/v1/time-slots` | Create slot (recurring or specific) |
| GET | `/api/v1/time-slots` | List slots with filters |
| GET | `/api/v1/doctors/:id/available-slots` | Get available slots |
| PUT | `/api/v1/time-slots/:id` | Update slot |
| DELETE | `/api/v1/time-slots/:id` | Delete slot |

---

## 🔍 Query Filters

### List Time Slots
```
GET /api/v1/time-slots?doctor_id=xxx&specific_date=2024-12-25
GET /api/v1/time-slots?doctor_id=xxx&day_of_week=1
GET /api/v1/time-slots?clinic_id=yyy&slot_type=online
```

### Get Available Slots
```
GET /api/v1/doctors/xxx/available-slots?specific_date=2024-12-25
GET /api/v1/doctors/xxx/available-slots?day_of_week=1
GET /api/v1/doctors/xxx/available-slots?slot_type=online
```

---

## ⚠️ Important Rules

1. **Must choose ONE:**
   - Either `day_of_week` (for recurring)
   - OR `specific_date` (for one-time)
   - ❌ Cannot use both!

2. **Day of Week Values:**
   - 0 = Sunday
   - 1 = Monday
   - 2 = Tuesday
   - 3 = Wednesday
   - 4 = Thursday
   - 5 = Friday
   - 6 = Saturday

3. **Date Format:**
   - Use `YYYY-MM-DD` format
   - Example: `2024-12-25`

4. **Time Format:**
   - Use 24-hour `HH:MM` format
   - Example: `09:00`, `17:30`

5. **Slot Types:**
   - `offline` - In-person consultation
   - `online` - Telemedicine/Video consultation

---

## 🔄 Migration Required

Run this migration file to enable specific date support:
```bash
migrations/009_add_specific_date_to_time_slots.sql
```

This adds:
- `specific_date` column
- Constraints to ensure either day_of_week OR specific_date
- Updated overlap checking for both types
- Performance indexes

---

## ✅ Response Examples

### Recurring Slot Response
```json
{
  "id": "uuid",
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "day_of_week": 1,
  "day_name": "Monday",
  "slot_type": "offline",
  "start_time": "09:00",
  "end_time": "17:00",
  "is_active": true,
  "max_patients": 20
}
```

### Specific Date Slot Response
```json
{
  "id": "uuid",
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "specific_date": "2024-12-25",
  "slot_type": "online",
  "start_time": "14:00",
  "end_time": "17:00",
  "is_active": true,
  "max_patients": 10
}
```

---

## 🛠️ Common Use Cases

### Use Case 1: Regular Weekly Schedule
Doctor works every Monday and Thursday:
- Create 2 slots with `day_of_week: 1` (Monday) and `day_of_week: 4` (Thursday)

### Use Case 2: Holiday Hours
Doctor has special hours on Christmas:
- Create 1 slot with `specific_date: "2024-12-25"`

### Use Case 3: Event-Specific Availability
Doctor available for health camp on a specific date:
- Create slot with `specific_date` for that event day

### Use Case 4: Override Regular Schedule
Doctor normally works Mondays, but on Jan 1 (a Monday) has different hours:
- Keep recurring Monday slot
- Add specific `"2024-01-01"` slot with different hours
- System will show both (appointment logic should prioritize specific date)

---

## 📱 Frontend Component Example

```javascript
// When user selects a date
async function onDateSelect(date) {
  // Format: YYYY-MM-DD
  const formattedDate = date.toISOString().split('T')[0];
  
  // Fetch slots for this date
  const response = await fetch(
    `/api/v1/time-slots?doctor_id=${doctorId}&specific_date=${formattedDate}`
  );
  const { time_slots } = await response.json();
  
  // Show existing slots or "Add Slot" button
  displaySlots(time_slots);
}

// When user adds a new slot
async function addSlot(slotData) {
  await fetch('/api/v1/time-slots', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      doctor_id: doctorId,
      clinic_id: clinicId,
      specific_date: selectedDate, // from date picker
      ...slotData
    })
  });
}
```

---

## 🐛 Error Handling

### 400 Bad Request
- Invalid date format
- Missing required fields
- Both day_of_week and specific_date provided
- Past date selected

### 404 Not Found
- Doctor or clinic doesn't exist
- Slot ID not found

### 409 Conflict
- Overlapping time slot exists
- Doctor already scheduled elsewhere at that time

---

## 📚 Full Documentation

For complete API documentation, see:
- **DOCTOR_TIME_SLOTS_ENHANCED_API.md** - Comprehensive guide

For older weekly-only documentation, see:
- **DOCTOR_TIME_SLOTS_API_GUIDE.md** - Original guide (now outdated)

---

**Last Updated:** After migration 009

