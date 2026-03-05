# Frontend Complete Implementation Guide 🎯

## 📋 **Complete Frontend Flow**

Your frontend needs to handle:
1. Patient click → Show appointments/follow-ups
2. Appointment click → Show details
3. Follow-up click → Book/Check eligibility

---

## 🔄 **Complete Flow**

### **Step 1: Get Clinic Patients**
```typescript
// Fetch all patients for the clinic
const response = await fetch(
  `/api/v1/organization/clinic/${clinicId}/patients?search=${searchTerm}`
);
const data = await response.json();
const patients = data.patients; // Array of patients
```

**Response Structure:**
```json
{
  "patients": [
    {
      "clinic_patient_id": "uuid",
      "first_name": "John",
      "last_name": "Doe",
      "phone": "1234567890",
      "mo_id": "MO123456",
      "is_active": true,
      
      // Follow-up status
      "current_followup_status": "active",
      "last_appointment_id": "appointment-uuid",
      "last_followup_id": "followup-uuid",
      
      // Arrays (never null)
      "appointments": [...],  // All appointments
      "follow_ups": [...]      // All follow-ups
    }
  ]
}
```

---

### **Step 2: Click Patient → Show Appointments**
```typescript
// Click handler for a patient
const handlePatientClick = (patient) => {
  console.log("Selected Patient:", patient.first_name);
  
  // Show patient's appointments
  const appointments = patient.appointments; // Already in response
  const followUps = patient.follow_ups;      // Already in response
  
  // Display appointments table
  displayAppointments(appointments, followUps);
};

// Display function
const displayAppointments = (appointments, followUps) => {
  appointments.forEach(apt => {
    console.log("Appointment:", apt.id);
    console.log("Date:", apt.appointment_date);
    console.log("Doctor:", apt.doctor_name);
    console.log("Status:", apt.status);
  });
  
  followUps.forEach(followUp => {
    console.log("Follow-up Status:", followUp.status);
    console.log("Days Remaining:", followUp.days_remaining);
  });
};
```

---

### **Step 3: Click Follow-Up → Check Eligibility**
```typescript
// Check if follow-up is eligible
const checkFollowUpEligibility = async (
  clinicPatientId,
  doctorId,
  departmentId
) => {
  const response = await fetch(
    `/api/v1/appointments/followup-eligibility?` +
    `clinic_patient_id=${clinicPatientId}` +
    `&doctor_id=${doctorId}` +
    `&department_id=${departmentId}`
  );
  
  const data = await response.json();
  
  // Check if free
  if (data.is_free && data.is_eligible) {
    // Show FREE follow-up button
    return { isFree: true, followUp: data.follow_up };
  } else {
    // Show PAID follow-up button
    return { isFree: false };
  }
};
```

**Response Structure:**
```json
{
  "is_free": true,
  "is_eligible": true,
  "message": "Patient is eligible for free follow-up",
  
  "follow_up": {
    "follow_up_logic_status": "new",
    "logic_notes": "Patient gets one free follow-up...",
    "days_remaining": 5,
    ...
  }
}
```

---

### **Step 4: Click Book Follow-Up → Create Appointment**
```typescript
// Book a follow-up appointment
const bookFollowUp = async (appointmentData) => {
  // First check eligibility
  const eligibility = await checkFollowUpEligibility(
    appointmentData.clinic_patient_id,
    appointmentData.doctor_id,
    appointmentData.department_id
  );
  
  // Determine if it's free or paid
  const isFree = eligibility.isFree && eligibility.is_eligible;
  
  // Set appointment type
  const appointmentInput = {
    clinic_id: appointmentData.clinic_id,
    clinic_patient_id: appointmentData.clinic_patient_id,
    doctor_id: appointmentData.doctor_id,
    department_id: appointmentData.department_id,
    slot_type: appointmentData.slot_type, // clinic_followup or video_followup
    appointment_date: appointmentData.appointment_date,
    
    // This will be a follow-up type
    is_followup_appointment: true
  };
  
  // Call create appointment API
  const response = await fetch('/api/v1/appointments/create-simple', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(appointmentInput)
  });
  
  const result = await response.json();
  
  // Check response
  if (result.message === "Appointment created successfully") {
    // Show success message
    alert("Follow-up booked successfully!");
    
    // Check if it was free or paid
    if (result.follow_up) {
      console.log("Follow-up details:", result.follow_up);
      console.log("Logic Status:", result.follow_up.follow_up_logic_status);
    }
  }
};
```

---

## 🎯 **Complete Example: React Component**

```typescript
import React, { useState, useEffect } from 'react';

interface Patient {
  clinic_patient_id: string;
  first_name: string;
  last_name: string;
  phone: string;
  mo_id: string;
  is_active: boolean;
  current_followup_status: string;
  appointments: Appointment[];
  follow_ups: FollowUp[];
}

interface Appointment {
  id: string;
  appointment_date: string;
  slot_type: string;
  doctor_name: string;
  status: string;
}

interface FollowUp {
  id: string;
  follow_up_status: string;
  follow_up_logic_status: string;
  logic_notes: string;
  days_remaining: number;
  is_free: boolean;
  valid_until: string;
}

const PatientAppointmentDashboard: React.FC = () => {
  const [patients, setPatients] = useState<Patient[]>([]);
  const [selectedPatient, setSelectedPatient] = useState<Patient | null>(null);
  const [followUpEligibility, setFollowUpEligibility] = useState<any>(null);
  
  // Fetch patients
  const fetchPatients = async () => {
    const response = await fetch(`/api/v1/organization/clinic/${clinicId}/patients`);
    const data = await response.json();
    setPatients(data.patients);
  };
  
  // Handle patient click
  const handlePatientClick = (patient: Patient) => {
    setSelectedPatient(patient);
    
    // Automatically check follow-up eligibility for latest appointment
    if (patient.appointments.length > 0) {
      const latestAppointment = patient.appointments[0];
      checkFollowUpStatus(patient, latestAppointment);
    }
  };
  
  // Check follow-up status
  const checkFollowUpStatus = async (patient: Patient, appointment: Appointment) => {
    try {
      const response = await fetch(
        `/api/v1/appointments/followup-eligibility?` +
        `clinic_patient_id=${patient.clinic_patient_id}` +
        `&doctor_id=${appointment.doctor_id}` +
        `&department_id=${appointment.department_id}`
      );
      
      const data = await response.json();
      setFollowUpEligibility(data);
    } catch (error) {
      console.error("Error checking follow-up eligibility:", error);
    }
  };
  
  // Book follow-up
  const handleBookFollowUp = async () => {
    if (!selectedPatient || !followUpEligibility) return;
    
    // Determine if free or paid
    const isFree = followUpEligibility.is_free && followUpEligibility.is_eligible;
    
    const appointmentData = {
      clinic_id: selectedPatient.clinic_id,
      clinic_patient_id: selectedPatient.clinic_patient_id,
      doctor_id: followUpEligibility.doctor_id,
      department_id: followUpEligibility.department_id,
      slot_type: "clinic_followup", // or "video_followup"
      appointment_date: selectedDate, // Your date picker value
      is_followup_appointment: true
    };
    
    // Call API
    const response = await fetch('/api/v1/appointments/create-simple', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(appointmentData)
    });
    
    const result = await response.json();
    
    if (result.message === "Appointment created successfully") {
      alert(`Follow-up booked ${isFree ? 'FREE' : 'PAID'}!`);
      fetchPatients(); // Refresh list
    }
  };
  
  return (
    <div>
      {/* Patient List */}
      <div>
        <h2>Patients</h2>
        {patients.map(patient => (
          <div key={patient.clinic_patient_id} onClick={() => handlePatientClick(patient)}>
            <h3>{patient.first_name} {patient.last_name}</h3>
            <p>Phone: {patient.phone}</p>
            <p>MO ID: {patient.mo_id}</p>
            
            {/* Follow-up status badge */}
            <span className={`badge badge-${patient.current_followup_status}`}>
              {patient.current_followup_status}
            </span>
          </div>
        ))}
      </div>
      
      {/* Selected Patient Details */}
      {selectedPatient && (
        <div>
          <h2>{selectedPatient.first_name}'s Appointments</h2>
          
          {/* Appointments List */}
          <div>
            <h3>All Appointments</h3>
            {selectedPatient.appointments.map(apt => (
              <div key={apt.id}>
                <p>Date: {apt.appointment_date}</p>
                <p>Doctor: {apt.doctor_name}</p>
                <p>Status: {apt.status}</p>
              </div>
            ))}
          </div>
          
          {/* Follow-ups List */}
          <div>
            <h3>Follow-ups</h3>
            {selectedPatient.follow_ups.map(followUp => (
              <div key={followUp.id}>
                <p>Status: {followUp.follow_up_status}</p>
                <p>Logic: {followUp.follow_up_logic_status}</p>
                <p>Days Remaining: {followUp.days_remaining}</p>
              </div>
            ))}
          </div>
          
          {/* Follow-up Eligibility Check */}
          {followUpEligibility && (
            <div>
              <h3>Book Follow-up</h3>
              <p>Status: {followUpEligibility.follow_up?.follow_up_logic_status}</p>
              <p>{followUpEligibility.follow_up?.logic_notes}</p>
              
              {/* Display FREE or PAID button */}
              <button onClick={handleBookFollowUp}>
                {followUpEligibility.is_free ? 'Book FREE Follow-up' : 'Book PAID Follow-up'}
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default PatientAppointmentDashboard;
```

---

## 🎯 **Key Methods to Use**

### **1. Get Clinic Patients**
```
GET /api/v1/organization/clinic/:clinicId/patients
```
Returns: Patients with appointments and follow-ups

### **2. Check Follow-Up Eligibility**
```
GET /api/v1/appointments/followup-eligibility
Query: clinic_patient_id, doctor_id, department_id
```
Returns: Eligibility status and follow-up details

### **3. Create Follow-Up Appointment**
```
POST /api/v1/appointments/create-simple
Body: Appointment data with is_followup_appointment: true
```
Returns: Created appointment with follow-up details

---

## ✅ **Complete Logic Flow**

```
Patient Click
  ↓
Show Appointments
  ↓
Click Follow-up
  ↓
Check Eligibility API
  ↓
Show FREE or PAID
  ↓
Book Follow-up
  ↓
Get Response
  ↓
Check follow_up_logic_status
  ↓
Update UI
```

---

## 🎉 **Ready to Use!**

Your frontend can now:
- ✅ Display patient list
- ✅ Show appointments and follow-ups
- ✅ Check eligibility
- ✅ Book free or paid follow-ups
- ✅ Handle all logic status values

**Perfect for production! 🚀**

