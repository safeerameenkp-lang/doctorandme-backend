# UI Implementation: Clinic-Wise Doctor Slots

## 🎯 Your Scenario

**One Doctor → Multiple Clinics → Different Slots**

```
Dr. John Doe:
├── ABC Clinic
│   └── Morning: 09:00 - 12:00 ✅
│
└── XYZ Clinic
    └── Afternoon: 14:00 - 17:00 ✅
```

**When User Clicks:**
- Click "Dr. John at ABC Clinic" → Show 09:00-12:00 only
- Click "Dr. John at XYZ Clinic" → Show 14:00-17:00 only

---

## 📱 UI Flow

### Step 1: Show Clinics List

```
┌──────────────────────────────┐
│   SELECT CLINIC              │
├──────────────────────────────┤
│  🏥 ABC Clinic      [Select] │
│  🏥 XYZ Clinic      [Select] │
└──────────────────────────────┘
```

### Step 2: User Selects ABC Clinic → Show Doctors

**API Call:**
```
GET /api/doctors/clinic/{abc-clinic-id}
```

**Shows:**
```
┌──────────────────────────────┐
│   DOCTORS AT ABC CLINIC      │
├──────────────────────────────┤
│  👨‍⚕️ Dr. John Doe   [View]   │
│  👨‍⚕️ Dr. Sarah Lee  [View]   │
└──────────────────────────────┘
```

### Step 3: User Clicks "Dr. John" → Show Slots for ABC Clinic

**API Call:**
```
GET /api/doctor-time-slots?doctor_id={john-id}&clinic_id={abc-clinic-id}
```

**Shows ONLY ABC Clinic Slots:**
```
┌──────────────────────────────┐
│   DR. JOHN AT ABC CLINIC     │
├──────────────────────────────┤
│  📅 Monday                   │
│     🌅 09:00 - 12:00 [Book] │
│                              │
│  📅 Wednesday                │
│     🌅 09:00 - 12:00 [Book] │
└──────────────────────────────┘
```

### Step 4: User Selects XYZ Clinic → Shows Different Slots

**API Call:**
```
GET /api/doctor-time-slots?doctor_id={john-id}&clinic_id={xyz-clinic-id}
```

**Shows ONLY XYZ Clinic Slots:**
```
┌──────────────────────────────┐
│   DR. JOHN AT XYZ CLINIC     │
├──────────────────────────────┤
│  📅 Monday                   │
│     🌆 14:00 - 17:00 [Book] │
│                              │
│  📅 Wednesday                │
│     🌆 14:00 - 17:00 [Book] │
└──────────────────────────────┘
```

---

## 💻 Frontend Implementation

### React/Vue/Angular Example

```javascript
// State variables
const [selectedClinic, setSelectedClinic] = useState(null);
const [selectedDoctor, setSelectedDoctor] = useState(null);
const [slots, setSlots] = useState([]);

// Step 1: User selects clinic
function handleClinicClick(clinic) {
  setSelectedClinic(clinic);
  
  // Get doctors for this clinic
  fetch(`/api/doctors/clinic/${clinic.id}`, {
    headers: { 'Authorization': `Bearer ${token}` }
  })
  .then(res => res.json())
  .then(data => {
    // Show doctors list
    setDoctors(data.doctors);
  });
}

// Step 2: User selects doctor
function handleDoctorClick(doctor) {
  setSelectedDoctor(doctor);
  
  // Get slots for THIS doctor at THIS clinic
  fetch(`/api/doctor-time-slots?doctor_id=${doctor.doctor_id}&clinic_id=${selectedClinic.id}`, {
    headers: { 'Authorization': `Bearer ${token}` }
  })
  .then(res => res.json())
  .then(data => {
    // Show clinic-specific slots
    setSlots(data.time_slots);
  });
}

// Render slots
{slots.map(slot => (
  <div key={slot.id}>
    <h4>{slot.day_name}</h4>
    <p>{slot.start_time} - {slot.end_time}</p>
    <button onClick={() => bookAppointment(slot)}>Book</button>
  </div>
))}
```

---

## 📋 Complete API Flow

### Flow 1: Book at ABC Clinic

```
User Journey                          API Call
────────────────────────────────────────────────────────────────

1. User opens app                     
                                      
2. User clicks "ABC Clinic"           GET /api/doctors/clinic/{abc-id}
   → Shows: Dr. John, Dr. Sarah       Response: [{doctor_id, full_name, ...}]

3. User clicks "Dr. John"             GET /api/doctor-time-slots
   → Shows ABC morning slots            ?doctor_id={john-id}
                                        &clinic_id={abc-id}
                                      
                                      Response: [
                                        { start_time: "09:00", 
                                          end_time: "12:00",
                                          clinic_name: "ABC Clinic" }
                                      ]

4. User clicks "09:00-12:00"          POST /api/appointments
   → Books appointment                  { doctor_id, clinic_id, time }
```

### Flow 2: Book at XYZ Clinic

```
User Journey                          API Call
────────────────────────────────────────────────────────────────

1. User clicks "XYZ Clinic"           GET /api/doctors/clinic/{xyz-id}
   → Shows: Dr. John, Dr. Mike        Response: [{doctor_id, full_name, ...}]

2. User clicks "Dr. John"             GET /api/doctor-time-slots
   → Shows XYZ afternoon slots          ?doctor_id={john-id}
                                        &clinic_id={xyz-id}
                                      
                                      Response: [
                                        { start_time: "14:00", 
                                          end_time: "17:00",
                                          clinic_name: "XYZ Clinic" }
                                      ]

3. User books afternoon slot          POST /api/appointments
```

---

## 🧪 PowerShell Test Example

```powershell
$token = "your-jwt-token"
$johnDoctorId = "doctor-john-uuid"
$abcClinicId = "abc-clinic-uuid"
$xyzClinicId = "xyz-clinic-uuid"

# ===== Scenario 1: User clicks ABC Clinic =====
Write-Host "User clicked: ABC Clinic" -ForegroundColor Cyan

# Get slots for Dr. John at ABC Clinic
$abcSlots = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots?doctor_id=$johnDoctorId&clinic_id=$abcClinicId" `
    -Headers @{"Authorization" = "Bearer $token"}

Write-Host "Showing slots for Dr. John at ABC Clinic:"
$abcSlots.time_slots | ForEach-Object {
    Write-Host "  🌅 $($_.day_name) $($_.start_time)-$($_.end_time)" -ForegroundColor Green
}

# Output:
# 🌅 Monday 09:00-12:00
# 🌅 Wednesday 09:00-12:00

# ===== Scenario 2: User clicks XYZ Clinic =====
Write-Host "`nUser clicked: XYZ Clinic" -ForegroundColor Cyan

# Get slots for Dr. John at XYZ Clinic
$xyzSlots = Invoke-RestMethod `
    -Uri "http://localhost:8081/api/doctor-time-slots?doctor_id=$johnDoctorId&clinic_id=$xyzClinicId" `
    -Headers @{"Authorization" = "Bearer $token"}

Write-Host "Showing slots for Dr. John at XYZ Clinic:"
$xyzSlots.time_slots | ForEach-Object {
    Write-Host "  🌆 $($_.day_name) $($_.start_time)-$($_.end_time)" -ForegroundColor Cyan
}

# Output:
# 🌆 Monday 14:00-17:00
# 🌆 Wednesday 14:00-17:00
```

---

## 🎨 UI Component Example (Pseudocode)

```
┌─────────────────────────────────────────────┐
│  CLINIC SELECTION                           │
│  ○ ABC Clinic                               │
│  ● XYZ Clinic  [Selected]                   │
└─────────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────────┐
│  DOCTORS AT XYZ CLINIC                      │
│  [Dr. John Doe]                             │
│  [Dr. Mike Smith]                           │
└─────────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────────┐
│  DR. JOHN DOE'S SLOTS AT XYZ CLINIC         │
│                                             │
│  Monday                                     │
│  ├─ 🌆 14:00 - 17:00  [Book Now]           │
│                                             │
│  Wednesday                                  │
│  ├─ 🌆 14:00 - 17:00  [Book Now]           │
└─────────────────────────────────────────────┘
```

---

## ✅ Key Points for Your UI

1. **Always pass both `doctor_id` AND `clinic_id`** when getting slots
2. **Clinic selection comes first** in your UI flow
3. **Then show doctors** for that clinic
4. **Then show slots** for that doctor at that specific clinic
5. **Slots automatically filter** based on clinic-doctor link

---

## 🔑 API Endpoints You Need

### 1. Get Doctors at a Clinic
```
GET /api/doctors/clinic/{clinic_id}
```
**Returns:** All doctors working at this clinic

### 2. Get Doctor's Slots at Specific Clinic
```
GET /api/doctor-time-slots?doctor_id={id}&clinic_id={id}
```
**Returns:** Only slots for this doctor at this clinic

### 3. Book Appointment
```
POST /api/appointments
{
  "doctor_id": "...",
  "clinic_id": "...",
  "slot_id": "...",
  "appointment_date": "2025-10-15",
  "appointment_time": "09:00"
}
```

---

## 📊 Visual Data Flow

```
┌──────────────┐
│ Select       │
│ ABC Clinic   │
└──────┬───────┘
       │
       ↓
┌──────────────────────────────┐
│ GET /api/doctors/clinic/abc  │
└──────┬───────────────────────┘
       │
       ↓ Returns: [Dr. John, Dr. Sarah]
       │
┌──────────────┐
│ Click        │
│ Dr. John     │
└──────┬───────┘
       │
       ↓
┌─────────────────────────────────────────┐
│ GET /api/doctor-time-slots              │
│   ?doctor_id=john-id                    │
│   &clinic_id=abc-id  ← IMPORTANT!       │
└──────┬──────────────────────────────────┘
       │
       ↓ Returns: Only ABC Clinic slots
       │
┌──────────────────┐
│ Show:            │
│ 🌅 09:00-12:00  │  ← Morning slots at ABC
│ 🌅 09:00-12:00  │
└──────────────────┘
```

---

## 💡 Implementation Checklist

- [ ] User selects clinic first
- [ ] Fetch doctors for selected clinic
- [ ] Store selected clinic ID in state
- [ ] When doctor clicked, fetch slots with BOTH doctor_id and clinic_id
- [ ] Display only the slots for that clinic
- [ ] When booking, include both doctor_id and clinic_id

---

## 🚀 Quick Start Code

```javascript
// 1. Clinic selection
const selectClinic = async (clinicId) => {
  const response = await fetch(`/api/doctors/clinic/${clinicId}`);
  const data = await response.json();
  return data.doctors;
};

// 2. Get clinic-specific slots for doctor
const getDoctorSlots = async (doctorId, clinicId) => {
  const response = await fetch(
    `/api/doctor-time-slots?doctor_id=${doctorId}&clinic_id=${clinicId}`
  );
  const data = await response.json();
  return data.time_slots;
};

// 3. Usage
const clinicId = 'abc-clinic-uuid';
const doctorId = 'john-doctor-uuid';

const slots = await getDoctorSlots(doctorId, clinicId);
// Returns: [{ start_time: "09:00", end_time: "12:00", ... }]
```

---

## ✅ Summary

**Your UI Flow:**
1. User selects clinic → Store `clinicId`
2. Show doctors for that clinic
3. User clicks doctor → Store `doctorId`
4. Fetch slots: `doctor_id={doctorId}&clinic_id={clinicId}`
5. Show clinic-specific slots
6. User books slot

**The API automatically filters slots based on:**
- ✅ Doctor-clinic link exists
- ✅ Link is active
- ✅ Slots match the clinic

**Result:**
- ABC Clinic → Shows morning slots
- XYZ Clinic → Shows afternoon slots
- Same doctor, different clinics, different slots! 🎉


