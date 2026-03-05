# Multiple Eligible Follow-Ups - Complete Guide ✅

## 🎯 **Core Rule**

**Each patient can have multiple FREE follow-ups - one per doctor+department combination!**

---

## ✅ **Key Concept**

Follow-up eligibility is tracked **per doctor AND department**, NOT globally per patient.

**Example:**
- Patient visits Dr. Smith (Cardiology) → Gets 1 free follow-up with Dr. Smith in Cardiology ✅
- Patient visits Dr. Lee (Cardiology) → Gets ANOTHER free follow-up with Dr. Lee in Cardiology ✅
- Patient visits Dr. Patel (Neurology) → Gets ANOTHER free follow-up with Dr. Patel in Neurology ✅

**Total:** 3 separate free follow-ups! Each is independent!

---

## 📊 **Response Structure**

```json
{
  "patient_id": "p12345",
  "patient_name": "John Doe",
  
  "appointments": [
    {
      "appointment_id": "a001",
      "doctor_id": "doctor-a",
      "doctor_name": "Dr. Smith",
      "department": "Cardiology",
      "appointment_date": "2025-10-17",
      "days_since": 1,
      "validity_days": 5,
      "remaining_days": 4,
      "status": "active",
      "follow_up_eligible": true,
      "free_follow_up_used": false,
      "note": "Eligible for free follow-up with Dr. Smith (Cardiology)"
    },
    {
      "appointment_id": "a002",
      "doctor_id": "doctor-b",
      "doctor_name": "Dr. Lee",
      "department": "Cardiology",
      "appointment_date": "2025-09-30",
      "days_since": 18,
      "validity_days": 5,
      "status": "expired",
      "follow_up_eligible": true,
      "free_follow_up_used": false,
      "note": "Follow-up period expired for Dr. Lee (Cardiology). Follow-up requires payment."
    },
    {
      "appointment_id": "a003",
      "doctor_id": "doctor-c",
      "doctor_name": "Dr. Patel",
      "department": "Neurology",
      "appointment_date": "2025-10-15",
      "days_since": 3,
      "validity_days": 5,
      "remaining_days": 2,
      "status": "active",
      "follow_up_eligible": true,
      "free_follow_up_used": false,
      "note": "Eligible for free follow-up with Dr. Patel (Neurology)"
    }
  ],
  
  "eligible_follow_ups": [
    {
      "appointment_id": "a001",
      "doctor_id": "doctor-a",
      "doctor_name": "Dr. Smith",
      "department": "Cardiology",
      "appointment_date": "2025-10-17",
      "remaining_days": 4,
      "note": "Eligible for free follow-up with Dr. Smith (Cardiology)"
    },
    {
      "appointment_id": "a003",
      "doctor_id": "doctor-c",
      "doctor_name": "Dr. Patel",
      "department": "Neurology",
      "appointment_date": "2025-10-15",
      "remaining_days": 2,
      "note": "Eligible for free follow-up with Dr. Patel (Neurology)"
    }
  ],
  
  "total_appointments": 3
}
```

---

## 🔑 **Field Explanations**

### appointments[] - Full History

| Field | Description |
|-------|-------------|
| `appointment_id` | Unique appointment ID |
| `doctor_id` | Doctor UUID |
| `doctor_name` | Doctor's full name |
| `department` | Department name |
| `appointment_date` | Date (YYYY-MM-DD) |
| `days_since` | Days since appointment (0=today, <0=future) |
| `validity_days` | Always 5 |
| `remaining_days` | Days left for free follow-up (null if expired/future) |
| `status` | `active`, `expired`, or `future` |
| `follow_up_eligible` | Can book follow-up? |
| `free_follow_up_used` | Free follow-up already used? |
| `note` | Human-readable explanation |

### eligible_follow_ups[] - Quick List

Array of **ALL** eligible FREE follow-ups. Each entry represents a different doctor+department combination where free follow-up is available.

---

## 🧪 **Example Scenarios**

### Scenario A: Patient with 3 Different Doctors ✅

**Patient Appointments:**
```
Oct 17: Dr. Smith → Cardiology (1 day ago)
Oct 16: Dr. Lee → Cardiology (2 days ago)
Oct 15: Dr. Patel → Neurology (3 days ago)
```

**Result:**
```json
{
  "eligible_follow_ups": [
    {
      "doctor_name": "Dr. Smith",
      "department": "Cardiology",
      "remaining_days": 4
    },
    {
      "doctor_name": "Dr. Lee",
      "department": "Cardiology",
      "remaining_days": 3
    },
    {
      "doctor_name": "Dr. Patel",
      "department": "Neurology",
      "remaining_days": 2
    }
  ]
}
```

**✅ Patient has 3 FREE follow-ups available!**

---

### Scenario B: Same Doctor, Different Departments ✅

**Patient Appointments:**
```
Oct 17: Dr. Smith → Cardiology (1 day ago)
Oct 16: Dr. Smith → Neurology (2 days ago)
```

**Result:**
```json
{
  "eligible_follow_ups": [
    {
      "doctor_name": "Dr. Smith",
      "department": "Cardiology",
      "remaining_days": 4
    },
    {
      "doctor_name": "Dr. Smith",
      "department": "Neurology",
      "remaining_days": 3
    }
  ]
}
```

**✅ Patient has 2 FREE follow-ups with same doctor but different departments!**

---

### Scenario C: Used Free Follow-Up ⚠️

**Patient Appointments:**
```
Oct 15: Dr. Smith → Cardiology (3 days ago)
Oct 16: Dr. Smith → Cardiology, Follow-up (FREE, used)
Oct 17: Dr. Lee → Cardiology (1 day ago)
```

**Result:**
```json
{
  "appointments": [
    {
      "doctor_name": "Dr. Smith",
      "department": "Cardiology",
      "free_follow_up_used": true,
      "note": "Free follow-up already used..."
    },
    {
      "doctor_name": "Dr. Lee",
      "department": "Cardiology",
      "free_follow_up_used": false,
      "note": "Eligible for free follow-up..."
    }
  ],
  "eligible_follow_ups": [
    {
      "doctor_name": "Dr. Lee",
      "department": "Cardiology",
      "remaining_days": 4
    }
  ]
}
```

**✅ Still has 1 FREE follow-up with Dr. Lee (different doctor)!**

---

## 🎨 **UI Integration**

### 1. Show All Eligible Follow-Ups

```dart
Widget buildEligibleFollowUpsList(Patient patient) {
  if (patient.eligibleFollowUps.isEmpty) {
    return Text('No free follow-ups available');
  }
  
  return Column(
    children: [
      Text('✅ ${patient.eligibleFollowUps.length} FREE Follow-Ups Available:'),
      ...patient.eligibleFollowUps.map((followUp) {
        return Card(
          color: Colors.green[50],
          child: ListTile(
            leading: Icon(Icons.check_circle, color: Colors.green),
            title: Text('${followUp.doctorName} - ${followUp.department}'),
            subtitle: Text('${followUp.remainingDays} days remaining'),
            trailing: ElevatedButton(
              onPressed: () => bookFollowUp(followUp),
              child: Text('Book FREE'),
            ),
          ),
        );
      }).toList(),
    ],
  );
}
```

---

### 2. Doctor Selection Dropdown

```dart
List<DropdownMenuItem<EligibleFollowUp>> buildDoctorOptions() {
  return patient.eligibleFollowUps.map((followUp) {
    return DropdownMenuItem(
      value: followUp,
      child: Row(
        children: [
          Icon(Icons.local_hospital, color: Colors.green),
          SizedBox(width: 8),
          Text('${followUp.doctorName} - ${followUp.department}'),
          Spacer(),
          Chip(
            label: Text('FREE'),
            backgroundColor: Colors.green,
          ),
        ],
      ),
    );
  }).toList();
}
```

---

### 3. Dashboard Summary

```dart
Widget buildFollowUpSummary(Patient patient) {
  final freeCount = patient.eligibleFollowUps.length;
  final expiredCount = patient.appointments
      .where((a) => a.status == 'expired')
      .length;
  final usedCount = patient.appointments
      .where((a) => a.freeFollowUpUsed == true)
      .length;
  
  return Card(
    child: Column(
      children: [
        Text('📊 Follow-Up Status', style: TextStyle(fontSize: 18)),
        SizedBox(height: 8),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceAround,
          children: [
            _buildStat('✅ FREE', freeCount, Colors.green),
            _buildStat('⚠️ USED', usedCount, Colors.orange),
            _buildStat('🕒 EXPIRED', expiredCount, Colors.grey),
          ],
        ),
      ],
    ),
  );
}
```

---

### 4. Appointment Cards with Color Coding

```dart
Widget buildAppointmentCard(Appointment apt) {
  Color cardColor;
  Icon statusIcon;
  
  if (apt.status == 'active' && apt.freeFollowUpUsed == false) {
    cardColor = Colors.green[50]!;
    statusIcon = Icon(Icons.check_circle, color: Colors.green);
  } else if (apt.status == 'active' && apt.freeFollowUpUsed == true) {
    cardColor = Colors.orange[50]!;
    statusIcon = Icon(Icons.warning, color: Colors.orange);
  } else if (apt.status == 'expired') {
    cardColor = Colors.grey[50]!;
    statusIcon = Icon(Icons.schedule, color: Colors.grey);
  } else {
    cardColor = Colors.blue[50]!;
    statusIcon = Icon(Icons.event, color: Colors.blue);
  }
  
  return Card(
    color: cardColor,
    child: ListTile(
      leading: statusIcon,
      title: Text('${apt.doctorName} - ${apt.department}'),
      subtitle: Text(apt.note),
      trailing: apt.followUpEligible && !apt.freeFollowUpUsed && apt.status == 'active'
          ? ElevatedButton(
              onPressed: () => bookFollowUp(apt),
              style: ElevatedButton.styleFrom(backgroundColor: Colors.green),
              child: Text('FREE'),
            )
          : null,
    ),
  );
}
```

---

## 🧪 **Testing Checklist**

### ✅ Test 1: Multiple Doctors, Same Department
- Book with Dr. A (Cardiology)
- Book with Dr. B (Cardiology)
- **Expected:** 2 eligible follow-ups

### ✅ Test 2: Same Doctor, Multiple Departments
- Book with Dr. A (Cardiology)
- Book with Dr. A (Neurology)
- **Expected:** 2 eligible follow-ups

### ✅ Test 3: Use Free Follow-Up
- Book with Dr. A (Cardiology)
- Book follow-up with Dr. A (Cardiology) - FREE
- **Expected:** 0 eligible follow-ups for Dr. A + Cardiology

### ✅ Test 4: Independent Tracking
- Book with Dr. A (Cardiology)
- Use free follow-up with Dr. A (Cardiology)
- Book with Dr. B (Cardiology)
- **Expected:** Still has 1 eligible follow-up with Dr. B

---

## 📊 **Summary**

| Aspect | Value |
|--------|-------|
| **Tracking** | Per Doctor + Department |
| **Multiple Follow-Ups** | ✅ Yes (one per doctor+dept) |
| **Response Field** | `eligible_follow_ups[]` (array) |
| **Note Field** | Human-readable explanation |
| **UI Benefits** | Easy dropdown, color-coded cards |

---

## ✅ **Benefits**

1. **Fair System:** Each doctor+department gets its own free follow-up
2. **Clear Display:** Array makes it easy to show all options
3. **Smart UI:** Can build dropdown from `eligible_follow_ups`
4. **User-Friendly:** Notes explain each appointment's status
5. **Scalable:** Works with any number of doctors/departments

---

**Result:** Patient can have multiple FREE follow-ups, each tracked independently per doctor+department! 🎉✅

