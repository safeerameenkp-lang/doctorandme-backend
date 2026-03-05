# Slot Availability Features - Implementation Summary

## Overview
Enhanced the Doctor Time Slots API to include real-time availability checking based on booked appointments. Patients can now see which slots are available, booking full, or unavailable.

---

## Key Features Added

### 1. Enhanced Slot Response Structure
```json
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
  "booked_patients": 3,        // NEW: Number of patients who booked
  "available_spots": 7,         // NEW: Available spots remaining
  "is_available": true,         // NEW: True if slots available
  "status": "available",       // NEW: "available", "booking_full", "unavailable"
  "notes": "Morning shift",
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 2. Availability Status Types
- **`available`**: Slots available for booking
- **`booking_full`**: All slots booked (max_patients reached)
- **`unavailable`**: Doctor doesn't work on this day

### 3. Real-time Appointment Counting
- Counts confirmed and completed appointments
- Filters by specific date for accurate availability
- Updates availability in real-time

---

## API Endpoints

### 1. List Slots with Availability
**GET** `/api/organizations/doctor-time-slots`

#### Features
- Shows availability for each slot
- Filters by date to show specific day availability
- Includes booked vs available patient counts

#### Example Request
```bash
GET /api/organizations/doctor-time-slots?doctor_id=uuid&clinic_id=uuid&slot_type=in-person&date=2024-10-15
```

#### Response
```json
{
  "slots": [
    {
      "id": "slot-uuid-1",
      "day_of_week": 1,
      "day_name": "Monday",
      "start_time": "09:00",
      "end_time": "12:00",
      "max_patients": 10,
      "booked_patients": 3,
      "available_spots": 7,
      "is_available": true,
      "status": "available"
    },
    {
      "id": "slot-uuid-2",
      "day_of_week": 1,
      "day_name": "Monday",
      "start_time": "14:00",
      "end_time": "17:00",
      "max_patients": 5,
      "booked_patients": 5,
      "available_spots": 0,
      "is_available": false,
      "status": "booking_full"
    }
  ],
  "total": 2
}
```

### 2. List Slots Grouped by Day
**GET** `/api/organizations/doctor-time-slots/grouped`

#### Features
- Shows all 7 days of the week
- Includes unavailable days
- Groups slots by day
- Shows day-level availability status

#### Example Request
```bash
GET /api/organizations/doctor-time-slots/grouped?doctor_id=uuid&clinic_id=uuid&slot_type=in-person&date=2024-10-15
```

#### Response
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
          "start_time": "09:00",
          "end_time": "12:00",
          "max_patients": 10,
          "booked_patients": 3,
          "available_spots": 7,
          "is_available": true,
          "status": "available"
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

## Frontend Integration Examples

### 1. Patient Booking UI
```jsx
import React, { useState, useEffect } from 'react';

function AppointmentBooking({ doctorId, clinicId }) {
  const [selectedDate, setSelectedDate] = useState('');
  const [consultationType, setConsultationType] = useState('in-person');
  const [slots, setSlots] = useState([]);
  const [loading, setLoading] = useState(false);

  const fetchSlots = async () => {
    if (!selectedDate) return;
    
    setLoading(true);
    try {
      const params = new URLSearchParams({
        doctor_id: doctorId,
        clinic_id: clinicId,
        slot_type: consultationType,
        date: selectedDate
      });

      const response = await fetch(`/api/organizations/doctor-time-slots?${params}`);
      const data = await response.json();
      setSlots(data.slots || []);
    } catch (error) {
      console.error('Error fetching slots:', error);
    } finally {
      setLoading(false);
    }
  };

  const getSlotStatusColor = (status) => {
    switch (status) {
      case 'available': return '#28a745';      // Green
      case 'booking_full': return '#dc3545';  // Red
      case 'unavailable': return '#6c757d';   // Gray
      default: return '#6c757d';
    }
  };

  const getSlotStatusText = (status) => {
    switch (status) {
      case 'available': return 'Available';
      case 'booking_full': return 'Booking Full';
      case 'unavailable': return 'Unavailable';
      default: return 'Unknown';
    }
  };

  return (
    <div style={{ padding: '20px', maxWidth: '800px', margin: '0 auto' }}>
      <h2>Book Appointment</h2>
      
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

      {/* Consultation Type */}
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

      {/* Available Slots */}
      {selectedDate && (
        <div style={{ marginBottom: '20px' }}>
          <h3>Available Slots for {selectedDate}</h3>
          {loading ? (
            <div>Loading slots...</div>
          ) : slots.length > 0 ? (
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(250px, 1fr))', gap: '15px' }}>
              {slots.map(slot => (
                <div 
                  key={slot.id} 
                  style={{
                    border: `2px solid ${getSlotStatusColor(slot.status)}`,
                    borderRadius: '8px',
                    padding: '15px',
                    backgroundColor: '#fff',
                    cursor: slot.is_available ? 'pointer' : 'not-allowed',
                    opacity: slot.is_available ? 1 : 0.6,
                    transition: 'all 0.2s'
                  }}
                  onClick={() => slot.is_available && bookAppointment(slot)}
                >
                  <div style={{ fontWeight: 'bold', color: '#2c5aa0', marginBottom: '8px' }}>
                    {slot.start_time} - {slot.end_time}
                  </div>
                  
                  <div style={{ marginBottom: '8px' }}>
                    <span style={{ 
                      backgroundColor: getSlotStatusColor(slot.status),
                      color: 'white',
                      padding: '4px 8px',
                      borderRadius: '4px',
                      fontSize: '0.8em',
                      fontWeight: 'bold'
                    }}>
                      {getSlotStatusText(slot.status)}
                    </span>
                  </div>
                  
                  <div style={{ color: '#666', fontSize: '0.9em', marginBottom: '5px' }}>
                    Capacity: {slot.max_patients} | Booked: {slot.booked_patients} | Available: {slot.available_spots}
                  </div>
                  
                  <div style={{ color: '#666', fontSize: '0.9em', marginBottom: '5px' }}>
                    {slot.slot_type === 'online' ? '🌐 Online' : 
                     slot.slot_type === 'video' ? '📹 Video' : '🏥 In-person'}
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
    </div>
  );
}

export default AppointmentBooking;
```

### 2. Weekly Schedule View (Admin UI)
```jsx
import React, { useState, useEffect } from 'react';

function WeeklyScheduleView({ doctorId, clinicId }) {
  const [schedule, setSchedule] = useState(null);
  const [loading, setLoading] = useState(false);

  const fetchWeeklySchedule = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        doctor_id: doctorId,
        clinic_id: clinicId,
        slot_type: 'in-person'
      });

      const response = await fetch(`/api/organizations/doctor-time-slots/grouped?${params}`);
      const data = await response.json();
      setSchedule(data);
    } catch (error) {
      console.error('Error fetching schedule:', error);
    } finally {
      setLoading(false);
    }
  };

  const getDayStatusColor = (status) => {
    switch (status) {
      case 'available': return '#28a745';      // Green
      case 'booking_full': return '#dc3545';  // Red
      case 'unavailable': return '#6c757d';   // Gray
      default: return '#6c757d';
    }
  };

  const getDayStatusText = (status) => {
    switch (status) {
      case 'available': return 'Available';
      case 'booking_full': return 'Booking Full';
      case 'unavailable': return 'Unavailable';
      default: return 'Unknown';
    }
  };

  return (
    <div style={{ padding: '20px', maxWidth: '1000px', margin: '0 auto' }}>
      <h2>Weekly Schedule</h2>
      
      {loading ? (
        <div>Loading schedule...</div>
      ) : schedule ? (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: '15px' }}>
          {schedule.days.map(day => (
            <div 
              key={day.day_of_week}
              style={{
                border: `2px solid ${getDayStatusColor(day.status)}`,
                borderRadius: '8px',
                padding: '15px',
                backgroundColor: '#fff',
                minHeight: '200px'
              }}
            >
              <div style={{ fontWeight: 'bold', marginBottom: '10px', textAlign: 'center' }}>
                {day.day_name}
              </div>
              
              <div style={{ marginBottom: '10px', textAlign: 'center' }}>
                <span style={{ 
                  backgroundColor: getDayStatusColor(day.status),
                  color: 'white',
                  padding: '4px 8px',
                  borderRadius: '4px',
                  fontSize: '0.8em',
                  fontWeight: 'bold'
                }}>
                  {getDayStatusText(day.status)}
                </span>
              </div>
              
              {day.has_slots ? (
                <div>
                  <div style={{ fontSize: '0.9em', color: '#666', marginBottom: '10px' }}>
                    {day.total_slots} slots | {day.available_slots} available
                  </div>
                  
                  <div style={{ fontSize: '0.8em' }}>
                    {day.slots.map(slot => (
                      <div key={slot.id} style={{ marginBottom: '5px', padding: '5px', backgroundColor: '#f8f9fa', borderRadius: '4px' }}>
                        <div style={{ fontWeight: 'bold' }}>
                          {slot.start_time} - {slot.end_time}
                        </div>
                        <div style={{ color: '#666' }}>
                          {slot.available_spots}/{slot.max_patients} available
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              ) : (
                <div style={{ textAlign: 'center', color: '#666', fontStyle: 'italic' }}>
                  No slots
                </div>
              )}
            </div>
          ))}
        </div>
      ) : (
        <div>No schedule data available</div>
      )}
    </div>
  );
}

export default WeeklyScheduleView;
```

---

## Database Query Enhancement

### Appointment Counting Query
```sql
SELECT 
    dts.id, dts.doctor_id, dts.clinic_id, dts.day_of_week,
    dts.slot_type, dts.start_time, dts.end_time, dts.max_patients, dts.notes,
    dts.is_active, dts.created_at, dts.updated_at,
    COALESCE(appointment_count.booked_count, 0) as booked_patients
FROM doctor_time_slots dts
LEFT JOIN (
    SELECT 
        slot_id,
        COUNT(*) as booked_count
    FROM appointments 
    WHERE status IN ('confirmed', 'completed') 
    AND appointment_date = $2
    GROUP BY slot_id
) appointment_count ON dts.id = appointment_count.slot_id
WHERE dts.doctor_id = $1 AND dts.is_active = true
ORDER BY dts.day_of_week, dts.start_time
```

---

## Key Benefits

### ✅ Real-time Availability
- **Live updates**: Availability changes as appointments are booked
- **Accurate counts**: Shows exact number of booked vs available spots
- **Date-specific**: Filters by specific date for accurate availability

### ✅ Enhanced User Experience
- **Clear status**: Visual indicators for available, booking full, unavailable
- **Capacity info**: Shows max patients, booked patients, available spots
- **Day-level view**: See all 7 days with availability status

### ✅ Better Appointment Management
- **Prevent overbooking**: Shows when slots are full
- **Flexible scheduling**: Multiple slots per day with individual capacity
- **Status tracking**: Real-time status updates

### ✅ Frontend-Friendly
- **Grouped data**: Slots organized by day
- **Status indicators**: Easy to implement visual status
- **Complete week view**: Shows all days, even unavailable ones

---

## Testing Examples

### 1. Check Availability for Specific Date
```bash
curl -X GET "http://localhost:8081/api/organizations/doctor-time-slots?doctor_id=uuid&clinic_id=uuid&slot_type=in-person&date=2024-10-15" \
  -H "Authorization: Bearer token"
```

### 2. Get Weekly Schedule Grouped by Day
```bash
curl -X GET "http://localhost:8081/api/organizations/doctor-time-slots/grouped?doctor_id=uuid&clinic_id=uuid&slot_type=in-person" \
  -H "Authorization: Bearer token"
```

### 3. Check Availability for Specific Day of Week
```bash
curl -X GET "http://localhost:8081/api/organizations/doctor-time-slots/grouped?doctor_id=uuid&date=2024-10-15" \
  -H "Authorization: Bearer token"
```

---

## Summary

The enhanced slot availability system provides:

- **Real-time availability**: Live updates based on booked appointments
- **Capacity management**: Shows booked vs available patient counts
- **Status indicators**: Clear visual status for each slot
- **Complete week view**: Shows all 7 days with availability
- **Date-specific filtering**: Accurate availability for specific dates
- **Frontend integration**: Easy to implement in booking UIs

This system ensures patients can see exactly which slots are available, when they're booking full, and which days the doctor doesn't work, providing a complete and transparent booking experience.

---

**Last Updated:** Slot availability features implementation
