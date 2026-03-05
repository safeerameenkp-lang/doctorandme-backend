# Frontend Appointment API - Quick Reference Card ⚡

## 🎯 **Quick Implementation Guide**

Fast reference for frontend developers to implement appointment creation.

---

## 📋 **API Endpoint**

```
POST /api/appointments/simple
Content-Type: application/json
Authorization: Bearer {token}
```

---

## 📥 **Request Format**

```typescript
{
  clinic_id: "string (UUID)",
  clinic_patient_id: "string (UUID)",
  doctor_id: "string (UUID)",
  department_id: "string (UUID) - optional",
  individual_slot_id: "string (UUID)",
  appointment_date: "YYYY-MM-DD",
  appointment_time: "HH:MM:SS",
  consultation_type: "clinic_visit | video_consultation | follow-up-via-clinic | follow-up-via-video",
  payment_method: "pay_now | pay_later | way_off - optional if free follow-up",
  payment_type: "cash | card | upi - optional if pay_later/way_off",
  reason: "string - optional",
  notes: "string - optional"
}
```

---

## 📤 **Response Format**

```typescript
{
  message: "Appointment created successfully",
  
  appointment: {
    id: "uuid",
    booking_number: "AP-2025-CL001-0101",
    token_number: 5,
    appointment_date: "2025-10-26",
    appointment_time: "2025-10-26T10:00:00Z",
    consultation_type: "clinic_visit",
    status: "confirmed",
    fee_amount: 250.00,
    payment_status: "paid",
    payment_mode: "upi"
  },
  
  follow_up: {  // Only if regular appointment
    id: "uuid",
    patient_name: "Ameen Khan",
    doctor_name: "Dr. Smith",
    department_name: "Cardiology",
    follow_up_status: "active",
    is_free: true,
    valid_from: "2025-10-26T10:00:00Z",
    valid_until: "2025-10-31T10:00:00Z",
    days_remaining: 5
  },
  
  is_regular_appointment: true,
  followup_granted: true,
  followup_valid_until: "2025-10-31"
}
```

---

## ⚡ **Quick Code Example**

```typescript
// Create appointment
const response = await fetch('/api/appointments/simple', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({
    clinic_id: 'your-clinic-id',
    clinic_patient_id: 'patient-id',
    doctor_id: 'doctor-id',
    individual_slot_id: 'slot-id',
    appointment_date: '2025-10-26',
    appointment_time: '10:30:00',
    consultation_type: 'clinic_visit',
    payment_method: 'pay_now',
    payment_type: 'upi'
  })
});

const data = await response.json();
console.log(data.message); // "Appointment created successfully"

// Check if follow-up granted
if (data.followup_granted) {
  console.log(`Free follow-up until: ${data.followup_valid_until}`);
}
```

---

## 🎨 **React Hook Example**

```typescript
import { useState } from 'react';

function useCreateAppointment() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  
  const createAppointment = async (data) => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch('/api/appointments/simple', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify(data)
      });
      
      if (!response.ok) {
        throw new Error('Failed to create appointment');
      }
      
      const result = await response.json();
      return result;
      
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };
  
  return { createAppointment, loading, error };
}

// Usage
function AppointmentForm() {
  const { createAppointment, loading, error } = useCreateAppointment();
  
  const handleSubmit = async (formData) => {
    try {
      const response = await createAppointment(formData);
      alert('Appointment created!');
      // Handle response.follow_up if needed
    } catch (err) {
      alert('Error: ' + err.message);
    }
  };
  
  return (
    <form onSubmit={handleSubmit}>
      {/* Your form fields */}
      <button type="submit" disabled={loading}>
        {loading ? 'Booking...' : 'Book Appointment'}
      </button>
    </form>
  );
}
```

---

## 🔑 **Key Points**

1. **Auto-detection:** `consultation_type` determines if it's a follow-up
2. **Payment optional:** Free follow-ups don't need payment
3. **Response includes:** Complete follow-up info if granted
4. **Status tracking:** `current_followup_status` updated automatically

---

## 🎉 **That's It!**

Simple, fast, and customer-friendly appointment creation! 🚀

