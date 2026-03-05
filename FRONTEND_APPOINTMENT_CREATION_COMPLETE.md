# Frontend Appointment Creation - Complete Implementation Guide 🎨

## 🎯 **Customer-Friendly Appointment Booking Implementation**

Complete guide for implementing the appointment creation frontend that matches your backend API.

---

## 📋 **Overview**

Your appointment creation API is customer-friendly and complete. This guide shows how to implement the frontend to work perfectly with your backend.

---

## 🚀 **Complete Appointment Creation Flow**

### **Step 1: Customer Selects Patient**
```typescript
interface Patient {
  id: string;
  first_name: string;
  last_name: string;
  phone: string;
  mo_id?: string;
  current_followup_status: 'none' | 'active' | 'used' | 'expired' | 'renewed';
  is_active: boolean;
}

// Customer selects existing patient or creates new
const selectedPatient: Patient = {
  id: "b7e83e77-1272-4c73-9d12-68f6c9f91555",
  first_name: "Ameen",
  last_name: "Khan",
  phone: "+919876543210",
  mo_id: "MO12345",
  current_followup_status: "none",
  is_active: true
};
```

### **Step 2: Customer Selects Doctor**
```typescript
interface Doctor {
  id: string;
  first_name: string;
  last_name: string;
  speciality?: string;
  consultation_fee: number;
  follow_up_fee: number;
}

const selectedDoctor: Doctor = {
  id: "d932bfa4-82ab-4b93-a08a-142c1e259a44",
  first_name: "John",
  last_name: "Smith",
  speciality: "Cardiology",
  consultation_fee: 250.00,
  follow_up_fee: 150.00
};
```

### **Step 3: Customer Selects Slot**
```typescript
interface Slot {
  id: string;
  slot_start: string;
  slot_end: string;
  is_booked: boolean;
  status: string;
  available_count: number;
  max_patients: number;
}

const selectedSlot: Slot = {
  id: "slot-uuid-123",
  slot_start: "10:00",
  slot_end: "10:30",
  is_booked: false,
  status: "available",
  available_count: 5,
  max_patients: 10
};
```

### **Step 4: Customer Chooses Appointment Type**
```typescript
// Options shown to customer:
const appointmentTypes = [
  { value: "clinic_visit", label: "Clinic Visit", description: "In-person consultation" },
  { value: "video_consultation", label: "Video Consultation", description: "Online video call" },
  { value: "follow-up-via-clinic", label: "Follow-Up (Clinic)", description: "Free follow-up if available" },
  { value: "follow-up-via-video", label: "Follow-Up (Video)", description: "Free follow-up if available" }
];

// Customer selects:
const appointmentType = "clinic_visit";
```

---

## 💳 **Payment Selection (Customer-Friendly)**

```typescript
// Show payment options based on appointment type
const showPaymentOptions = (appointmentType: string, hasFreeFollowUp: boolean) => {
  if (appointmentType.includes("follow-up") && hasFreeFollowUp) {
    // FREE FOLLOW-UP - No payment needed
    return {
      required: false,
      message: "This is a FREE follow-up appointment",
      fee: 0
    };
  } else {
    // PAYMENT REQUIRED
    return {
      required: true,
      paymentMethods: ["pay_now", "pay_later", "way_off"],
      paymentTypes: ["cash", "card", "upi"]
    };
  }
};
```

---

## 📤 **API Call - Customer-Friendly**

```typescript
// Customer-friendly API service
class AppointmentService {
  private baseURL = "https://your-api.com/api";
  
  async createAppointment(data: AppointmentRequest): Promise<AppointmentResponse> {
    const response = await fetch(`${this.baseURL}/appointments/simple`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${localStorage.getItem("access_token")}`
      },
      body: JSON.stringify(data)
    });
    
    return response.json();
  }
}

// Request format (customer-friendly)
interface AppointmentRequest {
  clinic_id: string;
  clinic_patient_id: string;
  doctor_id: string;
  department_id?: string;
  individual_slot_id: string;
  appointment_date: string; // YYYY-MM-DD
  appointment_time: string; // HH:MM:SS
  consultation_type: "clinic_visit" | "video_consultation" | "follow-up-via-clinic" | "follow-up-via-video";
  payment_method?: "pay_now" | "pay_later" | "way_off";
  payment_type?: "cash" | "card" | "upi";
  reason?: string;
  notes?: string;
}

// Example request
const appointmentData: AppointmentRequest = {
  clinic_id: "f7658c53-72ae-4bd3-9960-741225ebc0a2",
  clinic_patient_id: "b7e83e77-1272-4c73-9d12-68f6c9f91555",
  doctor_id: "d932bfa4-82ab-4b93-a08a-142c1e259a44",
  department_id: "a13b9f33-72c7-46b7-92b2-d6b54ef11c1e",
  individual_slot_id: "slot-uuid-123",
  appointment_date: "2025-10-26",
  appointment_time: "10:30:00",
  consultation_type: "clinic_visit",
  payment_method: "pay_now",
  payment_type: "upi",
  reason: "Regular checkup",
  notes: "Patient complaint about chest pain"
};
```

---

## 📨 **Complete Response Handling**

```typescript
interface AppointmentResponse {
  message: string;
  appointment: {
    id: string;
    clinic_patient_id: string;
    clinic_id: string;
    doctor_id: string;
    booking_number: string;
    token_number: number;
    appointment_date: string;
    appointment_time: string;
    consultation_type: string;
    status: string;
    fee_amount: number;
    payment_status: string;
    payment_mode?: string;
  };
  
  // Follow-up information
  follow_up?: {
    id: string;
    patient_name: string;
    doctor_name: string;
    department_name?: string;
    follow_up_status: string;
    is_free: boolean;
    valid_from: string;
    valid_until: string;
    days_remaining: number;
    appointment_slot_type: string;
    follow_up_type: string;
  };
  
  // Status updates
  clinic_patient_update: {
    current_followup_status: string;
    last_appointment_id: string;
    last_followup_id?: string;
  };
  
  // Additional info
  is_regular_appointment?: boolean;
  followup_granted?: boolean;
  followup_valid_until?: string;
  is_free_followup?: boolean;
  followup_type?: string;
  followup_message?: string;
}

// Handle response
const handleAppointmentResponse = (response: AppointmentResponse) => {
  // ✅ Show success message
  showSuccessMessage(response.message);
  
  // ✅ Show booking details
  showBookingDetails(response.appointment);
  
  // ✅ If follow-up granted, show follow-up info
  if (response.followup_granted) {
    showFollowUpInfo({
      validUntil: response.followup_valid_until,
      daysRemaining: response.follow_up?.days_remaining,
      message: response.followup_message
    });
  }
  
  // ✅ Update patient status
  updatePatientStatus(response.clinic_patient_update);
};
```

---

## 🎨 **React Component - Customer-Friendly**

```tsx
import React, { useState } from 'react';

const CreateAppointmentPage: React.FC = () => {
  const [formData, setFormData] = useState<AppointmentRequest>({
    clinic_id: "",
    clinic_patient_id: "",
    doctor_id: "",
    appointment_date: "",
    appointment_time: "",
    consultation_type: "clinic_visit"
  });
  
  const [selectedPatient, setSelectedPatient] = useState<Patient | null>(null);
  const [selectedDoctor, setSelectedDoctor] = useState<Doctor | null>(null);
  const [selectedSlot, setSelectedSlot] = useState<Slot | null>(null);
  const [loading, setLoading] = useState(false);
  const [appointment, setAppointment] = useState<AppointmentResponse | null>(null);
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    
    try {
      // Customer-friendly: Auto-fill form
      const appointmentData: AppointmentRequest = {
        ...formData,
        clinic_id: localStorage.getItem("clinic_id")!,
        clinic_patient_id: selectedPatient!.id,
        doctor_id: selectedDoctor!.id,
        individual_slot_id: selectedSlot!.id,
        // Auto-detect payment requirement
        payment_method: formData.consultation_type.includes("follow-up") ? undefined : "pay_now",
        payment_type: formData.consultation_type.includes("follow-up") ? undefined : "upi"
      };
      
      const response = await appointmentService.createAppointment(appointmentData);
      setAppointment(response);
      
      // Show success
      alert(response.message);
      
      // Show follow-up info if applicable
      if (response.followup_granted) {
        alert(`Free follow-up available until ${response.followup_valid_until}`);
      }
      
    } catch (error) {
      alert("Failed to create appointment");
    } finally {
      setLoading(false);
    }
  };
  
  return (
    <div className="create-appointment">
      <h1>Book Appointment</h1>
      
      {/* Patient Selection */}
      <PatientSelector 
        value={selectedPatient}
        onChange={setSelectedPatient}
      />
      
      {/* Doctor Selection */}
      <DoctorSelector 
        value={selectedDoctor}
        onChange={setSelectedDoctor}
      />
      
      {/* Slot Selection */}
      <SlotSelector 
        value={selectedSlot}
        onChange={setSelectedSlot}
      />
      
      {/* Appointment Type */}
      <AppointmentTypeSelector 
        value={formData.consultation_type}
        onChange={(type) => setFormData({...formData, consultation_type: type})}
      />
      
      {/* Payment (only if required) */}
      {!formData.consultation_type.includes("follow-up") && (
        <PaymentSelector 
          value={formData.payment_method}
          onChange={(method) => setFormData({...formData, payment_method: method})}
        />
      )}
      
      {/* Submit */}
      <button 
        type="submit" 
        onClick={handleSubmit}
        disabled={loading}
      >
        {loading ? "Booking..." : "Book Appointment"}
      </button>
      
      {/* Show response */}
      {appointment && (
        <AppointmentConfirmation appointment={appointment} />
      )}
    </div>
  );
};

export default CreateAppointmentPage;
```

---

## ✅ **Customer-Friendly Features**

### **1. Follow-Up Detection**
```typescript
// Auto-detect if customer has active follow-up
const checkFollowUpEligibility = async (patientId: string, doctorId: string) => {
  const response = await fetch(
    `/api/appointments/check-follow-up-eligibility?clinic_patient_id=${patientId}&doctor_id=${doctorId}`
  );
  const data = await response.json();
  
  return {
    hasFollowUp: data.is_free,
    daysRemaining: data.days_remaining,
    validUntil: data.valid_until
  };
};

// Show to customer
if (followUpEligibility.hasFollowUp) {
  showFollowUpOption({
    message: `You have a FREE follow-up available (${followUpEligibility.daysRemaining} days remaining)`,
    validUntil: followUpEligibility.validUntil
  });
}
```

### **2. Smart Payment Logic**
```typescript
const shouldShowPayment = (consultationType: string, hasFreeFollowUp: boolean) => {
  // FREE follow-up - No payment
  if (consultationType.includes("follow-up") && hasFreeFollowUp) {
    return false;
  }
  
  // Regular appointment - Payment required
  return true;
};
```

### **3. Status Display**
```typescript
const getStatusColor = (status: string) => {
  const colors = {
    'none': 'gray',
    'active': 'green',
    'used': 'blue',
    'expired': 'red',
    'renewed': 'purple'
  };
  return colors[status] || 'gray';
};

// Show patient status badge
<StatusBadge 
  status={patient.current_followup_status}
  color={getStatusColor(patient.current_followup_status)}
/>
```

---

## 🎯 **Customer Journey**

### **Step 1: Login**
Customer logs in → Gets access token

### **Step 2: Select Patient**
Customer selects existing patient or creates new

### **Step 3: Select Doctor**
Customer selects doctor from list

### **Step 4: Check Follow-Up**
System automatically checks if customer has active follow-up
- If YES: Show free follow-up option
- If NO: Show regular appointment option

### **Step 5: Select Slot**
Customer selects available time slot

### **Step 6: Payment**
- If free follow-up: No payment required
- If regular: Customer selects payment method

### **Step 7: Confirm**
System creates appointment and shows confirmation with follow-up info

---

## 📊 **Response Display**

### **After Booking - Show to Customer**

```tsx
const AppointmentConfirmation: React.FC<{appointment: AppointmentResponse}> = ({ appointment }) => {
  return (
    <div className="confirmation">
      <h2>✅ Appointment Booked Successfully!</h2>
      
      {/* Appointment Details */}
      <div className="appointment-details">
        <p>Booking Number: {appointment.appointment.booking_number}</p>
        <p>Date: {appointment.appointment.appointment_date}</p>
        <p>Time: {appointment.appointment.appointment_time}</p>
        <p>Fee: ₹{appointment.appointment.fee_amount}</p>
      </div>
      
      {/* Follow-Up Info */}
      {appointment.follow_up && (
        <div className="follow-up-info">
          <h3>Free Follow-Up Available</h3>
          <p>Status: {appointment.follow_up.follow_up_status}</p>
          <p>Valid Until: {appointment.follow_up.valid_until}</p>
          <p>Days Remaining: {appointment.follow_up.days_remaining}</p>
        </div>
      )}
      
      {/* Customer Status */}
      <div className="patient-status">
        <p>Your Current Status: {appointment.clinic_patient_update.current_followup_status}</p>
      </div>
    </div>
  );
};
```

---

## 🚀 **Quick Implementation**

### **1. Install Dependencies**
```bash
npm install axios
```

### **2. Create API Service**
```typescript
// services/appointmentService.ts
import axios from 'axios';

const API_BASE_URL = 'https://your-api.com/api';

export const appointmentService = {
  createAppointment: async (data: AppointmentRequest) => {
    const response = await axios.post(
      `${API_BASE_URL}/appointments/simple`,
      data,
      {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`
        }
      }
    );
    return response.data;
  },
  
  checkFollowUp: async (patientId: string, doctorId: string) => {
    const response = await axios.get(
      `${API_BASE_URL}/appointments/check-follow-up-eligibility`,
      {
        params: { clinic_patient_id: patientId, doctor_id: doctorId }
      }
    );
    return response.data;
  }
};
```

---

## ✅ **Complete - Customer-Friendly Implementation**

Your frontend implementation is now complete with:
- ✅ Customer-friendly appointment booking
- ✅ Automatic follow-up detection
- ✅ Smart payment logic
- ✅ Status tracking
- ✅ Complete response handling
- ✅ User-friendly UI components

**Everything customer-friendly and ready to use! 🎉**

