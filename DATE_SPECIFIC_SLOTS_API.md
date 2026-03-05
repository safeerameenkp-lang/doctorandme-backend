# Date-Specific Doctor Time Slots API

## Overview
Perfect for frontend date picker functionality! This API allows users to select a specific date and add multiple time slots for that date in a single request.

## Endpoint

### Create Date-Specific Slots
**POST** `/api/organizations/doctor-time-slots/date-specific`

Create multiple time slots for a doctor at a clinic for a specific date.

---

## Request Format

### Request Body
```json
{
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "specific_date": "2024-12-25",  // YYYY-MM-DD format
  "slots": [
    {
      "slot_type": "offline",     // Required: "offline" or "online"
      "start_time": "09:00",      // Required: HH:MM format
      "end_time": "12:00",        // Required: HH:MM format
      "max_patients": 10,         // Optional: defaults to 1
      "notes": "Morning shift"    // Optional
    },
    {
      "slot_type": "offline",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 5,
      "notes": "Afternoon shift"
    },
    {
      "slot_type": "online",
      "start_time": "19:00",
      "end_time": "21:00",
      "max_patients": 3,
      "notes": "Evening consultation"
    }
  ]
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `doctor_id` | UUID | Yes | ID of the doctor |
| `clinic_id` | UUID | Yes | ID of the clinic |
| `specific_date` | String | Yes | Date in YYYY-MM-DD format |
| `slots` | Array | Yes | Array of slot definitions (min 1) |

#### Slot Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `slot_type` | String | Yes | "offline" or "online" |
| `start_time` | String | Yes | HH:MM format (24-hour) |
| `end_time` | String | Yes | HH:MM format (24-hour) |
| `max_patients` | Integer | No | Defaults to 1 |
| `notes` | String | No | Optional notes |

---

## Response Format

### Success Response (201 Created)
```json
{
  "message": "Date-specific slots created. 3 created, 0 failed",
  "specific_date": "2024-12-25",
  "created_slots": [
    {
      "id": "uuid-1",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "specific_date": "2024-12-25",
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10,
      "notes": "Morning shift"
    },
    {
      "id": "uuid-2",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "specific_date": "2024-12-25",
      "slot_type": "offline",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 5,
      "notes": "Afternoon shift"
    },
    {
      "id": "uuid-3",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "specific_date": "2024-12-25",
      "slot_type": "online",
      "start_time": "19:00",
      "end_time": "21:00",
      "max_patients": 3,
      "notes": "Evening consultation"
    }
  ],
  "failed_slots": [],
  "total_created": 3,
  "total_failed": 0
}
```

### Partial Success Response (206 Partial Content)
```json
{
  "message": "Date-specific slots created. 2 created, 1 failed",
  "specific_date": "2024-12-25",
  "created_slots": [
    {
      "id": "uuid-1",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "specific_date": "2024-12-25",
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10,
      "notes": "Morning shift"
    },
    {
      "id": "uuid-2",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "specific_date": "2024-12-25",
      "slot_type": "offline",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 5,
      "notes": "Afternoon shift"
    }
  ],
  "failed_slots": [
    {
      "index": 2,
      "error": "Failed to create slot: Doctor already has a conflicting time slot on this specific date. A doctor cannot be available in multiple clinics at the same time."
    }
  ],
  "total_created": 2,
  "total_failed": 1
}
```

---

## Frontend Integration Example

### React Component with Date Picker
```jsx
import React, { useState } from 'react';

function DateSpecificSlotManager({ doctorId, clinicId }) {
  const [selectedDate, setSelectedDate] = useState('');
  const [slots, setSlots] = useState([]);
  const [newSlot, setNewSlot] = useState({
    slot_type: 'offline',
    start_time: '09:00',
    end_time: '12:00',
    max_patients: 10,
    notes: ''
  });

  const handleDateChange = (date) => {
    setSelectedDate(date);
    setSlots([]); // Clear slots when date changes
  };

  const addSlot = () => {
    if (newSlot.start_time && newSlot.end_time) {
      setSlots([...slots, { ...newSlot }]);
      setNewSlot({
        slot_type: 'offline',
        start_time: '',
        end_time: '',
        max_patients: 10,
        notes: ''
      });
    }
  };

  const removeSlot = (index) => {
    setSlots(slots.filter((_, i) => i !== index));
  };

  const createSlots = async () => {
    if (!selectedDate || slots.length === 0) {
      alert('Please select a date and add at least one slot');
      return;
    }

    try {
      const response = await fetch('/api/organizations/doctor-time-slots/date-specific', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          doctor_id: doctorId,
          clinic_id: clinicId,
          specific_date: selectedDate,
          slots: slots
        })
      });

      const result = await response.json();
      
      if (result.total_created > 0) {
        alert(`Successfully created ${result.total_created} slots for ${selectedDate}!`);
        setSlots([]); // Clear slots after successful creation
        if (result.total_failed > 0) {
          console.warn('Some slots failed:', result.failed_slots);
        }
      } else {
        alert('Failed to create slots. Check console for details.');
        console.error('All slots failed:', result.failed_slots);
      }
    } catch (error) {
      console.error('Error creating slots:', error);
      alert('Error creating slots');
    }
  };

  return (
    <div style={{ padding: '20px' }}>
      <h2>Add Time Slots for Specific Date</h2>
      
      {/* Date Picker */}
      <div style={{ marginBottom: '20px' }}>
        <label>
          Select Date:
          <input
            type="date"
            value={selectedDate}
            onChange={(e) => handleDateChange(e.target.value)}
            style={{ marginLeft: '10px', padding: '5px' }}
          />
        </label>
      </div>

      {selectedDate && (
        <>
          {/* Add New Slot Form */}
          <div style={{ border: '1px solid #ccc', padding: '15px', marginBottom: '20px' }}>
            <h3>Add New Slot</h3>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px', marginBottom: '10px' }}>
              <label>
                Slot Type:
                <select
                  value={newSlot.slot_type}
                  onChange={(e) => setNewSlot({...newSlot, slot_type: e.target.value})}
                >
                  <option value="offline">Offline</option>
                  <option value="online">Online</option>
                </select>
              </label>
              <label>
                Max Patients:
                <input
                  type="number"
                  value={newSlot.max_patients}
                  onChange={(e) => setNewSlot({...newSlot, max_patients: parseInt(e.target.value)})}
                  min="1"
                />
              </label>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px', marginBottom: '10px' }}>
              <label>
                Start Time:
                <input
                  type="time"
                  value={newSlot.start_time}
                  onChange={(e) => setNewSlot({...newSlot, start_time: e.target.value})}
                />
              </label>
              <label>
                End Time:
                <input
                  type="time"
                  value={newSlot.end_time}
                  onChange={(e) => setNewSlot({...newSlot, end_time: e.target.value})}
                />
              </label>
            </div>
            <label>
              Notes:
              <input
                type="text"
                value={newSlot.notes}
                onChange={(e) => setNewSlot({...newSlot, notes: e.target.value})}
                placeholder="Optional notes"
                style={{ width: '100%', marginTop: '5px' }}
              />
            </label>
            <button onClick={addSlot} style={{ marginTop: '10px', padding: '8px 16px' }}>
              Add Slot
            </button>
          </div>

          {/* Slots List */}
          {slots.length > 0 && (
            <div style={{ marginBottom: '20px' }}>
              <h3>Slots for {selectedDate}</h3>
              {slots.map((slot, index) => (
                <div key={index} style={{ border: '1px solid #ddd', padding: '10px', marginBottom: '10px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <strong>{slot.slot_type}</strong> - {slot.start_time} to {slot.end_time}
                    {slot.max_patients > 1 && <span> (Max: {slot.max_patients})</span>}
                    {slot.notes && <span> - {slot.notes}</span>}
                  </div>
                  <button onClick={() => removeSlot(index)} style={{ color: 'red' }}>
                    Remove
                  </button>
                </div>
              ))}
            </div>
          )}

          {/* Create Slots Button */}
          {slots.length > 0 && (
            <button
              onClick={createSlots}
              style={{
                backgroundColor: '#007bff',
                color: 'white',
                padding: '12px 24px',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
                fontSize: '16px'
              }}
            >
              Create {slots.length} Slot{slots.length > 1 ? 's' : ''} for {selectedDate}
            </button>
          )}
        </>
      )}
    </div>
  );
}

export default DateSpecificSlotManager;
```

---

## Use Cases

### Use Case 1: Holiday Special Hours
```json
{
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "specific_date": "2024-12-25",
  "slots": [
    {
      "slot_type": "online",
      "start_time": "10:00",
      "end_time": "14:00",
      "max_patients": 5,
      "notes": "Christmas Day - Online consultations only"
    },
    {
      "slot_type": "online",
      "start_time": "16:00",
      "end_time": "18:00",
      "max_patients": 3,
      "notes": "Emergency consultations"
    }
  ]
}
```

### Use Case 2: Conference Day Availability
```json
{
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "specific_date": "2024-11-15",
  "slots": [
    {
      "slot_type": "offline",
      "start_time": "08:00",
      "end_time": "10:00",
      "max_patients": 10,
      "notes": "Pre-conference consultations"
    },
    {
      "slot_type": "offline",
      "start_time": "18:00",
      "end_time": "20:00",
      "max_patients": 8,
      "notes": "Post-conference consultations"
    }
  ]
}
```

### Use Case 3: Weekend Special Hours
```json
{
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "specific_date": "2024-12-28",
  "slots": [
    {
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 15,
      "notes": "Weekend morning shift"
    },
    {
      "slot_type": "online",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 10,
      "notes": "Weekend afternoon online"
    }
  ]
}
```

---

## Validation Rules

**Minimal validation:**

1. **Doctor exists** and is active
2. **Clinic exists** and is active
3. **Date format** must be YYYY-MM-DD
4. **Slot type** must be "offline" or "online"
5. **Time format** must be HH:MM (24-hour)
6. **Max patients** defaults to 1 if not provided

**No complex validation:**
- No overlap checking (handled by database constraints)
- No past date validation
- No time range validation (end_time > start_time)

---

## Error Handling

The API processes each slot individually and reports:
- **Successfully created slots** in `created_slots` array
- **Failed slots** in `failed_slots` array with error details

Common error scenarios:
- Invalid date format (must be YYYY-MM-DD)
- Invalid slot type (must be offline/online)
- Invalid time format (must be HH:MM)
- Database constraint violations (overlaps, etc.)

---

## HTTP Status Codes

| Status | Description |
|--------|-------------|
| 201 Created | All slots created successfully |
| 206 Partial Content | Some slots created, some failed |
| 400 Bad Request | All slots failed or invalid request |
| 404 Not Found | Doctor or clinic not found |
| 401 Unauthorized | Authentication required |
| 403 Forbidden | Insufficient permissions |

---

## Testing Examples

### Test 1: Multiple Slots for Christmas Day
```bash
curl -X POST http://localhost:8081/api/organizations/doctor-time-slots/date-specific \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "doctor_id": "your-doctor-id",
    "clinic_id": "your-clinic-id",
    "specific_date": "2024-12-25",
    "slots": [
      {
        "slot_type": "online",
        "start_time": "10:00",
        "end_time": "12:00",
        "max_patients": 5,
        "notes": "Christmas Morning"
      },
      {
        "slot_type": "online",
        "start_time": "14:00",
        "end_time": "16:00",
        "max_patients": 5,
        "notes": "Christmas Afternoon"
      }
    ]
  }'
```

### Test 2: Single Slot for Specific Date
```bash
curl -X POST http://localhost:8081/api/organizations/doctor-time-slots/date-specific \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "doctor_id": "your-doctor-id",
    "clinic_id": "your-clinic-id",
    "specific_date": "2024-11-15",
    "slots": [
      {
        "slot_type": "offline",
        "start_time": "09:00",
        "end_time": "17:00",
        "max_patients": 20,
        "notes": "Full day availability"
      }
    ]
  }'
```

---

## Benefits

1. **Date-focused**: Perfect for date picker UI
2. **Simple**: Just specify date and slots
3. **Flexible**: Add any number of slots for a date
4. **Fast**: Minimal validation, quick processing
5. **Frontend-friendly**: Easy to integrate with date pickers
6. **Database-driven**: Constraints handled by database

---

## Frontend Flow

1. **User selects a date** from date picker
2. **User adds multiple slots** for that date
3. **User clicks "Create Slots"**
4. **API receives and saves** all slots for the selected date
5. **Response shows** which slots were created successfully

---

**Last Updated:** After implementation of date-specific slots API
