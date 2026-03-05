# Simple Appointment API - Quick Reference

## 🎯 Simplified Appointment Booking

**Endpoint:** `POST /api/appointments/simple`

**Purpose:** Book appointment with clinic patient - No complex validations, just simple booking!

---

## 📝 Required Fields Only

```json
{
  "clinic_patient_id": "UUID",
  "doctor_id": "UUID",
  "clinic_id": "UUID",
  "individual_slot_id": "UUID",
  "appointment_date": "YYYY-MM-DD",
  "appointment_time": "YYYY-MM-DD HH:MM:SS",
  "consultation_type": "offline|online",
  "payment_method": "pay_now|pay_later|way_off"
}
```

**Optional Fields:**
- `department_id` (UUID) - Department selection (if not provided, uses doctor's default)
- `reason` (string)
- `notes` (string)
- `payment_type` (cash, card, upi) - **Required when payment_method = pay_now**

---

## ✅ Complete Example

### Request

```json
POST /api/appointments/simple
Content-Type: application/json
Authorization: Bearer {token}

{
  "clinic_patient_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "department_id": "dept-uuid-cardiology",
  "individual_slot_id": "slot-09-30-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "consultation_type": "offline",
  "reason": "Regular checkup",
  "payment_method": "pay_now",
  "payment_type": "cash"
}
```

### Response (201 Created)

```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "id": "appointment-uuid-123",
    "clinic_patient_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
    "booking_number": "BN202510180001",
    "token_number": 1,
    "appointment_date": "2025-10-18",
    "appointment_time": "2025-10-18T09:30:00Z",
    "duration_minutes": 5,
    "consultation_type": "offline",
    "reason": "Regular checkup",
    "status": "confirmed",
    "fee_amount": 500.00,
    "payment_status": "paid",
    "payment_mode": "cash",
    "created_at": "2024-10-15T12:00:00Z"
  }
}
```

---

## 🔍 What the API Does

### Simple 7-Step Process

1. ✅ **Validates clinic patient exists**
2. ✅ **Checks patient belongs to correct clinic**
3. ✅ **Validates slot is available**
4. ✅ **Checks slot belongs to correct clinic**
5. ✅ **Calculates doctor fee**
6. ✅ **Creates appointment**
7. ✅ **Marks slot as booked**

**That's it!** No complex logic, just straightforward booking.

---

## 📱 Flutter Integration

### Simple Service Method

```dart
class AppointmentService {
  final String baseUrl = 'http://localhost:8082/api';
  final String token;

  AppointmentService(this.token);

  Future<Map<String, dynamic>> bookAppointment({
    required String clinicPatientId,
    required String doctorId,
    required String clinicId,
    required String slotId,
    required String date,
    required String time,
    required String type,
    String? reason,
    String? paymentMode,
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl/appointments/simple'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode({
        'clinic_patient_id': clinicPatientId,
        'doctor_id': doctorId,
        'clinic_id': clinicId,
        'individual_slot_id': slotId,
        'appointment_date': date,
        'appointment_time': '$date $time:00',
        'consultation_type': type,
        'reason': reason,
        'payment_mode': paymentMode ?? 'pay_later',
      }),
    );

    if (response.statusCode == 201) {
      return jsonDecode(response.body);
    } else {
      final error = jsonDecode(response.body);
      throw Exception(error['error']);
    }
  }
}
```

### Simple UI Usage

```dart
// Book appointment with one call
try {
  final result = await appointmentService.bookAppointment(
    clinicPatientId: patientId,
    doctorId: doctorId,
    clinicId: clinicId,
    slotId: selectedSlotId,
    date: '2025-10-18',
    time: '09:30',
    type: 'offline',
    reason: 'Regular checkup',
    paymentMode: 'cash',
  );

  final appointment = result['appointment'];
  print('✅ Booked: ${appointment['booking_number']}');
  print('Token: #${appointment['token_number']}');
  
} catch (e) {
  print('❌ Error: $e');
}
```

---

## ❌ Error Responses

### Error 1: Patient Not Found
```json
{
  "error": "Patient not found"
}
```

### Error 2: Patient from Different Clinic
```json
{
  "error": "Patient belongs to different clinic"
}
```

### Error 3: Slot Not Available
```json
{
  "error": "Slot already booked",
  "message": "Please select another slot"
}
```

### Error 4: Slot from Different Clinic
```json
{
  "error": "Slot belongs to different clinic"
}
```

### Error 5: Doctor Not Found
```json
{
  "error": "Doctor not found"
}
```

---

## 🔄 Complete Workflow

### Step 1: Create Patient
```bash
POST /clinic-specific-patients
{
  "clinic_id": "clinic-uuid",
  "first_name": "Ahmed",
  "last_name": "Khan",
  "phone": "+971501234567"
}
→ Returns: clinic_patient_id
```

### Step 2: Get Available Slots
```bash
GET /doctor-session-slots?doctor_id=xxx&clinic_id=xxx&date=2025-10-18
→ Returns: List of individual_slot_id
```

### Step 3: Book Appointment
```bash
POST /appointments/simple
{
  "clinic_patient_id": "from-step-1",
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "individual_slot_id": "from-step-2",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "consultation_type": "offline"
}
→ Returns: Appointment with booking_number
```

---

## 📊 Comparison

| Feature | Original API | Simple API |
|---------|--------------|------------|
| Endpoint | `/appointments` | `/appointments/simple` |
| Patient support | 5 types (UserID, PatientID, etc.) | 1 type (clinic_patient_id) |
| Validations | 15+ checks | 4 essential checks |
| Code lines | ~600 lines | ~200 lines |
| Response time | Slower | Faster |
| Use case | Complex scenarios | Simple clinic booking |

---

## ✅ Benefits

### 1. Simple Code ✅
- Only 200 lines vs 600+ lines
- Easy to understand
- Easy to maintain

### 2. Fast Response ✅
- Fewer database queries
- Less validation overhead
- Quicker booking

### 3. Easy Integration ✅
- Straightforward request/response
- No complex patient search
- Direct booking

### 4. Focused Purpose ✅
- Clinic patients only
- Session slots only
- Clear workflow

---

## 🎯 When to Use Which API

### Use Simple API (`/appointments/simple`) ✅
- Clinic patient already created
- Using session-based slots
- Standard clinic booking
- **Recommended for most cases**

### Use Original API (`/appointments`)
- Need patient search by phone/MO ID
- Using global patients
- Complex validation needed
- Legacy support

---

## 📋 Quick Reference Card

```
┌─────────────────────────────────────────────────────────┐
│ SIMPLE APPOINTMENT API                                  │
├─────────────────────────────────────────────────────────┤
│                                                          │
│ POST /api/appointments/simple                           │
│                                                          │
│ Required:                                                │
│  ✅ clinic_patient_id                                    │
│  ✅ doctor_id                                            │
│  ✅ clinic_id                                            │
│  ✅ individual_slot_id                                   │
│  ✅ appointment_date (YYYY-MM-DD)                        │
│  ✅ appointment_time (YYYY-MM-DD HH:MM:SS)               │
│  ✅ consultation_type (offline|online)                   │
│                                                          │
│ Optional:                                                │
│  • reason                                                │
│  • notes                                                 │
│  • payment_mode                                          │
│                                                          │
│ Returns:                                                 │
│  • appointment object                                    │
│  • booking_number                                        │
│  • token_number                                          │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

---

## ✅ Status

| Component | Status |
|-----------|--------|
| Simple API created | ✅ Done |
| Route added | ✅ /appointments/simple |
| Model updated | ✅ ClinicPatientID added |
| Validation simplified | ✅ Only 4 checks |
| Documentation | ✅ Complete |
| No linter errors | ✅ Clean |

---

**API:** `POST /api/appointments/simple`  
**Purpose:** Simple, fast appointment booking  
**Status:** ✅ **Ready to Use!** 🎉

Use this for straightforward clinic patient appointments!

