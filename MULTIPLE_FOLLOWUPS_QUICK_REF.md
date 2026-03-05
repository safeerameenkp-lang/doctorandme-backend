# Multiple Follow-Ups - Quick Reference 🚀

## ✅ **Core Rule**

**One FREE follow-up per (Doctor + Department) combination**

---

## 📊 **Response Format**

```json
{
  "patient_id": "p12345",
  "patient_name": "John Doe",
  
  "appointments": [
    {
      "appointment_id": "a001",
      "doctor_name": "Dr. Smith",
      "department": "Cardiology",
      "appointment_date": "2025-10-17",
      "days_since": 1,
      "remaining_days": 4,
      "status": "active",
      "follow_up_eligible": true,
      "note": "Eligible for free follow-up with Dr. Smith (Cardiology)"
    },
    {
      "appointment_id": "a002",
      "doctor_name": "Dr. Lee",
      "department": "Cardiology",
      "appointment_date": "2025-09-30",
      "days_since": 18,
      "status": "expired",
      "note": "Follow-up period expired for Dr. Lee"
    },
    {
      "appointment_id": "a003",
      "doctor_name": "Dr. Patel",
      "department": "Neurology",
      "appointment_date": "2025-10-15",
      "days_since": 3,
      "remaining_days": 2,
      "status": "active",
      "follow_up_eligible": true,
      "note": "Eligible for free follow-up with Dr. Patel (Neurology)"
    }
  ],
  
  "eligible_follow_ups": [
    {
      "doctor_id": "doctor-a",
      "doctor_name": "Dr. Smith",
      "department": "Cardiology",
      "remaining_days": 4
    },
    {
      "doctor_id": "doctor-c",
      "doctor_name": "Dr. Patel",
      "department": "Neurology",
      "remaining_days": 2
    }
  ]
}
```

---

## 🎯 **Key Points**

1. **`appointments[]`** - Shows ALL appointments with their status
2. **`eligible_follow_ups[]`** - Quick list of FREE follow-ups available
3. **Per Doctor+Dept** - Each combination tracked separately
4. **Multiple FREE** - Can have multiple at once!

---

## 💡 **Examples**

### Same Department, Different Doctors ✅
```
Dr. Smith (Cardiology) → FREE follow-up ✅
Dr. Lee (Cardiology) → FREE follow-up ✅
```

### Same Doctor, Different Departments ✅
```
Dr. Smith (Cardiology) → FREE follow-up ✅
Dr. Smith (Neurology) → FREE follow-up ✅
```

### Used Follow-Up ⚠️
```
Dr. Smith (Cardiology) → FREE used ❌
Dr. Lee (Cardiology) → FREE available ✅
```

---

## 🎨 **UI Integration**

```dart
// Show all eligible follow-ups
for (var followUp in patient.eligibleFollowUps) {
  showCard(
    '${followUp.doctorName} - ${followUp.department}',
    'FREE (${followUp.remainingDays} days left)',
    Colors.green,
  );
}

// Count by status
final freeCount = patient.eligibleFollowUps.length;
final activeCount = patient.appointments
    .where((a) => a.status == 'active')
    .length;
```

---

## ✅ **Summary**

- ✅ Multiple FREE follow-ups (one per doctor+dept)
- ✅ `eligible_follow_ups[]` array for easy display
- ✅ `note` field for human-readable explanation
- ✅ Tracked independently per combination

**Perfect for dropdown selection!** 🎉

