# Bulk Doctor Time Slots API

## Overview
The Bulk Doctor Time Slots API allows you to create multiple time slots for a doctor at a clinic in a single request. This is perfect for UI scenarios where users can select multiple days (like Monday, Tuesday, Wednesday, Friday) and add time slots for each selected day.

## New Endpoint

### Bulk Create Doctor Time Slots
**POST** `/api/v1/doctor-time-slots/bulk`

Create multiple time slots for a doctor at a specific clinic in one request.

---

## Request Format

### Request Body
```json
{
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "slots": [
    {
      "day_of_week": 1,           // For recurring: 0=Sunday, 1=Monday, etc.
      "slot_type": "offline",     // "offline" or "online"
      "start_time": "09:00",      // HH:MM format (24-hour)
      "end_time": "12:00",        // HH:MM format (24-hour)
      "max_patients": 5,          // Optional, defaults to 1
      "notes": "Morning shift"    // Optional
    },
    {
      "day_of_week": 3,           // Wednesday
      "slot_type": "offline",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 3,
      "notes": "Afternoon shift"
    },
    {
      "specific_date": "2024-12-25",  // For one-time slots: YYYY-MM-DD format
      "slot_type": "online",
      "start_time": "10:00",
      "end_time": "14:00",
      "max_patients": 10,
      "notes": "Holiday special hours"
    }
  ]
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `doctor_id` | UUID | Yes | ID of the doctor |
| `clinic_id` | UUID | Yes | ID of the clinic |
| `slots` | Array | Yes | Array of slot definitions (min 1) |

#### Slot Definition Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `day_of_week` | Integer | Conditional* | 0=Sunday, 1=Monday, ..., 6=Saturday |
| `specific_date` | String | Conditional* | YYYY-MM-DD format for one-time slots |
| `slot_type` | String | Yes | "offline" or "online" |
| `start_time` | String | Yes | HH:MM format (24-hour) |
| `end_time` | String | Yes | HH:MM format (24-hour) |
| `max_patients` | Integer | No | Defaults to 1, must be ≥ 1 |
| `notes` | String | No | Optional notes |

*Either `day_of_week` OR `specific_date` must be provided, but not both.

---

## Response Format

### Success Response (201 Created)
```json
{
  "message": "Bulk time slot creation completed. 3 created, 0 failed",
  "created_slots": [
    {
      "id": "uuid-1",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "day_of_week": 1,
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 5,
      "notes": "Morning shift"
    },
    {
      "id": "uuid-2",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "day_of_week": 3,
      "slot_type": "offline",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 3,
      "notes": "Afternoon shift"
    },
    {
      "id": "uuid-3",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "specific_date": "2024-12-25",
      "slot_type": "online",
      "start_time": "10:00",
      "end_time": "14:00",
      "max_patients": 10,
      "notes": "Holiday special hours"
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
  "message": "Bulk time slot creation completed. 2 created, 1 failed",
  "created_slots": [
    {
      "id": "uuid-1",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "day_of_week": 1,
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 5,
      "notes": "Morning shift"
    },
    {
      "id": "uuid-2",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "day_of_week": 3,
      "slot_type": "offline",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 3,
      "notes": "Afternoon shift"
    }
  ],
  "failed_slots": [
    {
      "index": 2,
      "error": "Doctor already has a conflicting time slot on this day/date. A doctor cannot be available in multiple clinics at the same time."
    }
  ],
  "total_created": 2,
  "total_failed": 1
}
```

### Complete Failure Response (400 Bad Request)
```json
{
  "message": "Bulk time slot creation completed. 0 created, 3 failed",
  "created_slots": [],
  "failed_slots": [
    {
      "index": 0,
      "error": "day_of_week must be 0-6 (0=Sunday, 6=Saturday)"
    },
    {
      "index": 1,
      "error": "end_time must be after start_time"
    },
    {
      "index": 2,
      "error": "Date must be in YYYY-MM-DD format"
    }
  ],
  "total_created": 0,
  "total_failed": 3
}
```

---

## Use Cases

### Use Case 1: Weekly Schedule Setup
Set up a doctor's weekly schedule with multiple slots per day:

```json
{
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "slots": [
    {
      "day_of_week": 1,
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10,
      "notes": "Monday Morning"
    },
    {
      "day_of_week": 1,
      "slot_type": "offline",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 10,
      "notes": "Monday Afternoon"
    },
    {
      "day_of_week": 3,
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10,
      "notes": "Wednesday Morning"
    },
    {
      "day_of_week": 5,
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "17:00",
      "max_patients": 20,
      "notes": "Friday Full Day"
    }
  ]
}
```

### Use Case 2: UI Checkbox Implementation
When user selects multiple days in UI (like your image shows):

```javascript
// User selects Monday, Tuesday, Wednesday, Friday
const selectedDays = [1, 2, 3, 5]; // Monday, Tuesday, Wednesday, Friday
const morningStart = "09:00";
const morningEnd = "12:00";
const afternoonStart = "14:00";
const afternoonEnd = "17:00";

// Build slots array
const slots = [];
selectedDays.forEach(day => {
  // Add morning slot
  slots.push({
    day_of_week: day,
    slot_type: "offline",
    start_time: morningStart,
    end_time: morningEnd,
    max_patients: 10,
    notes: "Morning shift"
  });
  
  // Add afternoon slot
  slots.push({
    day_of_week: day,
    slot_type: "offline",
    start_time: afternoonStart,
    end_time: afternoonEnd,
    max_patients: 10,
    notes: "Afternoon shift"
  });
});

// Send bulk request
const response = await fetch('/api/v1/doctor-time-slots/bulk', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    doctor_id: doctorId,
    clinic_id: clinicId,
    slots: slots
  })
});
```

### Use Case 3: Mixed Recurring and Specific Dates
Combine regular weekly slots with special date slots:

```json
{
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "slots": [
    {
      "day_of_week": 1,
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "17:00",
      "notes": "Regular Monday"
    },
    {
      "day_of_week": 3,
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "17:00",
      "notes": "Regular Wednesday"
    },
    {
      "specific_date": "2024-12-25",
      "slot_type": "online",
      "start_time": "10:00",
      "end_time": "14:00",
      "notes": "Christmas Day - Online Only"
    },
    {
      "specific_date": "2024-01-01",
      "slot_type": "offline",
      "start_time": "12:00",
      "end_time": "16:00",
      "notes": "New Year - Limited Hours"
    }
  ]
}
```

---

## Validation Rules

1. **Must provide EITHER** `day_of_week` OR `specific_date`, **not both**
2. `day_of_week` must be 0-6 (0=Sunday, 6=Saturday)
3. `specific_date` must be in YYYY-MM-DD format and cannot be in the past
4. `slot_type` must be "offline" or "online"
5. `end_time` must be after `start_time`
6. `max_patients` must be at least 1
7. System prevents overlapping time slots for the same doctor
8. Doctor and clinic must exist and be active

---

## Error Handling

The API processes each slot individually and reports:
- **Successfully created slots** in `created_slots` array
- **Failed slots** in `failed_slots` array with error details

Common error scenarios:
- Invalid day_of_week (must be 0-6)
- Invalid date format (must be YYYY-MM-DD)
- Past dates not allowed
- Overlapping time slots
- Invalid time format (must be HH:MM)
- End time before start time
- Invalid slot type (must be offline/online)

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

## Frontend Integration Example

### React Component Example
```jsx
function TimeSlotManager({ doctorId, clinicId }) {
  const [selectedDays, setSelectedDays] = useState([]);
  const [morningStart, setMorningStart] = useState("09:00");
  const [morningEnd, setMorningEnd] = useState("12:00");
  const [afternoonStart, setAfternoonStart] = useState("14:00");
  const [afternoonEnd, setAfternoonEnd] = useState("17:00");

  const handleDayToggle = (day) => {
    setSelectedDays(prev => 
      prev.includes(day) 
        ? prev.filter(d => d !== day)
        : [...prev, day]
    );
  };

  const handleBulkCreate = async () => {
    const slots = [];
    
    selectedDays.forEach(day => {
      // Add morning slot
      slots.push({
        day_of_week: day,
        slot_type: "offline",
        start_time: morningStart,
        end_time: morningEnd,
        max_patients: 10,
        notes: "Morning shift"
      });
      
      // Add afternoon slot
      slots.push({
        day_of_week: day,
        slot_type: "offline",
        start_time: afternoonStart,
        end_time: afternoonEnd,
        max_patients: 10,
        notes: "Afternoon shift"
      });
    });

    try {
      const response = await fetch('/api/v1/doctor-time-slots/bulk', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          doctor_id: doctorId,
          clinic_id: clinicId,
          slots: slots
        })
      });

      const result = await response.json();
      
      if (result.total_created > 0) {
        alert(`Successfully created ${result.total_created} time slots!`);
        if (result.total_failed > 0) {
          console.warn('Some slots failed:', result.failed_slots);
        }
      } else {
        alert('Failed to create time slots. Check console for details.');
        console.error('All slots failed:', result.failed_slots);
      }
    } catch (error) {
      console.error('Error creating time slots:', error);
      alert('Error creating time slots');
    }
  };

  return (
    <div>
      <h3>Select Days</h3>
      {['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'].map((day, index) => (
        <label key={day}>
          <input
            type="checkbox"
            checked={selectedDays.includes(index)}
            onChange={() => handleDayToggle(index)}
          />
          {day}
        </label>
      ))}
      
      <h3>Morning Shift</h3>
      <input
        type="time"
        value={morningStart}
        onChange={(e) => setMorningStart(e.target.value)}
      />
      <input
        type="time"
        value={morningEnd}
        onChange={(e) => setMorningEnd(e.target.value)}
      />
      
      <h3>Afternoon Shift</h3>
      <input
        type="time"
        value={afternoonStart}
        onChange={(e) => setAfternoonStart(e.target.value)}
      />
      <input
        type="time"
        value={afternoonEnd}
        onChange={(e) => setAfternoonEnd(e.target.value)}
      />
      
      <button onClick={handleBulkCreate} disabled={selectedDays.length === 0}>
        Create Time Slots ({selectedDays.length * 2} slots)
      </button>
    </div>
  );
}
```

---

## Testing Examples

### Test 1: Multiple Days with Morning and Afternoon Slots
```bash
curl -X POST http://localhost:8080/api/v1/doctor-time-slots/bulk \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "doctor_id": "your-doctor-id",
    "clinic_id": "your-clinic-id",
    "slots": [
      {
        "day_of_week": 1,
        "slot_type": "offline",
        "start_time": "09:00",
        "end_time": "12:00",
        "max_patients": 10,
        "notes": "Monday Morning"
      },
      {
        "day_of_week": 1,
        "slot_type": "offline",
        "start_time": "14:00",
        "end_time": "17:00",
        "max_patients": 10,
        "notes": "Monday Afternoon"
      },
      {
        "day_of_week": 3,
        "slot_type": "offline",
        "start_time": "09:00",
        "end_time": "12:00",
        "max_patients": 10,
        "notes": "Wednesday Morning"
      },
      {
        "day_of_week": 3,
        "slot_type": "offline",
        "start_time": "14:00",
        "end_time": "17:00",
        "max_patients": 10,
        "notes": "Wednesday Afternoon"
      }
    ]
  }'
```

### Test 2: Mixed Recurring and Specific Date Slots
```bash
curl -X POST http://localhost:8080/api/v1/doctor-time-slots/bulk \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "doctor_id": "your-doctor-id",
    "clinic_id": "your-clinic-id",
    "slots": [
      {
        "day_of_week": 1,
        "slot_type": "offline",
        "start_time": "09:00",
        "end_time": "17:00",
        "notes": "Regular Monday"
      },
      {
        "specific_date": "2024-12-25",
        "slot_type": "online",
        "start_time": "10:00",
        "end_time": "14:00",
        "notes": "Christmas Day"
      }
    ]
  }'
```

---

## Benefits

1. **Efficiency**: Create multiple slots in one API call
2. **UI-Friendly**: Perfect for checkbox-based day selection
3. **Flexible**: Mix recurring and specific date slots
4. **Robust**: Individual slot validation and error reporting
5. **Transactional**: Partial success handling
6. **Backward Compatible**: Original single-slot API still works

---

## Migration Notes

- This is a new endpoint, no database changes required
- Uses existing `doctor_time_slots` table
- Same validation rules as single-slot creation
- Same overlap prevention logic
- Same authentication and authorization requirements

---

**Last Updated:** After implementation of bulk time slots feature
