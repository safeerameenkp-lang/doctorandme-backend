# Doctor Time Slots UI Integration Guide

## Overview
Complete guide for integrating doctor time slots APIs with frontend UI components.

---

## API 1: Create Slots (Admin UI)

### Endpoint
**POST** `/api/organizations/doctor-time-slots/simple`

### Purpose
Create multiple time slots for a doctor at a clinic (weekly recurring slots).

---

## Request Format

### Request Body
```json
{
  "doctor_id": "uuid",
  "clinic_id": "uuid",
  "slots": [
    {
      "day_of_week": 1,           // 0=Sunday, 1=Monday, etc.
      "slot_type": "offline",     // "offline" or "online"
      "start_time": "09:00",      // HH:MM format
      "end_time": "12:00",        // HH:MM format
      "max_patients": 10,         // Optional, defaults to 1
      "notes": "Morning shift"    // Optional
    },
    {
      "day_of_week": 1,           // Monday
      "slot_type": "offline",
      "start_time": "14:00",      // Afternoon
      "end_time": "17:00",
      "max_patients": 10,
      "notes": "Afternoon shift"
    },
    {
      "day_of_week": 3,           // Wednesday
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "17:00",
      "max_patients": 20,
      "notes": "Full day"
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
| `day_of_week` | Integer | Yes | 0=Sunday, 1=Monday, ..., 6=Saturday |
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
  "message": "Slot creation completed. 3 created, 0 failed",
  "created_slots": [
    {
      "id": "slot-uuid-1",
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
      "id": "slot-uuid-2",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "day_of_week": 1,
      "slot_type": "offline",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 10,
      "notes": "Afternoon shift"
    },
    {
      "id": "slot-uuid-3",
      "doctor_id": "uuid",
      "clinic_id": "uuid",
      "day_of_week": 3,
      "slot_type": "offline",
      "start_time": "09:00",
      "end_time": "17:00",
      "max_patients": 20,
      "notes": "Full day"
    }
  ],
  "failed_slots": [],
  "total_created": 3,
  "total_failed": 0
}
```

---

## Frontend Integration (Admin UI)

### React Component for Creating Slots
```jsx
import React, { useState } from 'react';

function DoctorSlotManager({ doctorId, clinicId }) {
  const [selectedDays, setSelectedDays] = useState([]);
  const [morningStart, setMorningStart] = useState("09:00");
  const [morningEnd, setMorningEnd] = useState("12:00");
  const [afternoonStart, setAfternoonStart] = useState("14:00");
  const [afternoonEnd, setAfternoonEnd] = useState("17:00");
  const [slotType, setSlotType] = useState("offline");
  const [maxPatients, setMaxPatients] = useState(10);

  const dayNames = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];

  const handleDayToggle = (dayIndex) => {
    setSelectedDays(prev => 
      prev.includes(dayIndex) 
        ? prev.filter(d => d !== dayIndex)
        : [...prev, dayIndex]
    );
  };

  const createSlots = async () => {
    if (selectedDays.length === 0) {
      alert('Please select at least one day');
      return;
    }

    const slots = [];
    
    selectedDays.forEach(dayIndex => {
      // Add morning slot
      slots.push({
        day_of_week: dayIndex,
        slot_type: slotType,
        start_time: morningStart,
        end_time: morningEnd,
        max_patients: maxPatients,
        notes: "Morning shift"
      });
      
      // Add afternoon slot
      slots.push({
        day_of_week: dayIndex,
        slot_type: slotType,
        start_time: afternoonStart,
        end_time: afternoonEnd,
        max_patients: maxPatients,
        notes: "Afternoon shift"
      });
    });

    try {
      const response = await fetch('/api/organizations/doctor-time-slots/simple', {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({
          doctor_id: doctorId,
          clinic_id: clinicId,
          slots: slots
        })
      });

      const result = await response.json();
      
      if (result.total_created > 0) {
        alert(`Successfully created ${result.total_created} time slots for ${selectedDays.length} days!`);
        setSelectedDays([]); // Clear selection
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
    <div style={{ padding: '20px', maxWidth: '800px', margin: '0 auto' }}>
      <h2>Create Doctor Time Slots</h2>
      
      {/* Slot Type Selection */}
      <div style={{ marginBottom: '20px' }}>
        <label>
          Slot Type:
          <select 
            value={slotType} 
            onChange={(e) => setSlotType(e.target.value)}
            style={{ marginLeft: '10px', padding: '5px' }}
          >
            <option value="offline">Offline (In-person)</option>
            <option value="online">Online (Telemedicine)</option>
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
        onClick={createSlots}
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

      {/* Preview */}
      {selectedDays.length > 0 && (
        <div style={{ marginTop: '20px', padding: '15px', backgroundColor: '#f8f9fa', borderRadius: '4px' }}>
          <h4>Preview:</h4>
          <p>Will create {selectedDays.length * 2} slots for:</p>
          <ul>
            {selectedDays.map(dayIndex => (
              <li key={dayIndex}>
                {dayNames[dayIndex]}: {morningStart}-{morningEnd} & {afternoonStart}-{afternoonEnd}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}

export default DoctorSlotManager;
```

---

## API 2: List Slots (Patient UI)

### Endpoint
**GET** `/api/organizations/doctor-time-slots/categories?doctor_id=mmman-ggg-id&clinic_id=heasd-clinic-id&date=2024-10-13&slot_type=offline`

### Purpose
Get available time slots for appointment booking UI.

---

## Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `doctor_id` | UUID | Yes | ID of the doctor |
| `clinic_id` | UUID | No | ID of the clinic |
| `date` | String | No | Specific date (YYYY-MM-DD) |
| `slot_type` | String | No | "offline" or "online" |

---

## Response Format

### Success Response (200 OK)
```json
{
  "doctor_id": "mmman-ggg-id",
  "clinic_id": "heasd-clinic-id",
  "date": "2024-10-13",
  "days": [
    {
      "day_name": "Sunday",
      "day_of_week": 0,
      "night": [],
      "morning": [],
      "afternoon": [],
      "evening": [],
      "has_slots": false,
      "total_slots": 0,
      "is_available": false,
      "unavailable_message": "Doctor not available on this day"
    },
    {
      "day_name": "Monday",
      "day_of_week": 1,
      "night": [],
      "morning": [
        {
          "id": "slot-uuid-1",
          "doctor_id": "mmman-ggg-id",
          "clinic_id": "heasd-clinic-id",
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
      ],
      "afternoon": [
        {
          "id": "slot-uuid-2",
          "doctor_id": "mmman-ggg-id",
          "clinic_id": "heasd-clinic-id",
          "slot_type": "offline",
          "start_time": "14:00",
          "end_time": "17:00",
          "max_patients": 10,
          "capacity": 10,
          "notes": "Afternoon shift",
          "is_active": true,
          "created_at": "2024-01-01T00:00:00Z",
          "updated_at": "2024-01-01T00:00:00Z",
          "day_of_week": 1
        }
      ],
      "evening": [],
      "has_slots": true,
      "total_slots": 2,
      "is_available": true,
      "unavailable_message": ""
    },
    {
      "day_name": "Tuesday",
      "day_of_week": 2,
      "night": [],
      "morning": [],
      "afternoon": [],
      "evening": [],
      "has_slots": false,
      "total_slots": 0,
      "is_available": false,
      "unavailable_message": "Doctor not available on this day"
    }
    // ... other days
  ]
}
```

---

## Frontend Integration (Patient UI)

### React Component for Appointment Booking
```jsx
import React, { useState, useEffect } from 'react';

function AppointmentBooking({ doctorId, clinicId }) {
  const [selectedDate, setSelectedDate] = useState('');
  const [consultationType, setConsultationType] = useState('offline');
  const [slots, setSlots] = useState(null);
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
        date: selectedDate,
        slot_type: consultationType
      });

      const response = await fetch(`/api/organizations/doctor-time-slots/categories?${params}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });
      
      const data = await response.json();
      setSlots(data);
    } catch (error) {
      console.error('Error fetching slots:', error);
      alert('Error fetching available slots');
    } finally {
      setLoading(false);
    }
  };

  const renderTimeCategory = (category, slots) => {
    if (slots.length === 0) return null;

    return (
      <div className="time-category" style={{ marginBottom: '20px' }}>
        <h4 style={{ margin: '0 0 10px 0', color: '#333' }}>
          {category.charAt(0).toUpperCase() + category.slice(1)}
        </h4>
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
                {slot.slot_type === 'online' ? '🌐 Online' : '🏥 Offline'}
              </div>
              <div style={{ color: '#666', fontSize: '0.9em', marginBottom: '5px' }}>
                Capacity: {slot.capacity}
              </div>
              {slot.notes && (
                <div style={{ color: '#888', fontSize: '0.8em' }}>
                  {slot.notes}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    );
  };

  const bookAppointment = () => {
    if (!selectedSlot) {
      alert('Please select a time slot');
      return;
    }

    // Handle appointment booking
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
            <option value="offline">In-person</option>
            <option value="online">Online</option>
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
            min={new Date().toISOString().split('T')[0]} // Today's date
          />
        </label>
      </div>

      {/* Available Slots */}
      {selectedDate && (
        <div style={{ marginBottom: '20px' }}>
          <h3>Available Slots</h3>
          {loading ? (
            <div>Loading slots...</div>
          ) : slots ? (
            <div>
              {slots.days.map(day => (
                <div key={day.day_of_week} style={{ marginBottom: '30px' }}>
                  <h4 style={{ 
                    margin: '0 0 15px 0', 
                    padding: '10px',
                    backgroundColor: day.is_available ? '#e8f5e8' : '#ffe6e6',
                    color: day.is_available ? '#2d5a2d' : '#d32f2f',
                    borderRadius: '4px'
                  }}>
                    {day.day_name}
                    {day.has_slots && (
                      <span style={{ fontSize: '0.9em', fontWeight: 'normal' }}>
                        ({day.total_slots} slots)
                      </span>
                    )}
                  </h4>
                  
                  {day.is_available ? (
                    day.has_slots ? (
                      <div>
                        {renderTimeCategory('morning', day.morning)}
                        {renderTimeCategory('afternoon', day.afternoon)}
                        {renderTimeCategory('evening', day.evening)}
                        {renderTimeCategory('night', day.night)}
                      </div>
                    ) : (
                      <div style={{ color: '#666', fontStyle: 'italic', textAlign: 'center', padding: '20px' }}>
                        No time slots available for this date
                      </div>
                    )
                  ) : (
                    <div style={{ color: '#d32f2f', fontStyle: 'italic', textAlign: 'center', padding: '20px' }}>
                      {day.unavailable_message || 'Doctor not available on this day'}
                    </div>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <div>No slots available</div>
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

## Complete UI Flow

### 1. Admin UI (Create Slots)
```
Select Doctor → Select Clinic → Choose Days → Set Times → Create Slots
```

### 2. Patient UI (Book Appointments)
```
Select Consultation Type → Select Date → View Available Slots → Book Appointment
```

---

## Testing Examples

### Test 1: Create Slots
```bash
curl -X POST http://localhost:8081/api/organizations/doctor-time-slots/simple \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "doctor_id": "doctor-uuid",
    "clinic_id": "clinic-uuid",
    "slots": [
      {
        "day_of_week": 1,
        "slot_type": "offline",
        "start_time": "09:00",
        "end_time": "12:00",
        "max_patients": 10,
        "notes": "Morning shift"
      }
    ]
  }'
```

### Test 2: List Slots
```bash
curl -X GET "http://localhost:8081/api/organizations/doctor-time-slots/categories?doctor_id=doctor-uuid&clinic_id=clinic-uuid&date=2024-10-13&slot_type=offline" \
  -H "Authorization: Bearer your-token"
```

---

## Weekly Slot Creation & Unchecked Days Handling

### How It Works

#### 1. Admin Creates Weekly Slots
```javascript
// Admin selects days: Monday, Wednesday, Friday
const selectedDays = [1, 3, 5]; // Monday, Wednesday, Friday

// API creates slots for selected days only
const slots = selectedDays.map(day => ({
  day_of_week: day,
  slot_type: "offline",
  start_time: "09:00",
  end_time: "12:00"
}));
```

#### 2. Patient Views Available Slots
```javascript
// API returns all 7 days with availability status
{
  "days": [
    {
      "day_name": "Sunday",
      "is_available": false,
      "unavailable_message": "Doctor not available on this day"
    },
    {
      "day_name": "Monday", 
      "is_available": true,
      "has_slots": true,
      "morning": [...], // Available slots
      "afternoon": [...]
    },
    {
      "day_name": "Tuesday",
      "is_available": false,
      "unavailable_message": "Doctor not available on this day"
    }
    // ... continues for all 7 days
  ]
}
```

### UI Display Logic

#### Available Days (Green)
```jsx
// Days where doctor has slots
backgroundColor: '#e8f5e8',  // Green background
color: '#2d5a2d',            // Green text
```

#### Unavailable Days (Red)
```jsx
// Days where doctor doesn't work
backgroundColor: '#ffe6e6',  // Red background  
color: '#d32f2f',           // Red text
```

#### Messages
- **Available with slots**: Shows actual time slots
- **Available but no slots**: "No time slots available for this date"
- **Unavailable**: "Doctor not available on this day"

### Complete Flow Example

#### Step 1: Admin Setup
```
Admin selects: Monday ✓, Tuesday ✗, Wednesday ✓, Thursday ✗, Friday ✓
API creates: Slots for Monday, Wednesday, Friday only
```

#### Step 2: Patient Booking
```
Patient sees:
- Monday: Available slots (09:00-12:00, 14:00-17:00)
- Tuesday: "Doctor not available on this day" (Red)
- Wednesday: Available slots (09:00-12:00, 14:00-17:00)  
- Thursday: "Doctor not available on this day" (Red)
- Friday: Available slots (09:00-12:00, 14:00-17:00)
```

### Benefits
- ✅ **Clear availability**: Patients see which days doctor works
- ✅ **Visual distinction**: Green for available, red for unavailable
- ✅ **Informative messages**: Clear explanation for unavailable days
- ✅ **Flexible scheduling**: Admin can choose any combination of days
- ✅ **Weekly consistency**: Same times for selected days

---

## Key Features

### Admin UI Features
- ✅ Day selection (checkboxes)
- ✅ Time configuration (morning/afternoon)
- ✅ Slot type selection (offline/online)
- ✅ Max patients setting
- ✅ Preview before creation
- ✅ Bulk slot creation

### Patient UI Features
- ✅ Consultation type selection
- ✅ Date picker
- ✅ Time category display (morning/afternoon/evening/night)
- ✅ Slot selection
- ✅ Capacity information
- ✅ Empty state handling

---

**Last Updated:** Complete UI integration guide for doctor time slots APIs
