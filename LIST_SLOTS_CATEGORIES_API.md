# List Doctor Slots with Categories API

## Overview
Lists doctor time slots grouped by day with time-based categories (morning, afternoon, evening). Perfect for appointment booking UI that shows available slots organized by day and time.

## Endpoint

### List Doctor Slots with Categories
**GET** `/api/organizations/doctor-time-slots/categories`

List doctor time slots grouped by day with time categories (morning, afternoon, evening).

---

## Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `doctor_id` | UUID | Yes | ID of the doctor |
| `clinic_id` | UUID | No | ID of the clinic (optional filter) |
| `date` | String | No | Specific date in YYYY-MM-DD format |
| `slot_type` | String | No | Filter by slot type: "offline" or "online" |

---

## Response Format

### Success Response (200 OK)
```json
{
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "date": "2024-12-25",
  "days": [
    {
      "day_name": "Sunday",
      "day_of_week": 0,
      "night": [
        {
          "id": "slot-uuid-0",
          "doctor_id": "uuid",
          "clinic_id": "uuid",
          "slot_type": "online",
          "start_time": "23:00",
          "end_time": "01:00",
          "max_patients": 5,
          "capacity": 5,
          "notes": "Night emergency",
          "is_active": true,
          "created_at": "2024-01-01T00:00:00Z",
          "updated_at": "2024-01-01T00:00:00Z",
          "day_of_week": 0
        }
      ],
      "morning": [
        {
          "id": "slot-uuid-1",
          "doctor_id": "uuid",
          "clinic_id": "uuid",
          "slot_type": "offline",
          "start_time": "09:00",
          "end_time": "12:00",
          "max_patients": 10,
          "capacity": 10,
          "notes": "Morning shift",
          "is_active": true,
          "created_at": "2024-01-01T00:00:00Z",
          "updated_at": "2024-01-01T00:00:00Z",
          "day_of_week": 0
        }
      ],
      "afternoon": [
        {
          "id": "slot-uuid-2",
          "doctor_id": "uuid",
          "clinic_id": "uuid",
          "slot_type": "offline",
          "start_time": "14:00",
          "end_time": "17:00",
          "max_patients": 10,
          "capacity": 10,
          "notes": "Afternoon shift",
          "is_active": true,
          "created_at": "2024-01-01T00:00:00Z",
          "updated_at": "2024-01-01T00:00:00Z",
          "day_of_week": 0
        }
      ],
      "evening": [],
      "has_slots": true,
      "total_slots": 3
    },
    {
      "day_name": "Monday",
      "day_of_week": 1,
      "night": [],
      "morning": [],
      "afternoon": [],
      "evening": [],
      "has_slots": false,
      "total_slots": 0
    },
    {
      "day_name": "Tuesday",
      "day_of_week": 2,
      "night": [],
      "morning": [
        {
          "id": "slot-uuid-3",
          "doctor_id": "uuid",
          "clinic_id": "uuid",
          "slot_type": "online",
          "start_time": "10:00",
          "end_time": "11:00",
          "max_patients": 5,
          "capacity": 5,
          "notes": "Online consultation",
          "is_active": true,
          "created_at": "2024-01-01T00:00:00Z",
          "updated_at": "2024-01-01T00:00:00Z",
          "day_of_week": 2
        }
      ],
      "afternoon": [],
      "evening": [],
      "has_slots": true,
      "total_slots": 1
    },
    {
      "day_name": "Wednesday",
      "day_of_week": 3,
      "night": [],
      "morning": [],
      "afternoon": [],
      "evening": [],
      "has_slots": false,
      "total_slots": 0
    },
    {
      "day_name": "Thursday",
      "day_of_week": 4,
      "night": [],
      "morning": [],
      "afternoon": [],
      "evening": [],
      "has_slots": false,
      "total_slots": 0
    },
    {
      "day_name": "Friday",
      "day_of_week": 5,
      "night": [],
      "morning": [],
      "afternoon": [],
      "evening": [],
      "has_slots": false,
      "total_slots": 0
    },
    {
      "day_name": "Saturday",
      "day_of_week": 6,
      "night": [],
      "morning": [],
      "afternoon": [],
      "evening": [],
      "has_slots": false,
      "total_slots": 0
    }
  ]
}
```

---

## Time Categories

### Night Slots
- **Time Range:** 22:00 - 05:59
- **Category:** `night`
- **Examples:** 23:00-01:00, 02:00-04:00

### Morning Slots
- **Time Range:** 06:00 - 11:59
- **Category:** `morning`
- **Examples:** 09:00-12:00, 08:00-10:00

### Afternoon Slots
- **Time Range:** 12:00 - 16:59
- **Category:** `afternoon`
- **Examples:** 14:00-17:00, 13:00-15:00

### Evening Slots
- **Time Range:** 17:00 - 21:59
- **Category:** `evening`
- **Examples:** 18:00-20:00, 19:00-21:00

---

## Use Cases

### Use Case 1: All Doctor Slots
```bash
GET /api/organizations/doctor-time-slots/categories?doctor_id=uuid
```

### Use Case 2: Doctor Slots for Specific Clinic
```bash
GET /api/organizations/doctor-time-slots/categories?doctor_id=uuid&clinic_id=uuid
```

### Use Case 3: Doctor Slots for Specific Date
```bash
GET /api/organizations/doctor-time-slots/categories?doctor_id=uuid&date=2024-12-25
```

### Use Case 4: Doctor Slots for Specific Date and Clinic
```bash
GET /api/organizations/doctor-time-slots/categories?doctor_id=uuid&clinic_id=uuid&date=2024-12-25
```
**Note:** Requires doctor to be linked to the clinic via `clinic_doctor_links` table

### Use Case 5: Only Offline Slots
```bash
GET /api/organizations/doctor-time-slots/categories?doctor_id=uuid&slot_type=offline
```

### Use Case 6: Only Online Slots
```bash
GET /api/organizations/doctor-time-slots/categories?doctor_id=uuid&slot_type=online
```

### Use Case 7: Night Doctor Slots Only
```bash
GET /api/organizations/doctor-time-slots/categories?doctor_id=uuid&slot_type=online&date=2024-12-25
```

---

## Frontend Integration Example

### React Component for Appointment Booking
```jsx
import React, { useState, useEffect } from 'react';

function AppointmentBooking({ doctorId, clinicId, selectedDate }) {
  const [slots, setSlots] = useState(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetchSlots();
  }, [doctorId, clinicId, selectedDate]);

  const fetchSlots = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        doctor_id: doctorId,
        ...(clinicId && { clinic_id: clinicId }),
        ...(selectedDate && { date: selectedDate })
      });

      const response = await fetch(`/api/organizations/doctor-time-slots/categories?${params}`);
      const data = await response.json();
      setSlots(data);
    } catch (error) {
      console.error('Error fetching slots:', error);
    } finally {
      setLoading(false);
    }
  };

  const renderTimeCategory = (category, slots) => {
    if (slots.length === 0) return null;

    return (
      <div className="time-category">
        <h4>{category.charAt(0).toUpperCase() + category.slice(1)}</h4>
        <div className="slots-grid">
          {slots.map(slot => (
            <div key={slot.id} className="slot-card">
              <div className="slot-time">
                {slot.start_time} - {slot.end_time}
              </div>
              <div className={`slot-type ${slot.slot_type}`}>
                {slot.slot_type === 'online' ? '🌐 Online' : '🏥 Offline'}
              </div>
              <div className="slot-capacity">
                Capacity: {slot.capacity}
              </div>
              {slot.notes && (
                <div className="slot-notes">
                  {slot.notes}
                </div>
              )}
              <button 
                className="book-slot-btn"
                onClick={() => bookSlot(slot)}
              >
                Book Slot
              </button>
            </div>
          ))}
        </div>
      </div>
    );
  };

  const bookSlot = (slot) => {
    // Handle slot booking
    console.log('Booking slot:', slot);
  };

  if (loading) return <div>Loading slots...</div>;
  if (!slots) return <div>No slots available</div>;

  return (
    <div className="appointment-booking">
      <h2>Available Time Slots</h2>
      
      {slots.days.map(day => (
        <div key={day.day_of_week} className="day-section">
          <h3 className={`day-header ${day.has_slots ? 'has-slots' : 'no-slots'}`}>
            {day.day_name}
            {day.has_slots && (
              <span className="slot-count">({day.total_slots} slots)</span>
            )}
          </h3>
          
          {day.has_slots ? (
            <div className="day-slots">
              {renderTimeCategory('night', day.night)}
              {renderTimeCategory('morning', day.morning)}
              {renderTimeCategory('afternoon', day.afternoon)}
              {renderTimeCategory('evening', day.evening)}
            </div>
          ) : (
            <div className="no-slots-message">
              No available slots for this day
            </div>
          )}
        </div>
      ))}
    </div>
  );
}

export default AppointmentBooking;
```

---

## CSS Styling Example
```css
.appointment-booking {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.day-section {
  margin-bottom: 30px;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 15px;
}

.day-header {
  margin: 0 0 15px 0;
  padding: 10px;
  border-radius: 4px;
}

.day-header.has-slots {
  background-color: #e8f5e8;
  color: #2d5a2d;
}

.day-header.no-slots {
  background-color: #f5f5f5;
  color: #666;
}

.slot-count {
  font-size: 0.9em;
  font-weight: normal;
}

.time-category {
  margin-bottom: 20px;
}

.time-category h4 {
  margin: 0 0 10px 0;
  color: #333;
  font-size: 1.1em;
}

.slots-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 10px;
}

.slot-card {
  border: 1px solid #ccc;
  border-radius: 6px;
  padding: 12px;
  background-color: #fff;
  transition: box-shadow 0.2s;
}

.slot-card:hover {
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.slot-time {
  font-weight: bold;
  color: #2c5aa0;
  margin-bottom: 5px;
}

.slot-type {
  color: #666;
  font-size: 0.9em;
  margin-bottom: 5px;
}

.slot-capacity {
  color: #666;
  font-size: 0.9em;
  margin-bottom: 5px;
}

.slot-notes {
  color: #888;
  font-size: 0.8em;
  margin-bottom: 10px;
}

.book-slot-btn {
  background-color: #007bff;
  color: white;
  border: none;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9em;
}

.book-slot-btn:hover {
  background-color: #0056b3;
}

.no-slots-message {
  color: #666;
  font-style: italic;
  text-align: center;
  padding: 20px;
}
```

---

## Testing Examples

### Test 1: All Doctor Slots
```bash
curl -X GET "http://localhost:8081/api/organizations/doctor-time-slots/categories?doctor_id=your-doctor-id" \
  -H "Authorization: Bearer your-token"
```

### Test 2: Doctor Slots for Specific Date
```bash
curl -X GET "http://localhost:8081/api/organizations/doctor-time-slots/categories?doctor_id=your-doctor-id&date=2024-12-25" \
  -H "Authorization: Bearer your-token"
```

### Test 3: Doctor Slots for Specific Clinic
```bash
curl -X GET "http://localhost:8081/api/organizations/doctor-time-slots/categories?doctor_id=your-doctor-id&clinic_id=your-clinic-id" \
  -H "Authorization: Bearer your-token"
```

### Test 4: Only Offline Slots
```bash
curl -X GET "http://localhost:8081/api/organizations/doctor-time-slots/categories?doctor_id=your-doctor-id&slot_type=offline" \
  -H "Authorization: Bearer your-token"
```

### Test 5: Only Online Slots
```bash
curl -X GET "http://localhost:8081/api/organizations/doctor-time-slots/categories?doctor_id=your-doctor-id&slot_type=online" \
  -H "Authorization: Bearer your-token"
```

### Test 6: Night Doctor Slots
```bash
curl -X GET "http://localhost:8081/api/organizations/doctor-time-slots/categories?doctor_id=your-doctor-id&slot_type=online&date=2024-12-25" \
  -H "Authorization: Bearer your-token"
```

---

## Clinic-Doctor Link Validation

### Important Security Feature
When `clinic_id` is provided, the API validates that the doctor is actually linked to that clinic through the `clinic_doctor_links` table.

### Validation Logic
```sql
SELECT EXISTS(
    SELECT 1 FROM clinic_doctor_links cdl
    WHERE cdl.clinic_id = $1 
    AND cdl.doctor_id = $2 
    AND cdl.is_active = true
)
```

### Error Response
If doctor is not linked to clinic:
```json
{
  "error": "Doctor is not linked to this clinic",
  "message": "The specified doctor is not associated with this clinic"
}
```

### Use Cases
- **Multi-clinic doctors**: Doctors can work at multiple clinics
- **Clinic-specific access**: Only show slots for clinics where doctor is linked
- **Security**: Prevents unauthorized access to other clinic's data
- **Data isolation**: Ensures proper multi-tenant data separation

---

## Benefits

1. **Organized by Day**: All 7 days of the week included
2. **Time Categories**: Night, morning, afternoon, evening slots separated
3. **Slot Type Filtering**: Filter by offline or online slots
4. **Capacity Information**: Shows max_patients for each slot
5. **Flexible Filtering**: By doctor, clinic, date, and slot type
6. **Empty State Handling**: Shows when no slots available
7. **Frontend-Friendly**: Perfect for appointment booking UI
8. **Night Doctor Support**: Handles night shift slots (22:00-05:59)

---

## Response Structure

### Day Object
```json
{
  "day_name": "Monday",
  "day_of_week": 1,
  "night": [...],       // Array of night slots (22:00-05:59)
  "morning": [...],     // Array of morning slots (06:00-11:59)
  "afternoon": [...],   // Array of afternoon slots (12:00-16:59)
  "evening": [...],     // Array of evening slots (17:00-21:59)
  "has_slots": true,    // Boolean indicating if any slots exist
  "total_slots": 4      // Total number of slots for this day
}
```

### Slot Object
```json
{
  "id": "uuid",
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "slot_type": "offline",
  "start_time": "09:00",
  "end_time": "12:00",
  "max_patients": 10,
  "capacity": 10,
  "notes": "Morning shift",
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "day_of_week": 1
}
```

---

**Last Updated:** After implementation of list slots with categories API
