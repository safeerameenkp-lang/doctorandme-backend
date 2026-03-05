# Follow-Up Frontend Integration - Quick Guide 🚀

## 🎯 **What Changed**

Follow-up eligibility is now **context-aware** based on the **selected doctor and department**.

---

## ✅ **How to Use**

### Step 1: User Selects Doctor & Department

```dart
class AppointmentBookingScreen extends StatefulWidget {
  String? selectedDoctorId;
  String? selectedDepartmentId;
  
  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        // Doctor dropdown
        DropdownButton<String>(
          value: selectedDoctorId,
          items: doctors.map((doctor) => 
            DropdownMenuItem(
              value: doctor.id,
              child: Text(doctor.name),
            )
          ).toList(),
          onChanged: (value) {
            setState(() {
              selectedDoctorId = value;
            });
          },
        ),
        
        // Department dropdown
        DropdownButton<String>(
          value: selectedDepartmentId,
          items: departments.map((dept) => 
            DropdownMenuItem(
              value: dept.id,
              child: Text(dept.name),
            )
          ).toList(),
          onChanged: (value) {
            setState(() {
              selectedDepartmentId = value;
            });
          },
        ),
      ],
    );
  }
}
```

---

### Step 2: Search Patients with Doctor+Department Context

```dart
Future<List<Patient>> searchPatients(String query) async {
  // ✅ IMPORTANT: Pass selected doctor_id and department_id!
  final url = Uri.parse(
    '$baseUrl/clinic-specific-patients'
    '?clinic_id=$clinicId'
    '&doctor_id=$selectedDoctorId'           // ✅ NEW
    '&department_id=$selectedDepartmentId'   // ✅ NEW
    '&search=$query'
  );
  
  final response = await http.get(url, headers: {
    'Authorization': 'Bearer $token',
  });
  
  if (response.statusCode == 200) {
    final data = json.decode(response.body);
    return (data['patients'] as List)
      .map((p) => Patient.fromJson(p))
      .toList();
  }
  
  throw Exception('Failed to load patients');
}
```

---

### Step 3: Display Follow-Up Eligibility

```dart
Widget buildPatientCard(Patient patient) {
  final eligibility = patient.followUpEligibility;
  
  return Card(
    child: ListTile(
      title: Text('${patient.firstName} ${patient.lastName}'),
      subtitle: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Phone: ${patient.phone}'),
          
          // ✅ Show last appointment info
          if (patient.lastAppointment != null) ...[
            SizedBox(height: 8),
            Text(
              'Last Visit: ${patient.lastAppointment!.date} '
              '(${patient.lastAppointment!.daysSince} days ago)',
              style: TextStyle(fontSize: 12, color: Colors.grey[600]),
            ),
            Text(
              'Doctor: ${patient.lastAppointment!.doctorName}',
              style: TextStyle(fontSize: 12, color: Colors.grey[600]),
            ),
            Text(
              'Department: ${patient.lastAppointment!.department}',
              style: TextStyle(fontSize: 12, color: Colors.grey[600]),
            ),
          ],
          
          // ✅ Show eligibility status
          if (eligibility != null) ...[
            SizedBox(height: 8),
            Container(
              padding: EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: eligibility.isFree 
                  ? Colors.green[50] 
                  : Colors.orange[50],
                borderRadius: BorderRadius.circular(4),
              ),
              child: Row(
                children: [
                  Icon(
                    eligibility.isFree 
                      ? Icons.check_circle 
                      : Icons.info,
                    size: 16,
                    color: eligibility.isFree 
                      ? Colors.green 
                      : Colors.orange,
                  ),
                  SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      eligibility.message ?? '',
                      style: TextStyle(fontSize: 12),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
      trailing: buildActionButton(patient),
    ),
  );
}
```

---

### Step 4: Action Button Logic

```dart
Widget buildActionButton(Patient patient) {
  final eligibility = patient.followUpEligibility;
  
  // ✅ Case 1: Eligible for FREE follow-up
  if (eligibility?.eligible == true && eligibility?.isFree == true) {
    return ElevatedButton(
      onPressed: () => bookFollowUp(patient.id, isFree: true),
      style: ElevatedButton.styleFrom(
        backgroundColor: Colors.green,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text('Follow-Up', style: TextStyle(fontSize: 12)),
          Text('FREE', style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold)),
        ],
      ),
    );
  }
  
  // ✅ Case 2: Eligible for PAID follow-up
  if (eligibility?.eligible == true && eligibility?.isFree == false) {
    return ElevatedButton(
      onPressed: () => bookFollowUp(patient.id, isFree: false),
      style: ElevatedButton.styleFrom(
        backgroundColor: Colors.orange,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text('Follow-Up', style: TextStyle(fontSize: 12)),
          Text('₹200', style: TextStyle(fontSize: 10)),
        ],
      ),
    );
  }
  
  // ✅ Case 3: Not eligible (new appointment)
  return ElevatedButton(
    onPressed: () => bookNewAppointment(patient.id),
    style: ElevatedButton.styleFrom(
      backgroundColor: Colors.blue,
    ),
    child: Text('New Appointment', style: TextStyle(fontSize: 12)),
  );
}
```

---

### Step 5: Book Follow-Up

```dart
Future<void> bookFollowUp(String patientId, {required bool isFree}) async {
  final body = {
    'clinic_patient_id': patientId,
    'doctor_id': selectedDoctorId,          // ✅ Selected doctor
    'clinic_id': clinicId,
    'department_id': selectedDepartmentId,  // ✅ Selected department
    'individual_slot_id': selectedSlotId,
    'appointment_date': selectedDate,
    'appointment_time': selectedTime,
    'consultation_type': 'follow-up-via-clinic',  // ✅ Follow-up type
  };
  
  // ✅ Add payment info if NOT free
  if (!isFree) {
    body['payment_method'] = 'pay_now';
    body['payment_type'] = 'cash';  // or card, upi
  }
  
  final response = await http.post(
    Uri.parse('$baseUrl/appointments/simple'),
    headers: {
      'Authorization': 'Bearer $token',
      'Content-Type': 'application/json',
    },
    body: json.encode(body),
  );
  
  if (response.statusCode == 201) {
    // ✅ Success!
    showSuccessDialog('Appointment booked successfully!');
  } else {
    // ❌ Error
    final error = json.decode(response.body);
    showErrorDialog(error['message'] ?? 'Failed to book appointment');
  }
}
```

---

## 📊 **API Response Structure**

```json
{
  "clinic_id": "xxx",
  "total": 1,
  "patients": [
    {
      "id": "patient-uuid",
      "first_name": "John",
      "last_name": "Doe",
      "phone": "1234567890",
      
      "last_appointment": {
        "id": "appointment-uuid",
        "doctor_id": "doctor-uuid",
        "doctor_name": "Dr. ABC",
        "department_id": "dept-uuid",
        "department": "Cardiology",
        "date": "2025-10-18",
        "days_since": 2,
        "status": "completed"
      },
      
      "follow_up_eligibility": {
        "eligible": true,
        "is_free": true,
        "days_remaining": 3,
        "message": "You have one FREE follow-up available with this doctor in this department",
        "reason": null
      },
      
      "total_appointments": 5
    }
  ]
}
```

---

## 🧪 **Testing Checklist**

### ✅ Test 1: Same Doctor, Same Department (Within 5 Days)
- **Select:** Doctor A, Cardiology
- **Search:** Patient with last appointment: Doctor A → Cardiology (2 days ago)
- **Expected:** ✅ "FREE Follow-Up" button

### ✅ Test 2: Different Doctor
- **Select:** Doctor B, Cardiology
- **Search:** Patient with last appointment: Doctor A → Cardiology (2 days ago)
- **Expected:** ❌ "New Appointment" button only

### ✅ Test 3: Different Department
- **Select:** Doctor A, Neurology
- **Search:** Patient with last appointment: Doctor A → Cardiology (2 days ago)
- **Expected:** ❌ "New Appointment" button only

### ✅ Test 4: Free Follow-Up Already Used
- **Select:** Doctor A, Cardiology
- **Search:** Patient with 2 appointments: Doctor A → Cardiology (1 day ago, FREE follow-up used)
- **Expected:** ⚠️ "Follow-Up (₹200)" button (paid)

### ✅ Test 5: After 5 Days
- **Select:** Doctor A, Cardiology
- **Search:** Patient with last appointment: Doctor A → Cardiology (6 days ago)
- **Expected:** ⚠️ "Follow-Up (₹200)" button OR "New Appointment"

---

## 💡 **Pro Tips**

### 1. Always Pass Context
```dart
// ❌ BAD: No context
GET /clinic-specific-patients?clinic_id=xxx&search=John

// ✅ GOOD: With context
GET /clinic-specific-patients?clinic_id=xxx&doctor_id=yyy&department_id=zzz&search=John
```

### 2. Show Clear Messages
```dart
if (eligibility.isFree) {
  showBanner('🎉 This patient is eligible for a FREE follow-up!');
} else if (eligibility.eligible) {
  showBanner('ℹ️ Follow-up available (₹200)');
} else {
  showBanner('New appointment required');
}
```

### 3. Disable Follow-Up if Not Eligible
```dart
ElevatedButton(
  onPressed: eligibility?.eligible == true 
    ? () => bookFollowUp() 
    : null,  // ✅ Disable button
  child: Text('Book Follow-Up'),
)
```

---

## 🚨 **Common Mistakes**

### ❌ Mistake 1: Not Passing doctor_id/department_id
```dart
// This will show patient's LAST appointment (any doctor/department)
// NOT the selected doctor+department!
GET /clinic-specific-patients?clinic_id=xxx&search=John
```

### ❌ Mistake 2: Ignoring Eligibility
```dart
// User clicks "Follow-Up" → Backend error: "Doctor mismatch"
// Always check eligibility BEFORE showing the button!
```

### ❌ Mistake 3: Wrong consultation_type
```dart
// ❌ Wrong
'consultation_type': 'offline'

// ✅ Correct
'consultation_type': 'follow-up-via-clinic'
```

---

## ✅ **Summary**

| Step | Action |
|------|--------|
| 1 | User selects doctor + department |
| 2 | Pass `doctor_id` and `department_id` to patient search API |
| 3 | Check `follow_up_eligibility` in response |
| 4 | Show appropriate button (FREE/PAID/NEW) |
| 5 | Book with `consultation_type: 'follow-up-via-clinic'` |

---

**Result:** Users can see **accurate follow-up eligibility** before booking! 🎉✅

