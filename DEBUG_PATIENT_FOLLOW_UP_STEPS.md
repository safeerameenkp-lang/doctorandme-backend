# Debug Patient Follow-Up - Step-by-Step Guide 🔍

## 🎯 **Quick Diagnosis**

If a patient appears **in red** (not eligible) after booking an appointment, follow these steps:

---

## Step 1: Check Patient's Appointments

```sql
-- Replace UUIDs with your actual values
SELECT 
    a.id,
    a.appointment_date,
    a.appointment_time,
    a.status,
    a.consultation_type,
    a.doctor_id,
    a.department_id,
    dept.name as department,
    a.payment_status,
    CASE 
        WHEN a.appointment_date < CURRENT_DATE THEN 'PAST'
        WHEN a.appointment_date = CURRENT_DATE THEN 'TODAY'
        ELSE 'FUTURE'
    END as timing
FROM appointments a
LEFT JOIN departments dept ON dept.id = a.department_id
WHERE a.clinic_patient_id = 'YOUR_PATIENT_ID'  -- ← Replace this
  AND a.clinic_id = 'YOUR_CLINIC_ID'           -- ← Replace this
ORDER BY a.appointment_date DESC, a.appointment_time DESC;
```

### ✅ What to Check:
1. **Status:** Must be `confirmed` or `completed`
2. **Timing:** Is it TODAY, PAST, or FUTURE?
3. **Consultation Type:** Should be `video_consultation` or `clinic_visit` (NOT follow-up)
4. **Department:** Does it match what you selected in UI?

---

## Step 2: Test API Call

```bash
# Replace with your actual values
curl -X GET 'http://localhost:8081/api/clinic-specific-patients?clinic_id=YOUR_CLINIC_ID&doctor_id=YOUR_DOCTOR_ID&department_id=YOUR_DEPT_ID&search=patient_name' \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

### ✅ Expected Response:
```json
{
  "patients": [
    {
      "id": "patient-uuid",
      "first_name": "John",
      "last_appointment": {
        "doctor_id": "doctor-uuid",
        "department_id": "dept-uuid",
        "date": "2025-10-20",
        "days_since": 0
      },
      "follow_up_eligibility": {
        "eligible": true,    ← Should be true
        "is_free": true,     ← Should be true (if within 5 days)
        "message": "You have one FREE follow-up available..."
      }
    }
  ]
}
```

---

## Step 3: Common Issues & Fixes

### Issue A: Appointment is FUTURE ⏳

**Symptom:**
```json
{
  "follow_up_eligibility": {
    "eligible": false,
    "reason": "Last appointment is scheduled for the future..."
  }
}
```

**Cause:** You booked appointment for TOMORROW or later

**Fix:** 
- ✅ **Option 1:** Change appointment date to TODAY or PAST
- ✅ **Option 2:** Wait until appointment date arrives

```sql
-- Change appointment date to TODAY (for testing)
UPDATE appointments
SET appointment_date = CURRENT_DATE
WHERE id = 'APPOINTMENT_ID';
```

---

### Issue B: Wrong Doctor/Department 🔄

**Symptom:**
```json
{
  "last_appointment": null,
  "follow_up_eligibility": {
    "eligible": false,
    "reason": "No previous appointment found"
  }
}
```

**Cause:** 
- Patient's last appointment was with different doctor
- OR patient's last appointment was in different department

**Check:**
```sql
-- See which doctor+department the appointment is with
SELECT 
    doctor_id,
    department_id,
    appointment_date
FROM appointments
WHERE clinic_patient_id = 'YOUR_PATIENT_ID'
ORDER BY appointment_date DESC
LIMIT 1;
```

**Fix:**
- ✅ Search patient with the SAME doctor_id and department_id from the appointment

---

### Issue C: Status Not Confirmed ❌

**Symptom:**
API returns `eligible: false` even though appointment is TODAY

**Cause:** Appointment status is `pending` not `confirmed`

**Check:**
```sql
SELECT status
FROM appointments
WHERE id = 'APPOINTMENT_ID';
```

**Fix:**
```sql
-- Update status to confirmed
UPDATE appointments
SET status = 'confirmed'
WHERE id = 'APPOINTMENT_ID';
```

---

### Issue D: Consultation Type is Follow-Up 🔄

**Symptom:**
Patient has appointment but it doesn't show in eligibility

**Cause:** The appointment itself is a follow-up, which shouldn't be the base for another follow-up

**Check:**
```sql
SELECT consultation_type
FROM appointments
WHERE clinic_patient_id = 'YOUR_PATIENT_ID'
ORDER BY appointment_date DESC
LIMIT 1;
```

**Fix:**
- ✅ Patient needs a REGULAR appointment first (clinic_visit or video_consultation)
- ❌ Can't do follow-up based on another follow-up

---

### Issue E: Department_id is NULL 🔍

**Symptom:**
Appointment exists but doesn't match when searching with department_id

**Cause:** Appointment has NULL department_id

**Check:**
```sql
SELECT 
    id,
    doctor_id,
    department_id,
    CASE WHEN department_id IS NULL THEN 'NO DEPARTMENT' ELSE 'HAS DEPARTMENT' END
FROM appointments
WHERE clinic_patient_id = 'YOUR_PATIENT_ID';
```

**Fix Option 1: Add Department to Appointment**
```sql
UPDATE appointments
SET department_id = 'YOUR_DEPT_ID'
WHERE id = 'APPOINTMENT_ID'
  AND department_id IS NULL;
```

**Fix Option 2: Search Without Department**
```bash
# Don't pass department_id parameter
curl -X GET 'http://localhost:8081/api/clinic-specific-patients?clinic_id=xxx&doctor_id=yyy&search=name'
```

---

## Step 4: Check Frontend API Call

### ✅ Correct API Call:
```dart
// When doctor and department are selected:
final response = await http.get(
  Uri.parse(
    '$baseUrl/clinic-specific-patients'
    '?clinic_id=$clinicId'
    '&doctor_id=$selectedDoctorId'       // ✅ MUST PASS
    '&department_id=$selectedDepartmentId'  // ✅ MUST PASS
    '&search=$searchQuery'
  ),
  headers: {'Authorization': 'Bearer $token'},
);
```

### ❌ Wrong API Call:
```dart
// Missing doctor_id and department_id
final response = await http.get(
  Uri.parse(
    '$baseUrl/clinic-specific-patients'
    '?clinic_id=$clinicId'
    '&search=$searchQuery'  // ❌ Missing context!
  ),
);
```

**Result:** Will show patient's LAST appointment with ANY doctor/department, not the SELECTED one

---

## Step 5: Check UI State Refresh

### Ensure API is Called After Selection:

```dart
class PatientSearchScreen extends StatefulWidget {
  String? selectedDoctorId;
  String? selectedDepartmentId;
  
  // ✅ Call API when doctor/department changes
  void onDoctorChanged(String? newDoctorId) {
    setState(() {
      selectedDoctorId = newDoctorId;
      searchPatients();  // ✅ Refresh patient list
    });
  }
  
  void onDepartmentChanged(String? newDeptId) {
    setState(() {
      selectedDepartmentId = newDeptId;
      searchPatients();  // ✅ Refresh patient list
    });
  }
  
  Future<void> searchPatients() async {
    // ✅ Always pass selected doctor and department
    final response = await api.getPatients(
      clinicId: clinicId,
      doctorId: selectedDoctorId,
      departmentId: selectedDepartmentId,
      search: searchQuery,
    );
    
    setState(() {
      patients = response.patients;
    });
  }
}
```

---

## 🎯 **Quick Checklist**

When patient shows **in red** (not eligible):

- [ ] Check if appointment date is TODAY or PAST (not future)
- [ ] Check if appointment status is `confirmed` or `completed`
- [ ] Check if appointment consultation_type is `clinic_visit` or `video_consultation` (not follow-up)
- [ ] Check if doctor_id matches between appointment and search
- [ ] Check if department_id matches between appointment and search
- [ ] Check if frontend passes `doctor_id` and `department_id` to API
- [ ] Check if UI refreshes after selecting doctor/department
- [ ] Rebuild services (`docker-compose build organization-service`)
- [ ] Restart services (`docker-compose up -d`)

---

## 🚀 **Test Your Fix**

### Test 1: Create Test Appointment
```sql
-- Create a test appointment for TODAY
INSERT INTO appointments (
  id, clinic_id, clinic_patient_id, doctor_id, department_id,
  appointment_date, appointment_time, consultation_type,
  status, payment_status, fee_amount
) VALUES (
  gen_random_uuid(),
  'YOUR_CLINIC_ID',
  'YOUR_PATIENT_ID',
  'YOUR_DOCTOR_ID',
  'YOUR_DEPT_ID',
  CURRENT_DATE,  -- ✅ TODAY
  CURRENT_TIME,
  'video_consultation',  -- ✅ Regular appointment
  'confirmed',  -- ✅ Confirmed status
  'paid',
  500.00
);
```

### Test 2: Search Patient
```bash
curl -X GET 'http://localhost:8081/api/clinic-specific-patients?clinic_id=XXX&doctor_id=YYY&department_id=ZZZ&search=patient_name' \
  -H 'Authorization: Bearer TOKEN'
```

### Test 3: Verify Response
```json
{
  "follow_up_eligibility": {
    "eligible": true,  ← ✅ Should be true
    "is_free": true,   ← ✅ Should be true
    "message": "You have one FREE follow-up available..."
  }
}
```

---

## ✅ **Success Indicators**

Patient is eligible when:
- ✅ Name shows in **GREEN** (or not red)
- ✅ Badge shows **"FREE Follow-Up"**
- ✅ API returns `eligible: true, is_free: true`
- ✅ Button enabled for booking follow-up

---

**If still showing red after all checks, share the SQL query results and API response for further diagnosis!** 🔍

