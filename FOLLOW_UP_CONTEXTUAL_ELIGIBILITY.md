# Follow-Up Contextual Eligibility - Complete Guide ✅

## 🎯 **Feature: Context-Aware Follow-Up Eligibility**

When a user **selects a doctor and department** before searching for a patient, the system now checks follow-up eligibility **specifically for that doctor+department combination**.

---

## ❌ **Old Behavior (Generic)**

```
1. User selects: Doctor A, Department Cardiology
2. User searches: Patient John
3. System shows: "Eligible for follow-up" ❌
   (Based on patient's LAST appointment with ANY doctor/department)
4. User tries to book follow-up
5. System ERROR: "Doctor mismatch" ❌
```

**Problem:** Frontend showed eligibility based on patient's last appointment (which might be different doctor/department), causing confusion.

---

## ✅ **New Behavior (Contextual)**

```
1. User selects: Doctor A, Department Cardiology
2. User searches: Patient John
3. System checks: "Did John's LAST appointment with Doctor A in Cardiology happen within 5 days?"
   - If YES → ✅ "Eligible for FREE follow-up"
   - If NO → ❌ "Not eligible" or "New appointment required"
4. User books with confidence!
```

**Benefit:** Follow-up eligibility is **specific to the selected doctor+department** before even displaying the patient!

---

## 📊 **How It Works**

### API Flow:

```
GET /api/clinic-specific-patients?clinic_id=xxx&doctor_id=yyy&department_id=zzz&search=John
```

**Parameters:**
- `clinic_id` (required): Clinic UUID
- `doctor_id` (optional): Selected doctor UUID
- `department_id` (optional): Selected department UUID
- `search` (optional): Patient name/phone/MO ID

**Response:**
```json
{
  "clinic_id": "clinic-uuid",
  "total": 1,
  "patients": [
    {
      "id": "patient-uuid",
      "first_name": "John",
      "last_name": "Doe",
      "phone": "1234567890",
      "last_appointment": {
        "doctor_id": "doctor-a-uuid",
        "doctor_name": "Dr. ABC",
        "department_id": "cardiology-uuid",
        "department": "Cardiology",
        "date": "2025-10-18",
        "days_since": 2
      },
      "follow_up_eligibility": {
        "eligible": true,
        "is_free": true,
        "days_remaining": 3,
        "message": "You have one FREE follow-up available with this doctor in this department"
      }
    }
  ]
}
```

---

## 🔍 **Logic Details**

### Without `doctor_id` & `department_id`:

```sql
-- Returns patient's LAST appointment (any doctor, any department)
SELECT * FROM appointments
WHERE clinic_patient_id = ?
  AND status IN ('completed', 'confirmed')
ORDER BY appointment_date DESC
LIMIT 1
```

**Use Case:** General patient list (show all patients with their last appointment)

---

### With `doctor_id` & `department_id`:

```sql
-- Returns patient's LAST appointment with THIS doctor in THIS department
SELECT * FROM appointments
WHERE clinic_patient_id = ?
  AND doctor_id = ?          -- ✅ Filter by selected doctor
  AND department_id = ?       -- ✅ Filter by selected department
  AND status IN ('completed', 'confirmed')
ORDER BY appointment_date DESC
LIMIT 1
```

**Use Case:** Booking screen (check if patient can book follow-up with selected doctor+department)

---

## 🧪 **Example Scenarios**

### Scenario 1: Eligible for Follow-Up ✅

**Patient History:**
```
Oct 18: Doctor A → Cardiology → Completed
```

**User Action:**
```
1. Select: Doctor A
2. Select: Cardiology
3. Search: Patient John
```

**API Call:**
```
GET /api/clinic-specific-patients?clinic_id=xxx&doctor_id=doctor-a&department_id=cardiology&search=John
```

**Response:**
```json
{
  "last_appointment": {
    "doctor_id": "doctor-a",
    "doctor_name": "Dr. A",
    "department": "Cardiology",
    "date": "2025-10-18",
    "days_since": 2
  },
  "follow_up_eligibility": {
    "eligible": true,
    "is_free": true,
    "days_remaining": 3,
    "message": "You have one FREE follow-up available..."
  }
}
```

**UI Action:**
- ✅ Enable "Book Follow-Up" button
- ✅ Show "FREE" badge

---

### Scenario 2: Different Doctor (Not Eligible) ❌

**Patient History:**
```
Oct 18: Doctor A → Cardiology → Completed
```

**User Action:**
```
1. Select: Doctor B  ← Different doctor!
2. Select: Cardiology
3. Search: Patient John
```

**API Call:**
```
GET /api/clinic-specific-patients?clinic_id=xxx&doctor_id=doctor-b&department_id=cardiology&search=John
```

**Response:**
```json
{
  "last_appointment": null,  // No appointment with Doctor B in Cardiology
  "follow_up_eligibility": {
    "eligible": false,
    "is_free": false,
    "reason": "No previous appointment found"
  }
}
```

**UI Action:**
- ❌ Disable "Book Follow-Up" button
- ✅ Enable "Book New Appointment" button

---

### Scenario 3: Different Department (Not Eligible) ❌

**Patient History:**
```
Oct 18: Doctor A → Cardiology → Completed
```

**User Action:**
```
1. Select: Doctor A
2. Select: Neurology  ← Different department!
3. Search: Patient John
```

**API Call:**
```
GET /api/clinic-specific-patients?clinic_id=xxx&doctor_id=doctor-a&department_id=neurology&search=John
```

**Response:**
```json
{
  "last_appointment": null,  // No appointment with Doctor A in Neurology
  "follow_up_eligibility": {
    "eligible": false,
    "is_free": false,
    "reason": "No previous appointment found"
  }
}
```

**UI Action:**
- ❌ Disable "Book Follow-Up" button
- ✅ Enable "Book New Appointment" button

---

### Scenario 4: Free Follow-Up Already Used ❌

**Patient History:**
```
Oct 15: Doctor A → Cardiology → Completed
Oct 16: Doctor A → Cardiology → Follow-up (FREE)
```

**User Action:**
```
1. Select: Doctor A
2. Select: Cardiology
3. Search: Patient John
```

**API Call:**
```
GET /api/clinic-specific-patients?clinic_id=xxx&doctor_id=doctor-a&department_id=cardiology&search=John
```

**Response:**
```json
{
  "last_appointment": {
    "doctor_id": "doctor-a",
    "doctor_name": "Dr. A",
    "department": "Cardiology",
    "date": "2025-10-16",
    "days_since": 4
  },
  "follow_up_eligibility": {
    "eligible": true,
    "is_free": false,  // ❌ Already used
    "message": "Free follow-up already used. Additional follow-ups require payment."
  }
}
```

**UI Action:**
- ✅ Enable "Book Follow-Up" button
- ❌ Remove "FREE" badge
- ✅ Show fee amount (₹200)

---

### Scenario 5: Multiple Departments, Each Eligible ✅

**Patient History:**
```
Oct 18: Doctor A → Cardiology → Completed
Oct 19: Doctor A → Neurology → Completed
```

**Test A - Cardiology:**
```
GET /api/clinic-specific-patients?doctor_id=doctor-a&department_id=cardiology
```
**Result:** ✅ Eligible for FREE follow-up (within 5 days, no free used in Cardiology)

**Test B - Neurology:**
```
GET /api/clinic-specific-patients?doctor_id=doctor-a&department_id=neurology
```
**Result:** ✅ Eligible for FREE follow-up (within 5 days, no free used in Neurology)

**Key Point:** Each department gets its own free follow-up! ✅

---

## 🎨 **Frontend Integration**

### Step 1: User Selects Doctor & Department

```dart
String selectedDoctorId = 'doctor-a-uuid';
String selectedDepartmentId = 'cardiology-uuid';
```

---

### Step 2: Search Patients with Context

```dart
final response = await http.get(Uri.parse(
  '$baseUrl/clinic-specific-patients'
  '?clinic_id=$clinicId'
  '&doctor_id=$selectedDoctorId'      // ✅ Pass selected doctor
  '&department_id=$selectedDepartmentId'  // ✅ Pass selected department
  '&search=$searchQuery'
));
```

---

### Step 3: Display Eligibility

```dart
if (patient.followUpEligibility?.eligible == true) {
  if (patient.followUpEligibility?.isFree == true) {
    // ✅ Show "FREE Follow-Up" button
    return ElevatedButton(
      child: Text('Book Follow-Up (FREE)'),
      style: ButtonStyle(backgroundColor: Colors.green),
      onPressed: () => bookFollowUp(patient.id),
    );
  } else {
    // ✅ Show "Paid Follow-Up" button
    return ElevatedButton(
      child: Text('Book Follow-Up (₹200)'),
      onPressed: () => bookFollowUp(patient.id),
    );
  }
} else {
  // ❌ Show "New Appointment" button only
  return ElevatedButton(
    child: Text('Book New Appointment'),
    onPressed: () => bookNewAppointment(patient.id),
  );
}
```

---

## 📋 **API Changes**

### Updated Endpoints:

#### 1. List Patients
```
GET /api/clinic-specific-patients
  ?clinic_id=xxx
  &doctor_id=yyy      ← NEW (optional)
  &department_id=zzz  ← NEW (optional)
  &search=...
```

#### 2. Get Single Patient
```
GET /api/clinic-specific-patients/:id
  ?doctor_id=yyy      ← NEW (optional)
  &department_id=zzz  ← NEW (optional)
```

---

## ✅ **Benefits**

| Benefit | Description |
|---------|-------------|
| **Accurate Eligibility** | Shows eligibility for THE SELECTED doctor+department |
| **No Confusion** | Frontend and backend always in sync |
| **Better UX** | User knows eligibility BEFORE clicking |
| **Prevents Errors** | No more "Doctor mismatch" errors |
| **Flexible** | Works with or without doctor/department context |
| **Scalable** | Supports multiple doctors, multiple departments |

---

## 🔄 **Backward Compatibility**

### Without Parameters (Still Works):
```
GET /api/clinic-specific-patients?clinic_id=xxx
```
**Result:** Shows patient's LAST appointment (any doctor/department) - same as before ✅

### With Parameters (New Feature):
```
GET /api/clinic-specific-patients?clinic_id=xxx&doctor_id=yyy&department_id=zzz
```
**Result:** Shows patient's LAST appointment with THAT doctor+department ✅

---

## 🚀 **Deployment**

### 1. Build (In Progress)
```bash
docker-compose build organization-service
```

### 2. Deploy
```bash
docker-compose up -d organization-service
```

### 3. Update Frontend
- Add `doctor_id` and `department_id` to patient search API calls
- Update UI to show contextual eligibility

---

## ✅ **Summary**

| Aspect | Value |
|--------|-------|
| **Feature** | Contextual Follow-Up Eligibility |
| **Parameters** | `doctor_id`, `department_id` (optional) |
| **Scope** | Per selected doctor+department |
| **Backward Compatible** | ✅ Yes |
| **Frontend Impact** | Must pass selected doctor+department IDs |
| **Status** | ✅ **COMPLETE** |

---

**Result:** Follow-up eligibility is now **context-aware** based on selected doctor+department! 🎉✅

