# Department Selection in Appointments 🏥

## 🎯 Overview

The appointment creation API now supports **optional department selection**. The UI can show a dropdown of departments for the user to select.

---

## 📋 API Flow

### Step 1: Get Departments List

**Endpoint:** `GET /api/organizations/departments`

**Purpose:** Get all departments for dropdown

**Request:**
```bash
GET /api/organizations/departments
Authorization: Bearer {token}
```

**Response:**
```json
{
  "departments": [
    {
      "id": "dept-uuid-1",
      "name": "Cardiology",
      "description": "Heart and cardiovascular care",
      "is_active": true
    },
    {
      "id": "dept-uuid-2",
      "name": "Orthopedics",
      "description": "Bone and joint care",
      "is_active": true
    },
    {
      "id": "dept-uuid-3",
      "name": "General Medicine",
      "description": "General health consultation",
      "is_active": true
    }
  ]
}
```

---

### Step 2: Create Appointment with Department

**Endpoint:** `POST /api/appointments/simple`

**Request (With Department):**
```json
{
  "clinic_patient_id": "patient-uuid",
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "department_id": "dept-uuid-1",        // ✅ Optional: Cardiology selected
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "consultation_type": "offline",
  "payment_method": "pay_now",
  "payment_type": "cash"
}
```

**Request (Without Department - Uses Doctor's Default):**
```json
{
  "clinic_patient_id": "patient-uuid",
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  // department_id not provided - uses doctor's default department
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "consultation_type": "offline",
  "payment_method": "pay_now",
  "payment_type": "cash"
}
```

---

## 📱 Flutter Integration

### 1. Department Model

```dart
class Department {
  final String id;
  final String name;
  final String? description;
  final bool isActive;
  
  Department({
    required this.id,
    required this.name,
    this.description,
    required this.isActive,
  });
  
  factory Department.fromJson(Map<String, dynamic> json) {
    return Department(
      id: json['id'],
      name: json['name'],
      description: json['description'],
      isActive: json['is_active'] ?? true,
    );
  }
}
```

---

### 2. Get Departments Service

```dart
class DepartmentService {
  final String baseUrl = 'http://localhost:8081/api/organizations';
  final String token;
  
  DepartmentService(this.token);
  
  Future<List<Department>> getDepartments() async {
    final response = await http.get(
      Uri.parse('$baseUrl/departments'),
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
    );
    
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      final List departments = data['departments'] ?? [];
      
      return departments
          .map((json) => Department.fromJson(json))
          .toList();
    } else {
      throw Exception('Failed to load departments');
    }
  }
}
```

---

### 3. UI - Department Dropdown

```dart
class AppointmentBookingForm extends StatefulWidget {
  @override
  _AppointmentBookingFormState createState() => _AppointmentBookingFormState();
}

class _AppointmentBookingFormState extends State<AppointmentBookingForm> {
  List<Department> departments = [];
  Department? selectedDepartment;
  bool isLoadingDepartments = true;
  
  @override
  void initState() {
    super.initState();
    loadDepartments();
  }
  
  Future<void> loadDepartments() async {
    try {
      final service = DepartmentService(yourToken);
      final data = await service.getDepartments();
      
      setState(() {
        departments = data;
        isLoadingDepartments = false;
      });
    } catch (e) {
      setState(() => isLoadingDepartments = false);
      print('Error loading departments: $e');
    }
  }
  
  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        // Department Dropdown
        DropdownButtonFormField<Department>(
          decoration: InputDecoration(
            labelText: 'Department (Optional)',
            hintText: 'Select department',
            border: OutlineInputBorder(),
          ),
          value: selectedDepartment,
          items: departments.map((dept) {
            return DropdownMenuItem<Department>(
              value: dept,
              child: Text(dept.name),
            );
          }).toList(),
          onChanged: (Department? value) {
            setState(() {
              selectedDepartment = value;
            });
          },
        ),
        
        // Other form fields...
        
        // Submit Button
        ElevatedButton(
          onPressed: () async {
            await createAppointment(
              departmentId: selectedDepartment?.id,  // ✅ Optional
            );
          },
          child: Text('Book Appointment'),
        ),
      ],
    );
  }
  
  Future<void> createAppointment({String? departmentId}) async {
    final body = {
      'clinic_patient_id': patientId,
      'doctor_id': doctorId,
      'clinic_id': clinicId,
      'individual_slot_id': slotId,
      'appointment_date': appointmentDate,
      'appointment_time': appointmentTime,
      'consultation_type': consultationType,
      'payment_method': paymentMethod,
    };
    
    // Add department_id only if selected
    if (departmentId != null) {
      body['department_id'] = departmentId;
    }
    
    // Add payment_type if pay_now
    if (paymentMethod == 'pay_now' && paymentType != null) {
      body['payment_type'] = paymentType;
    }
    
    final response = await http.post(
      Uri.parse('$baseUrl/appointments/simple'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode(body),
    );
    
    // Handle response...
  }
}
```

---

## 📊 Department List API

### Get All Departments

**Endpoint:** `GET /api/organizations/departments`

**Response:**
```json
{
  "departments": [
    {
      "id": "dept-uuid-1",
      "name": "Cardiology",
      "description": "Heart and cardiovascular care",
      "is_active": true
    },
    {
      "id": "dept-uuid-2",
      "name": "Orthopedics",
      "description": "Bone and joint care",
      "is_active": true
    }
  ],
  "total": 2
}
```

---

## 📝 Complete Example

### Scenario: Patient Booking Appointment with Department Selection

**Step 1: Load Departments**
```bash
GET /api/organizations/departments
→ Returns: Cardiology, Orthopedics, General Medicine
```

**Step 2: User Selects Department**
```
UI Dropdown: "Cardiology" selected
department_id = "dept-uuid-cardiology"
```

**Step 3: Create Appointment**
```json
POST /api/appointments/simple
{
  "clinic_patient_id": "752590e9-deda-4043-a5e2-7f9366f00cfc",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "department_id": "dept-uuid-cardiology",        // ✅ Selected department
  "individual_slot_id": "bbf60035-c9e3-4a80-b88c-10cbb0925d06",
  "appointment_date": "2025-10-17",
  "appointment_time": "2025-10-17 13:07:00",
  "consultation_type": "offline",
  "payment_method": "way_off"
}
```

**Response:**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "id": "appointment-uuid",
    "booking_number": "BN202510170001",
    "token_number": 1,
    "department_id": "dept-uuid-cardiology",  // ✅ Department saved
    "status": "confirmed"
  }
}
```

**Step 4: Appointment List Shows Department**
```json
GET /api/appointments/simple-list?clinic_id=xxx&date=2025-10-17

{
  "appointments": [
    {
      "token_number": 1,
      "patient_name": "Ahmed Ali",
      "doctor_name": "Dr. Sara Ahmed",
      "department": "Cardiology",               // ✅ Department name displayed
      "status": "confirmed"
    }
  ]
}
```

---

## 🎨 UI Design Suggestions

### Department Dropdown Widget

```dart
DropdownButtonFormField<String>(
  decoration: InputDecoration(
    labelText: 'Department',
    hintText: 'Select department (Optional)',
    prefixIcon: Icon(Icons.local_hospital),
    border: OutlineInputBorder(),
  ),
  value: selectedDepartmentId,
  items: [
    DropdownMenuItem(
      value: null,
      child: Text('Doctor\'s Default Department'),
    ),
    ...departments.map((dept) => DropdownMenuItem(
      value: dept.id,
      child: Row(
        children: [
          Icon(Icons.business, size: 16),
          SizedBox(width: 8),
          Text(dept.name),
        ],
      ),
    )),
  ],
  onChanged: (value) {
    setState(() => selectedDepartmentId = value);
  },
)
```

---

## 📊 When to Use Department Selection

| Scenario | Department Selection | Behavior |
|----------|---------------------|----------|
| Doctor has single department | Optional | Uses doctor's default |
| Doctor works in multiple departments | Required | User must select |
| Department-specific fees | Required | For accurate billing |
| General consultation | Optional | Can skip |

---

## ✅ Summary

| Feature | Status |
|---------|--------|
| Department selection | ✅ Added to API |
| Optional field | ✅ Can be omitted |
| Department list API | ✅ Available |
| Saves to appointments | ✅ Yes |
| Shows in appointment list | ✅ Yes |

---

**API Updates:**

```json
// Request Input
{
  "department_id": "dept-uuid"  // ✅ NEW: Optional department
}

// Database Storage
appointments table:
  department_id: "dept-uuid"    // ✅ Saved

// List Response
{
  "department": "Cardiology"    // ✅ Department name shown
}
```

---

**Status:** ✅ **Department selection integrated! UI can now select departments from dropdown!** 🏥

