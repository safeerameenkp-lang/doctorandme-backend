# Frontend Consultation Type Fix 🔧

## ❌ Current Error

```
Request failed with status 400: {
  "details": "Key: 'SimpleAppointmentInput.ConsultationType' Error:Field validation for 'ConsultationType' failed on the 'oneof' tag",
  "error": "Invalid input"
}
```

**Reason:** Frontend is sending old values that backend no longer accepts!

---

## 🔍 Root Cause

### Backend Changed (Lines 27):
```go
ConsultationType string `json:"consultation_type" binding:"required,oneof=clinic_visit video_consultation in_person follow_up"`
```

### Frontend Still Sending:
```dart
"consultation_type": "offline"  // ❌ NOT VALID ANYMORE
// or
"consultation_type": "online"   // ❌ NOT VALID ANYMORE
```

---

## ✅ Solution: Update Frontend

### Step 1: Find Your Consultation Type Constants

Look for where you define consultation types in your Flutter code. It might be in:
- `constants.dart`
- `appointment_model.dart`
- `appointment_booking_page.dart`

---

### Step 2: Update Values

**❌ OLD (Remove these):**
```dart
static const String offline = 'offline';
static const String online = 'online';
```

**✅ NEW (Use these):**
```dart
static const String clinicVisit = 'clinic_visit';
static const String videoConsultation = 'video_consultation';
```

---

### Step 3: Update Dropdown/Selection

**❌ OLD:**
```dart
final consultationTypes = [
  {'value': 'offline', 'label': 'Offline Visit'},
  {'value': 'online', 'label': 'Online Consultation'},
  {'value': 'in_person', 'label': 'In Person'},
  {'value': 'follow_up', 'label': 'Follow Up'},
];
```

**✅ NEW:**
```dart
final consultationTypes = [
  {'value': 'clinic_visit', 'label': '🏥 Clinic Visit'},
  {'value': 'video_consultation', 'label': '💻 Video Consultation'},
  {'value': 'in_person', 'label': '👤 In Person'},
  {'value': 'follow_up', 'label': '🔄 Follow Up'},
];
```

---

### Step 4: Update Model Class

```dart
class AppointmentRequest {
  final String clinicPatientId;
  final String doctorId;
  final String clinicId;
  final String? departmentId;
  final String individualSlotId;
  final String appointmentDate;
  final String appointmentTime;
  final String consultationType;  // ✅ Update this field
  final bool isFollowUp;
  final String? reason;
  final String? notes;
  final String? paymentMethod;
  final String? paymentType;

  AppointmentRequest({
    required this.clinicPatientId,
    required this.doctorId,
    required this.clinicId,
    this.departmentId,
    required this.individualSlotId,
    required this.appointmentDate,
    required this.appointmentTime,
    required this.consultationType,
    this.isFollowUp = false,
    this.reason,
    this.notes,
    this.paymentMethod,
    this.paymentType,
  });

  Map<String, dynamic> toJson() {
    return {
      'clinic_patient_id': clinicPatientId,
      'doctor_id': doctorId,
      'clinic_id': clinicId,
      'department_id': departmentId,
      'individual_slot_id': individualSlotId,
      'appointment_date': appointmentDate,
      'appointment_time': appointmentTime,
      'consultation_type': consultationType,  // ✅ Send new value
      'is_follow_up': isFollowUp,
      'reason': reason,
      'notes': notes,
      'payment_method': paymentMethod,
      'payment_type': paymentType,
    };
  }
}
```

---

### Step 5: Update Booking Logic

**❌ OLD:**
```dart
String consultationType = 'offline'; // Default

if (selectedType == 'online') {
  consultationType = 'online';
} else {
  consultationType = 'offline';
}
```

**✅ NEW:**
```dart
String consultationType = 'clinic_visit'; // Default

if (selectedType == 'video') {
  consultationType = 'video_consultation';
} else {
  consultationType = 'clinic_visit';
}
```

---

## 📋 Complete Valid Values

| Backend Expects | Frontend Should Send | UI Label |
|----------------|---------------------|----------|
| `clinic_visit` | `'clinic_visit'` | 🏥 Clinic Visit |
| `video_consultation` | `'video_consultation'` | 💻 Video Consultation |
| `in_person` | `'in_person'` | 👤 In Person |
| `follow_up` | `'follow_up'` | 🔄 Follow-Up |

---

## 🔍 Debugging: Print Request Body

Add this before API call to see what you're sending:

```dart
Future<void> createAppointment() async {
  final body = {
    'clinic_patient_id': patientId,
    'doctor_id': doctorId,
    'clinic_id': clinicId,
    'individual_slot_id': slotId,
    'appointment_date': date,
    'appointment_time': time,
    'consultation_type': consultationType,  // ✅ Check this value!
    'is_follow_up': isFollowUp,
    'payment_method': paymentMethod,
    'payment_type': paymentType,
  };
  
  // ✅ PRINT TO SEE WHAT'S BEING SENT
  print('📤 Request Body: ${jsonEncode(body)}');
  print('📍 consultation_type = $consultationType');
  
  final response = await http.post(
    Uri.parse('$baseUrl/appointments/simple'),
    headers: {
      'Authorization': 'Bearer $token',
      'Content-Type': 'application/json',
    },
    body: jsonEncode(body),
  );
  
  print('📥 Response: ${response.statusCode} - ${response.body}');
}
```

---

## 🧪 Quick Test

### Test 1: Check Current Value
Add this debug code:
```dart
print('Current consultation_type: $consultationType');
// If it prints "offline" or "online" → WRONG!
// Should print "clinic_visit" or "video_consultation"
```

### Test 2: Hardcode Correct Value
Temporarily hardcode to test:
```dart
final body = {
  'consultation_type': 'clinic_visit',  // ✅ Hardcoded to test
  // ... other fields
};
```

If this works, then you know the issue is where you set `consultationType`.

---

## 🎯 Common Mistakes to Avoid

### ❌ Mistake 1: Using Old Constants
```dart
// ❌ DON'T DO THIS
String type = AppointmentTypes.offline;  // If this is 'offline'
```

### ✅ Fix:
```dart
// ✅ DO THIS
String type = AppointmentTypes.clinicVisit;  // Should be 'clinic_visit'
```

---

### ❌ Mistake 2: Conditional with Old Values
```dart
// ❌ DON'T DO THIS
if (isOnline) {
  consultationType = 'online';  // Wrong!
} else {
  consultationType = 'offline';  // Wrong!
}
```

### ✅ Fix:
```dart
// ✅ DO THIS
if (isOnline) {
  consultationType = 'video_consultation';  // Correct!
} else {
  consultationType = 'clinic_visit';  // Correct!
}
```

---

### ❌ Mistake 3: Dropdown Value Not Updated
```dart
// ❌ DON'T DO THIS
DropdownMenuItem(
  value: 'offline',  // Wrong value!
  child: Text('Clinic Visit'),
)
```

### ✅ Fix:
```dart
// ✅ DO THIS
DropdownMenuItem(
  value: 'clinic_visit',  // Correct value!
  child: Text('🏥 Clinic Visit'),
)
```

---

## 📱 Complete Flutter Example

```dart
class AppointmentBookingPage extends StatefulWidget {
  @override
  _AppointmentBookingPageState createState() => _AppointmentBookingPageState();
}

class _AppointmentBookingPageState extends State<AppointmentBookingPage> {
  // ✅ NEW: Use correct values
  String selectedConsultationType = 'clinic_visit';
  
  // ✅ NEW: Define valid options
  final List<Map<String, String>> consultationOptions = [
    {
      'value': 'clinic_visit',
      'label': '🏥 Clinic Visit',
      'description': 'In-person visit to clinic',
    },
    {
      'value': 'video_consultation',
      'label': '💻 Video Consultation',
      'description': 'Online video call',
    },
    {
      'value': 'in_person',
      'label': '👤 In Person',
      'description': 'Direct in-person meeting',
    },
    {
      'value': 'follow_up',
      'label': '🔄 Follow-Up',
      'description': 'Return visit',
    },
  ];
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          // ✅ Dropdown with correct values
          DropdownButtonFormField<String>(
            decoration: InputDecoration(
              labelText: 'Consultation Type',
              border: OutlineInputBorder(),
            ),
            value: selectedConsultationType,
            items: consultationOptions.map((option) {
              return DropdownMenuItem<String>(
                value: option['value'],  // ✅ Correct: 'clinic_visit', not 'offline'
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      option['label']!,
                      style: TextStyle(fontWeight: FontWeight.bold),
                    ),
                    Text(
                      option['description']!,
                      style: TextStyle(fontSize: 12, color: Colors.grey),
                    ),
                  ],
                ),
              );
            }).toList(),
            onChanged: (value) {
              setState(() {
                selectedConsultationType = value!;
              });
            },
          ),
          
          SizedBox(height: 20),
          
          // Book button
          ElevatedButton(
            onPressed: _bookAppointment,
            child: Text('Book Appointment'),
          ),
        ],
      ),
    );
  }
  
  Future<void> _bookAppointment() async {
    final body = {
      'clinic_patient_id': widget.patientId,
      'doctor_id': widget.doctorId,
      'clinic_id': widget.clinicId,
      'individual_slot_id': widget.slotId,
      'appointment_date': widget.date,
      'appointment_time': widget.time,
      'consultation_type': selectedConsultationType,  // ✅ Will be 'clinic_visit' or 'video_consultation'
      'is_follow_up': widget.isFollowUp,
      'payment_method': selectedPaymentMethod,
      'payment_type': selectedPaymentType,
    };
    
    // ✅ Debug print
    print('📤 Sending: consultation_type = $selectedConsultationType');
    
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/appointments/simple'),
        headers: {
          'Authorization': 'Bearer $token',
          'Content-Type': 'application/json',
        },
        body: jsonEncode(body),
      );
      
      if (response.statusCode == 201) {
        // Success!
        print('✅ Appointment created successfully');
      } else {
        // Error
        final error = jsonDecode(response.body);
        print('❌ Error: ${error['error']}');
        print('❌ Details: ${error['details']}');
      }
    } catch (e) {
      print('❌ Exception: $e');
    }
  }
}
```

---

## ✅ Checklist

Before testing, verify:

- [ ] Constants updated (`offline` → `clinic_visit`, `online` → `video_consultation`)
- [ ] Dropdown values updated
- [ ] Model class updated
- [ ] Default values updated
- [ ] Conditional logic updated
- [ ] No hardcoded `'offline'` or `'online'` strings
- [ ] Debug print added to check sent value

---

## 🧪 Test

1. **Run your Flutter app**
2. **Open appointment booking page**
3. **Check console for debug print:**
   ```
   📤 Sending: consultation_type = clinic_visit  ✅ CORRECT
   ```
   NOT:
   ```
   📤 Sending: consultation_type = offline  ❌ WRONG
   ```

4. **Try booking appointment**
5. **Should now work!** ✅

---

## 📞 Still Having Issues?

If it still doesn't work:

1. **Search your entire Flutter project for:**
   - `'offline'` (in quotes)
   - `'online'` (in quotes)
   - `consultation_type`

2. **Replace ALL occurrences of:**
   - `'offline'` → `'clinic_visit'`
   - `'online'` → `'video_consultation'`

3. **Restart your Flutter app** (hot reload might not be enough)

---

**Quick Fix:** Just replace these two values everywhere in your Flutter code and you're done! ✅

