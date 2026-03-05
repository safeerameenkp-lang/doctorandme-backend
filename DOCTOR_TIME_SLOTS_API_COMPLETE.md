# Doctor Time Slots API - Complete Guide

## Overview
Complete CRUD API system for managing doctor time slots with support for weekly recurring slots and specific date slots.

---

## API Endpoints

### 1. Create Doctor Time Slots
**POST** `/api/organizations/doctor-time-slots`

#### Purpose
Create multiple weekly recurring time slots for a doctor at a clinic.

#### Request Body
```json
{
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "slot_type": "in-person",  // "in-person", "online", "video"
  "slots": [
    {
      "day_of_week": 1,           // 0=Sunday, 1=Monday, etc.
      "start_time": "09:00",      // HH:MM format
      "end_time": "12:00",        // HH:MM format
      "max_patients": 10,         // Optional, defaults to 1
      "notes": "Morning shift"    // Optional
    },
    {
      "day_of_week": 1,           // Monday
      "start_time": "14:00",      // Afternoon
      "end_time": "17:00",
      "max_patients": 10,
      "notes": "Afternoon shift"
    },
    {
      "day_of_week": 3,           // Wednesday
      "start_time": "10:00",
      "end_time": "11:00",
      "max_patients": 5,
      "notes": "Consultation hour"
    }
  ]
}
```

#### Response (201 Created)
```json
{
  "message": "Slot creation completed. 3 created, 0 failed",
  "created_slots": [
    {
      "id": "slot-uuid-1",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "day_of_week": 1,
      "day_name": "Monday",
      "slot_type": "in-person",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10,
      "notes": "Morning shift",
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "failed_slots": [],
  "total_created": 3,
  "total_failed": 0
}
```

---

### 2. List Doctor Time Slots
**GET** `/api/organizations/doctor-time-slots`

#### Purpose
List time slots with filtering by doctor, clinic, slot type, and date. Includes availability information based on booked appointments.

#### Query Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `doctor_id` | UUID | Yes | ID of the doctor |
| `clinic_id` | UUID | No | ID of the clinic |
| `slot_type` | String | No | "in-person", "online", "video" |
| `date` | String | No | Specific date (YYYY-MM-DD) - filters by day of week |

#### Example Requests
```bash
# List all slots for a doctor
GET /api/organizations/doctor-time-slots?doctor_id=uuid

# List slots for a specific doctor, clinic, and slot type
GET /api/organizations/doctor-time-slots?doctor_id=uuid&clinic_id=uuid&slot_type=in-person

# List slots for a specific date
GET /api/organizations/doctor-time-slots?doctor_id=uuid&date=2024-10-15

# List slots for a specific doctor, clinic, slot type, and date
GET /api/organizations/doctor-time-slots?doctor_id=uuid&clinic_id=uuid&slot_type=in-person&date=2024-10-15
```

#### Response (200 OK)
```json
{
  "slots": [
    {
      "id": "slot-uuid-1",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "day_of_week": 1,
      "day_name": "Monday",
      "slot_type": "in-person",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10,
      "booked_patients": 3,
      "available_spots": 7,
      "is_available": true,
      "status": "available",
      "notes": "Morning shift",
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": "slot-uuid-2",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "day_of_week": 3,
      "day_name": "Wednesday",
      "slot_type": "in-person",
      "start_time": "10:00",
      "end_time": "11:00",
      "max_patients": 5,
      "booked_patients": 5,
      "available_spots": 0,
      "is_available": false,
      "status": "booking_full",
      "notes": "Consultation hour",
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 2,
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "slot_type": "in-person",
  "date": "2024-10-15"
}
```

---

### 3. List Doctor Time Slots Grouped by Day
**GET** `/api/organizations/doctor-time-slots/grouped`

#### Purpose
List time slots grouped by day with availability information. Shows all 7 days of the week, including unavailable days.

#### Query Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `doctor_id` | UUID | Yes | ID of the doctor |
| `clinic_id` | UUID | No | ID of the clinic |
| `slot_type` | String | No | "in-person", "online", "video" |
| `date` | String | No | Specific date (YYYY-MM-DD) - filters by day of week |

#### Example Request
```bash
GET /api/organizations/doctor-time-slots/grouped?doctor_id=uuid&clinic_id=uuid&slot_type=in-person&date=2024-10-15
```

#### Response (200 OK)
```json
{
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "slot_type": "in-person",
  "date": "2024-10-15",
  "days": [
    {
      "day_name": "Sunday",
      "day_of_week": 0,
      "slots": [],
      "has_slots": false,
      "total_slots": 0,
      "available_slots": 0,
      "is_available": false,
      "status": "unavailable"
    },
    {
      "day_name": "Monday",
      "day_of_week": 1,
      "slots": [
        {
          "id": "slot-uuid-1",
          "doctor_id": "uuid",
          "clinic_id": "uuid",
          "day_of_week": 1,
          "day_name": "Monday",
          "slot_type": "in-person",
          "start_time": "09:00",
          "end_time": "12:00",
          "max_patients": 10,
          "booked_patients": 3,
          "available_spots": 7,
          "is_available": true,
          "status": "available",
          "notes": "Morning shift",
          "is_active": true,
          "created_at": "2024-01-01T00:00:00Z",
          "updated_at": "2024-01-01T00:00:00Z"
        }
      ],
      "has_slots": true,
      "total_slots": 1,
      "available_slots": 1,
      "is_available": true,
      "status": "available"
    },
    {
      "day_name": "Tuesday",
      "day_of_week": 2,
      "slots": [],
      "has_slots": false,
      "total_slots": 0,
      "available_slots": 0,
      "is_available": false,
      "status": "unavailable"
    }
  ]
}
```

---

### 4. Get Single Time Slot
**GET** `/api/organizations/doctor-time-slots/:id`

#### Purpose
Get details of a single time slot by ID.

#### Response (200 OK)
```json
{
  "slot": {
    "id": "slot-uuid-1",
    "doctor_id": "uuid",
    "clinic_id": "uuid",
    "day_of_week": 1,
    "day_name": "Monday",
    "slot_type": "in-person",
    "start_time": "09:00",
    "end_time": "12:00",
    "max_patients": 10,
    "notes": "Morning shift",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 5. Update Time Slot
**PUT** `/api/organizations/doctor-time-slots/:id`

#### Purpose
Update an existing time slot. All fields are optional.

#### Request Body
```json
{
  "slot_type": "online",        // Optional
  "start_time": "10:00",        // Optional
  "end_time": "13:00",          // Optional
  "max_patients": 15,           // Optional
  "notes": "Updated shift",      // Optional
  "is_active": true             // Optional
}
```

#### Response (200 OK)
```json
{
  "message": "Time slot updated successfully",
    "slot": {
      "id": "slot-uuid-1",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "day_of_week": 1,
      "day_name": "Monday",
      "slot_type": "online",
      "start_time": "10:00",
      "end_time": "13:00",
      "max_patients": 15,
      "notes": "Updated shift",
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
}
```

---

### 6. Delete Time Slot
**DELETE** `/api/organizations/doctor-time-slots/:id`

#### Purpose
Soft delete a time slot by setting `is_active` to false.

#### Response (200 OK)
```json
{
  "message": "Time slot deleted successfully",
  "slot_id": "slot-uuid-1"
}
```

---

## Frontend Integration Examples

### 1. Weekly Slot Creation (Admin UI)
```jsx
import React, { useState } from 'react';

function WeeklySlotManager({ doctorId, clinicId }) {
  const [selectedDays, setSelectedDays] = useState([]);
  const [morningStart, setMorningStart] = useState("09:00");
  const [morningEnd, setMorningEnd] = useState("12:00");
  const [afternoonStart, setAfternoonStart] = useState("14:00");
  const [afternoonEnd, setAfternoonEnd] = useState("17:00");
  const [slotType, setSlotType] = useState("in-person");
  const [maxPatients, setMaxPatients] = useState(10);

  const dayNames = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];

  const handleDayToggle = (dayIndex) => {
    setSelectedDays(prev => 
      prev.includes(dayIndex) 
        ? prev.filter(d => d !== dayIndex)
        : [...prev, dayIndex]
    );
  };

  const createWeeklySlots = async () => {
    if (selectedDays.length === 0) {
      alert('Please select at least one day');
      return;
    }

    const slots = [];
    
    selectedDays.forEach(dayIndex => {
      // Add morning slot
      slots.push({
        day_of_week: dayIndex,
        start_time: morningStart,
        end_time: morningEnd,
        max_patients: maxPatients,
        notes: "Morning shift"
      });
      
      // Add afternoon slot
      slots.push({
        day_of_week: dayIndex,
        start_time: afternoonStart,
        end_time: afternoonEnd,
        max_patients: maxPatients,
        notes: "Afternoon shift"
      });
    });

    try {
      const response = await fetch('/api/organizations/doctor-time-slots', {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({
          doctor_id: doctorId,
          clinic_id: clinicId,
          slot_type: slotType,
          slots: slots
        })
      });

      const result = await response.json();
      
      if (result.total_created > 0) {
        alert(`Successfully created ${result.total_created} time slots for ${selectedDays.length} days!`);
        setSelectedDays([]);
      } else {
        alert('Failed to create time slots');
        console.error('All slots failed:', result.failed_slots);
      }
    } catch (error) {
      console.error('Error creating time slots:', error);
      alert('Error creating time slots');
    }
  };

  return (
    <div style={{ padding: '20px', maxWidth: '800px', margin: '0 auto' }}>
      <h2>Weekly Consultation Timing</h2>
      
      {/* Slot Type Selection */}
      <div style={{ marginBottom: '20px' }}>
        <label>
          Consultation Type:
          <select 
            value={slotType} 
            onChange={(e) => setSlotType(e.target.value)}
            style={{ marginLeft: '10px', padding: '5px' }}
          >
            <option value="in-person">In-person</option>
            <option value="online">Online</option>
            <option value="video">Video</option>
          </select>
        </label>
      </div>

      {/* Day Selection */}
      <div style={{ marginBottom: '20px' }}>
        <h3>Select Days</h3>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: '10px' }}>
          {dayNames.map((day, index) => (
            <label key={day} style={{ display: 'flex', alignItems: 'center' }}>
              <input
                type="checkbox"
                checked={selectedDays.includes(index)}
                onChange={() => handleDayToggle(index)}
                style={{ marginRight: '5px' }}
              />
              {day}
            </label>
          ))}
        </div>
      </div>

      {/* Time Configuration */}
      <div style={{ marginBottom: '20px' }}>
        <h3>Time Configuration</h3>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '20px' }}>
          <div>
            <h4>Morning Shift</h4>
            <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
              <label>
                Start:
                <input
                  type="time"
                  value={morningStart}
                  onChange={(e) => setMorningStart(e.target.value)}
                  style={{ marginLeft: '5px' }}
                />
              </label>
              <label>
                End:
                <input
                  type="time"
                  value={morningEnd}
                  onChange={(e) => setMorningEnd(e.target.value)}
                  style={{ marginLeft: '5px' }}
                />
              </label>
            </div>
          </div>
          
          <div>
            <h4>Afternoon Shift</h4>
            <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
              <label>
                Start:
                <input
                  type="time"
                  value={afternoonStart}
                  onChange={(e) => setAfternoonStart(e.target.value)}
                  style={{ marginLeft: '5px' }}
                />
              </label>
              <label>
                End:
                <input
                  type="time"
                  value={afternoonEnd}
                  onChange={(e) => setAfternoonEnd(e.target.value)}
                  style={{ marginLeft: '5px' }}
                />
              </label>
            </div>
          </div>
        </div>
      </div>

      {/* Max Patients */}
      <div style={{ marginBottom: '20px' }}>
        <label>
          Max Patients per Slot:
          <input
            type="number"
            value={maxPatients}
            onChange={(e) => setMaxPatients(parseInt(e.target.value))}
            min="1"
            style={{ marginLeft: '10px', padding: '5px' }}
          />
        </label>
      </div>

      {/* Create Button */}
      <button
        onClick={createWeeklySlots}
        disabled={selectedDays.length === 0}
        style={{
          backgroundColor: selectedDays.length > 0 ? '#007bff' : '#ccc',
          color: 'white',
          padding: '12px 24px',
          border: 'none',
          borderRadius: '4px',
          cursor: selectedDays.length > 0 ? 'pointer' : 'not-allowed',
          fontSize: '16px'
        }}
      >
        Create {selectedDays.length * 2} Slots ({selectedDays.length} days × 2 shifts)
      </button>
    </div>
  );
}

export default WeeklySlotManager;
```

### 2. Slot List & Booking (Patient UI)
```jsx
import React, { useState, useEffect } from 'react';

function AppointmentBooking({ doctorId, clinicId }) {
  const [selectedDate, setSelectedDate] = useState('');
  const [consultationType, setConsultationType] = useState('in-person');
  const [slots, setSlots] = useState([]);
  const [loading, setLoading] = useState(false);
  const [selectedSlot, setSelectedSlot] = useState(null);

  // Fetch slots when date or consultation type changes
  useEffect(() => {
    if (selectedDate) {
      fetchSlots();
    }
  }, [selectedDate, consultationType]);

  const fetchSlots = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        doctor_id: doctorId,
        clinic_id: clinicId,
        slot_type: consultationType,
        date: selectedDate
      });

      const response = await fetch(`/api/organizations/doctor-time-slots?${params}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });
      
      const data = await response.json();
      setSlots(data.slots || []);
    } catch (error) {
      console.error('Error fetching slots:', error);
      alert('Error fetching available slots');
    } finally {
      setLoading(false);
    }
  };

  const bookAppointment = () => {
    if (!selectedSlot) {
      alert('Please select a time slot');
      return;
    }

    console.log('Booking appointment:', selectedSlot);
    alert(`Appointment booked for ${selectedDate} at ${selectedSlot.start_time} - ${selectedSlot.end_time}`);
  };

  return (
    <div style={{ padding: '20px', maxWidth: '800px', margin: '0 auto' }}>
      <h2>Book Appointment</h2>
      
      {/* Consultation Type Selection */}
      <div style={{ marginBottom: '20px' }}>
        <label>
          Consultation Type:
          <select 
            value={consultationType} 
            onChange={(e) => setConsultationType(e.target.value)}
            style={{ marginLeft: '10px', padding: '5px' }}
          >
            <option value="in-person">In-person</option>
            <option value="online">Online</option>
            <option value="video">Video</option>
          </select>
        </label>
      </div>

      {/* Date Selection */}
      <div style={{ marginBottom: '20px' }}>
        <label>
          Select Date:
          <input
            type="date"
            value={selectedDate}
            onChange={(e) => setSelectedDate(e.target.value)}
            style={{ marginLeft: '10px', padding: '5px' }}
            min={new Date().toISOString().split('T')[0]}
          />
        </label>
      </div>

      {/* Available Slots */}
      {selectedDate && (
        <div style={{ marginBottom: '20px' }}>
          <h3>Available Slots</h3>
          {loading ? (
            <div>Loading slots...</div>
          ) : slots.length > 0 ? (
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: '10px' }}>
              {slots.map(slot => (
                <div 
                  key={slot.id} 
                  className={`slot-card ${selectedSlot?.id === slot.id ? 'selected' : ''}`}
                  onClick={() => setSelectedSlot(slot)}
                  style={{
                    border: selectedSlot?.id === slot.id ? '2px solid #007bff' : '1px solid #ddd',
                    borderRadius: '6px',
                    padding: '12px',
                    backgroundColor: '#fff',
                    cursor: 'pointer',
                    transition: 'all 0.2s'
                  }}
                >
                  <div style={{ fontWeight: 'bold', color: '#2c5aa0', marginBottom: '5px' }}>
                    {slot.start_time} - {slot.end_time}
                  </div>
                  <div style={{ color: '#666', fontSize: '0.9em', marginBottom: '5px' }}>
                    {slot.slot_type === 'online' ? '🌐 Online' : 
                     slot.slot_type === 'video' ? '📹 Video' : '🏥 In-person'}
                  </div>
                  <div style={{ color: '#666', fontSize: '0.9em', marginBottom: '5px' }}>
                    Capacity: {slot.max_patients}
                  </div>
                  {slot.notes && (
                    <div style={{ color: '#888', fontSize: '0.8em' }}>
                      {slot.notes}
                    </div>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <div style={{ color: '#666', fontStyle: 'italic', textAlign: 'center', padding: '20px' }}>
              No time slots available for this date
            </div>
          )}
        </div>
      )}

      {/* Book Appointment Button */}
      {selectedSlot && (
        <div style={{ marginTop: '20px' }}>
          <button
            onClick={bookAppointment}
            style={{
              backgroundColor: '#28a745',
              color: 'white',
              padding: '12px 24px',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
              fontSize: '16px'
            }}
          >
            Book Appointment - {selectedSlot.start_time} to {selectedSlot.end_time}
          </button>
        </div>
      )}
    </div>
  );
}

export default AppointmentBooking;
```

---

## Testing Examples

### 1. Create Weekly Slots
```bash
curl -X POST http://localhost:8081/api/organizations/doctor-time-slots \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "doctor_id": "doctor-uuid",
    "clinic_id": "clinic-uuid",
    "slot_type": "in-person",
    "slots": [
      {
        "day_of_week": 1,
        "start_time": "09:00",
        "end_time": "12:00",
        "max_patients": 10,
        "notes": "Morning shift"
      },
      {
        "day_of_week": 1,
        "start_time": "14:00",
        "end_time": "17:00",
        "max_patients": 10,
        "notes": "Afternoon shift"
      }
    ]
  }'
```

### 2. List Slots for Specific Date
```bash
curl -X GET "http://localhost:8081/api/organizations/doctor-time-slots?doctor_id=doctor-uuid&clinic_id=clinic-uuid&slot_type=in-person&date=2024-10-15" \
  -H "Authorization: Bearer your-token"
```

### 3. Update Slot
```bash
curl -X PUT http://localhost:8081/api/organizations/doctor-time-slots/slot-uuid \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "start_time": "10:00",
    "end_time": "13:00",
    "max_patients": 15
  }'
```

### 4. Delete Slot
```bash
curl -X DELETE http://localhost:8081/api/organizations/doctor-time-slots/slot-uuid \
  -H "Authorization: Bearer your-token"
```

---

## Key Features

### ✅ Complete CRUD Operations
- **Create**: Multiple slots in one request
- **Read**: List with filtering, Get single slot
- **Update**: Partial updates for any field
- **Delete**: Soft delete (sets is_active = false)

### ✅ Flexible Slot Types
- **in-person**: Physical consultation
- **online**: Online consultation
- **video**: Video call consultation

### ✅ Weekly Recurring Slots
- **Weekly recurring**: Using `day_of_week` (0-6)
- **Day names**: Automatic conversion (0=Sunday, 1=Monday, etc.)

### ✅ Advanced Filtering
- Filter by doctor, clinic, slot type, and date
- Perfect for appointment booking UI

### ✅ Validation & Security
- UUID validation for all IDs
- Clinic-doctor link validation
- Time format validation
- Role-based access control

### ✅ Error Handling
- Detailed error messages
- Partial success reporting for bulk operations
- Proper HTTP status codes

---

**Last Updated:** Complete doctor time slots API system with CRUD operations
