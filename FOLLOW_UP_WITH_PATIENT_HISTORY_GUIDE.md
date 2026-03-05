# Follow-Up Appointments with Patient History - Complete Guide 🔄

## 🎯 Better Approach

Instead of a separate eligibility check API, **appointment history and follow-up eligibility are now included directly with clinic patient data**. This is more efficient and provides better UX!

---

## 📋 What Changed

### ❌ Old Approach (Removed):
```bash
# Separate API call needed
GET /api/appointments/check-follow-up-eligibility?
  clinic_patient_id=xxx&clinic_id=xxx
```

### ✅ New Approach (Better):
```bash
# Appointment history included automatically with patient data
GET /api/organizations/clinic-specific-patients?clinic_id=xxx

# Or single patient
GET /api/organizations/clinic-specific-patients/{patient_id}
```

**Result:** Patient data now includes `last_appointment`, `follow_up_eligibility`, and `total_appointments`!

---

## 🌐 Updated Patient Response

### List Clinic Patients

**Request:**
```bash
GET /api/organizations/clinic-specific-patients?clinic_id={uuid}
Authorization: Bearer {token}
```

**Response:**
```json
{
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "total": 2,
  "patients": [
    {
      "id": "patient-uuid-1",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "first_name": "Ahmed",
      "last_name": "Ali",
      "phone": "1234567890",
      "email": "ahmed@example.com",
      "mo_id": "MO2024100001",
      "is_active": true,
      
      // ✅ NEW: Last appointment details
      "last_appointment": {
        "id": "appointment-uuid",
        "doctor_id": "doctor-uuid",
        "doctor_name": "Dr. Sara Ahmed",
        "department_id": "dept-uuid",
        "department": "Cardiology",
        "date": "2025-10-17",
        "status": "completed",
        "days_since": 2
      },
      
      // ✅ NEW: Follow-up eligibility (auto-calculated)
      "follow_up_eligibility": {
        "eligible": true,
        "days_remaining": 3
      },
      
      // ✅ NEW: Total appointment count
      "total_appointments": 5
    },
    {
      "id": "patient-uuid-2",
      "first_name": "Fatima",
      "last_name": "Hassan",
      "phone": "0987654321",
      
      // ✅ NEW: No previous appointment
      "last_appointment": null,
      
      // ✅ NEW: Not eligible (no previous appointment)
      "follow_up_eligibility": {
        "eligible": false,
        "reason": "No previous appointment found"
      },
      
      "total_appointments": 0
    }
  ]
}
```

---

## 📊 Follow-Up Eligibility States

### 1. Eligible (Within 5 Days) ✅

```json
{
  "last_appointment": {
    "id": "uuid",
    "doctor_id": "doctor-uuid",
    "doctor_name": "Dr. Sara Ahmed",
    "department_id": "dept-uuid",
    "department": "Cardiology",
    "date": "2025-10-17",
    "days_since": 2
  },
  "follow_up_eligibility": {
    "eligible": true,
    "days_remaining": 3  // 5 - 2 = 3 days left
  }
}
```

**UI Action:** Show "Book Follow-Up" button

---

### 2. Not Eligible - Expired (> 5 Days) ❌

```json
{
  "last_appointment": {
    "id": "uuid",
    "doctor_id": "doctor-uuid",
    "doctor_name": "Dr. Ahmed Ibrahim",
    "date": "2025-10-10",
    "days_since": 9
  },
  "follow_up_eligibility": {
    "eligible": false,
    "reason": "Follow-up period expired (must book within 5 days)"
  }
}
```

**UI Action:** Show only "Book Regular Appointment" button

---

### 3. Not Eligible - No Previous Appointment ❌

```json
{
  "last_appointment": null,
  "follow_up_eligibility": {
    "eligible": false,
    "reason": "No previous appointment found"
  },
  "total_appointments": 0
}
```

**UI Action:** Show only "Book Regular Appointment" button

---

## 💻 Flutter Integration

### Step 1: Load Patients with History

```dart
class PatientListPage extends StatefulWidget {
  final String clinicId;
  
  @override
  _PatientListPageState createState() => _PatientListPageState();
}

class _PatientListPageState extends State<PatientListPage> {
  List<Patient> patients = [];
  
  @override
  void initState() {
    super.initState();
    _loadPatients();
  }
  
  Future<void> _loadPatients() async {
    final response = await http.get(
      Uri.parse('$baseUrl/organizations/clinic-specific-patients?'
        'clinic_id=${widget.clinicId}'),
      headers: {'Authorization': 'Bearer $token'},
    );
    
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      setState(() {
        patients = (data['patients'] as List)
          .map((json) => Patient.fromJson(json))
          .toList();
      });
    }
  }
}
```

---

### Step 2: Patient Model with History

```dart
class Patient {
  final String id;
  final String firstName;
  final String lastName;
  final String phone;
  
  // ✅ NEW: Appointment history fields
  final LastAppointment? lastAppointment;
  final FollowUpEligibility followUpEligibility;
  final int totalAppointments;
  
  Patient({
    required this.id,
    required this.firstName,
    required this.lastName,
    required this.phone,
    this.lastAppointment,
    required this.followUpEligibility,
    required this.totalAppointments,
  });
  
  factory Patient.fromJson(Map<String, dynamic> json) {
    return Patient(
      id: json['id'],
      firstName: json['first_name'],
      lastName: json['last_name'],
      phone: json['phone'],
      lastAppointment: json['last_appointment'] != null
        ? LastAppointment.fromJson(json['last_appointment'])
        : null,
      followUpEligibility: FollowUpEligibility.fromJson(
        json['follow_up_eligibility']
      ),
      totalAppointments: json['total_appointments'] ?? 0,
    );
  }
  
  // ✅ Helper: Check if eligible for follow-up
  bool get canBookFollowUp => followUpEligibility.eligible;
}

class LastAppointment {
  final String id;
  final String doctorId;
  final String doctorName;
  final String? departmentId;
  final String? department;
  final String date;
  final int daysSince;
  
  LastAppointment({
    required this.id,
    required this.doctorId,
    required this.doctorName,
    this.departmentId,
    this.department,
    required this.date,
    required this.daysSince,
  });
  
  factory LastAppointment.fromJson(Map<String, dynamic> json) {
    return LastAppointment(
      id: json['id'],
      doctorId: json['doctor_id'],
      doctorName: json['doctor_name'],
      departmentId: json['department_id'],
      department: json['department'],
      date: json['date'],
      daysSince: json['days_since'],
    );
  }
}

class FollowUpEligibility {
  final bool eligible;
  final String? reason;
  final int? daysRemaining;
  
  FollowUpEligibility({
    required this.eligible,
    this.reason,
    this.daysRemaining,
  });
  
  factory FollowUpEligibility.fromJson(Map<String, dynamic> json) {
    return FollowUpEligibility(
      eligible: json['eligible'],
      reason: json['reason'],
      daysRemaining: json['days_remaining'],
    );
  }
}
```

---

### Step 3: UI with Follow-Up Option

```dart
Widget _buildPatientCard(Patient patient) {
  return Card(
    child: ListTile(
      title: Text('${patient.firstName} ${patient.lastName}'),
      subtitle: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Phone: ${patient.phone}'),
          if (patient.lastAppointment != null)
            Text(
              'Last visit: ${patient.lastAppointment!.date} '
              '(${patient.lastAppointment!.daysSince} days ago)',
              style: TextStyle(fontSize: 12, color: Colors.grey[600]),
            ),
          Text('Total appointments: ${patient.totalAppointments}'),
          
          // ✅ Show eligibility status
          if (patient.canBookFollowUp)
            Container(
              margin: EdgeInsets.only(top: 4),
              padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: Colors.green[50],
                borderRadius: BorderRadius.circular(4),
              ),
              child: Text(
                '✅ Eligible for Follow-Up (${patient.followUpEligibility.daysRemaining} days left)',
                style: TextStyle(
                  color: Colors.green[700],
                  fontWeight: FontWeight.bold,
                  fontSize: 12,
                ),
              ),
            ),
          if (!patient.canBookFollowUp && patient.followUpEligibility.reason != null)
            Text(
              patient.followUpEligibility.reason!,
              style: TextStyle(fontSize: 11, color: Colors.grey),
            ),
        ],
      ),
      trailing: ElevatedButton(
        onPressed: () => _bookAppointment(patient),
        child: Text('Book Appointment'),
      ),
    ),
  );
}
```

---

### Step 4: Booking with Auto-Detection

```dart
void _bookAppointment(Patient patient) {
  // ✅ Automatically determine if it's a follow-up
  bool isFollowUp = patient.canBookFollowUp;
  
  Navigator.push(
    context,
    MaterialPageRoute(
      builder: (context) => AppointmentBookingPage(
        patient: patient,
        isFollowUp: isFollowUp,
        
        // ✅ If follow-up, pre-fill doctor & department
        preselectedDoctorId: isFollowUp 
          ? patient.lastAppointment!.doctorId 
          : null,
        preselectedDepartmentId: isFollowUp 
          ? patient.lastAppointment!.departmentId 
          : null,
      ),
    ),
  );
}
```

---

### Step 5: Appointment Booking Page

```dart
class AppointmentBookingPage extends StatefulWidget {
  final Patient patient;
  final bool isFollowUp;
  final String? preselectedDoctorId;
  final String? preselectedDepartmentId;
  
  @override
  _AppointmentBookingPageState createState() => _AppointmentBookingPageState();
}

class _AppointmentBookingPageState extends State<AppointmentBookingPage> {
  String? selectedDoctorId;
  String? selectedDepartmentId;
  
  @override
  void initState() {
    super.initState();
    
    // ✅ Pre-fill if follow-up
    if (widget.isFollowUp) {
      selectedDoctorId = widget.preselectedDoctorId;
      selectedDepartmentId = widget.preselectedDepartmentId;
    }
  }
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.isFollowUp 
          ? 'Book Follow-Up Appointment' 
          : 'Book Appointment'),
      ),
      body: Column(
        children: [
          // ✅ Show follow-up banner
          if (widget.isFollowUp)
            Container(
              padding: EdgeInsets.all(16),
              color: Colors.green[50],
              child: Row(
                children: [
                  Icon(Icons.check_circle, color: Colors.green),
                  SizedBox(width: 8),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Follow-Up Appointment',
                          style: TextStyle(
                            fontWeight: FontWeight.bold,
                            color: Colors.green[700],
                          ),
                        ),
                        Text(
                          '✅ FREE - No payment required\n'
                          'Doctor: ${widget.patient.lastAppointment!.doctorName}\n'
                          'Department: ${widget.patient.lastAppointment!.department ?? 'N/A'}',
                          style: TextStyle(fontSize: 12),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          
          // ✅ Doctor selector (disabled for follow-ups)
          DropdownButtonFormField<String>(
            decoration: InputDecoration(labelText: 'Doctor'),
            value: selectedDoctorId,
            enabled: !widget.isFollowUp,  // Disabled for follow-ups
            items: doctors.map((doctor) => DropdownMenuItem(
              value: doctor.id,
              child: Text(doctor.name),
            )).toList(),
            onChanged: widget.isFollowUp ? null : (value) {
              setState(() => selectedDoctorId = value);
            },
          ),
          
          // ... rest of booking form
          
          ElevatedButton(
            onPressed: _submitBooking,
            child: Text(widget.isFollowUp 
              ? 'Book Follow-Up (Free)' 
              : 'Book Appointment'),
          ),
        ],
      ),
    );
  }
  
  Future<void> _submitBooking() async {
    final body = {
      'clinic_patient_id': widget.patient.id,
      'doctor_id': selectedDoctorId,
      'department_id': selectedDepartmentId,
      'clinic_id': clinicId,
      'individual_slot_id': selectedSlotId,
      'appointment_date': selectedDate,
      'appointment_time': selectedTime,
      'consultation_type': widget.isFollowUp ? 'follow_up' : 'offline',
      'is_follow_up': widget.isFollowUp,  // ✅ Key flag
    };
    
    // ✅ Only add payment for regular appointments
    if (!widget.isFollowUp) {
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
      _showSuccess();
    }
  }
}
```

---

## 📊 Backend Logic

### Helper Function (Auto-Called)

```go
func populateAppointmentHistory(patient *ClinicPatientResponse, db *sql.DB) {
    // Query last appointment
    // Calculate days since
    // Determine follow-up eligibility
    // Count total appointments
}
```

**Called in:**
- ✅ `ListClinicPatients` - For each patient in list
- ✅ `GetClinicPatient` - For single patient

---

## ✅ Benefits of New Approach

| Aspect | Old Approach | New Approach ✅ |
|--------|--------------|-----------------|
| API calls | 2 (list + eligibility) | 1 (list with history) |
| Performance | Slower (2 requests) | Faster (1 request) |
| UX | Delayed eligibility check | Instant display |
| Code | Separate eligibility API | Integrated automatically |
| Maintenance | More endpoints to maintain | Simpler architecture |

---

## 🧪 Testing

### Test 1: List Patients with History

```bash
GET /api/organizations/clinic-specific-patients?clinic_id=xxx

# Should return patients with:
# - last_appointment (if exists)
# - follow_up_eligibility (auto-calculated)
# - total_appointments
```

### Test 2: Eligible Patient

```json
{
  "follow_up_eligibility": {
    "eligible": true,
    "days_remaining": 3
  }
}
```

### Test 3: Expired Follow-Up

```json
{
  "last_appointment": {
    "days_since": 7
  },
  "follow_up_eligibility": {
    "eligible": false,
    "reason": "Follow-up period expired (must book within 5 days)"
  }
}
```

---

## 📝 Files Changed

| File | Changes |
|------|---------|
| `clinic_patient.controller.go` | Added LastAppointmentInfo, FollowUpEligibility structs, populateAppointmentHistory helper |
| `appointment_simple.controller.go` | Removed CheckFollowUpEligibility API (not needed) |
| `appointment.routes.go` | Removed eligibility check route |

---

## ✅ Status

**Implementation:** ✅ **COMPLETE**

**Features:**
- ✅ Appointment history in patient data
- ✅ Auto-calculated follow-up eligibility
- ✅ Total appointment count
- ✅ No separate API needed
- ✅ Better performance
- ✅ Cleaner code

**Ready for:**
- ✅ UI Integration
- ✅ Testing
- ✅ Production

---

**Done!** 🔄✅🎉

