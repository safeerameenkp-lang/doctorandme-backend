# Appointment History - Quick Reference 🚀

## 📊 **Response Structure**

```json
{
  "patient_id": "...",
  "appointments": [
    {
      "appointment_id": "...",
      "doctor_id": "...",
      "doctor_name": "Dr. Smith",
      "department": "Cardiology",
      "appointment_date": "2025-10-18",
      "days_since": 2,
      "remaining_days": 3,
      "status": "active",
      "follow_up_eligible": true,
      "free_follow_up_used": false
    }
  ],
  "eligible_follow_up": {
    "doctor_id": "...",
    "doctor_name": "Dr. Smith",
    "department": "Cardiology",
    "remaining_days": 3
  }
}
```

---

## 🎨 **Status Colors**

```dart
Color getStatusColor(status, freeUsed) {
  if (status == 'active' && !freeUsed) return Colors.green;  // ✅ Free
  if (status == 'active' && freeUsed) return Colors.orange;  // ⚠️ Used
  if (status == 'expired') return Colors.grey;               // 🕒 Expired
  if (status == 'future') return Colors.blue;                // ⏳ Future
}
```

---

## 🔍 **Quick Filters**

```dart
// Get all FREE follow-ups available
final free = appointments.where((a) => 
  a.status == 'active' && !a.freeFollowUpUsed
).toList();

// Get appointments by doctor
final drSmith = appointments.where((a) => 
  a.doctorId == 'doctor-id'
).toList();

// Get expired appointments
final expired = appointments.where((a) => 
  a.status == 'expired'
).toList();
```

---

## ✅ **Button Logic**

```dart
if (apt.status == 'active' && !apt.freeFollowUpUsed) {
  showButton('Book FREE Follow-Up', Colors.green);
} else if (apt.followUpEligible) {
  showButton('Book Follow-Up (₹200)', Colors.orange);
} else {
  showButton('New Appointment', Colors.blue);
}
```

---

## 📋 **Summary Widget**

```dart
Text('✅ ${freeCount} FREE follow-ups available');
Text('⚠️ ${usedCount} free follow-ups used');
Text('🕒 ${expiredCount} appointments expired');
```

---

## 🚀 **API Endpoints**

```
GET /api/clinic-specific-patients?clinic_id=xxx&search=...
GET /api/clinic-specific-patients/:id
```

Both return full appointment history!

---

**Quick Tip:** Use `eligible_follow_up` to show the **best** option to user! ✅

