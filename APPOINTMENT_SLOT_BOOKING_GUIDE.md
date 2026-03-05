# Appointment Slot-Based Booking System

## 🎯 Overview

The appointment system now supports **slot-based booking** where appointments are linked to pre-defined time slots. This ensures:
- **Automatic capacity management** - Slots track available spots
- **Prevents overbooking** - Validates max_patients limit
- **Real-time availability** - Shows when slots are full
- **Better scheduling control** - Doctors can manage availability via slots

---

## 📋 How It Works

### 1. **Doctor Creates Time Slots**
```json
POST /organizations/doctor-time-slots
{
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "slot_type": "offline",
  "date": "2025-01-20",
  "slots": [
    {
      "start_time": "09:00",
      "end_time": "09:30",
      "max_patients": 3  // ✅ Max 3 patients can book this slot
    }
  ]
}
```

**Response:**
```json
{
  "created_slots": [
    {
      "id": "slot-uuid-123",
      "date": "2025-01-20",
      "start_time": "09:00",
      "end_time": "09:30",
      "max_patients": 3,
      "booked_patients": 0,
      "available_spots": 3,
      "is_available": true,
      "status": "available"
    }
  ]
}
```

---

### 2. **Create Appointment with Slot**

#### Option A: With existing patient
```json
POST /appointments
{
  "patient_id": "patient-uuid",
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "slot_id": "slot-uuid-123",  // ✅ Link to time slot
  "appointment_date": "2025-01-20",
  "appointment_time": "2025-01-20 09:00:00",
  "consultation_type": "offline"
}
```

#### Option B: Create patient + appointment with slot
```json
POST /appointments/patient-appointment
{
  // Patient details
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  
  // Appointment details
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "slot_id": "slot-uuid-123",  // ✅ Link to time slot
  "appointment_date": "2025-01-20",
  "appointment_time": "2025-01-20 09:00:00",
  "consultation_type": "offline"
}
```

---

### 3. **System Validates Slot Availability**

When creating an appointment with `slot_id`, the system automatically:

✅ **Checks if slot exists**
```sql
SELECT * FROM doctor_time_slots WHERE id = 'slot-uuid-123'
```

✅ **Validates slot is active**
```
if !slot.is_active → Error: "Time slot is not active"
```

✅ **Verifies slot belongs to correct doctor/clinic**
```
if slot.doctor_id != appointment.doctor_id → Error: "Slot mismatch"
```

✅ **Counts current bookings**
```sql
SELECT COUNT(*) FROM appointments 
WHERE slot_id = 'slot-uuid-123' 
AND status IN ('confirmed', 'completed')
```

✅ **Checks if slot is full**
```
if booked_count >= max_patients → Error: "Slot is fully booked"
```

---

## 📊 Slot Capacity Management

### Example: 3-Patient Slot

**Initial State:**
```json
{
  "id": "slot-123",
  "max_patients": 3,
  "booked_patients": 0,
  "available_spots": 3,
  "is_available": true,
  "status": "available"
}
```

**After 1st Appointment:**
```json
{
  "max_patients": 3,
  "booked_patients": 1,
  "available_spots": 2,
  "is_available": true,
  "status": "available"
}
```

**After 2nd Appointment:**
```json
{
  "max_patients": 3,
  "booked_patients": 2,
  "available_spots": 1,
  "is_available": true,
  "status": "available"
}
```

**After 3rd Appointment (FULL):**
```json
{
  "max_patients": 3,
  "booked_patients": 3,
  "available_spots": 0,
  "is_available": false,
  "status": "booking_full"
}
```

**4th Attempt → REJECTED ❌**
```json
{
  "error": "Slot is fully booked",
  "message": "This time slot has reached maximum capacity",
  "max_patients": 3,
  "booked_patients": 3
}
```

---

## 🚫 Error Scenarios

### Error 1: Slot Not Found
```json
{
  "error": "Time slot not found"
}
```
**Cause**: Invalid `slot_id` or slot was deleted

---

### Error 2: Slot Inactive
```json
{
  "error": "Time slot is not active"
}
```
**Cause**: Doctor deactivated the slot

---

### Error 3: Slot Mismatch
```json
{
  "error": "Slot mismatch",
  "message": "The selected time slot does not belong to this doctor or clinic"
}
```
**Cause**: `slot.doctor_id` ≠ `appointment.doctor_id` OR `slot.clinic_id` ≠ `appointment.clinic_id`

---

### Error 4: Slot Fully Booked
```json
{
  "error": "Slot is fully booked",
  "message": "This time slot has reached maximum capacity",
  "max_patients": 3,
  "booked_patients": 3
}
```
**Cause**: Slot reached `max_patients` limit

---

## 🔍 Get Slot Availability

### Check Available Slots
```
GET /organizations/doctor-time-slots?doctor_id=xxx&clinic_id=xxx&date=2025-01-20
```

**Response:**
```json
{
  "slots": [
    {
      "id": "slot-1",
      "start_time": "09:00",
      "end_time": "09:30",
      "max_patients": 3,
      "booked_patients": 2,
      "available_spots": 1,
      "is_available": true,
      "status": "available"
    },
    {
      "id": "slot-2",
      "start_time": "10:00",
      "end_time": "10:30",
      "max_patients": 3,
      "booked_patients": 3,
      "available_spots": 0,
      "is_available": false,
      "status": "booking_full"
    }
  ]
}
```

---

## 💡 UI Integration Guide

### Step 1: Fetch Available Slots
```javascript
async function getAvailableSlots(doctorId, clinicId, date) {
  const response = await fetch(
    `/organizations/doctor-time-slots?doctor_id=${doctorId}&clinic_id=${clinicId}&date=${date}`
  );
  const data = await response.json();
  
  // Filter only available slots
  return data.slots.filter(slot => slot.is_available && slot.available_spots > 0);
}
```

### Step 2: Display Slots to User
```javascript
function SlotSelector({ slots }) {
  return (
    <div>
      {slots.map(slot => (
        <div key={slot.id} className={slot.is_available ? 'available' : 'full'}>
          <span>{slot.start_time} - {slot.end_time}</span>
          <span>{slot.available_spots}/{slot.max_patients} available</span>
          <button 
            disabled={!slot.is_available}
            onClick={() => bookAppointment(slot.id)}
          >
            {slot.is_available ? 'Book' : 'Full'}
          </button>
        </div>
      ))}
    </div>
  );
}
```

### Step 3: Create Appointment with Selected Slot
```javascript
async function bookAppointment(slotId) {
  try {
    const response = await fetch('/appointments', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        patient_id: patientId,
        doctor_id: doctorId,
        clinic_id: clinicId,
        slot_id: slotId,  // ✅ Link to slot
        appointment_date: selectedDate,
        appointment_time: `${selectedDate} ${slotStartTime}`,
        consultation_type: 'offline'
      })
    });
    
    if (response.ok) {
      alert('Appointment booked successfully!');
    } else {
      const error = await response.json();
      if (error.error === 'Slot is fully booked') {
        alert('Sorry, this slot just got filled. Please select another slot.');
      }
    }
  } catch (error) {
    console.error('Booking failed:', error);
  }
}
```

---

## 📈 Database Schema

### appointments table (slot_id column added)
```sql
ALTER TABLE appointments 
ADD COLUMN slot_id UUID REFERENCES doctor_time_slots(id) ON DELETE SET NULL;
```

### Query to get slot booking count
```sql
SELECT 
  dts.id,
  dts.max_patients,
  COUNT(a.id) FILTER (WHERE a.status IN ('confirmed', 'completed')) as booked_patients,
  dts.max_patients - COUNT(a.id) FILTER (WHERE a.status IN ('confirmed', 'completed')) as available_spots
FROM doctor_time_slots dts
LEFT JOIN appointments a ON a.slot_id = dts.id
WHERE dts.id = 'slot-uuid'
GROUP BY dts.id;
```

---

## ⚙️ Optional vs Required

### `slot_id` is **OPTIONAL**

**Without slot_id (traditional booking):**
```json
{
  "patient_id": "...",
  "doctor_id": "...",
  "appointment_date": "2025-01-20",
  "appointment_time": "2025-01-20 09:00:00"
  // No slot_id - works as before
}
```

**With slot_id (slot-based booking):**
```json
{
  "patient_id": "...",
  "doctor_id": "...",
  "slot_id": "slot-uuid",  // ✅ Validates and tracks capacity
  "appointment_date": "2025-01-20",
  "appointment_time": "2025-01-20 09:00:00"
}
```

---

## 🎯 Benefits

### For Clinics
- ✅ **Better capacity management** - Know exactly how many spots available
- ✅ **Prevents overbooking** - Automatic validation
- ✅ **Real-time visibility** - See slot status instantly
- ✅ **Flexible scheduling** - Different max_patients per slot

### For Patients
- ✅ **See availability** - Know which slots are open
- ✅ **Fair booking** - First-come, first-served
- ✅ **Instant confirmation** - No manual approval needed

### For Developers
- ✅ **Simple integration** - Just add `slot_id` to request
- ✅ **Automatic validation** - Built-in checks
- ✅ **Clear error messages** - Easy to handle in UI

---

## 📝 Complete Example Flow

### 1. Doctor Creates Morning Slots
```bash
POST /organizations/doctor-time-slots
{
  "doctor_id": "dr-smith",
  "clinic_id": "main-clinic",
  "date": "2025-01-20",
  "slot_type": "offline",
  "slots": [
    { "start_time": "09:00", "end_time": "09:30", "max_patients": 2 },
    { "start_time": "09:30", "end_time": "10:00", "max_patients": 2 },
    { "start_time": "10:00", "end_time": "10:30", "max_patients": 2 }
  ]
}
```

### 2. Patient 1 Books 09:00 Slot
```bash
POST /appointments
{
  "patient_id": "patient-1",
  "slot_id": "slot-09-00",
  ...
}
# ✅ Success! Booked 1/2
```

### 3. Patient 2 Books 09:00 Slot
```bash
POST /appointments
{
  "patient_id": "patient-2",
  "slot_id": "slot-09-00",
  ...
}
# ✅ Success! Booked 2/2 (FULL)
```

### 4. Patient 3 Tries to Book 09:00 Slot
```bash
POST /appointments
{
  "patient_id": "patient-3",
  "slot_id": "slot-09-00",
  ...
}
# ❌ Error: "Slot is fully booked"
# Patient 3 must choose 09:30 or 10:00 slot
```

---

## ✅ Status

| Feature | Status | Description |
|---------|--------|-------------|
| Slot validation | ✅ Complete | Checks slot exists and is valid |
| Capacity check | ✅ Complete | Prevents overbooking |
| Real-time count | ✅ Complete | Shows current booked_patients |
| Error handling | ✅ Complete | Clear error messages |
| Both APIs updated | ✅ Complete | CreateAppointment & CreatePatientWithAppointment |

---

**Last Updated**: October 15, 2025  
**Version**: 1.0  
**Status**: ✅ Production Ready

