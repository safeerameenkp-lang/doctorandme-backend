# Doctor Time Slots API - Enhanced with Specific Date Support

## Overview
The Doctor Time Slots system now supports **two types of slots**:
1. **Recurring Weekly Slots** - Slots that repeat every week on specific days (e.g., every Monday, every Thursday)
2. **Specific Date Slots** - One-time slots for particular dates (e.g., December 25, 2024)

This allows flexible scheduling where doctors can:
- Set regular weekly availability
- Add special slots for specific dates
- Override recurring slots with date-specific availability

---

## API Endpoints

### 1. Create Doctor Time Slot
**POST** `/api/v1/time-slots`

Create a time slot for a doctor at a specific clinic. Can be either recurring (weekly) or date-specific.

#### Request Body

**For Recurring Weekly Slots:**
```json
{
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "day_of_week": 1,           // 0=Sunday, 1=Monday, ..., 6=Saturday
  "slot_type": "offline",     // "offline" or "online"
  "start_time": "09:00",      // HH:MM format (24-hour)
  "end_time": "12:00",        // HH:MM format (24-hour)
  "max_patients": 5,          // Optional, defaults to 1
  "notes": "Morning shift"    // Optional
}
```

**For Specific Date Slots:**
```json
{
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "specific_date": "2024-12-25",  // YYYY-MM-DD format
  "slot_type": "online",          // "offline" or "online"
  "start_time": "14:00",          // HH:MM format (24-hour)
  "end_time": "17:00",            // HH:MM format (24-hour)
  "max_patients": 3,              // Optional, defaults to 1
  "notes": "Holiday special hours" // Optional
}
```

#### Validation Rules
- **Must provide EITHER** `day_of_week` OR `specific_date`, but **not both**
- `day_of_week` must be 0-6 (0=Sunday, 6=Saturday)
- `specific_date` must be in YYYY-MM-DD format and cannot be in the past
- `slot_type` must be "offline" or "online"
- `end_time` must be after `start_time`
- `max_patients` must be at least 1
- System prevents overlapping time slots for the same doctor

#### Success Response (201 Created)
```json
{
  "message": "Time slot created successfully",
  "slot_id": "uuid",
  "slot": {
    "id": "uuid",
    "doctor_id": "uuid",
    "clinic_id": "uuid",
    "day_of_week": 1,           // OR
    "specific_date": "2024-12-25",
    "slot_type": "offline",
    "start_time": "09:00",
    "end_time": "12:00",
    "max_patients": 5,
    "notes": "Morning shift"
  }
}
```

#### Error Responses
- **400 Bad Request**: Invalid input or validation error
- **404 Not Found**: Doctor or clinic not found
- **409 Conflict**: Overlapping time slot exists

---

### 2. List Doctor Time Slots
**GET** `/api/v1/time-slots`

List time slots with various filters.

#### Query Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `doctor_id` | UUID | Filter by doctor |
| `clinic_id` | UUID | Filter by clinic |
| `day_of_week` | 0-6 | Filter by day (0=Sunday, 6=Saturday) |
| `specific_date` | YYYY-MM-DD | Filter by specific date |
| `slot_type` | string | Filter by "offline" or "online" |
| `only_active` | boolean | Show only active slots (default: true) |

#### Example Requests

**Get all Monday slots for a doctor:**
```
GET /api/v1/time-slots?doctor_id=xxx&day_of_week=1
```

**Get slots for a specific date:**
```
GET /api/v1/time-slots?doctor_id=xxx&specific_date=2024-12-25
```

**Get all slots for a doctor at a clinic:**
```
GET /api/v1/time-slots?doctor_id=xxx&clinic_id=yyy
```

#### Success Response (200 OK)
```json
{
  "time_slots": [
    {
      "id": "uuid",
      "doctor_id": "uuid",
      "doctor_name": "Dr. John Doe",
      "clinic_id": "uuid",
      "clinic_name": "City Hospital",
      "day_of_week": 1,
      "day_name": "Monday",
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "12:00",
      "is_active": true,
      "max_patients": 5,
      "notes": "Morning shift",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": "uuid",
      "doctor_id": "uuid",
      "doctor_name": "Dr. John Doe",
      "clinic_id": "uuid",
      "clinic_name": "City Hospital",
      "specific_date": "2024-12-25",
      "slot_type": "online",
      "start_time": "14:00",
      "end_time": "17:00",
      "is_active": true,
      "max_patients": 3,
      "notes": "Holiday hours",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total_count": 2
}
```

---

### 3. Get Available Time Slots (For Appointments)
**GET** `/api/v1/doctors/:doctor_id/available-slots`

Get available time slots for a specific doctor. Used by the appointment booking system.

#### Query Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `clinic_id` | UUID | Filter by clinic |
| `day_of_week` | 0-6 | Filter by day of week |
| `specific_date` | YYYY-MM-DD | Filter by specific date |
| `slot_type` | string | Filter by "offline" or "online" |

#### Example Requests

**Get available slots for a doctor on Mondays:**
```
GET /api/v1/doctors/xxx/available-slots?day_of_week=1
```

**Get available slots for December 25, 2024:**
```
GET /api/v1/doctors/xxx/available-slots?specific_date=2024-12-25
```

**Get online consultation slots:**
```
GET /api/v1/doctors/xxx/available-slots?slot_type=online
```

#### Success Response (200 OK)
```json
{
  "doctor_id": "uuid",
  "available_slots": [
    {
      "id": "uuid",
      "clinic_id": "uuid",
      "clinic_name": "City Hospital",
      "day_of_week": 1,
      "day_name": "Monday",
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 5,
      "notes": "Morning shift"
    },
    {
      "id": "uuid",
      "clinic_id": "uuid",
      "clinic_name": "City Hospital",
      "specific_date": "2024-12-25",
      "slot_type": "online",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 3,
      "notes": "Holiday hours"
    }
  ],
  "total_count": 2
}
```

---

### 4. Update Doctor Time Slot
**PUT** `/api/v1/time-slots/:id`

Update an existing time slot. Can update times, type, status, etc.

#### Request Body (All fields optional)
```json
{
  "start_time": "10:00",
  "end_time": "13:00",
  "slot_type": "online",
  "is_active": true,
  "max_patients": 10,
  "notes": "Updated shift"
}
```

**Note:** You cannot change `day_of_week` or `specific_date` after creation. To change the day/date, delete and create a new slot.

#### Success Response (200 OK)
```json
{
  "message": "Time slot updated successfully"
}
```

#### Error Responses
- **400 Bad Request**: Invalid input
- **404 Not Found**: Slot not found
- **409 Conflict**: Update causes overlap with another slot

---

### 5. Delete Doctor Time Slot
**DELETE** `/api/v1/time-slots/:id`

Delete a time slot.

#### Success Response (200 OK)
```json
{
  "message": "Time slot deleted successfully"
}
```

#### Error Response
- **404 Not Found**: Slot not found

---

## Use Cases

### Use Case 1: Regular Weekly Schedule
A doctor works at a clinic every Monday and Thursday from 9 AM to 5 PM.

**Create Monday slot:**
```json
POST /api/v1/time-slots
{
  "doctor_id": "xxx",
  "clinic_id": "yyy",
  "day_of_week": 1,
  "slot_type": "offline",
  "start_time": "09:00",
  "end_time": "17:00",
  "max_patients": 20
}
```

**Create Thursday slot:**
```json
POST /api/v1/time-slots
{
  "doctor_id": "xxx",
  "clinic_id": "yyy",
  "day_of_week": 4,
  "slot_type": "offline",
  "start_time": "09:00",
  "end_time": "17:00",
  "max_patients": 20
}
```

---

### Use Case 2: Special Date-Specific Availability
A doctor wants to add special hours on December 25, 2024 for emergencies only.

```json
POST /api/v1/time-slots
{
  "doctor_id": "xxx",
  "clinic_id": "yyy",
  "specific_date": "2024-12-25",
  "slot_type": "online",
  "start_time": "10:00",
  "end_time": "14:00",
  "max_patients": 5,
  "notes": "Holiday emergency consultations - online only"
}
```

---

### Use Case 3: UI Date Picker Implementation

#### Frontend Flow:
1. **User selects a date from date picker** (e.g., December 25, 2024)
2. **UI makes API call to check existing slots:**
   ```
   GET /api/v1/time-slots?doctor_id=xxx&clinic_id=yyy&specific_date=2024-12-25
   ```
3. **If no slots exist, show "Add Slot" button**
4. **User clicks "Add Slot" and fills in:**
   - Start time (e.g., 09:00)
   - End time (e.g., 12:00)
   - Slot type (offline/online)
   - Max patients
   - Notes (optional)
5. **UI sends create request:**
   ```json
   POST /api/v1/time-slots
   {
     "doctor_id": "xxx",
     "clinic_id": "yyy",
     "specific_date": "2024-12-25",
     "slot_type": "offline",
     "start_time": "09:00",
     "end_time": "12:00",
     "max_patients": 10
   }
   ```
6. **UI refreshes the slot list for that date**

---

### Use Case 4: Combining Recurring and Specific Slots

**Scenario:** Doctor has regular Monday slots, but on Monday, January 1, 2024 (New Year), they want different hours.

1. **Regular Monday slot (recurring):**
   ```json
   {
     "day_of_week": 1,
     "start_time": "09:00",
     "end_time": "17:00"
   }
   ```

2. **Special New Year Monday (specific date overrides recurring):**
   ```json
   {
     "specific_date": "2024-01-01",
     "start_time": "10:00",
     "end_time": "13:00"
   }
   ```

**Result:** On January 1, 2024 (which is a Monday), both slots will be available in the system, but the appointment booking logic should prioritize specific date slots over recurring ones.

---

## Database Schema

```sql
CREATE TABLE doctor_time_slots (
    id UUID PRIMARY KEY,
    doctor_id UUID NOT NULL,
    clinic_id UUID NOT NULL,
    day_of_week INTEGER,              -- For recurring: 0-6
    specific_date DATE,                -- For one-time slots
    slot_type VARCHAR(20) NOT NULL,    -- 'offline' or 'online'
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    max_patients INTEGER DEFAULT 1,
    notes TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    
    -- Constraint: Either day_of_week OR specific_date must be set
    CONSTRAINT valid_slot_type_constraint CHECK (
        (day_of_week IS NOT NULL AND specific_date IS NULL) OR 
        (day_of_week IS NULL AND specific_date IS NOT NULL)
    )
);
```

---

## Important Notes

1. **Overlap Prevention:** The system automatically prevents a doctor from having overlapping time slots, even across different clinics. A doctor cannot be in two places at once.

2. **Recurring vs Specific:** 
   - Recurring slots (day_of_week) repeat every week
   - Specific slots (specific_date) are one-time only
   - You must choose one type; cannot use both in the same slot

3. **Past Dates:** Cannot create specific date slots for dates in the past.

4. **Active vs Inactive:** Slots can be deactivated without deletion using the `is_active` flag.

5. **Clinic Linking:** Only doctors who are actively linked to a clinic can have time slots at that clinic.

---

## UI Implementation Guidelines

### Date Picker Component
```jsx
// Example React component structure
function TimeSlotManager({ doctorId, clinicId }) {
  const [selectedDate, setSelectedDate] = useState(null);
  const [slots, setSlots] = useState([]);

  const handleDateSelect = async (date) => {
    setSelectedDate(date);
    // Fetch slots for this specific date
    const response = await fetch(
      `/api/v1/time-slots?doctor_id=${doctorId}&clinic_id=${clinicId}&specific_date=${date}`
    );
    const data = await response.json();
    setSlots(data.time_slots);
  };

  const handleAddSlot = async (slotData) => {
    await fetch('/api/v1/time-slots', {
      method: 'POST',
      body: JSON.stringify({
        doctor_id: doctorId,
        clinic_id: clinicId,
        specific_date: selectedDate,
        ...slotData
      })
    });
    // Refresh slots
    handleDateSelect(selectedDate);
  };

  return (
    <div>
      <DatePicker onChange={handleDateSelect} />
      {selectedDate && (
        <div>
          <h3>Slots for {selectedDate}</h3>
          <SlotList slots={slots} />
          <AddSlotButton onClick={handleAddSlot} />
        </div>
      )}
    </div>
  );
}
```

---

## Migration Instructions

Run the migration file `009_add_specific_date_to_time_slots.sql` to add specific date support to your existing system.

This migration:
1. Adds `specific_date` column
2. Adds constraint ensuring either day_of_week OR specific_date is used
3. Updates indexes for performance
4. Updates overlap checking to handle both types of slots

---

## Testing Examples

### Test 1: Create Recurring Slot
```bash
curl -X POST http://localhost:8080/api/v1/time-slots \
  -H "Content-Type: application/json" \
  -d '{
    "doctor_id": "your-doctor-id",
    "clinic_id": "your-clinic-id",
    "day_of_week": 1,
    "slot_type": "offline",
    "start_time": "09:00",
    "end_time": "12:00"
  }'
```

### Test 2: Create Specific Date Slot
```bash
curl -X POST http://localhost:8080/api/v1/time-slots \
  -H "Content-Type: application/json" \
  -d '{
    "doctor_id": "your-doctor-id",
    "clinic_id": "your-clinic-id",
    "specific_date": "2024-12-25",
    "slot_type": "online",
    "start_time": "14:00",
    "end_time": "17:00"
  }'
```

### Test 3: List Specific Date Slots
```bash
curl http://localhost:8080/api/v1/time-slots?specific_date=2024-12-25
```

---

## Support

For issues or questions, refer to the main API documentation or contact the development team.

