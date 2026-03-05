# Simple Appointment List API Guide 📋

## 🎯 Overview

A simple API endpoint that returns appointments in a table-ready format with all necessary columns for display.

---

## 📝 API Endpoint

**Method:** `GET`  
**Endpoint:** `/api/appointments/simple-list`

**Query Parameters:**
- `clinic_id` (required) - Filter by clinic
- `date` (optional) - Filter by specific date (YYYY-MM-DD)

---

## 📊 Response Fields

| Field | Type | Description | Display Column |
|-------|------|-------------|----------------|
| `id` | UUID | Appointment ID | (Hidden) |
| `token_number` | Integer | Queue token | **Token** |
| `mo_id` | String | Patient MO ID | **Mo ID** |
| `patient_name` | String | Patient full name | **Patient Name** |
| `doctor_name` | String | Doctor full name | **Doctor's Name** |
| `department` | String | Doctor's department | **Department** |
| `consultation_type` | String | Type of consultation | **Consultation Type** |
| `appointment_date` | Date | Appointment date | **Appointment Date** |
| `appointment_time` | DateTime | Appointment time | **Appointment Time** |
| `status` | String | Appointment status | **STATUS** |
| `fee_amount` | Float | Consultation fee | (For calculations) |
| `payment_status` | String | Payment status | **Fee Status** |
| `booking_number` | String | Booking reference | (For details) |

---

## ✅ Example Request & Response

### Example 1: Get All Appointments for Clinic

**Request:**
```bash
GET /api/appointments/simple-list?clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "success": true,
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "date": "",
  "total": 5,
  "appointments": [
    {
      "id": "apt-uuid-1",
      "token_number": 1,
      "mo_id": "MO2024100001",
      "patient_name": "Ahmed Ali",
      "doctor_name": "Dr. Sara Ahmed",
      "department": "Cardiology",
      "consultation_type": "offline",
      "appointment_date": "2025-10-17",
      "appointment_time": "2025-10-17 09:00:00",
      "status": "confirmed",
      "fee_amount": 500.00,
      "payment_status": "paid",
      "booking_number": "BN202510170001",
      "reason": "Regular checkup",
      "notes": null
    },
    {
      "id": "apt-uuid-2",
      "token_number": 2,
      "mo_id": "MO2024100002",
      "patient_name": "Fatima Hassan",
      "doctor_name": "Dr. Sara Ahmed",
      "department": "Cardiology",
      "consultation_type": "offline",
      "appointment_date": "2025-10-17",
      "appointment_time": "2025-10-17 09:05:00",
      "status": "confirmed",
      "fee_amount": 500.00,
      "payment_status": "pending",
      "booking_number": "BN202510170002",
      "reason": "Follow-up",
      "notes": "Needs ECG report"
    },
    {
      "id": "apt-uuid-3",
      "token_number": 1,
      "mo_id": "MO2024100003",
      "patient_name": "Mohammed Khalil",
      "doctor_name": "Dr. Ahmed Ibrahim",
      "department": "General Medicine",
      "consultation_type": "online",
      "appointment_date": "2025-10-17",
      "appointment_time": "2025-10-17 10:00:00",
      "status": "confirmed",
      "fee_amount": 300.00,
      "payment_status": "paid",
      "booking_number": "BN202510170003",
      "reason": "Video consultation",
      "notes": null
    }
  ]
}
```

---

### Example 2: Get Appointments for Specific Date

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
  "appointments": [...]
}
```

---

## 📱 Flutter Integration

### 1. Model Class

```dart
class AppointmentListItem {
  final String id;
  final int? tokenNumber;
  final String? moId;
  final String patientName;
  final String doctorName;
  final String? department;
  final String consultationType;
  final String? appointmentDate;
  final String appointmentTime;
  final String status;
  final double? feeAmount;
  final String paymentStatus;
  final String bookingNumber;
  final String? reason;
  final String? notes;

  AppointmentListItem({
    required this.id,
    this.tokenNumber,
    this.moId,
    required this.patientName,
    required this.doctorName,
    this.department,
    required this.consultationType,
    this.appointmentDate,
    required this.appointmentTime,
    required this.status,
    this.feeAmount,
    required this.paymentStatus,
    required this.bookingNumber,
    this.reason,
    this.notes,
  });

  factory AppointmentListItem.fromJson(Map<String, dynamic> json) {
    return AppointmentListItem(
      id: json['id'],
      tokenNumber: json['token_number'],
      moId: json['mo_id'],
      patientName: json['patient_name'] ?? 'Unknown',
      doctorName: json['doctor_name'] ?? 'Unknown Doctor',
      department: json['department'],
      consultationType: json['consultation_type'],
      appointmentDate: json['appointment_date'],
      appointmentTime: json['appointment_time'],
      status: json['status'],
      feeAmount: json['fee_amount']?.toDouble(),
      paymentStatus: json['payment_status'],
      bookingNumber: json['booking_number'],
      reason: json['reason'],
      notes: json['notes'],
    );
  }
}
```

---

### 2. Service Class

```dart
class AppointmentService {
  final String baseUrl = 'http://localhost:8082/api';
  final String token;

  AppointmentService(this.token);

  Future<List<AppointmentListItem>> getSimpleAppointmentList({
    required String clinicId,
    String? date,
  }) async {
    String url = '$baseUrl/appointments/simple-list?clinic_id=$clinicId';
    
    if (date != null && date.isNotEmpty) {
      url += '&date=$date';
    }

    final response = await http.get(
      Uri.parse(url),
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      final List appointments = data['appointments'] ?? [];
      
      return appointments
          .map((json) => AppointmentListItem.fromJson(json))
          .toList();
    } else {
      throw Exception('Failed to load appointments');
    }
  }
}
```

---

### 3. DataTable Widget

```dart
class AppointmentListTable extends StatefulWidget {
  final String clinicId;
  final String? selectedDate;

  const AppointmentListTable({
    required this.clinicId,
    this.selectedDate,
  });

  @override
  _AppointmentListTableState createState() => _AppointmentListTableState();
}

class _AppointmentListTableState extends State<AppointmentListTable> {
  List<AppointmentListItem> appointments = [];
  bool isLoading = true;

  @override
  void initState() {
    super.initState();
    loadAppointments();
  }

  Future<void> loadAppointments() async {
    setState(() => isLoading = true);
    
    try {
      final service = AppointmentService(yourToken);
      final data = await service.getSimpleAppointmentList(
        clinicId: widget.clinicId,
        date: widget.selectedDate,
      );
      
      setState(() {
        appointments = data;
        isLoading = false;
      });
    } catch (e) {
      setState(() => isLoading = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error: $e')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    if (isLoading) {
      return Center(child: CircularProgressIndicator());
    }

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: DataTable(
        columns: [
          DataColumn(label: Text('Token')),
          DataColumn(label: Text('Mo ID')),
          DataColumn(label: Text('Patient Name')),
          DataColumn(label: Text('Doctor\'s Name')),
          DataColumn(label: Text('Department')),
          DataColumn(label: Text('Consultation Type')),
          DataColumn(label: Text('Appointment Date & Time')),
          DataColumn(label: Text('STATUS')),
          DataColumn(label: Text('Fee Status')),
        ],
        rows: appointments.map((appointment) {
          return DataRow(
            cells: [
              // Token
              DataCell(Text(
                appointment.tokenNumber?.toString() ?? '-',
                style: TextStyle(fontWeight: FontWeight.bold),
              )),
              
              // Mo ID
              DataCell(Text(appointment.moId ?? '-')),
              
              // Patient Name
              DataCell(Text(appointment.patientName)),
              
              // Doctor's Name
              DataCell(Text(appointment.doctorName)),
              
              // Department
              DataCell(Text(appointment.department ?? '-')),
              
              // Consultation Type
              DataCell(
                Container(
                  padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: appointment.consultationType == 'online'
                        ? Colors.blue.shade100
                        : Colors.green.shade100,
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    appointment.consultationType.toUpperCase(),
                    style: TextStyle(fontSize: 12),
                  ),
                ),
              ),
              
              // Appointment Date & Time
              DataCell(
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      appointment.appointmentDate ?? '-',
                      style: TextStyle(fontWeight: FontWeight.w500),
                    ),
                    Text(
                      appointment.appointmentTime.split(' ')[1],
                      style: TextStyle(fontSize: 12, color: Colors.grey),
                    ),
                  ],
                ),
              ),
              
              // Status
              DataCell(
                _buildStatusChip(appointment.status),
              ),
              
              // Fee Status
              DataCell(
                _buildPaymentStatusChip(appointment.paymentStatus),
              ),
            ],
          );
        }).toList(),
      ),
    );
  }

  Widget _buildStatusChip(String status) {
    Color color;
    switch (status.toLowerCase()) {
      case 'confirmed':
        color = Colors.green;
        break;
      case 'completed':
        color = Colors.blue;
        break;
      case 'cancelled':
        color = Colors.red;
        break;
      case 'no_show':
        color = Colors.orange;
        break;
      default:
        color = Colors.grey;
    }

    return Container(
      padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        border: Border.all(color: color),
        borderRadius: BorderRadius.circular(4),
      ),
      child: Text(
        status.toUpperCase(),
        style: TextStyle(color: color, fontSize: 12, fontWeight: FontWeight.bold),
      ),
    );
  }

  Widget _buildPaymentStatusChip(String paymentStatus) {
    Color color;
    String text;
    
    switch (paymentStatus.toLowerCase()) {
      case 'paid':
        color = Colors.green;
        text = 'PAID';
        break;
      case 'pending':
        color = Colors.orange;
        text = 'PENDING';
        break;
      case 'waived':
        color = Colors.blue;
        text = 'WAIVED';
        break;
      default:
        color = Colors.grey;
        text = paymentStatus.toUpperCase();
    }

    return Container(
      padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        border: Border.all(color: color),
        borderRadius: BorderRadius.circular(4),
      ),
      child: Text(
        text,
        style: TextStyle(color: color, fontSize: 12, fontWeight: FontWeight.bold),
      ),
    );
  }
}
```

---

### 4. Usage with Date Filter

```dart
class AppointmentListScreen extends StatefulWidget {
  @override
  _AppointmentListScreenState createState() => _AppointmentListScreenState();
}

class _AppointmentListScreenState extends State<AppointmentListScreen> {
  String? selectedDate;
  final clinicId = 'your-clinic-id';

  Future<void> _selectDate(BuildContext context) async {
    final DateTime? picked = await showDatePicker(
      context: context,
      initialDate: DateTime.now(),
      firstDate: DateTime(2024),
      lastDate: DateTime(2026),
    );
    
    if (picked != null) {
      setState(() {
        selectedDate = picked.toString().split(' ')[0]; // YYYY-MM-DD
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Appointments'),
        actions: [
          IconButton(
            icon: Icon(Icons.calendar_today),
            onPressed: () => _selectDate(context),
          ),
          if (selectedDate != null)
            IconButton(
              icon: Icon(Icons.clear),
              onPressed: () => setState(() => selectedDate = null),
            ),
        ],
      ),
      body: Column(
        children: [
          if (selectedDate != null)
            Padding(
              padding: EdgeInsets.all(8.0),
              child: Chip(
                label: Text('Filtered: $selectedDate'),
                onDeleted: () => setState(() => selectedDate = null),
              ),
            ),
          Expanded(
            child: AppointmentListTable(
              clinicId: clinicId,
              selectedDate: selectedDate,
            ),
          ),
        ],
      ),
    );
  }
}
```

---

## 🎨 Status Colors

### Appointment Status

| Status | Color | Display |
|--------|-------|---------|
| `confirmed` | 🟢 Green | CONFIRMED |
| `completed` | 🔵 Blue | COMPLETED |
| `cancelled` | 🔴 Red | CANCELLED |
| `no_show` | 🟠 Orange | NO SHOW |

### Payment Status

| Status | Color | Display |
|--------|-------|---------|
| `paid` | 🟢 Green | PAID |
| `pending` | 🟠 Orange | PENDING |
| `waived` | 🔵 Blue | WAIVED |

---

## 📊 Table Example

```
┌───────┬─────────────┬──────────────┬─────────────────┬─────────────┬──────────────────┬────────────────────┬───────────┬────────────┐
│ Token │ Mo ID       │ Patient Name │ Doctor's Name   │ Department  │ Consultation Type│ Date & Time        │ STATUS    │ Fee Status │
├───────┼─────────────┼──────────────┼─────────────────┼─────────────┼──────────────────┼────────────────────┼───────────┼────────────┤
│   1   │ MO2024100001│ Ahmed Ali    │ Dr. Sara Ahmed  │ Cardiology  │ OFFLINE          │ 2025-10-17 09:00  │ CONFIRMED │ PAID       │
│   2   │ MO2024100002│ Fatima Hassan│ Dr. Sara Ahmed  │ Cardiology  │ OFFLINE          │ 2025-10-17 09:05  │ CONFIRMED │ PENDING    │
│   1   │ MO2024100003│ Mohammed K.  │ Dr. Ahmed I.    │ General Med │ ONLINE           │ 2025-10-17 10:00  │ CONFIRMED │ PAID       │
└───────┴─────────────┴──────────────┴─────────────────┴─────────────┴──────────────────┴────────────────────┴───────────┴────────────┘
```

---

## ✅ Features

| Feature | Status |
|---------|--------|
| Filter by clinic | ✅ Required |
| Filter by date | ✅ Optional |
| Token number | ✅ Included |
| MO ID | ✅ Included |
| Patient name | ✅ Included |
| Doctor name | ✅ Included |
| Department | ✅ Included |
| Consultation type | ✅ Included |
| Date & Time | ✅ Included |
| Appointment status | ✅ Included |
| Payment status | ✅ Included |
| Sorted by date/time | ✅ DESC order |

---

## 🔍 Error Handling

### Error 1: Missing clinic_id

**Response (400):**
```json
{
  "error": "clinic_id is required"
}
```

### Error 2: Invalid date format

**Response (500):**
```json
{
  "error": "Failed to fetch appointments",
  "details": "parsing time \"invalid-date\""
}
```

---

## 📋 Quick Reference

**Endpoint:** `GET /api/appointments/simple-list`

**Required:** `clinic_id`

**Optional:** `date` (YYYY-MM-DD format)

**Returns:** List of appointments with all display fields

**Status:** ✅ Ready for Flutter DataTable integration!

---

**Perfect for building appointment management dashboards!** 📊✨

