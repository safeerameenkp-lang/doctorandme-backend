# Follow-Up Appointments - Quick Reference 🔄

## 🎯 Key Rules

1. **Previous Patient Only** - Must have completed/confirmed appointment
2. **5-Day Window** - Must book within 5 days of last appointment
3. **Same Doctor** - Must be with the same doctor
4. **Same Department** - Must be in the same department
5. **No Payment** - Follow-ups are FREE

---

## 🌐 APIs

### Check Eligibility
```
GET /api/appointments/check-follow-up-eligibility?
  clinic_patient_id={uuid}&
  clinic_id={uuid}
```

### Book Follow-Up
```
POST /api/appointments/simple
{
  "is_follow_up": true,  // ✅ Key flag
  "doctor_id": "same-doctor-uuid",
  "department_id": "same-dept-uuid",
  // NO payment_method required
}
```

---

## 📊 Response Examples

### Eligible
```json
{
  "eligible": true,
  "days_remaining": 3,
  "last_appointment": {
    "doctor_id": "uuid",
    "doctor_name": "Dr. Ahmed",
    "department": "Cardiology"
  }
}
```

### Not Eligible
```json
{
  "eligible": false,
  "reason": "Follow-up period expired",
  "days_since_last": 7
}
```

---

## 💻 Flutter Quick Code

```dart
// Check eligibility
final response = await http.get(
  Uri.parse('$baseUrl/appointments/check-follow-up-eligibility?'
    'clinic_patient_id=$patientId&clinic_id=$clinicId')
);
final data = jsonDecode(response.body);
bool isEligible = data['eligible'];

// Book follow-up
final body = {
  'is_follow_up': true,  // ✅ Key
  'doctor_id': data['last_appointment']['doctor_id'],
  // NO payment fields
};
```

---

## ✅ Validation Flow

```
Request → Check Previous Appointment
       → Check 5-Day Window
       → Check Doctor Match
       → Check Department Match
       → Check Slot Availability
       → Create (Payment Waived, Fee = 0)
```

---

## 🧪 Quick Test

```bash
# Check eligibility
GET /check-follow-up-eligibility?clinic_patient_id=xxx&clinic_id=xxx

# Book follow-up
POST /appointments/simple
{
  "is_follow_up": true,
  "doctor_id": "same-doctor",
  "department_id": "same-dept",
  "individual_slot_id": "slot-uuid",
  ...
}
# ✅ No payment fields needed
```

---

**Status:** ✅ Complete!

