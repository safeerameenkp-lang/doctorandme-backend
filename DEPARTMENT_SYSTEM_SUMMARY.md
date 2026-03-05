# Department System - Complete Implementation ✅

## 🎯 How Department Works

The system now supports **two levels of department assignment**:
1. **Doctor's Default Department** (from doctors table)
2. **Appointment-Specific Department** (selected during booking)

---

## 📊 Department Priority Logic

### Query Logic:

```sql
SELECT COALESCE(dept_appt.name, dept_doc.name) as department
FROM appointments a
LEFT JOIN departments dept_appt ON dept_appt.id = a.department_id      -- Appointment's department
LEFT JOIN departments dept_doc ON dept_doc.id = d.department_id        -- Doctor's department
```

**Priority:**
1. **First:** Use appointment's `department_id` (if user selected)
2. **Fallback:** Use doctor's `department_id` (if no selection)
3. **Default:** `NULL` (if neither exists)

---

## 📋 Scenarios

### Scenario 1: User Selects Department During Booking

**Request:**
```json
POST /api/appointments/simple
{
  "doctor_id": "doctor-uuid",
  "department_id": "cardiology-uuid",  // ✅ User selected Cardiology
  ...
}
```

**Database:**
```sql
appointments table:
  department_id = 'cardiology-uuid'  ✅

doctors table:
  department_id = 'general-medicine-uuid'  (doctor's default)
```

**Appointment List Response:**
```json
{
  "department": "Cardiology"  // ✅ Shows user-selected department
}
```

---

### Scenario 2: User Does NOT Select Department

**Request:**
```json
POST /api/appointments/simple
{
  "doctor_id": "doctor-uuid",
  // department_id not provided
  ...
}
```

**Database:**
```sql
appointments table:
  department_id = NULL  (not selected)

doctors table:
  department_id = 'general-medicine-uuid'  ✅ Doctor's default
```

**Appointment List Response:**
```json
{
  "department": "General Medicine"  // ✅ Shows doctor's default department
}
```

---

### Scenario 3: Neither Department Set

**Database:**
```sql
appointments table:
  department_id = NULL

doctors table:
  department_id = NULL
```

**Appointment List Response:**
```json
{
  "department": null  // or could default to "-"
}
```

---

## 🔄 Data Flow

### Creating Appointment

```
UI Dropdown:
  Select Department: "Cardiology"
           ↓
API Request:
  department_id: "cardiology-uuid"
           ↓
Database INSERT:
  appointments.department_id = "cardiology-uuid"
           ↓
✅ Saved
```

---

### Listing Appointments

```
Database Query:
  COALESCE(dept_appt.name, dept_doc.name)
           ↓
Priority Check:
  1. appointments.department_id → "Cardiology" ✅
  2. doctors.department_id → "General Medicine"
           ↓
API Response:
  "department": "Cardiology"  ✅
           ↓
UI Table:
  Shows "Cardiology" in Department column
```

---

## 📊 Database Structure

### Appointments Table
```sql
appointments:
  department_id UUID  -- Optional: User-selected department
```

### Doctors Table
```sql
doctors:
  department_id UUID  -- Doctor's default department
```

### Departments Table
```sql
departments:
  id UUID
  name VARCHAR  -- "Cardiology", "Orthopedics", etc.
```

---

## 🎨 UI Implementation

### Department Dropdown

```dart
// Load departments for dropdown
Future<List<Department>> getDepartments() async {
  final response = await http.get(
    Uri.parse('$baseUrl/departments'),
    headers: {'Authorization': 'Bearer $token'},
  );
  
  final data = jsonDecode(response.body);
  return (data['departments'] as List)
      .map((json) => Department.fromJson(json))
      .toList();
}

// Show in UI
DropdownButtonFormField<String>(
  decoration: InputDecoration(
    labelText: 'Department',
    hintText: 'Select department (Optional)',
  ),
  value: selectedDepartmentId,
  items: [
    // Option 1: Use doctor's default
    DropdownMenuItem(
      value: null,
      child: Text('Use Doctor\'s Default Department'),
    ),
    // Option 2+: All available departments
    ...departments.map((dept) => DropdownMenuItem(
      value: dept.id,
      child: Text(dept.name),  // Cardiology, Orthopedics, etc.
    )),
  ],
  onChanged: (value) {
    setState(() => selectedDepartmentId = value);
  },
)

// Include in API request
final body = {
  'doctor_id': doctorId,
  'clinic_id': clinicId,
  ...
};

// Only add if user selected a department
if (selectedDepartmentId != null) {
  body['department_id'] = selectedDepartmentId;
}
```

---

## 📝 Complete Example

### Create Appointment with Department

**Request:**
```bash
POST /api/appointments/simple
Content-Type: application/json
Authorization: Bearer {token}

{
  "clinic_patient_id": "752590e9-deda-4043-a5e2-7f9366f00cfc",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "department_id": "dept-cardiology-uuid",    // ✅ User selected Cardiology
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
    "department_id": "dept-cardiology-uuid",
    "status": "confirmed"
  }
}
```

---

### List Appointments with Department

**Request:**
```bash
GET /api/appointments/simple-list?clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&date=2025-10-17
Authorization: Bearer {token}
```

**Response:**
```json
{
  "success": true,
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "date": "2025-10-17",
  "total": 3,
  "appointments": [
    {
      "id": "apt-uuid-1",
      "token_number": 1,
      "mo_id": "MO2024100001",
      "patient_name": "Ahmed Ali",
      "doctor_name": "Dr. Sara Ahmed",
      "department": "Cardiology",              // ✅ Department shown
      "consultation_type": "offline",
      "appointment_date_time": "2025-10-17 13:07:00",
      "status": "confirmed",
      "fee_status": "waived",
      "fee_amount": 500.00,
      "payment_status": "waived",
      "booking_number": "BN202510170001"
    },
    {
      "id": "apt-uuid-2",
      "token_number": 1,
      "mo_id": "MO2024100002",
      "patient_name": "Fatima Hassan",
      "doctor_name": "Dr. Ahmed Ibrahim",
      "department": "General Medicine",        // ✅ Doctor's default (user didn't select)
      "consultation_type": "offline",
      "appointment_date_time": "2025-10-17 14:00:00",
      "status": "confirmed",
      "fee_status": "paid",
      "payment_status": "paid"
    }
  ]
}
```

---

## ✅ Summary

### API Changes

| API | Department Support | Status |
|-----|-------------------|--------|
| `POST /appointments/simple` | ✅ Accepts `department_id` (optional) | ✅ Updated |
| `GET /appointments/simple-list` | ✅ Returns department name | ✅ Updated |

### Department Logic

| Priority | Source | Field |
|----------|--------|-------|
| 1️⃣ First | Appointment's department | `a.department_id` |
| 2️⃣ Fallback | Doctor's department | `d.department_id` |
| 3️⃣ Default | NULL | - |

### Query Structure

```sql
-- Two JOINs for departments:
LEFT JOIN departments dept_appt ON dept_appt.id = a.department_id      -- Appointment's dept
LEFT JOIN departments dept_doc ON dept_doc.id = d.department_id        -- Doctor's dept

-- Priority selection:
COALESCE(dept_appt.name, dept_doc.name) as department
```

---

## 🧪 Testing

### Test 1: Create with Department Selection

```bash
POST /api/appointments/simple
{
  "department_id": "cardiology-uuid"
}

# Then check list:
GET /api/appointments/simple-list?clinic_id=xxx

# Should show:
{
  "department": "Cardiology"  ✅
}
```

---

### Test 2: Create without Department Selection

```bash
POST /api/appointments/simple
{
  // No department_id
}

# Then check list:
GET /api/appointments/simple-list?clinic_id=xxx

# Should show:
{
  "department": "General Medicine"  ✅ (Doctor's default)
}
```

---

## 📋 Files Updated

| File | Change | Status |
|------|--------|--------|
| `appointment_simple.controller.go` | Added `department_id` input | ✅ Done |
| `appointment_list_simple.controller.go` | Enhanced department JOIN | ✅ Done |
| `SIMPLE_APPOINTMENT_API_GUIDE.md` | Added department docs | ✅ Done |
| `DEPARTMENT_SELECTION_GUIDE.md` | Complete guide | ✅ Created |
| `DEPARTMENT_SYSTEM_SUMMARY.md` | This summary | ✅ Created |

---

**Status:** ✅ **Department system fully integrated!**

**Create:** Can select department or use doctor's default  
**List:** Shows correct department name in table  
**Ready:** ✅ All fixes applied! 🏥🎉

