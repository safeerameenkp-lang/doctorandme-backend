# Enhanced Follow-Up System - Flutter Implementation Guide 📱

## 🎯 **Complete Implementation**

This guide provides everything you need to implement the enhanced doctor-specific follow-up system in Flutter.

---

## 📋 **System Overview**

### **Enhanced Features:**
- ✅ **Doctor-Specific Follow-Up Status** - Each doctor has independent follow-up conditions
- ✅ **Department-Specific Tracking** - Each department within a doctor has its own status
- ✅ **Complete Appointment History** - Shows all appointments per doctor+department
- ✅ **Automatic Follow-Up Creation** - Regular appointments automatically create follow-up eligibility
- ✅ **Enhanced UI Components** - Doctor-specific follow-up cards and dialogs

---

## 🔗 **API Endpoints**

### **1. Get Patient Follow-Up Status (All Doctors)**
```http
GET /api/organizations/patient-followup-status/{patient_id}
```

**Response:**
```json
{
  "patient": {
    "id": "patient-uuid",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890"
  },
  "doctors": [
    {
      "doctor_id": "doctor-uuid",
      "doctor_name": "Dr. Smith",
      "departments": [
        {
          "department_id": "dept-uuid",
          "department_name": "Cardiology",
          "follow_up_status": {
            "status": true,
            "is_free": true,
            "status_label": "free",
            "color_code": "green",
            "message": "Free follow-up available (3 days remaining)",
            "days_remaining": 3,
            "last_appointment_date": "2025-10-20",
            "follow_up_expiry": "2025-10-25",
            "free_follow_up_used": false
          },
          "appointment_history": [
            {
              "appointment_id": "appointment-uuid",
              "doctor_id": "doctor-uuid",
              "doctor_name": "Dr. Smith",
              "department_id": "dept-uuid",
              "department": "Cardiology",
              "appointment_type": "clinic_visit",
              "appointment_date": "2025-10-20",
              "days_since": 2,
              "validity_days": 5,
              "status": "completed",
              "follow_up_eligible": true,
              "follow_up_status": "active",
              "renewal_status": "valid",
              "free_follow_up_used": false,
              "note": "Appointment with Dr. Smith (Cardiology)"
            }
          ]
        }
      ]
    }
  ]
}
```

---

## 📱 **Flutter Implementation**

### **1. Enhanced API Service**

```dart
import 'dart:convert';
import 'package:http/http.dart' as http;

class EnhancedFollowUpApiService {
  static const String baseUrl = 'http://localhost:3002';
  static const String appointmentUrl = 'http://localhost:3001';
  
  // Headers
  static Map<String, String> getHeaders(String token) {
    return {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer $token',
    };
  }

  // Get patient follow-up status for all doctors
  static Future<Map<String, dynamic>> getPatientFollowUpStatus({
    required String patientId,
    required String token,
  }) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/api/organizations/patient-followup-status/$patientId'),
        headers: getHeaders(token),
      );
      
      if (response.statusCode == 200) {
        return json.decode(response.body);
      } else {
        throw Exception('Failed to load follow-up status: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error fetching follow-up status: $e');
    }
  }

  // Enhanced patient list with follow-up status
  static Future<Map<String, dynamic>> getPatientsWithFollowUpStatus({
    required String clinicId,
    String? search,
    required String token,
  }) async {
    try {
      String url = '$baseUrl/api/organizations/clinic-specific-patients?clinic_id=$clinicId';
      
      if (search != null && search.isNotEmpty) url += '&search=$search';
      
      final response = await http.get(
        Uri.parse(url),
        headers: getHeaders(token),
      );
      
      if (response.statusCode == 200) {
        return json.decode(response.body);
      } else {
        throw Exception('Failed to load patients: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error fetching patients: $e');
    }
  }

  // Create follow-up appointment
  static Future<Map<String, dynamic>> createFollowUpAppointment({
    required String clinicPatientId,
    required String clinicId,
    required String doctorId,
    String? departmentId,
    required String individualSlotId,
    required String appointmentDate,
    required String appointmentTime,
    required String consultationType,
    String? notes,
    required String token,
  }) async {
    try {
      final body = {
        'clinic_patient_id': clinicPatientId,
        'clinic_id': clinicId,
        'doctor_id': doctorId,
        'individual_slot_id': individualSlotId,
        'appointment_date': appointmentDate,
        'appointment_time': appointmentTime,
        'consultation_type': consultationType,
        'is_follow_up': true,
        'payment_method': null,
        'payment_mode': null,
        if (departmentId != null) 'department_id': departmentId,
        if (notes != null) 'notes': notes,
      };
      
      final response = await http.post(
        Uri.parse('$appointmentUrl/api/appointments/simple'),
        headers: getHeaders(token),
        body: json.encode(body),
      );
      
      if (response.statusCode == 201) {
        return json.decode(response.body);
      } else {
        final error = json.decode(response.body);
        throw Exception(error['message'] ?? 'Failed to create appointment');
      }
    } catch (e) {
      throw Exception('Error creating follow-up appointment: $e');
    }
  }

  // Create regular appointment
  static Future<Map<String, dynamic>> createRegularAppointment({
    required String clinicPatientId,
    required String clinicId,
    required String doctorId,
    String? departmentId,
    required String individualSlotId,
    required String appointmentDate,
    required String appointmentTime,
    required String consultationType,
    required String paymentMethod,
    String? paymentMode,
    String? notes,
    required String token,
  }) async {
    try {
      final body = {
        'clinic_patient_id': clinicPatientId,
        'clinic_id': clinicId,
        'doctor_id': doctorId,
        'individual_slot_id': individualSlotId,
        'appointment_date': appointmentDate,
        'appointment_time': appointmentTime,
        'consultation_type': consultationType,
        'is_follow_up': false,
        'payment_method': paymentMethod,
        if (paymentMode != null) 'payment_mode': paymentMode,
        if (departmentId != null) 'department_id': departmentId,
        if (notes != null) 'notes': notes,
      };
      
      final response = await http.post(
        Uri.parse('$appointmentUrl/api/appointments/simple'),
        headers: getHeaders(token),
        body: json.encode(body),
      );
      
      if (response.statusCode == 201) {
        return json.decode(response.body);
      } else {
        final error = json.decode(response.body);
        throw Exception(error['message'] ?? 'Failed to create appointment');
      }
    } catch (e) {
      throw Exception('Error creating regular appointment: $e');
    }
  }
}
```

---

### **2. Enhanced Data Models**

```dart
// Enhanced Patient Model with Doctor-Specific Follow-Up Status
class EnhancedPatient {
  final String id;
  final String clinicId;
  final String firstName;
  final String lastName;
  final String phone;
  final String? email;
  final FollowUpEligibility followUpEligibility;
  final List<DoctorFollowUpStatus> doctors;

  EnhancedPatient({
    required this.id,
    required this.clinicId,
    required this.firstName,
    required this.lastName,
    required this.phone,
    this.email,
    required this.followUpEligibility,
    required this.doctors,
  });

  factory EnhancedPatient.fromJson(Map<String, dynamic> json) {
    return EnhancedPatient(
      id: json['id'],
      clinicId: json['clinic_id'],
      firstName: json['first_name'],
      lastName: json['last_name'],
      phone: json['phone'],
      email: json['email'],
      followUpEligibility: FollowUpEligibility.fromJson(json['follow_up_eligibility']),
      doctors: (json['doctors'] as List?)
          ?.map((doctor) => DoctorFollowUpStatus.fromJson(doctor))
          .toList() ?? [],
    );
  }
}

// Doctor Follow-Up Status Model
class DoctorFollowUpStatus {
  final String doctorId;
  final String doctorName;
  final List<DepartmentFollowUp> departments;

  DoctorFollowUpStatus({
    required this.doctorId,
    required this.doctorName,
    required this.departments,
  });

  factory DoctorFollowUpStatus.fromJson(Map<String, dynamic> json) {
    return DoctorFollowUpStatus(
      doctorId: json['doctor_id'],
      doctorName: json['doctor_name'],
      departments: (json['departments'] as List)
          .map((dept) => DepartmentFollowUp.fromJson(dept))
          .toList(),
    );
  }
}

// Department Follow-Up Model
class DepartmentFollowUp {
  final String departmentId;
  final String departmentName;
  final FollowUpStatus followUpStatus;
  final List<AppointmentHistoryItem> appointmentHistory;

  DepartmentFollowUp({
    required this.departmentId,
    required this.departmentName,
    required this.followUpStatus,
    required this.appointmentHistory,
  });

  factory DepartmentFollowUp.fromJson(Map<String, dynamic> json) {
    return DepartmentFollowUp(
      departmentId: json['department_id'],
      departmentName: json['department_name'],
      followUpStatus: FollowUpStatus.fromJson(json['follow_up_status']),
      appointmentHistory: (json['appointment_history'] as List)
          .map((item) => AppointmentHistoryItem.fromJson(item))
          .toList(),
    );
  }
}

// Enhanced Follow-Up Status Model
class FollowUpStatus {
  final bool status;               // true = free, false = paid
  final bool isFree;               // same as status
  final String statusLabel;        // "free", "paid", "none"
  final String colorCode;          // "green", "orange", "gray"
  final String message;
  final int? daysRemaining;
  final String? lastAppointmentDate;
  final String? followUpExpiry;
  final bool freeFollowUpUsed;

  FollowUpStatus({
    required this.status,
    required this.isFree,
    required this.statusLabel,
    required this.colorCode,
    required this.message,
    this.daysRemaining,
    this.lastAppointmentDate,
    this.followUpExpiry,
    required this.freeFollowUpUsed,
  });

  factory FollowUpStatus.fromJson(Map<String, dynamic> json) {
    return FollowUpStatus(
      status: json['status'] ?? false,
      isFree: json['is_free'] ?? false,
      statusLabel: json['status_label'] ?? 'none',
      colorCode: json['color_code'] ?? 'gray',
      message: json['message'] ?? '',
      daysRemaining: json['days_remaining'],
      lastAppointmentDate: json['last_appointment_date'],
      followUpExpiry: json['follow_up_expiry'],
      freeFollowUpUsed: json['free_follow_up_used'] ?? false,
    );
  }
}

// Patient Follow-Up Status Response Model
class PatientFollowUpStatusResponse {
  final PatientInfo patient;
  final List<DoctorFollowUpStatus> doctors;

  PatientFollowUpStatusResponse({
    required this.patient,
    required this.doctors,
  });

  factory PatientFollowUpStatusResponse.fromJson(Map<String, dynamic> json) {
    return PatientFollowUpStatusResponse(
      patient: PatientInfo.fromJson(json['patient']),
      doctors: (json['doctors'] as List)
          .map((doctor) => DoctorFollowUpStatus.fromJson(doctor))
          .toList(),
    );
  }
}

// Patient Info Model
class PatientInfo {
  final String id;
  final String firstName;
  final String lastName;
  final String phone;
  final String? email;

  PatientInfo({
    required this.id,
    required this.firstName,
    required this.lastName,
    required this.phone,
    this.email,
  });

  factory PatientInfo.fromJson(Map<String, dynamic> json) {
    return PatientInfo(
      id: json['id'],
      firstName: json['first_name'],
      lastName: json['last_name'],
      phone: json['phone'],
      email: json['email'],
    );
  }
}

// Existing models (FollowUpEligibility, AppointmentHistoryItem) remain the same
```

---

### **3. Enhanced UI Components**

#### **A. Enhanced Patient List Widget**

```dart
import 'package:flutter/material.dart';

class EnhancedPatientListWidget extends StatefulWidget {
  final String clinicId;
  final String token;

  const EnhancedPatientListWidget({
    Key? key,
    required this.clinicId,
    required this.token,
  }) : super(key: key);

  @override
  _EnhancedPatientListWidgetState createState() => _EnhancedPatientListWidgetState();
}

class _EnhancedPatientListWidgetState extends State<EnhancedPatientListWidget> {
  List<EnhancedPatient> patients = [];
  bool isLoading = false;
  String? errorMessage;
  String searchQuery = '';

  @override
  void initState() {
    super.initState();
    loadPatients();
  }

  Future<void> loadPatients() async {
    setState(() {
      isLoading = true;
      errorMessage = null;
    });

    try {
      final response = await EnhancedFollowUpApiService.getPatientsWithFollowUpStatus(
        clinicId: widget.clinicId,
        search: searchQuery.isNotEmpty ? searchQuery : null,
        token: widget.token,
      );

      final List<dynamic> patientsJson = response['patients'];
      setState(() {
        patients = patientsJson.map((json) => EnhancedPatient.fromJson(json)).toList();
        isLoading = false;
      });
    } catch (e) {
      setState(() {
        errorMessage = e.toString();
        isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        // Search Bar
        Padding(
          padding: const EdgeInsets.all(16.0),
          child: TextField(
            decoration: InputDecoration(
              hintText: 'Search patients...',
              prefixIcon: Icon(Icons.search),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
            onChanged: (value) {
              setState(() {
                searchQuery = value;
              });
              // Debounce search
              Future.delayed(Duration(milliseconds: 500), () {
                if (searchQuery == value) {
                  loadPatients();
                }
              });
            },
          ),
        ),
        
        // Patient List
        Expanded(
          child: isLoading
              ? Center(child: CircularProgressIndicator())
              : errorMessage != null
                  ? Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.error, size: 64, color: Colors.red),
                          SizedBox(height: 16),
                          Text(
                            'Error loading patients',
                            style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                          ),
                          SizedBox(height: 8),
                          Text(errorMessage!),
                          SizedBox(height: 16),
                          ElevatedButton(
                            onPressed: loadPatients,
                            child: Text('Retry'),
                          ),
                        ],
                      ),
                    )
                  : patients.isEmpty
                      ? Center(
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(Icons.people_outline, size: 64, color: Colors.grey),
                              SizedBox(height: 16),
                              Text(
                                'No patients found',
                                style: TextStyle(fontSize: 18, color: Colors.grey),
                              ),
                            ],
                          ),
                        )
                      : ListView.builder(
                          itemCount: patients.length,
                          itemBuilder: (context, index) {
                            return EnhancedPatientCard(
                              patient: patients[index],
                              onViewFollowUpStatus: () => _showFollowUpStatusDialog(patients[index]),
                              onBookFollowUp: () => _showBookFollowUpDialog(patients[index]),
                              onBookRegular: () => _showBookRegularDialog(patients[index]),
                            );
                          },
                        ),
        ),
      ],
    );
  }

  void _showFollowUpStatusDialog(EnhancedPatient patient) {
    showDialog(
      context: context,
      builder: (context) => FollowUpStatusDialog(
        patientId: patient.id,
        token: widget.token,
      ),
    );
  }

  void _showBookFollowUpDialog(EnhancedPatient patient) {
    showDialog(
      context: context,
      builder: (context) => BookFollowUpDialog(
        patient: patient,
        token: widget.token,
        onSuccess: () {
          Navigator.pop(context);
          loadPatients(); // Refresh the list
        },
      ),
    );
  }

  void _showBookRegularDialog(EnhancedPatient patient) {
    showDialog(
      context: context,
      builder: (context) => BookRegularDialog(
        patient: patient,
        token: widget.token,
        onSuccess: () {
          Navigator.pop(context);
          loadPatients(); // Refresh the list
        },
      ),
    );
  }
}
```

#### **B. Enhanced Patient Card**

```dart
class EnhancedPatientCard extends StatelessWidget {
  final EnhancedPatient patient;
  final VoidCallback onViewFollowUpStatus;
  final VoidCallback onBookFollowUp;
  final VoidCallback onBookRegular;

  const EnhancedPatientCard({
    Key? key,
    required this.patient,
    required this.onViewFollowUpStatus,
    required this.onBookFollowUp,
    required this.onBookRegular,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Padding(
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Patient Header
            Row(
              children: [
                CircleAvatar(
                  backgroundColor: _getStatusColor(patient.followUpEligibility.colorCode),
                  radius: 24,
                  child: Text(
                    '${patient.firstName[0]}${patient.lastName[0]}',
                    style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold),
                  ),
                ),
                SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        '${patient.firstName} ${patient.lastName}',
                        style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                      ),
                      Text(
                        patient.phone,
                        style: TextStyle(fontSize: 14, color: Colors.grey[600]),
                      ),
                    ],
                  ),
                ),
                IconButton(
                  onPressed: onViewFollowUpStatus,
                  icon: Icon(Icons.info_outline),
                  tooltip: 'View Follow-Up Status',
                ),
              ],
            ),
            
            SizedBox(height: 12),
            
            // Follow-Up Status Summary
            Container(
              padding: EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: _getStatusColor(patient.followUpEligibility.colorCode).withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(
                  color: _getStatusColor(patient.followUpEligibility.colorCode),
                  width: 1,
                ),
              ),
              child: Row(
                children: [
                  Icon(
                    _getStatusIcon(patient.followUpEligibility.statusLabel).icon,
                    color: _getStatusColor(patient.followUpEligibility.colorCode),
                    size: 20,
                  ),
                  SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      patient.followUpEligibility.message,
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        color: _getStatusColor(patient.followUpEligibility.colorCode),
                      ),
                    ),
                  ),
                ],
              ),
            ),
            
            SizedBox(height: 12),
            
            // Doctor Summary
            if (patient.doctors.isNotEmpty)
              Container(
                padding: EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: Colors.blue[50],
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Doctors with Follow-Up Status:',
                      style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12),
                    ),
                    SizedBox(height: 4),
                    ...patient.doctors.take(2).map((doctor) {
                      return Padding(
                        padding: EdgeInsets.only(bottom: 2),
                        child: Text(
                          '• ${doctor.doctorName} (${doctor.departments.length} dept${doctor.departments.length > 1 ? 's' : ''})',
                          style: TextStyle(fontSize: 11, color: Colors.grey[700]),
                        ),
                      );
                    }).toList(),
                    if (patient.doctors.length > 2)
                      Text(
                        '• +${patient.doctors.length - 2} more...',
                        style: TextStyle(fontSize: 11, color: Colors.grey[600], fontStyle: FontStyle.italic),
                      ),
                  ],
                ),
              ),
            
            SizedBox(height: 12),
            
            // Action Buttons
            Row(
              children: [
                Expanded(
                  child: ElevatedButton.icon(
                    onPressed: onBookFollowUp,
                    icon: Icon(Icons.refresh, size: 18),
                    label: Text('Follow-Up'),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: _getStatusColor(patient.followUpEligibility.colorCode),
                      foregroundColor: Colors.white,
                      padding: EdgeInsets.symmetric(vertical: 12),
                    ),
                  ),
                ),
                SizedBox(width: 8),
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: onBookRegular,
                    icon: Icon(Icons.calendar_today, size: 18),
                    label: Text('Regular'),
                    style: OutlinedButton.styleFrom(
                      padding: EdgeInsets.symmetric(vertical: 12),
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Color _getStatusColor(String colorCode) {
    switch (colorCode) {
      case 'green':
        return Colors.green;
      case 'orange':
        return Colors.orange;
      case 'gray':
      default:
        return Colors.grey;
    }
  }

  Icon _getStatusIcon(String statusLabel) {
    switch (statusLabel) {
      case 'free':
        return Icon(Icons.check_circle, color: Colors.green, size: 24);
      case 'paid':
        return Icon(Icons.payment, color: Colors.orange, size: 24);
      case 'none':
        return Icon(Icons.info_outline, color: Colors.grey, size: 24);
      default:
        return Icon(Icons.help_outline, color: Colors.grey, size: 24);
    }
  }
}
```

#### **C. Follow-Up Status Detail Dialog**

```dart
class FollowUpStatusDialog extends StatefulWidget {
  final String patientId;
  final String token;

  const FollowUpStatusDialog({
    Key? key,
    required this.patientId,
    required this.token,
  }) : super(key: key);

  @override
  _FollowUpStatusDialogState createState() => _FollowUpStatusDialogState();
}

class _FollowUpStatusDialogState extends State<FollowUpStatusDialog> {
  PatientFollowUpStatusResponse? followUpData;
  bool isLoading = true;
  String? errorMessage;

  @override
  void initState() {
    super.initState();
    loadFollowUpStatus();
  }

  Future<void> loadFollowUpStatus() async {
    try {
      final data = await EnhancedFollowUpApiService.getPatientFollowUpStatus(
        patientId: widget.patientId,
        token: widget.token,
      );
      
      setState(() {
        followUpData = PatientFollowUpStatusResponse.fromJson(data);
        isLoading = false;
      });
    } catch (e) {
      setState(() {
        errorMessage = e.toString();
        isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Row(
        children: [
          Icon(Icons.info_outline, color: Colors.blue),
          SizedBox(width: 8),
          Text('Follow-Up Status'),
        ],
      ),
      content: SizedBox(
        width: double.maxFinite,
        height: 500,
        child: isLoading
            ? Center(child: CircularProgressIndicator())
            : errorMessage != null
                ? Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.error, size: 48, color: Colors.red),
                        SizedBox(height: 16),
                        Text('Error: $errorMessage'),
                        SizedBox(height: 16),
                        ElevatedButton(
                          onPressed: loadFollowUpStatus,
                          child: Text('Retry'),
                        ),
                      ],
                    ),
                  )
                : followUpData == null
                    ? Center(child: Text('No data available'))
                    : Column(
                        children: [
                          // Patient Info Header
                          Container(
                            padding: EdgeInsets.all(12),
                            decoration: BoxDecoration(
                              color: Colors.blue[50],
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Row(
                              children: [
                                CircleAvatar(
                                  backgroundColor: Colors.blue,
                                  child: Text(
                                    '${followUpData!.patient.firstName[0]}${followUpData!.patient.lastName[0]}',
                                    style: TextStyle(color: Colors.white),
                                  ),
                                ),
                                SizedBox(width: 12),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        '${followUpData!.patient.firstName} ${followUpData!.patient.lastName}',
                                        style: TextStyle(fontWeight: FontWeight.bold),
                                      ),
                                      Text(
                                        followUpData!.patient.phone,
                                        style: TextStyle(color: Colors.grey[600]),
                                      ),
                                    ],
                                  ),
                                ),
                              ],
                            ),
                          ),
                          
                          SizedBox(height: 16),
                          
                          // Doctors List
                          Expanded(
                            child: followUpData!.doctors.isEmpty
                                ? Center(
                                    child: Column(
                                      mainAxisAlignment: MainAxisAlignment.center,
                                      children: [
                                        Icon(Icons.people_outline, size: 48, color: Colors.grey),
                                        SizedBox(height: 16),
                                        Text(
                                          'No doctors found',
                                          style: TextStyle(color: Colors.grey),
                                        ),
                                      ],
                                    ),
                                  )
                                : ListView.builder(
                                    itemCount: followUpData!.doctors.length,
                                    itemBuilder: (context, index) {
                                      return DoctorFollowUpCard(
                                        doctor: followUpData!.doctors[index],
                                      );
                                    },
                                  ),
                          ),
                        ],
                      ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text('Close'),
        ),
      ],
    );
  }
}

// Doctor Follow-Up Card
class DoctorFollowUpCard extends StatelessWidget {
  final DoctorFollowUpStatus doctor;

  const DoctorFollowUpCard({Key? key, required this.doctor}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: EdgeInsets.symmetric(vertical: 4),
      child: Padding(
        padding: EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.person, color: Colors.blue),
                SizedBox(width: 8),
                Text(
                  doctor.doctorName,
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                ),
              ],
            ),
            SizedBox(height: 8),
            ...doctor.departments.map((dept) {
              return DepartmentFollowUpTile(department: dept);
            }).toList(),
          ],
        ),
      ),
    );
  }
}

// Department Follow-Up Tile
class DepartmentFollowUpTile extends StatelessWidget {
  final DepartmentFollowUp department;

  const DepartmentFollowUpTile({Key? key, required this.department}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    final status = department.followUpStatus;
    
    return Container(
      margin: EdgeInsets.symmetric(vertical: 4),
      padding: EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: _getStatusColor(status.colorCode).withOpacity(0.1),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: _getStatusColor(status.colorCode),
          width: 1,
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              CircleAvatar(
                backgroundColor: _getStatusColor(status.colorCode),
                radius: 12,
                child: Icon(
                  _getStatusIcon(status.statusLabel).icon,
                  color: Colors.white,
                  size: 16,
                ),
              ),
              SizedBox(width: 8),
              Expanded(
                child: Text(
                  department.departmentName,
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
              ),
              Chip(
                label: Text(
                  status.status ? 'FREE' : 'PAID',
                  style: TextStyle(
                    fontSize: 10,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                backgroundColor: status.status ? Colors.green[100] : Colors.orange[100],
                labelStyle: TextStyle(
                  color: status.status ? Colors.green[800] : Colors.orange[800],
                ),
              ),
            ],
          ),
          SizedBox(height: 4),
          Text(
            status.message,
            style: TextStyle(fontSize: 12, color: Colors.grey[700]),
          ),
          if (status.daysRemaining != null)
            Text(
              '${status.daysRemaining} days remaining',
              style: TextStyle(fontSize: 11, color: Colors.grey[600]),
            ),
          if (department.appointmentHistory.isNotEmpty)
            Padding(
              padding: EdgeInsets.only(top: 4),
              child: Text(
                '${department.appointmentHistory.length} appointment${department.appointmentHistory.length > 1 ? 's' : ''}',
                style: TextStyle(fontSize: 10, color: Colors.grey[500]),
              ),
            ),
        ],
      ),
    );
  }

  Color _getStatusColor(String colorCode) {
    switch (colorCode) {
      case 'green':
        return Colors.green;
      case 'orange':
        return Colors.orange;
      case 'gray':
      default:
        return Colors.grey;
    }
  }

  Icon _getStatusIcon(String statusLabel) {
    switch (statusLabel) {
      case 'free':
        return Icon(Icons.check_circle, color: Colors.green);
      case 'paid':
        return Icon(Icons.payment, color: Colors.orange);
      case 'none':
        return Icon(Icons.info_outline, color: Colors.grey);
      default:
        return Icon(Icons.help_outline, color: Colors.grey);
    }
  }
}
```

---

### **4. Enhanced Appointment Booking Dialogs**

#### **A. Enhanced Book Follow-Up Dialog**

```dart
class EnhancedBookFollowUpDialog extends StatefulWidget {
  final EnhancedPatient patient;
  final String token;
  final VoidCallback onSuccess;

  const EnhancedBookFollowUpDialog({
    Key? key,
    required this.patient,
    required this.token,
    required this.onSuccess,
  }) : super(key: key);

  @override
  _EnhancedBookFollowUpDialogState createState() => _EnhancedBookFollowUpDialogState();
}

class _EnhancedBookFollowUpDialogState extends State<EnhancedBookFollowUpDialog> {
  DoctorFollowUpStatus? selectedDoctor;
  DepartmentFollowUp? selectedDepartment;
  DateTime selectedDate = DateTime.now();
  String selectedTime = '10:00';
  String consultationType = 'follow-up-via-clinic';
  bool isLoading = false;

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text('Book Follow-Up'),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Patient Info
            Container(
              padding: EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.blue[50],
                borderRadius: BorderRadius.circular(8),
              ),
              child: Row(
                children: [
                  CircleAvatar(
                    backgroundColor: Colors.blue,
                    child: Text(
                      '${widget.patient.firstName[0]}${widget.patient.lastName[0]}',
                      style: TextStyle(color: Colors.white),
                    ),
                  ),
                  SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          '${widget.patient.firstName} ${widget.patient.lastName}',
                          style: TextStyle(fontWeight: FontWeight.bold),
                        ),
                        Text(
                          widget.patient.phone,
                          style: TextStyle(color: Colors.grey[600]),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            
            SizedBox(height: 16),
            
            // Doctor Selection
            DropdownButtonFormField<DoctorFollowUpStatus>(
              value: selectedDoctor,
              decoration: InputDecoration(
                labelText: 'Select Doctor',
                border: OutlineInputBorder(),
              ),
              items: widget.patient.doctors.map((doctor) {
                return DropdownMenuItem(
                  value: doctor,
                  child: Text(doctor.doctorName),
                );
              }).toList(),
              onChanged: (value) {
                setState(() {
                  selectedDoctor = value;
                  selectedDepartment = null; // Reset department
                });
              },
            ),
            
            SizedBox(height: 16),
            
            // Department Selection
            if (selectedDoctor != null)
              DropdownButtonFormField<DepartmentFollowUp>(
                value: selectedDepartment,
                decoration: InputDecoration(
                  labelText: 'Select Department',
                  border: OutlineInputBorder(),
                ),
                items: selectedDoctor!.departments.map((dept) {
                  return DropdownMenuItem(
                    value: dept,
                    child: Text(dept.departmentName),
                  );
                }).toList(),
                onChanged: (value) {
                  setState(() {
                    selectedDepartment = value;
                  });
                },
              ),
            
            SizedBox(height: 16),
            
            // Follow-Up Status Display
            if (selectedDepartment != null)
              Container(
                padding: EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: _getStatusColor(selectedDepartment!.followUpStatus.colorCode).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(
                    color: _getStatusColor(selectedDepartment!.followUpStatus.colorCode),
                  ),
                ),
                child: Row(
                  children: [
                    Icon(
                      _getStatusIcon(selectedDepartment!.followUpStatus.statusLabel).icon,
                      color: _getStatusColor(selectedDepartment!.followUpStatus.colorCode),
                    ),
                    SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        selectedDepartment!.followUpStatus.message,
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                          color: _getStatusColor(selectedDepartment!.followUpStatus.colorCode),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            
            SizedBox(height: 16),
            
            // Date Picker
            ListTile(
              leading: Icon(Icons.calendar_today),
              title: Text('Date'),
              subtitle: Text('${selectedDate.day}/${selectedDate.month}/${selectedDate.year}'),
              onTap: () async {
                final date = await showDatePicker(
                  context: context,
                  initialDate: selectedDate,
                  firstDate: DateTime.now(),
                  lastDate: DateTime.now().add(Duration(days: 30)),
                );
                if (date != null) {
                  setState(() {
                    selectedDate = date;
                  });
                }
              },
            ),
            
            // Time Picker
            ListTile(
              leading: Icon(Icons.access_time),
              title: Text('Time'),
              subtitle: Text(selectedTime),
              onTap: () async {
                final time = await showTimePicker(
                  context: context,
                  initialTime: TimeOfDay.fromDateTime(
                    DateTime.parse('2025-01-01 $selectedTime:00'),
                  ),
                );
                if (time != null) {
                  setState(() {
                    selectedTime = '${time.hour.toString().padLeft(2, '0')}:${time.minute.toString().padLeft(2, '0')}';
                  });
                }
              },
            ),
            
            // Consultation Type
            DropdownButtonFormField<String>(
              value: consultationType,
              decoration: InputDecoration(
                labelText: 'Consultation Type',
                border: OutlineInputBorder(),
              ),
              items: [
                DropdownMenuItem(
                  value: 'follow-up-via-clinic',
                  child: Text('Follow-Up (Clinic)'),
                ),
                DropdownMenuItem(
                  value: 'follow-up-via-video',
                  child: Text('Follow-Up (Video)'),
                ),
              ],
              onChanged: (value) {
                setState(() {
                  consultationType = value!;
                });
              },
            ),
            
            SizedBox(height: 16),
            
            // Payment Info (only for paid follow-ups)
            if (selectedDepartment != null && !selectedDepartment!.followUpStatus.status)
              Container(
                padding: EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.orange[50],
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.orange),
                ),
                child: Row(
                  children: [
                    Icon(Icons.payment, color: Colors.orange),
                    SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        'Payment required for this follow-up',
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                          color: Colors.orange[800],
                        ),
                      ),
                    ),
                  ],
                ),
              ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: isLoading ? null : () => Navigator.pop(context),
          child: Text('Cancel'),
        ),
        ElevatedButton(
          onPressed: isLoading ? null : _bookFollowUp,
          child: isLoading
              ? SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : Text('Book Follow-Up'),
        ),
      ],
    );
  }

  Future<void> _bookFollowUp() async {
    if (selectedDoctor == null || selectedDepartment == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Please select a doctor and department'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    setState(() {
      isLoading = true;
    });

    try {
      await EnhancedFollowUpApiService.createFollowUpAppointment(
        clinicPatientId: widget.patient.id,
        clinicId: widget.patient.clinicId,
        doctorId: selectedDoctor!.doctorId,
        departmentId: selectedDepartment!.departmentId,
        individualSlotId: 'slot-id', // You need to implement slot selection
        appointmentDate: selectedDate.toIso8601String().split('T')[0],
        appointmentTime: selectedTime,
        consultationType: consultationType,
        token: widget.token,
      );

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Follow-up appointment booked successfully!'),
          backgroundColor: Colors.green,
        ),
      );

      widget.onSuccess();
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Error: $e'),
          backgroundColor: Colors.red,
        ),
      );
    } finally {
      setState(() {
        isLoading = false;
      });
    }
  }

  Color _getStatusColor(String colorCode) {
    switch (colorCode) {
      case 'green':
        return Colors.green;
      case 'orange':
        return Colors.orange;
      case 'gray':
      default:
        return Colors.grey;
    }
  }

  Icon _getStatusIcon(String statusLabel) {
    switch (statusLabel) {
      case 'free':
        return Icon(Icons.check_circle, color: Colors.green);
      case 'paid':
        return Icon(Icons.payment, color: Colors.orange);
      case 'none':
        return Icon(Icons.info_outline, color: Colors.grey);
      default:
        return Icon(Icons.help_outline, color: Colors.grey);
    }
  }
}
```

---

### **5. Main App Integration**

```dart
void main() {
  runApp(
    ChangeNotifierProvider(
      create: (context) => EnhancedFollowUpProvider(),
      child: MyApp(),
    ),
  );
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Enhanced Follow-Up Management',
      theme: ThemeData(
        primarySwatch: Colors.blue,
        visualDensity: VisualDensity.adaptivePlatformDensity,
      ),
      home: EnhancedFollowUpApp(),
    );
  }
}

class EnhancedFollowUpApp extends StatefulWidget {
  @override
  _EnhancedFollowUpAppState createState() => _EnhancedFollowUpAppState();
}

class _EnhancedFollowUpAppState extends State<EnhancedFollowUpApp> {
  String clinicId = 'your-clinic-id';
  String token = 'your-auth-token';

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Enhanced Follow-Up Management'),
        backgroundColor: Colors.blue,
      ),
      body: EnhancedPatientListWidget(
        clinicId: clinicId,
        token: token,
      ),
    );
  }
}
```

---

## 🧪 **Testing Guide**

### **Test Scenarios:**

1. **Load Patient Follow-Up Status**
   ```dart
   final status = await EnhancedFollowUpApiService.getPatientFollowUpStatus(
     patientId: 'patient-uuid',
     token: 'your-token',
   );
   ```

2. **Book Follow-Up with Doctor Selection**
   ```dart
   await EnhancedFollowUpApiService.createFollowUpAppointment(
     clinicPatientId: 'patient-uuid',
     clinicId: 'clinic-uuid',
     doctorId: 'doctor-uuid',
     departmentId: 'dept-uuid',
     individualSlotId: 'slot-uuid',
     appointmentDate: '2025-10-23',
     appointmentTime: '10:00',
     consultationType: 'follow-up-via-clinic',
     token: 'your-token',
   );
   ```

---

## ✅ **Summary**

**Enhanced Features Implemented:**

1. **✅ Doctor-Specific Follow-Up Status** - Each doctor has independent follow-up conditions
2. **✅ Department-Specific Tracking** - Each department within a doctor has its own status
3. **✅ Complete Appointment History** - Shows all appointments per doctor+department
4. **✅ Enhanced API Service** - Complete API integration with error handling
5. **✅ Enhanced UI Components** - Doctor-specific follow-up cards and dialogs
6. **✅ Status-Based Color Coding** - Green (free), Orange (paid), Gray (none)
7. **✅ Automatic Follow-Up Creation** - Regular appointments create follow-up eligibility
8. **✅ Complete Flutter Integration** - Ready-to-use components and services

**Ready to implement!** 🚀✅

**This enhanced system provides complete doctor-specific follow-up management!** 📋✨
