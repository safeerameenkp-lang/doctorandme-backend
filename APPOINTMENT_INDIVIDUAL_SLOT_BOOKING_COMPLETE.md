## 🎯 Complete Appointment Booking with Session-Based Slots

## Overview

Complete integration between appointments and session-based individual slots. When a patient books a 5-minute slot, it automatically becomes unavailable.

---

## 📊 Complete Workflow

### Step 1: Doctor Creates Session-Based Slots

**Request:**
```json
POST /api/organizations/doctor-session-slots
Content-Type: application/json
Authorization: Bearer {token}

{
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "slot_type": "offline",
  "slot_duration": 5,
  "date": "2025-10-18",
  "sessions": [
    {
      "session_name": "Morning Session",
      "start_time": "09:30",
      "end_time": "11:30",
      "max_patients": 24,
      "slot_interval_minutes": 5
    },
    {
      "session_name": "Afternoon Session",
      "start_time": "13:30",
      "end_time": "18:30",
      "max_patients": 60,
      "slot_interval_minutes": 5
    }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Doctor time slots created successfully",
  "data": {
    "id": "timeslot-sat-uuid",
    "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "date": "2025-10-18",
    "day_of_week": 6,
    "slot_type": "offline",
    "sessions": [
      {
        "id": "morning-session-uuid",
        "session_name": "Morning Session",
        "start_time": "09:30",
        "end_time": "11:30",
        "generated_slots": 24,
        "available_slots": 24,
        "booked_slots": 0,
        "slots": [
          {
            "id": "slot-09-30-uuid",
            "slot_start": "09:30",
            "slot_end": "09:35",
            "is_booked": false,
            "status": "available"
          },
          {
            "id": "slot-09-35-uuid",
            "slot_start": "09:35",
            "slot_end": "09:40",
            "is_booked": false,
            "status": "available"
          }
          // ... 22 more morning slots
        ]
      },
      {
        "id": "afternoon-session-uuid",
        "session_name": "Afternoon Session",
        "start_time": "13:30",
        "end_time": "18:30",
        "generated_slots": 60,
        "available_slots": 60,
        "booked_slots": 0,
        "slots": [
          {
            "id": "slot-13-30-uuid",
            "slot_start": "13:30",
            "slot_end": "13:35",
            "is_booked": false,
            "status": "available"
          }
          // ... 59 more afternoon slots
        ]
      }
    ]
  }
}
```

**Result:**
- ✅ Created 84 individual 5-minute slots
- ✅ All slots status = "available"
- ✅ All slots is_booked = false

---

### Step 2: Patient Views Available Slots

**Request:**
```
GET /api/organizations/doctor-session-slots?doctor_id=85394ce8-94f7-4dca-a536-34305c46a98e&clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&date=2025-10-18&slot_type=offline
Authorization: Bearer {token}
```

**Response:**
```json
{
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "date": "2025-10-18",
  "slot_type": "offline",
  "total": 1,
  "slots": [
    {
      "date": "2025-10-18",
      "day_of_week": 6,
      "sessions": [
        {
          "session_name": "Morning Session",
          "start_time": "09:30",
          "end_time": "11:30",
          "available_slots": 24,
          "booked_slots": 0,
          "slots": [
            {
              "id": "slot-09-30-uuid",
              "slot_start": "09:30",
              "slot_end": "09:35",
              "is_booked": false,
              "status": "available"
            },
            {
              "id": "slot-09-35-uuid",
              "slot_start": "09:35",
              "slot_end": "09:40",
              "is_booked": false,
              "status": "available"
            }
            // ... 22 more available slots
          ]
        },
        {
          "session_name": "Afternoon Session",
          "available_slots": 60,
          "booked_slots": 0,
          "slots": [/* 60 available afternoon slots */]
        }
      ]
    }
  ]
}
```

---

### Step 3: Patient Books Specific 5-Minute Slot

**Request:**
```json
POST /api/appointments
Content-Type: application/json
Authorization: Bearer {token}

{
  "patient_id": "patient-uuid-123",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "individual_slot_id": "slot-09-30-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "consultation_type": "offline",
  "reason": "Regular checkup",
  "payment_mode": "pay_later"
}
```

**What the API Does:**
1. ✅ Validates individual_slot_id exists
2. ✅ Checks slot belongs to correct clinic
3. ✅ Verifies slot is_booked = false and status = "available"
4. ✅ Creates appointment record
5. ✅ **Automatically updates slot:**
   ```sql
   UPDATE doctor_individual_slots
   SET is_booked = true,
       booked_patient_id = 'patient-uuid-123',
       booked_appointment_id = 'new-appointment-uuid',
       status = 'booked'
   WHERE id = 'slot-09-30-uuid'
   ```

**Response (201 Created):**
```json
{
  "appointment": {
    "id": "appointment-uuid-new",
    "patient_id": "patient-uuid-123",
    "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "booking_number": "BN202510180001",
    "token_number": 1,
    "appointment_date": "2025-10-18",
    "appointment_time": "2025-10-18T09:30:00Z",
    "duration_minutes": 12,
    "consultation_type": "offline",
    "status": "confirmed",
    "fee_amount": 500.00,
    "payment_status": "pending",
    "payment_mode": "pay_later",
    "created_at": "2024-10-15T14:30:00Z"
  }
}
```

---

### Step 4: Verify Slot is Now Unavailable

**Request:**
```
GET /api/organizations/doctor-session-slots?doctor_id=85394ce8-94f7-4dca-a536-34305c46a98e&date=2025-10-18
```

**Response:**
```json
{
  "slots": [
    {
      "sessions": [
        {
          "session_name": "Morning Session",
          "generated_slots": 24,
          "available_slots": 23,
          "booked_slots": 1,
          "slots": [
            {
              "id": "slot-09-30-uuid",
              "slot_start": "09:30",
              "slot_end": "09:35",
              "is_booked": true,
              "booked_patient_id": "patient-uuid-123",
              "booked_appointment_id": "appointment-uuid-new",
              "status": "booked"
            },
            {
              "id": "slot-09-35-uuid",
              "slot_start": "09:35",
              "slot_end": "09:40",
              "is_booked": false,
              "status": "available"
            }
            // ... 22 more slots
          ]
        }
      ]
    }
  ]
}
```

**Changes:**
- ✅ available_slots decreased from 24 → 23
- ✅ booked_slots increased from 0 → 1
- ✅ slot-09-30 is_booked = true
- ✅ slot-09-30 status = "booked"
- ✅ booked_patient_id and booked_appointment_id set

---

### Step 5: Another Patient Tries to Book Same Slot

**Request:**
```json
POST /api/appointments
{
  "patient_id": "different-patient-uuid",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "individual_slot_id": "slot-09-30-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "consultation_type": "offline"
}
```

**Response (409 Conflict):**
```json
{
  "error": "Slot already booked",
  "message": "This 5-minute slot is no longer available",
  "slot_start": "09:30",
  "slot_end": "09:35",
  "current_status": "booked"
}
```

✅ **Prevents double-booking!**

---

## 📋 Complete API Examples

### Example 1: Book with Existing Patient

**Request:**
```json
POST /api/appointments
Content-Type: application/json
Authorization: Bearer {token}

{
  "patient_id": "patient-abc-123",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "individual_slot_id": "slot-14-00-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 14:00:00",
  "duration_minutes": 5,
  "consultation_type": "offline",
  "reason": "Follow-up consultation",
  "notes": "Diabetes checkup",
  "is_priority": false,
  "payment_mode": "cash"
}
```

**Response (201 Created):**
```json
{
  "appointment": {
    "id": "appt-uuid-xyz",
    "patient_id": "patient-abc-123",
    "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "booking_number": "BN202510180002",
    "token_number": 2,
    "appointment_date": "2025-10-18",
    "appointment_time": "2025-10-18T14:00:00Z",
    "duration_minutes": 5,
    "consultation_type": "offline",
    "reason": "Follow-up consultation",
    "notes": "Diabetes checkup",
    "status": "confirmed",
    "fee_amount": 300.00,
    "payment_status": "paid",
    "payment_mode": "cash",
    "is_priority": false,
    "created_at": "2024-10-15T14:45:00Z"
  }
}
```

**Database Changes:**
```sql
-- Appointment created
INSERT INTO appointments (...) VALUES (...);

-- Individual slot marked as booked
UPDATE doctor_individual_slots
SET is_booked = true,
    booked_patient_id = 'patient-abc-123',
    booked_appointment_id = 'appt-uuid-xyz',
    status = 'booked'
WHERE id = 'slot-14-00-uuid';
```

---

### Example 2: Create Patient + Book Appointment

**Request:**
```json
POST /api/appointments/patient-appointment
Content-Type: application/json
Authorization: Bearer {token}

{
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "email": "john.doe@example.com",
  "date_of_birth": "1990-05-15",
  "gender": "male",
  "mo_id": "MO123456",
  "blood_group": "O+",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "individual_slot_id": "slot-09-35-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:35:00",
  "duration_minutes": 5,
  "consultation_type": "new",
  "reason": "First visit - general checkup",
  "payment_mode": "card"
}
```

**Response (201 Created):**
```json
{
  "user": {
    "id": "user-uuid-new",
    "username": "johndoe",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890",
    "email": "john.doe@example.com",
    "date_of_birth": "1990-05-15",
    "gender": "male",
    "is_active": true
  },
  "patient": {
    "id": "patient-uuid-new",
    "user_id": "user-uuid-new",
    "mo_id": "MO123456",
    "blood_group": "O+",
    "is_active": true
  },
  "appointment": {
    "id": "appointment-uuid-123",
    "patient_id": "patient-uuid-new",
    "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "booking_number": "BN202510180003",
    "appointment_date": "2025-10-18",
    "appointment_time": "2025-10-18T09:35:00Z",
    "status": "confirmed",
    "consultation_type": "new",
    "fee_amount": 500.00,
    "payment_status": "paid"
  }
}
```

**What Happened:**
1. ✅ Created new user account for John Doe
2. ✅ Created patient record linked to user
3. ✅ Created appointment
4. ✅ **Marked slot-09-35 as booked**
5. ✅ Linked patient and appointment to the slot

---

### Step 3: Check Updated Availability

**Request:**
```
GET /doctor-session-slots?doctor_id=85394ce8-94f7-4dca-a536-34305c46a98e&date=2025-10-18
```

**Response:**
```json
{
  "slots": [
    {
      "sessions": [
        {
          "session_name": "Morning Session",
          "generated_slots": 24,
          "available_slots": 22,
          "booked_slots": 2,
          "slots": [
            {
              "id": "slot-09-30-uuid",
              "slot_start": "09:30",
              "slot_end": "09:35",
              "is_booked": true,
              "booked_patient_id": "patient-abc-123",
              "status": "booked"
            },
            {
              "id": "slot-09-35-uuid",
              "slot_start": "09:35",
              "slot_end": "09:40",
              "is_booked": true,
              "booked_patient_id": "patient-uuid-new",
              "status": "booked"
            },
            {
              "id": "slot-09-40-uuid",
              "slot_start": "09:40",
              "slot_end": "09:45",
              "is_booked": false,
              "status": "available"
            }
            // ... 21 more available slots
          ]
        }
      ]
    }
  ]
}
```

**Summary:**
- ✅ 2 slots booked (09:30 and 09:35)
- ✅ 22 slots still available
- ✅ Real-time availability tracking

---

## 🎨 UI Integration Example

### React Component for Slot Booking

```tsx
import React, { useState, useEffect } from 'react';

interface IndividualSlot {
  id: string;
  slot_start: string;
  slot_end: string;
  is_booked: boolean;
  status: string;
}

interface Session {
  id: string;
  session_name: string;
  available_slots: number;
  booked_slots: number;
  slots: IndividualSlot[];
}

const SlotBookingComponent: React.FC<{
  doctorId: string;
  clinicId: string;
  patientId: string;
}> = ({ doctorId, clinicId, patientId }) => {
  const [selectedDate, setSelectedDate] = useState('2025-10-18');
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadAvailableSlots();
  }, [selectedDate]);

  const loadAvailableSlots = async () => {
    setLoading(true);
    try {
      const response = await fetch(
        `/api/organizations/doctor-session-slots?` +
        `doctor_id=${doctorId}&` +
        `clinic_id=${clinicId}&` +
        `date=${selectedDate}&` +
        `slot_type=offline`
      );
      const data = await response.json();
      
      if (data.slots && data.slots.length > 0) {
        setSessions(data.slots[0].sessions || []);
      }
    } catch (error) {
      console.error('Failed to load slots:', error);
    } finally {
      setLoading(false);
    }
  };

  const bookSlot = async (slotId: string, slotTime: string) => {
    try {
      const response = await fetch('/api/appointments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          patient_id: patientId,
          doctor_id: doctorId,
          clinic_id: clinicId,
          individual_slot_id: slotId,
          appointment_date: selectedDate,
          appointment_time: `${selectedDate} ${slotTime}`,
          consultation_type: 'offline',
          payment_mode: 'pay_later'
        })
      });

      if (response.ok) {
        alert('Appointment booked successfully!');
        loadAvailableSlots(); // Refresh to show updated availability
      } else {
        const error = await response.json();
        if (error.error === 'Slot already booked') {
          alert('Sorry, this slot was just booked by someone else. Please select another slot.');
          loadAvailableSlots(); // Refresh
        } else {
          alert(`Booking failed: ${error.error}`);
        }
      }
    } catch (error) {
      console.error('Booking error:', error);
      alert('Failed to book appointment');
    }
  };

  return (
    <div className="slot-booking">
      <h2>Book Appointment</h2>
      
      <div className="date-picker">
        <label>Select Date:</label>
        <input
          type="date"
          value={selectedDate}
          onChange={(e) => setSelectedDate(e.target.value)}
        />
      </div>

      {loading ? (
        <p>Loading available slots...</p>
      ) : (
        <div className="sessions">
          {sessions.map(session => (
            <div key={session.id} className="session-card">
              <h3>{session.session_name}</h3>
              <p className="availability">
                <span className="available">{session.available_slots} available</span>
                <span className="booked">{session.booked_slots} booked</span>
              </p>
              
              <div className="slots-grid">
                {session.slots
                  .filter(slot => !slot.is_booked && slot.status === 'available')
                  .map(slot => (
                    <button
                      key={slot.id}
                      className="slot-button available"
                      onClick={() => bookSlot(slot.id, slot.slot_start)}
                    >
                      <div className="time">{slot.slot_start}</div>
                      <div className="duration">5 min</div>
                    </button>
                  ))}
                
                {session.slots
                  .filter(slot => slot.is_booked)
                  .slice(0, 3) // Show first 3 booked slots
                  .map(slot => (
                    <button
                      key={slot.id}
                      className="slot-button booked"
                      disabled
                    >
                      <div className="time">{slot.slot_start}</div>
                      <div className="status">Booked</div>
                    </button>
                  ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default SlotBookingComponent;
```

**CSS Example:**
```css
.session-card {
  border: 1px solid #ddd;
  padding: 20px;
  margin: 10px 0;
  border-radius: 8px;
}

.slots-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 10px;
  margin-top: 15px;
}

.slot-button {
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
}

.slot-button.available {
  background-color: #4CAF50;
  color: white;
}

.slot-button.available:hover {
  background-color: #45a049;
  transform: scale(1.05);
}

.slot-button.booked {
  background-color: #ccc;
  color: #666;
  cursor: not-allowed;
}

.availability {
  display: flex;
  gap: 15px;
  margin: 10px 0;
}

.available {
  color: #4CAF50;
  font-weight: bold;
}

.booked {
  color: #ff9800;
  font-weight: bold;
}
```

---

## 🔍 Error Scenarios

### Error 1: Slot Not Found

**Request:**
```json
{
  "individual_slot_id": "invalid-uuid",
  ...
}
```

**Response (404):**
```json
{
  "error": "Individual slot not found"
}
```

---

### Error 2: Slot Belongs to Different Clinic

**Request:**
```json
{
  "clinic_id": "clinic-a-uuid",
  "individual_slot_id": "slot-from-clinic-b",
  ...
}
```

**Response (400):**
```json
{
  "error": "Slot mismatch",
  "message": "The selected slot does not belong to this clinic"
}
```

---

### Error 3: Slot Already Booked

**Request:**
```json
{
  "individual_slot_id": "already-booked-slot-uuid",
  ...
}
```

**Response (409):**
```json
{
  "error": "Slot already booked",
  "message": "This 5-minute slot is no longer available",
  "slot_start": "09:30",
  "slot_end": "09:35",
  "current_status": "booked"
}
```

---

## 📊 Real-Time Availability Tracking

### Scenario: Morning Session (24 slots)

#### Initial State
```json
{
  "session_name": "Morning Session",
  "generated_slots": 24,
  "available_slots": 24,
  "booked_slots": 0
}
```

#### After 1st Booking
```json
{
  "session_name": "Morning Session",
  "generated_slots": 24,
  "available_slots": 23,
  "booked_slots": 1
}
```

#### After 10 Bookings
```json
{
  "session_name": "Morning Session",
  "generated_slots": 24,
  "available_slots": 14,
  "booked_slots": 10
}
```

#### After 24 Bookings (FULL)
```json
{
  "session_name": "Morning Session",
  "generated_slots": 24,
  "available_slots": 0,
  "booked_slots": 24
}
```

**Next attempt → "Slot already booked" error** ❌

---

## 🎯 Complete Flow Diagram

```
1. Doctor Creates Slots
   POST /doctor-session-slots
   ↓
   Creates 84 individual 5-min slots (all available)

2. Patient 1 Views Slots
   GET /doctor-session-slots
   ↓
   See 84 available slots

3. Patient 1 Books 09:30 Slot
   POST /appointments { individual_slot_id: "slot-09-30" }
   ↓
   Slot marked as booked
   83 slots available

4. Patient 2 Views Slots
   GET /doctor-session-slots
   ↓
   Sees 09:30 is booked
   83 available slots

5. Patient 2 Books 09:35 Slot
   POST /appointments { individual_slot_id: "slot-09-35" }
   ↓
   Slot marked as booked
   82 slots available

6. Patient 3 Tries 09:30 (Already Booked)
   POST /appointments { individual_slot_id: "slot-09-30" }
   ↓
   ❌ Error: "Slot already booked"
   Must choose different slot
```

---

## 📝 Quick Reference

### Request Fields
```json
{
  "patient_id": "UUID (required if existing patient)",
  "doctor_id": "UUID (required)",
  "clinic_id": "UUID (required)",
  "individual_slot_id": "UUID (optional - for session-based booking)",
  "slot_id": "UUID (optional - for simple slot booking)",
  "appointment_date": "YYYY-MM-DD (required)",
  "appointment_time": "YYYY-MM-DD HH:MM:SS (required)",
  "consultation_type": "offline|online|... (required)"
}
```

### System Behavior with individual_slot_id
1. ✅ Validates slot exists
2. ✅ Checks slot belongs to correct clinic
3. ✅ Verifies slot is_booked = false
4. ✅ Verifies status = "available"
5. ✅ Creates appointment
6. ✅ **Updates slot:**
   - is_booked = true
   - booked_patient_id = patient_id
   - booked_appointment_id = appointment_id
   - status = "booked"

---

## ✅ Status

| Feature | Status | Description |
|---------|--------|-------------|
| individual_slot_id support | ✅ Complete | Both APIs support it |
| Slot availability check | ✅ Complete | Validates before booking |
| Auto-mark slot as booked | ✅ Complete | Updates on successful booking |
| Prevent double-booking | ✅ Complete | Returns error if already booked |
| Track patient link | ✅ Complete | booked_patient_id stored |
| Track appointment link | ✅ Complete | booked_appointment_id stored |
| Real-time counts | ✅ Complete | available_slots/booked_slots |

---

**Status:** ✅ **Complete & Production Ready!**  
**Last Updated:** October 15, 2025  
**Version:** 1.0

Now when you book an appointment with `individual_slot_id`, the 5-minute slot automatically becomes unavailable! 🎉

