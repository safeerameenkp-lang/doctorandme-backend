# Frontend Quick Implementation - Complete Flow 🎯

## 🎯 **Your Question Answered!**

You asked: **"How do I handle patient click, appointment, daily follow-up in frontend?"**

Here's the **complete answer** with **exact methods** to use:

---

## 📋 **Complete Frontend Flow**

### **Step 1: Click Patient → Get Details**

```typescript
// When user clicks a patient
const handlePatientClick = (patient) => {
  // Patient object already has ALL data you need!
  console.log("Patient:", patient.first_name);
  console.log("Appointments:", patient.appointments); // ✅ Already loaded!
  console.log("Follow-ups:", patient.follow_ups);     // ✅ Already loaded!
  
  // No need to call API again - data is already there!
}
```

**Why?** The `/api/v1/organization/clinic/:clinicId/patients` endpoint already returns:
- ✅ All appointments (never null)
- ✅ All follow-ups (never null)
- ✅ Current follow-up status
- ✅ Last appointment and follow-up IDs

---

### **Step 2: Display Appointments**

```typescript
// Display appointments from the patient object
const AppointmentList = ({ patient }) => {
  return (
    <div>
      <h3>Appointments</h3>
      {patient.appointments.map(apt => (
        <div key={apt.id}>
          <p>Date: {apt.appointment_date}</p>
          <p>Doctor: {apt.doctor_name}</p>
          <p>Status: {apt.status}</p>
          <p>Type: {apt.slot_type}</p>
        </div>
      ))}
    </div>
  );
};
```

---

### **Step 3: Display Follow-Ups**

```typescript
// Display follow-ups from the patient object
const FollowUpList = ({ patient }) => {
  return (
    <div>
      <h3>Follow-ups</h3>
      {patient.follow_ups.map(followUp => (
        <div key={followUp.id}>
          <p>Status: {followUp.status}</p>
          <p>Logic Status: {followUp.follow_up_logic_status}</p>
          <p>Days Remaining: {followUp.days_remaining}</p>
          <p>Is Free: {followUp.is_free ? 'Yes' : 'No'}</p>
          <p>Valid Until: {followUp.valid_until}</p>
        </div>
      ))}
    </div>
  );
};
```

---

### **Step 4: Book Follow-Up Appointment**

```typescript
// When user wants to book a follow-up
const bookFollowUp = async (clinicPatientId, doctorId, departmentId, slotData) => {
  
  // Create appointment data
  const appointmentData = {
    clinic_id: clinicId,
    clinic_patient_id: clinicPatientId,
    doctor_id: doctorId,
    department_id: departmentId,
    individual_slot_id: slotData.slot_id,
    appointment_date: slotData.date,
    appointment_time: slotData.time,
    consultation_type: "follow-up-via-clinic" // or "follow-up-via-video"
    // No payment needed - it will be FREE or PAID based on eligibility
  };
  
  // Call API
  const response = await fetch('/api/v1/appointments/create-simple', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${accessToken}`
    },
    body: JSON.stringify(appointmentData)
  });
  
  const result = await response.json();
  
  if (result.message === "Appointment created successfully") {
    alert("Follow-up booked successfully!");
    
    // Check if it was free or paid
    if (result.is_free_followup) {
      console.log("FREE follow-up used!");
    } else {
      console.log("PAID follow-up booked");
    }
    
    // Refresh patient list
    fetchPatients();
  }
};
```

---

## 🎯 **Which Methods to Use?**

### **Method 1: Get All Patients (Already has appointments & follow-ups)**

```typescript
GET /api/v1/organization/clinic/:clinicId/patients
```
**Returns:**
```json
{
  "patients": [
    {
      "clinic_patient_id": "...",
      "first_name": "John",
      "last_name": "Doe",
      "appointments": [...],  // ✅ Already here!
      "follow_ups": [...],     // ✅ Already here!
      "current_followup_status": "active"
    }
  ]
}
```

**✅ Use this for:** Displaying patient list with all data

---

### **Method 2: Book Follow-Up (Auto-detects FREE or PAID)**

```typescript
POST /api/v1/appointments/create-simple
```
**Request:**
```json
{
  "clinic_id": "...",
  "clinic_patient_id": "...",
  "doctor_id": "...",
  "consultation_type": "follow-up-via-clinic" // or "follow-up-via-video"
  // ... other appointment fields
}
```

**Response:**
```json
{
  "message": "Appointment created successfully",
  "is_free_followup": true,  // or false
  "follow_up": {
    "follow_up_logic_status": "used",
    "logic_notes": "..."
  }
}
```

**✅ Use this for:** Booking follow-up appointments

---

## 🔄 **Complete Flow Example**

```typescript
// 1. Fetch patients (has appointments + follow-ups)
const patients = await fetchPatients();

// 2. User clicks patient
const handlePatientClick = (patient) => {
  // Show appointments (already in patient object)
  showAppointments(patient.appointments);
  
  // Show follow-ups (already in patient object)
  showFollowUps(patient.follow_ups);
};

// 3. User wants to book follow-up
const handleBookFollowUp = async () => {
  // Check which follow-up type to book
  const consultationType = selectedType === 'clinic' 
    ? 'follow-up-via-clinic' 
    : 'follow-up-via-video';
  
  // Book follow-up
  const result = await bookFollowUp({
    clinic_id: clinicId,
    clinic_patient_id: selectedPatientId,
    doctor_id: doctorId,
    consultation_type: consultationType,
    // ... other appointment data
  });
  
  if (result.is_free_followup) {
    alert("FREE follow-up booked!");
  } else {
    alert("PAID follow-up booked!");
  }
};
```

---

## ✅ **Complete Answer**

**Your question:** "Which method use?"

**Answer:**
1. **For Patient List:** Use `GET /api/v1/organization/clinic/:clinicId/patients`
   - ✅ Has all appointments
   - ✅ Has all follow-ups
   - ✅ Has follow-up status
   
2. **For Booking Follow-Up:** Use `POST /api/v1/appointments/create-simple`
   - ✅ Auto-detects FREE or PAID
   - ✅ Sets `consultation_type` to "follow-up-via-clinic" or "follow-up-via-video"
   - ✅ Returns follow-up details in response

**That's it! Just use these 2 methods! 🚀**

