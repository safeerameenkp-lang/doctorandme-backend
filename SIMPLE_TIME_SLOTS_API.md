# Simple Doctor Time Slots API

## Overview
Simplified API for creating doctor time slots with minimal validation. Focuses on accepting and storing slot data without complex business logic.

## Endpoint

### Simple Create Doctor Time Slots
**POST** `/api/organizations/doctor-time-slots/simple`

Create multiple time slots for a doctor at a clinic with minimal validation.

---

## Request Format

### Request Body
```json
{
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "slots": [
    {
      "day_of_week": 1,           // Optional: 0=Sunday, 1=Monday, etc.
      "specific_date": "2024-12-25",  // Optional: YYYY-MM-DD format
      "slot_type": "offline",     // Required: "offline" or "online"
      "start_time": "09:00",      // Required: HH:MM format
      "end_time": "12:00",        // Required: HH:MM format
      "max_patients": 10,         // Optional: defaults to 1
      "notes": "Morning shift"    // Optional
    },
    {
      "day_of_week": 2,           // Tuesday
      "slot_type": "offline",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 5,
      "notes": "Afternoon shift"
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

#### Slot Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `day_of_week` | Integer | No | 0=Sunday, 1=Monday, ..., 6=Saturday |
| `specific_date` | String | No | YYYY-MM-DD format for specific dates |
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
  "message": "Slot creation completed. 2 created, 0 failed",
  "created_slots": [
    {
      "id": "uuid-1",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "day_of_week": 1,
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
      "day_of_week": 2,
      "slot_type": "offline",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 5,
      "notes": "Afternoon shift"
    }
  ],
  "failed_slots": [],
  "total_created": 2,
  "total_failed": 0
}
```

### Partial Success Response (206 Partial Content)
```json
{
  "message": "Slot creation completed. 1 created, 1 failed",
  "created_slots": [
    {
      "id": "uuid-1",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "day_of_week": 1,
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10,
      "notes": "Morning shift"
    }
  ],
  "failed_slots": [
    {
      "index": 1,
      "error": "Slot type must be 'offline' or 'online'"
    }
  ],
  "total_created": 1,
  "total_failed": 1
}
```

---

## Validation Rules

**Minimal validation only:**

1. **Doctor exists** and is active
2. **Clinic exists** and is active
3. **Slot type** must be "offline" or "online"
4. **Time format** must be HH:MM (24-hour)
5. **Date format** must be YYYY-MM-DD (if provided)
6. **Max patients** defaults to 1 if not provided

**No complex validation:**
- No overlap checking (handled by database constraints)
- No past date validation
- No time range validation (end_time > start_time)
- No day_of_week vs specific_date conflict checking

---

## Use Cases

### Use Case 1: Simple Weekly Schedule
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
      "end_time": "17:00",
      "max_patients": 20,
      "notes": "Wednesday Full Day"
    }
  ]
}
```

### Use Case 2: Mixed Days and Dates
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
      "specific_date": "2024-12-25",
      "slot_type": "online",
      "start_time": "10:00",
      "end_time": "14:00",
      "notes": "Christmas Day"
    },
    {
      "day_of_week": 5,
      "slot_type": "offline",
      "start_time": "08:00",
      "end_time": "12:00",
      "notes": "Friday Morning"
    }
  ]
}
```

### Use Case 3: UI Checkbox Implementation
```javascript
// When user selects multiple days and clicks "Add Slots"
const selectedDays = [1, 3, 5]; // Monday, Wednesday, Friday
const slots = [];

selectedDays.forEach(day => {
  // Add morning slot
  slots.push({
    day_of_week: day,
    slot_type: "offline",
    start_time: "09:00",
    end_time: "12:00",
    max_patients: 10,
    notes: "Morning shift"
  });
  
  // Add afternoon slot
  slots.push({
    day_of_week: day,
    slot_type: "offline",
    start_time: "14:00",
    end_time: "17:00",
    max_patients: 10,
    notes: "Afternoon shift"
  });
});

// Send simple request
fetch('/api/organizations/doctor-time-slots/simple', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    doctor_id: doctorId,
    clinic_id: clinicId,
    slots: slots
  })
});
```

---

## Error Handling

The API processes each slot individually and reports:
- **Successfully created slots** in `created_slots` array
- **Failed slots** in `failed_slots` array with error details

Common error scenarios:
- Invalid slot type (must be offline/online)
- Invalid time format (must be HH:MM)
- Invalid date format (must be YYYY-MM-DD)
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

## Frontend Integration Example

### React Component Example
```jsx
function SimpleTimeSlotManager({ doctorId, clinicId }) {
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

  const handleCreateSlots = async () => {
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
      const response = await fetch('/api/organizations/doctor-time-slots/simple', {
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
      
      <button onClick={handleCreateSlots} disabled={selectedDays.length === 0}>
        Create Time Slots ({selectedDays.length * 2} slots)
      </button>
    </div>
  );
}
```

---

## Testing Examples

### Test 1: Multiple Days
```bash
curl -X POST http://localhost:8081/api/organizations/doctor-time-slots/simple \
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
      }
    ]
  }'
```

### Test 2: Mixed Days and Dates
```bash
curl -X POST http://localhost:8081/api/organizations/doctor-time-slots/simple \
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

1. **Simple**: Minimal validation, easy to use
2. **Fast**: No complex business logic
3. **Flexible**: Accepts any valid slot data
4. **Frontend-friendly**: Perfect for UI implementations
5. **Database-driven**: Constraints handled by database
6. **Error-tolerant**: Partial success handling

---

## Differences from Complex APIs

| Feature | Simple API | Complex API |
|---------|------------|-------------|
| Validation | Minimal | Extensive |
| Overlap checking | Database only | API + Database |
| Past date validation | None | Yes |
| Time range validation | None | Yes |
| Business logic | None | Complex |
| Performance | Fast | Slower |
| Error handling | Basic | Detailed |

---

**Last Updated:** After implementation of simplified time slots API
