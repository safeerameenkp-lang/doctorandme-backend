# Follow-Up Appointments - Complete Implementation Guide 🔄

## 🎯 Overview

Implemented a comprehensive follow-up appointment system that ensures follow-ups are only available for previous patients within a 5-day window, with the same doctor and department, and **no payment required**.

---

## 📋 Key Features

| Feature | Description |
|---------|-------------|
| ✅ Previous Patient Only | Must have a completed/confirmed appointment |
| ✅ 5-Day Window | Must book within 5 days of last appointment |
| ✅ Doctor Match | Must be with the same doctor |
| ✅ Department Match | Must be in the same department |
| ✅ Slot Validation | Validates doctor's available schedule |
| ✅ No Payment | Follow-ups are free (payment status: "waived") |
| ✅ Eligibility Check API | Check eligibility before showing UI |

---

## 🔄 How It Works

### Follow-Up Appointment Flow:

```
1. Patient checks follow-up eligibility
   ↓
2. System checks:
   ✅ Previous appointment exists
   ✅ Within 5-day window
   ✅ Same doctor/department
   ↓
3. If eligible → Show follow-up option
   ↓
4. Patient selects follow-up slot
   ↓
5. System validates:
   ✅ Doctor match
   ✅ Department match
   ✅ Slot availability
   ↓
6. Appointment created:
   - Payment: Waived
   - Fee: 0.00
   - Status: Confirmed
```

---

## 🌐 APIs

### 1. Check Follow-Up Eligibility ⭐

**Endpoint:**
```
GET /api/appointments/check-follow-up-eligibility
```

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `clinic_patient_id` | UUID | ✅ Yes | Patient's UUID |
| `clinic_id` | UUID | ✅ Yes | Clinic's UUID |

**Request Example:**
```bash
GET /api/appointments/check-follow-up-eligibility?
  clinic_patient_id=752590e9-deda-4043-a5e2-7f9366f00cfc&
  clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2

Authorization: Bearer {token}
```

**Response (Eligible):**
```json
{
  "eligible": true,
  "message": "Patient is eligible for follow-up appointment",
  "days_remaining": 3,
  "days_since_last": 2,
  "last_appointment": {
    "id": "appointment-uuid",
    "doctor_id": "doctor-uuid",
    "doctor_name": "Dr. Ahmed Ibrahim",
    "department_id": "dept-uuid",
    "department": "Cardiology",
    "date": "2025-10-17",
    "status": "completed"
  },
  "note": "Follow-up must be with the same doctor and department"
}
```

**Response (Not Eligible - No Previous Appointment):**
```json
{
  "eligible": false,
  "reason": "No previous appointment found",
  "message": "Follow-up appointments are only available for patients with a previous appointment"
}
```

**Response (Not Eligible - Expired):**
```json
{
  "eligible": false,
  "reason": "Follow-up period expired",
  "message": "Follow-up appointments must be booked within 5 days of your previous appointment",
  "days_since_last": 7,
  "last_appointment": {
    "id": "appointment-uuid",
    "date": "2025-10-12",
    "doctor": "Dr. Ahmed Ibrahim",
    "department": "Cardiology"
  }
}
```

---

### 2. Create Follow-Up Appointment ⭐

**Endpoint:**
```
POST /api/appointments/simple
```

**Request Body:**
```json
{
  "clinic_patient_id": "752590e9-deda-4043-a5e2-7f9366f00cfc",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "department_id": "dept-uuid",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2025-10-20",
  "appointment_time": "2025-10-20 10:00:00",
  "consultation_type": "follow_up",
  "is_follow_up": true,   // ✅ NEW: Triggers follow-up validation
  "reason": "Follow-up checkup",
  "notes": "Second visit"
  // ✅ NOTE: payment_method NOT required for follow-ups
}
```

**Response (Success):**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "id": "new-appointment-uuid",
    "booking_number": "BN202510200001",
    "token_number": 1,
    "clinic_patient_id": "752590e9-deda-4043-a5e2-7f9366f00cfc",
    "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "appointment_date": "2025-10-20",
    "appointment_time": "2025-10-20T10:00:00Z",
    "consultation_type": "follow_up",
    "status": "confirmed",
    "fee_amount": 0.00,              // ✅ Zero fee
    "payment_status": "waived",       // ✅ Payment waived
    "payment_mode": null
  }
}
```

**Response (Error - No Previous Appointment):**
```json
{
  "error": "No previous appointment found",
  "message": "Follow-up appointments are only available for patients with a previous appointment"
}
```

**Response (Error - Expired):**
```json
{
  "error": "Follow-up period expired",
  "message": "Follow-up appointments must be booked within 5 days of your previous appointment. Please book a regular appointment instead.",
  "days_since_last_appointment": 7
}
```

**Response (Error - Doctor Mismatch):**
```json
{
  "error": "Doctor mismatch",
  "message": "Follow-up appointments must be with the same doctor as your previous appointment"
}
```

**Response (Error - Department Mismatch):**
```json
{
  "error": "Department mismatch",
  "message": "Follow-up appointments must be in the same department as your previous appointment"
}
```

---

## 📊 Validation Rules

### 1. Previous Appointment Check ✅

```sql
SELECT a.doctor_id, a.department_id, a.appointment_date
FROM appointments a
WHERE a.clinic_patient_id = $1
  AND a.clinic_id = $2
  AND a.status IN ('completed', 'confirmed')
  AND a.appointment_date <= CURRENT_DATE
ORDER BY a.appointment_date DESC, a.appointment_time DESC
LIMIT 1
```

**Checks:**
- Patient has at least one previous appointment
- Appointment is completed or confirmed
- Appointment date is not in future

---

### 2. 5-Day Window Check ✅

```go
daysSinceLastAppointment := time.Since(previousAppointmentDate).Hours() / 24

if daysSinceLastAppointment > 5 {
    return error("Follow-up period expired")
}
```

**Rule:** Must be ≤ 5 days since last appointment

---

### 3. Doctor Match Check ✅

```go
if previousDoctorID != currentDoctorID {
    return error("Doctor mismatch")
}
```

**Rule:** Follow-up must be with same doctor

---

### 4. Department Match Check ✅

```go
if previousDepartmentID != null && currentDepartmentID != previousDepartmentID {
    return error("Department mismatch")
}
```

**Rule:** Follow-up must be in same department (if original had one)

---

### 5. Slot Availability Check ✅

```sql
SELECT available_count, status
FROM doctor_individual_slots
WHERE id = $1
```

**Rule:** Slot must be available (same as regular appointments)

---

## 💻 Flutter UI Integration

### Step 1: Check Eligibility

```dart
class AppointmentBookingPage extends StatefulWidget {
  final String clinicPatientId;
  final String clinicId;
  
  @override
  _AppointmentBookingPageState createState() => _AppointmentBookingPageState();
}

class _AppointmentBookingPageState extends State<AppointmentBookingPage> {
  bool? isFollowUpEligible;
  Map<String, dynamic>? followUpInfo;
  bool isFollowUpMode = false;
  
  @override
  void initState() {
    super.initState();
    _checkFollowUpEligibility();
  }
  
  Future<void> _checkFollowUpEligibility() async {
    final url = Uri.parse(
      '$baseUrl/appointments/check-follow-up-eligibility?'
      'clinic_patient_id=${widget.clinicPatientId}&'
      'clinic_id=${widget.clinicId}'
    );
    
    final response = await http.get(
      url,
      headers: {'Authorization': 'Bearer $token'},
    );
    
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      setState(() {
        isFollowUpEligible = data['eligible'];
        followUpInfo = data;
      });
    }
  }
  
  // ... rest of the page
}
```

---

### Step 2: Show Follow-Up Option

```dart
Widget _buildAppointmentTypeSelector() {
  return Card(
    child: Column(
      children: [
        // Regular Appointment Option
        RadioListTile<bool>(
          title: Text('Regular Appointment'),
          subtitle: Text('Standard consultation with payment'),
          value: false,
          groupValue: isFollowUpMode,
          onChanged: (value) {
            setState(() => isFollowUpMode = value!);
          },
        ),
        
        // Follow-Up Option (only if eligible)
        if (isFollowUpEligible == true)
          RadioListTile<bool>(
            title: Text('Follow-Up Appointment'),
            subtitle: Text(
              '✅ FREE - No payment required\n'
              'Doctor: ${followUpInfo!['last_appointment']['doctor_name']}\n'
              'Days remaining: ${followUpInfo!['days_remaining']}'
            ),
            value: true,
            groupValue: isFollowUpMode,
            onChanged: (value) {
              setState(() => isFollowUpMode = value!);
            },
          ),
        
        // Show message if not eligible
        if (isFollowUpEligible == false)
          ListTile(
            leading: Icon(Icons.info, color: Colors.grey),
            title: Text('Follow-up not available'),
            subtitle: Text(followUpInfo!['message']),
          ),
      ],
    ),
  );
}
```

---

### Step 3: Filter Slots (For Follow-Up)

```dart
Future<void> _loadSlots() async {
  // If follow-up, must use same doctor
  String doctorId = isFollowUpMode 
    ? followUpInfo!['last_appointment']['doctor_id']
    : selectedDoctorId;
  
  // If follow-up, must use same department
  String? departmentId = isFollowUpMode 
    ? followUpInfo!['last_appointment']['department_id']
    : selectedDepartmentId;
  
  final url = Uri.parse(
    '$baseUrl/organizations/doctor-session-slots?'
    'doctor_id=$doctorId&'
    'clinic_id=${widget.clinicId}&'
    'date=$selectedDate'
  );
  
  final response = await http.get(
    url,
    headers: {'Authorization': 'Bearer $token'},
  );
  
  // ... process slots
}
```

---

### Step 4: Book Appointment

```dart
Future<void> _bookAppointment() async {
  final body = {
    'clinic_patient_id': widget.clinicPatientId,
    'doctor_id': isFollowUpMode 
      ? followUpInfo!['last_appointment']['doctor_id']
      : selectedDoctorId,
    'clinic_id': widget.clinicId,
    'department_id': isFollowUpMode 
      ? followUpInfo!['last_appointment']['department_id']
      : selectedDepartmentId,
    'individual_slot_id': selectedSlotId,
    'appointment_date': selectedDate,
    'appointment_time': selectedTime,
    'consultation_type': isFollowUpMode ? 'follow_up' : 'offline',
    'is_follow_up': isFollowUpMode,  // ✅ Key flag
    'reason': reasonController.text,
    'notes': notesController.text,
  };
  
  // ✅ Only add payment for regular appointments
  if (!isFollowUpMode) {
    body['payment_method'] = selectedPaymentMethod;
    if (selectedPaymentMethod == 'pay_now') {
      body['payment_type'] = selectedPaymentType;
    }
  }
  
  final response = await http.post(
    Uri.parse('$baseUrl/appointments/simple'),
    headers: {
      'Authorization': 'Bearer $token',
      'Content-Type': 'application/json',
    },
    body: jsonEncode(body),
  );
  
  if (response.statusCode == 201) {
    final data = jsonDecode(response.body);
    _showSuccessDialog(data['appointment']);
  } else {
    final error = jsonDecode(response.body);
    _showErrorDialog(error['message'] ?? error['error']);
  }
}
```

---

### Step 5: Success Dialog

```dart
void _showSuccessDialog(Map<String, dynamic> appointment) {
  showDialog(
    context: context,
    builder: (context) => AlertDialog(
      title: Text('✅ Appointment Booked'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (appointment['payment_status'] == 'waived')
            Container(
              padding: EdgeInsets.all(8),
              color: Colors.green[50],
              child: Row(
                children: [
                  Icon(Icons.check_circle, color: Colors.green),
                  SizedBox(width: 8),
                  Text(
                    'Follow-Up - NO PAYMENT REQUIRED',
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      color: Colors.green[700],
                    ),
                  ),
                ],
              ),
            ),
          SizedBox(height: 16),
          Text('Booking Number: ${appointment['booking_number']}'),
          Text('Token Number: ${appointment['token_number']}'),
          Text('Date: ${appointment['appointment_date']}'),
          Text('Time: ${appointment['appointment_time']}'),
          Text('Fee: ₹${appointment['fee_amount']}'),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text('OK'),
        ),
      ],
    ),
  );
}
```

---

## 🧪 Testing Scenarios

### Test 1: Eligible Patient ✅

**Setup:**
- Patient has appointment on 2025-10-17
- Current date: 2025-10-19 (2 days later)
- Same doctor & department

**Request:**
```json
{
  "is_follow_up": true,
  "doctor_id": "same-doctor-uuid",
  "department_id": "same-dept-uuid",
  ...
}
```

**Expected:** ✅ Success - Appointment created with payment waived

---

### Test 2: Not Eligible - No Previous Appointment ❌

**Setup:**
- New patient with no previous appointments

**Request:**
```json
{
  "is_follow_up": true,
  ...
}
```

**Expected:** ❌ Error 400 - "No previous appointment found"

---

### Test 3: Not Eligible - Expired Window ❌

**Setup:**
- Patient has appointment on 2025-10-10
- Current date: 2025-10-20 (10 days later)

**Request:**
```json
{
  "is_follow_up": true,
  ...
}
```

**Expected:** ❌ Error 400 - "Follow-up period expired"

---

### Test 4: Wrong Doctor ❌

**Setup:**
- Patient has appointment with Doctor A
- Trying to book follow-up with Doctor B

**Request:**
```json
{
  "is_follow_up": true,
  "doctor_id": "different-doctor-uuid",
  ...
}
```

**Expected:** ❌ Error 400 - "Doctor mismatch"

---

### Test 5: Wrong Department ❌

**Setup:**
- Patient has appointment in Cardiology
- Trying to book follow-up in Orthopedics

**Request:**
```json
{
  "is_follow_up": true,
  "department_id": "different-dept-uuid",
  ...
}
```

**Expected:** ❌ Error 400 - "Department mismatch"

---

### Test 6: Regular Appointment (Non-Follow-Up) ✅

**Request:**
```json
{
  "is_follow_up": false,
  "payment_method": "pay_now",
  "payment_type": "cash",
  ...
}
```

**Expected:** ✅ Success - Normal appointment with payment

---

## 📝 Files Changed

| File | Changes |
|------|---------|
| `appointment_simple.controller.go` | Added follow-up validation logic |
| `appointment.routes.go` | Added eligibility check route |

---

## ✅ Implementation Checklist

| Feature | Status |
|---------|--------|
| Previous patient validation | ✅ Done |
| 5-day window check | ✅ Done |
| Doctor match validation | ✅ Done |
| Department match validation | ✅ Done |
| Slot availability check | ✅ Done |
| No payment for follow-ups | ✅ Done |
| Zero fee for follow-ups | ✅ Done |
| Eligibility check API | ✅ Done |
| Route added | ✅ Done |
| No linter errors | ✅ Done |
| Documentation | ✅ Done |

---

## 🎉 Status

**Implementation:** ✅ **COMPLETE**

**Features:**
- ✅ All validation rules implemented
- ✅ Payment waived for follow-ups
- ✅ Eligibility check API
- ✅ Flutter integration guide
- ✅ Comprehensive documentation

**Ready for:**
- ✅ Testing
- ✅ UI integration
- ✅ Production deployment

---

**Done!** 🔄✅🎉

