# Time Slot Day Validation - UI Integration Guide

## 🎯 Overview

The API now supports **optional day_of_week validation** in each slot to help UI prevent mismatched dates. This ensures that when users select a date in the UI, the day of the week matches.

## 📋 Day of Week Format (ISO 8601)

```
1 = Monday
2 = Tuesday
3 = Wednesday
4 = Thursday
5 = Friday
6 = Saturday
7 = Sunday
```

---

## ✅ Valid Request Example

### Scenario: Creating slots for Monday, January 20, 2025

```json
POST /doctor-time-slots
{
  "doctor_id": "123e4567-e89b-12d3-a456-426614174000",
  "clinic_id": "123e4567-e89b-12d3-a456-426614174001",
  "slot_type": "offline",
  "date": "2025-01-20",           // This is a Monday
  "slots": [
    {
      "start_time": "09:00",
      "end_time": "09:30",
      "max_patients": 1,
      "day_of_week": 1            // ✅ 1 = Monday (matches the date!)
    },
    {
      "start_time": "09:30",
      "end_time": "10:00",
      "max_patients": 1,
      "day_of_week": 1            // ✅ 1 = Monday
    }
  ]
}
```

**Response (201 Created):**
```json
{
  "message": "Slot creation completed. 2 created, 0 failed",
  "created_slots": [
    {
      "id": "slot-uuid-1",
      "doctor_id": "123e4567-e89b-12d3-a456-426614174000",
      "clinic_id": "123e4567-e89b-12d3-a456-426614174001",
      "date": "2025-01-20",
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "09:30",
      "max_patients": 1,
      "booked_patients": 0,
      "available_spots": 1,
      "is_available": true,
      "status": "available",
      "is_active": true,
      "created_at": "2024-10-15T12:00:00Z",
      "updated_at": "2024-10-15T12:00:00Z"
    },
    {
      "id": "slot-uuid-2",
      "date": "2025-01-20",
      "start_time": "09:30",
      "end_time": "10:00",
      ...
    }
  ],
  "total_created": 2,
  "total_failed": 0
}
```

---

## ❌ Invalid Request Examples

### Error 1: Day of Week Doesn't Match Date

```json
POST /doctor-time-slots
{
  "doctor_id": "123e4567-e89b-12d3-a456-426614174000",
  "clinic_id": "123e4567-e89b-12d3-a456-426614174001",
  "slot_type": "offline",
  "date": "2025-01-20",           // This is Monday
  "slots": [
    {
      "start_time": "09:00",
      "end_time": "09:30",
      "max_patients": 1,
      "day_of_week": 2            // ❌ 2 = Tuesday (doesn't match!)
    }
  ]
}
```

**Response (400 Bad Request):**
```json
{
  "message": "Slot creation completed. 0 created, 1 failed",
  "created_slots": null,
  "failed_slots": [
    {
      "index": 0,
      "error": "Date 2025-01-20 is a Monday, but day_of_week is set to 2 (Tuesday)"
    }
  ],
  "total_created": 0,
  "total_failed": 1
}
```

---

### Error 2: Invalid day_of_week Value

```json
POST /doctor-time-slots
{
  "doctor_id": "123e4567-e89b-12d3-a456-426614174000",
  "clinic_id": "123e4567-e89b-12d3-a456-426614174001",
  "slot_type": "offline",
  "date": "2025-01-20",
  "slots": [
    {
      "start_time": "09:00",
      "end_time": "09:30",
      "max_patients": 1,
      "day_of_week": 8            // ❌ Invalid! Must be 1-7
    }
  ]
}
```

**Response (400 Bad Request):**
```json
{
  "message": "Slot creation completed. 0 created, 1 failed",
  "created_slots": null,
  "failed_slots": [
    {
      "index": 0,
      "error": "Invalid day_of_week in slot. Must be between 1 (Monday) and 7 (Sunday)"
    }
  ],
  "total_created": 0,
  "total_failed": 1
}
```

---

## 🎨 UI Integration Examples

### Example 1: React/JavaScript Date Picker

```javascript
// When user selects a date in the UI
function handleDateSelect(selectedDate) {
  // Get day of week (JavaScript: 0=Sunday, 6=Saturday)
  const jsDay = selectedDate.getDay();
  
  // Convert to ISO 8601 format (1=Monday, 7=Sunday)
  const isoDay = jsDay === 0 ? 7 : jsDay;
  
  // Format date for API
  const dateString = selectedDate.toISOString().split('T')[0]; // "2025-01-20"
  
  // Prepare API payload
  const payload = {
    doctor_id: selectedDoctor.id,
    clinic_id: selectedClinic.id,
    slot_type: "offline",
    date: dateString,
    slots: [
      {
        start_time: "09:00",
        end_time: "09:30",
        max_patients: 1,
        day_of_week: isoDay  // ✅ Validates date matches day
      }
    ]
  };
  
  // Send to API
  await createTimeSlots(payload);
}
```

---

### Example 2: Display Day Name with Validation

```javascript
// Helper function to get day name
function getDayName(isoDay) {
  const days = ['', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];
  return days[isoDay];
}

// When building the UI
function SlotCreationForm({ selectedDate }) {
  const jsDay = selectedDate.getDay();
  const isoDay = jsDay === 0 ? 7 : jsDay;
  const dayName = getDayName(isoDay);
  
  return (
    <div>
      <p>Creating slots for: {selectedDate.toLocaleDateString()}</p>
      <p>Day: {dayName} (day_of_week: {isoDay})</p>
      
      <button onClick={() => {
        const payload = {
          date: selectedDate.toISOString().split('T')[0],
          slots: slots.map(slot => ({
            ...slot,
            day_of_week: isoDay  // Auto-include for validation
          }))
        };
        createSlots(payload);
      }}>
        Create Slots
      </button>
    </div>
  );
}
```

---

### Example 3: Convert JavaScript Day to ISO 8601

```javascript
/**
 * Converts JavaScript day (0=Sunday) to ISO 8601 (1=Monday, 7=Sunday)
 */
function jsToIsoDay(jsDay) {
  return jsDay === 0 ? 7 : jsDay;
}

/**
 * Converts ISO 8601 day to JavaScript day
 */
function isoToJsDay(isoDay) {
  return isoDay === 7 ? 0 : isoDay;
}

// Usage:
const date = new Date('2025-01-20'); // Monday
const jsDay = date.getDay();         // Returns 1
const isoDay = jsToIsoDay(jsDay);    // Returns 1 (Monday)

const sundayDate = new Date('2025-01-19'); // Sunday
const jsSunday = sundayDate.getDay();      // Returns 0
const isoSunday = jsToIsoDay(jsSunday);    // Returns 7 (Sunday in ISO)
```

---

## 📊 Complete Working Example

### Full UI Flow

```javascript
// Component for creating time slots
import React, { useState } from 'react';
import axios from 'axios';

function CreateTimeSlotsForm({ doctorId, clinicId }) {
  const [selectedDate, setSelectedDate] = useState(new Date());
  const [slots, setSlots] = useState([
    { start_time: '09:00', end_time: '09:30', max_patients: 1 }
  ]);
  
  const handleSubmit = async () => {
    try {
      // Get ISO day of week (1-7)
      const jsDay = selectedDate.getDay();
      const isoDay = jsDay === 0 ? 7 : jsDay;
      
      // Format date as YYYY-MM-DD
      const dateString = selectedDate.toISOString().split('T')[0];
      
      // Prepare payload
      const payload = {
        doctor_id: doctorId,
        clinic_id: clinicId,
        slot_type: 'offline',
        date: dateString,
        slots: slots.map(slot => ({
          ...slot,
          day_of_week: isoDay  // Include for validation
        }))
      };
      
      // Send to API
      const response = await axios.post('/doctor-time-slots', payload);
      
      console.log('Success!', response.data);
      alert(`Created ${response.data.total_created} slots`);
      
    } catch (error) {
      if (error.response?.data?.failed_slots) {
        const errors = error.response.data.failed_slots
          .map(f => f.error)
          .join('\n');
        alert('Error creating slots:\n' + errors);
      }
    }
  };
  
  // Get day name for display
  const dayNames = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
  const dayName = dayNames[selectedDate.getDay()];
  
  return (
    <div>
      <h2>Create Time Slots</h2>
      
      <label>
        Select Date:
        <input 
          type="date" 
          value={selectedDate.toISOString().split('T')[0]}
          onChange={(e) => setSelectedDate(new Date(e.target.value))}
        />
      </label>
      
      <p>Selected: {selectedDate.toLocaleDateString()} ({dayName})</p>
      
      {/* Slot configuration UI */}
      {slots.map((slot, index) => (
        <div key={index}>
          <input 
            type="time" 
            value={slot.start_time}
            onChange={(e) => {
              const newSlots = [...slots];
              newSlots[index].start_time = e.target.value;
              setSlots(newSlots);
            }}
          />
          <input 
            type="time" 
            value={slot.end_time}
            onChange={(e) => {
              const newSlots = [...slots];
              newSlots[index].end_time = e.target.value;
              setSlots(newSlots);
            }}
          />
        </div>
      ))}
      
      <button onClick={handleSubmit}>Create Slots</button>
    </div>
  );
}

export default CreateTimeSlotsForm;
```

---

## 🔍 Date Validation Reference

### January 2025 Calendar Example

| Date       | Day of Week | ISO day_of_week |
|------------|-------------|-----------------|
| 2025-01-13 | Monday      | 1               |
| 2025-01-14 | Tuesday     | 2               |
| 2025-01-15 | Wednesday   | 3               |
| 2025-01-16 | Thursday    | 4               |
| 2025-01-17 | Friday      | 5               |
| 2025-01-18 | Saturday    | 6               |
| 2025-01-19 | Sunday      | 7               |
| 2025-01-20 | Monday      | 1               |

---

## ⚙️ Optional vs Required

### `day_of_week` in slots is **OPTIONAL**

**Without day_of_week (still works):**
```json
{
  "date": "2025-01-20",
  "slots": [
    {
      "start_time": "09:00",
      "end_time": "09:30"
      // No day_of_week - API accepts this
    }
  ]
}
```

**With day_of_week (validates):**
```json
{
  "date": "2025-01-20",
  "slots": [
    {
      "start_time": "09:00",
      "end_time": "09:30",
      "day_of_week": 1  // Validates that 2025-01-20 is indeed Monday
    }
  ]
}
```

**Recommendation:** Include `day_of_week` in UI requests for extra validation and better error messages.

---

## 🎯 Benefits for UI

1. **Prevents User Errors**: Catches when users accidentally select wrong dates
2. **Better Error Messages**: Shows exactly what went wrong ("Date is Monday but you set Tuesday")
3. **Frontend Validation**: Can validate before sending to API
4. **Type Safety**: Ensures date and day are in sync

---

## 📝 Quick Reference

### JavaScript Date to ISO Day of Week

```javascript
const date = new Date('2025-01-20');
const jsDay = date.getDay();           // 0-6 (0=Sunday)
const isoDay = jsDay === 0 ? 7 : jsDay; // 1-7 (1=Monday, 7=Sunday)
```

### API Payload Structure

```json
{
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "slot_type": "offline|online",
  "date": "YYYY-MM-DD",                    // Required for date-specific slots
  "slots": [
    {
      "start_time": "HH:MM",               // Required
      "end_time": "HH:MM",                 // Required
      "max_patients": 1,                   // Optional (default: 1)
      "notes": "string",                   // Optional
      "day_of_week": 1-7                   // Optional (for validation)
    }
  ]
}
```

---

**Status**: ✅ Implemented and Ready for UI Integration  
**Last Updated**: October 15, 2025

