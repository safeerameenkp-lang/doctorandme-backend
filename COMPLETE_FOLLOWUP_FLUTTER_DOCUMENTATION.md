# Complete Follow-Up System - Flutter API & UI Documentation 📱

## 🎯 **Complete System Overview**

This document provides everything you need to implement the follow-up system in Flutter, including API calls, UI components, and complete workflow.

---

## 📋 **Table of Contents**

1. [API Endpoints](#api-endpoints)
2. [Flutter API Functions](#flutter-api-functions)
3. [UI Components](#ui-components)
4. [Complete Workflow](#complete-workflow)
5. [State Management](#state-management)
6. [Error Handling](#error-handling)
7. [Testing Guide](#testing-guide)

---

## 🔗 **API Endpoints**

### **1. List Patients with Follow-Up Status**
```http
GET /api/organizations/clinic-specific-patients
```

**Query Parameters:**
- `clinic_id` (required): Clinic UUID
- `doctor_id` (optional): Doctor UUID for specific doctor
- `department_id` (optional): Department UUID for specific department
- `search` (optional): Search term for patient name/phone

**Response:**
```json
{
  "clinic_id": "clinic-uuid",
  "total": 5,
  "patients": [
    {
      "id": "patient-uuid",
      "first_name": "John",
      "last_name": "Doe",
      "phone": "+1234567890",
      "follow_up_eligibility": {
        "eligible": true,
        "is_free": true,
        "status_label": "free",
        "color_code": "green",
        "message": "Free follow-up available (3 days remaining)",
        "days_remaining": 3,
        "reason": null
      },
      "last_appointment": {
        "id": "appointment-uuid",
        "doctor_id": "doctor-uuid",
        "doctor_name": "Dr. Smith",
        "department_id": "dept-uuid",
        "department": "Cardiology",
        "date": "2025-10-20",
        "status": "completed",
        "days_since": 2
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
          "remaining_days": 3,
          "status": "active",
          "follow_up_eligible": true,
          "follow_up_status": "active",
          "renewal_status": "valid",
          "free_follow_up_used": false,
          "next_followup_expiry": "2025-10-25",
          "note": "Eligible for free follow-up with Dr. Smith (Cardiology)"
        }
      ],
      "eligible_follow_ups": [
        {
          "appointment_id": "appointment-uuid",
          "doctor_id": "doctor-uuid",
          "doctor_name": "Dr. Smith",
          "department_id": "dept-uuid",
          "department": "Cardiology",
          "appointment_date": "2025-10-20",
          "remaining_days": 3,
          "next_followup_expiry": "2025-10-25",
          "note": "Eligible for free follow-up with Dr. Smith (Cardiology)"
        }
      ],
      "expired_followups": []
    }
  ]
}
```

---

### **2. Create Simple Appointment**
```http
POST /api/appointments/simple
```

**Request Body:**
```json
{
  "clinic_patient_id": "patient-uuid",
  "clinic_id": "clinic-uuid",
  "doctor_id": "doctor-uuid",
  "department_id": "dept-uuid",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2025-10-23",
  "appointment_time": "10:00",
  "consultation_type": "follow-up-via-clinic",
  "is_follow_up": true,
  "payment_method": null,
  "payment_mode": null,
  "notes": "Follow-up consultation"
}
```

**Response:**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "id": "appointment-uuid",
    "clinic_patient_id": "patient-uuid",
    "doctor_id": "doctor-uuid",
    "appointment_date": "2025-10-23",
    "appointment_time": "10:00",
    "consultation_type": "follow-up-via-clinic",
    "fee_amount": 0,
    "payment_status": "waived",
    "status": "confirmed"
  },
  "is_free_followup": true,
  "followup_type": "free",
  "followup_message": "This is a FREE follow-up (renewed after regular appointment)"
}
```

---

## 📱 **Flutter API Functions**

### **1. API Service Class**

```dart
import 'dart:convert';
import 'package:http/http.dart' as http;

class FollowUpApiService {
  static const String baseUrl = 'http://localhost:3002';
  static const String appointmentUrl = 'http://localhost:3001';
  
  // Headers
  static Map<String, String> getHeaders(String token) {
    return {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer $token',
    };
  }

  // 1. Get Patients with Follow-Up Status
  static Future<Map<String, dynamic>> getPatientsWithFollowUpStatus({
    required String clinicId,
    String? doctorId,
    String? departmentId,
    String? search,
    required String token,
  }) async {
    try {
      String url = '$baseUrl/api/organizations/clinic-specific-patients?clinic_id=$clinicId';
      
      if (doctorId != null) url += '&doctor_id=$doctorId';
      if (departmentId != null) url += '&department_id=$departmentId';
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

  // 2. Create Follow-Up Appointment
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

  // 3. Create Regular Appointment
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

### **2. Data Models**

```dart
// Follow-Up Eligibility Model
class FollowUpEligibility {
  final bool eligible;
  final bool isFree;
  final String statusLabel;
  final String colorCode;
  final String message;
  final int? daysRemaining;
  final String? reason;

  FollowUpEligibility({
    required this.eligible,
    required this.isFree,
    required this.statusLabel,
    required this.colorCode,
    required this.message,
    this.daysRemaining,
    this.reason,
  });

  factory FollowUpEligibility.fromJson(Map<String, dynamic> json) {
    return FollowUpEligibility(
      eligible: json['eligible'] ?? false,
      isFree: json['is_free'] ?? false,
      statusLabel: json['status_label'] ?? 'none',
      colorCode: json['color_code'] ?? 'gray',
      message: json['message'] ?? '',
      daysRemaining: json['days_remaining'],
      reason: json['reason'],
    );
  }
}

// Patient Model
class Patient {
  final String id;
  final String clinicId;
  final String firstName;
  final String lastName;
  final String phone;
  final String? email;
  final FollowUpEligibility followUpEligibility;
  final LastAppointmentInfo? lastAppointment;
  final List<AppointmentHistoryItem> appointmentHistory;
  final List<EligibleFollowUp> eligibleFollowUps;
  final List<ExpiredFollowUp> expiredFollowUps;

  Patient({
    required this.id,
    required this.clinicId,
    required this.firstName,
    required this.lastName,
    required this.phone,
    this.email,
    required this.followUpEligibility,
    this.lastAppointment,
    required this.appointmentHistory,
    required this.eligibleFollowUps,
    required this.expiredFollowUps,
  });

  factory Patient.fromJson(Map<String, dynamic> json) {
    return Patient(
      id: json['id'],
      clinicId: json['clinic_id'],
      firstName: json['first_name'],
      lastName: json['last_name'],
      phone: json['phone'],
      email: json['email'],
      followUpEligibility: FollowUpEligibility.fromJson(json['follow_up_eligibility']),
      lastAppointment: json['last_appointment'] != null 
          ? LastAppointmentInfo.fromJson(json['last_appointment'])
          : null,
      appointmentHistory: (json['appointment_history'] as List?)
          ?.map((item) => AppointmentHistoryItem.fromJson(item))
          .toList() ?? [],
      eligibleFollowUps: (json['eligible_follow_ups'] as List?)
          ?.map((item) => EligibleFollowUp.fromJson(item))
          .toList() ?? [],
      expiredFollowUps: (json['expired_followups'] as List?)
          ?.map((item) => ExpiredFollowUp.fromJson(item))
          .toList() ?? [],
    );
  }
}

// Last Appointment Info Model
class LastAppointmentInfo {
  final String id;
  final String doctorId;
  final String doctorName;
  final String? departmentId;
  final String? department;
  final String date;
  final String status;
  final int daysSince;

  LastAppointmentInfo({
    required this.id,
    required this.doctorId,
    required this.doctorName,
    this.departmentId,
    this.department,
    required this.date,
    required this.status,
    required this.daysSince,
  });

  factory LastAppointmentInfo.fromJson(Map<String, dynamic> json) {
    return LastAppointmentInfo(
      id: json['id'],
      doctorId: json['doctor_id'],
      doctorName: json['doctor_name'],
      departmentId: json['department_id'],
      department: json['department'],
      date: json['date'],
      status: json['status'],
      daysSince: json['days_since'],
    );
  }
}

// Appointment History Item Model
class AppointmentHistoryItem {
  final String id;
  final String doctorId;
  final String doctorName;
  final String? departmentId;
  final String? department;
  final String appointmentType;
  final String appointmentDate;
  final int daysSince;
  final int validityDays;
  final int? remainingDays;
  final String status;
  final bool followUpEligible;
  final String followUpStatus;
  final String renewalStatus;
  final bool freeFollowUpUsed;
  final String? nextFollowUpExpiry;
  final String note;

  AppointmentHistoryItem({
    required this.id,
    required this.doctorId,
    required this.doctorName,
    this.departmentId,
    this.department,
    required this.appointmentType,
    required this.appointmentDate,
    required this.daysSince,
    required this.validityDays,
    this.remainingDays,
    required this.status,
    required this.followUpEligible,
    required this.followUpStatus,
    required this.renewalStatus,
    required this.freeFollowUpUsed,
    this.nextFollowUpExpiry,
    required this.note,
  });

  factory AppointmentHistoryItem.fromJson(Map<String, dynamic> json) {
    return AppointmentHistoryItem(
      id: json['appointment_id'],
      doctorId: json['doctor_id'],
      doctorName: json['doctor_name'],
      departmentId: json['department_id'],
      department: json['department'],
      appointmentType: json['appointment_type'],
      appointmentDate: json['appointment_date'],
      daysSince: json['days_since'],
      validityDays: json['validity_days'],
      remainingDays: json['remaining_days'],
      status: json['status'],
      followUpEligible: json['follow_up_eligible'],
      followUpStatus: json['follow_up_status'],
      renewalStatus: json['renewal_status'],
      freeFollowUpUsed: json['free_follow_up_used'],
      nextFollowUpExpiry: json['next_followup_expiry'],
      note: json['note'] ?? '',
    );
  }
}

// Eligible Follow-Up Model
class EligibleFollowUp {
  final String appointmentId;
  final String doctorId;
  final String doctorName;
  final String? departmentId;
  final String? department;
  final String appointmentDate;
  final int remainingDays;
  final String nextFollowUpExpiry;
  final String note;

  EligibleFollowUp({
    required this.appointmentId,
    required this.doctorId,
    required this.doctorName,
    this.departmentId,
    this.department,
    required this.appointmentDate,
    required this.remainingDays,
    required this.nextFollowUpExpiry,
    required this.note,
  });

  factory EligibleFollowUp.fromJson(Map<String, dynamic> json) {
    return EligibleFollowUp(
      appointmentId: json['appointment_id'],
      doctorId: json['doctor_id'],
      doctorName: json['doctor_name'],
      departmentId: json['department_id'],
      department: json['department'],
      appointmentDate: json['appointment_date'],
      remainingDays: json['remaining_days'],
      nextFollowUpExpiry: json['next_followup_expiry'],
      note: json['note'] ?? '',
    );
  }
}

// Expired Follow-Up Model
class ExpiredFollowUp {
  final String appointmentId;
  final String doctorId;
  final String doctorName;
  final String? departmentId;
  final String? department;
  final String expiredOn;
  final String note;

  ExpiredFollowUp({
    required this.appointmentId,
    required this.doctorId,
    required this.doctorName,
    this.departmentId,
    this.department,
    required this.expiredOn,
    required this.note,
  });

  factory ExpiredFollowUp.fromJson(Map<String, dynamic> json) {
    return ExpiredFollowUp(
      appointmentId: json['appointment_id'],
      doctorId: json['doctor_id'],
      doctorName: json['doctor_name'],
      departmentId: json['department_id'],
      department: json['department'],
      expiredOn: json['expired_on'],
      note: json['note'],
    );
  }
}
```

---

## 🎨 **UI Components**

### **1. Patient List Widget**

```dart
import 'package:flutter/material.dart';

class PatientListWidget extends StatefulWidget {
  final String clinicId;
  final String? doctorId;
  final String? departmentId;
  final String token;

  const PatientListWidget({
    Key? key,
    required this.clinicId,
    this.doctorId,
    this.departmentId,
    required this.token,
  }) : super(key: key);

  @override
  _PatientListWidgetState createState() => _PatientListWidgetState();
}

class _PatientListWidgetState extends State<PatientListWidget> {
  List<Patient> patients = [];
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
      final response = await FollowUpApiService.getPatientsWithFollowUpStatus(
        clinicId: widget.clinicId,
        doctorId: widget.doctorId,
        departmentId: widget.departmentId,
        search: searchQuery.isNotEmpty ? searchQuery : null,
        token: widget.token,
      );

      final List<dynamic> patientsJson = response['patients'];
      setState(() {
        patients = patientsJson.map((json) => Patient.fromJson(json)).toList();
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
                            return PatientCard(
                              patient: patients[index],
                              onBookFollowUp: () => _showBookFollowUpDialog(patients[index]),
                              onBookRegular: () => _showBookRegularDialog(patients[index]),
                            );
                          },
                        ),
        ),
      ],
    );
  }

  void _showBookFollowUpDialog(Patient patient) {
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

  void _showBookRegularDialog(Patient patient) {
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

---

### **2. Patient Card Widget**

```dart
class PatientCard extends StatelessWidget {
  final Patient patient;
  final VoidCallback onBookFollowUp;
  final VoidCallback onBookRegular;

  const PatientCard({
    Key? key,
    required this.patient,
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
                // Avatar with status color
                CircleAvatar(
                  backgroundColor: _getStatusColor(patient.followUpEligibility.colorCode),
                  radius: 24,
                  child: Text(
                    '${patient.firstName[0]}${patient.lastName[0]}',
                    style: TextStyle(
                      color: Colors.white,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                SizedBox(width: 16),
                
                // Patient Info
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        '${patient.firstName} ${patient.lastName}',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      SizedBox(height: 4),
                      Text(
                        patient.phone,
                        style: TextStyle(
                          fontSize: 14,
                          color: Colors.grey[600],
                        ),
                      ),
                    ],
                  ),
                ),
                
                // Status Icon
                _getStatusIcon(patient.followUpEligibility.statusLabel),
              ],
            ),
            
            SizedBox(height: 12),
            
            // Follow-Up Status
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
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          _getStatusLabel(patient.followUpEligibility.statusLabel),
                          style: TextStyle(
                            fontWeight: FontWeight.bold,
                            color: _getStatusColor(patient.followUpEligibility.colorCode),
                          ),
                        ),
                        if (patient.followUpEligibility.daysRemaining != null)
                          Text(
                            '${patient.followUpEligibility.daysRemaining} days remaining',
                            style: TextStyle(
                              fontSize: 12,
                              color: Colors.grey[600],
                            ),
                          ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            
            SizedBox(height: 12),
            
            // Action Buttons
            Row(
              children: [
                // Follow-Up Button
                if (_canBookFollowUp(patient.followUpEligibility.statusLabel))
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
                
                if (_canBookFollowUp(patient.followUpEligibility.statusLabel))
                  SizedBox(width: 8),
                
                // Regular Appointment Button
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
            
            // Additional Info
            if (patient.followUpEligibility.message.isNotEmpty)
              Padding(
                padding: EdgeInsets.only(top: 8),
                child: Text(
                  patient.followUpEligibility.message,
                  style: TextStyle(
                    fontSize: 12,
                    color: Colors.grey[600],
                    fontStyle: FontStyle.italic,
                  ),
                ),
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
      case 'needs_selection':
        return Icon(Icons.person_search, color: Colors.grey, size: 24);
      default:
        return Icon(Icons.help_outline, color: Colors.grey, size: 24);
    }
  }

  String _getStatusLabel(String statusLabel) {
    switch (statusLabel) {
      case 'free':
        return 'Free Follow-Up';
      case 'paid':
        return 'Paid Follow-Up';
      case 'none':
        return 'No History';
      case 'needs_selection':
        return 'Select Doctor';
      default:
        return 'Unknown';
    }
  }

  bool _canBookFollowUp(String statusLabel) {
    return statusLabel == 'free' || statusLabel == 'paid';
  }
}
```

---

### **3. Book Follow-Up Dialog**

```dart
class BookFollowUpDialog extends StatefulWidget {
  final Patient patient;
  final String token;
  final VoidCallback onSuccess;

  const BookFollowUpDialog({
    Key? key,
    required this.patient,
    required this.token,
    required this.onSuccess,
  }) : super(key: key);

  @override
  _BookFollowUpDialogState createState() => _BookFollowUpDialogState();
}

class _BookFollowUpDialogState extends State<BookFollowUpDialog> {
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
                    backgroundColor: _getStatusColor(widget.patient.followUpEligibility.colorCode),
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
            
            // Follow-Up Status
            Container(
              padding: EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: _getStatusColor(widget.patient.followUpEligibility.colorCode).withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(
                  color: _getStatusColor(widget.patient.followUpEligibility.colorCode),
                ),
              ),
              child: Row(
                children: [
                  Icon(
                    _getStatusIcon(widget.patient.followUpEligibility.statusLabel).icon,
                    color: _getStatusColor(widget.patient.followUpEligibility.colorCode),
                  ),
                  SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      widget.patient.followUpEligibility.message,
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        color: _getStatusColor(widget.patient.followUpEligibility.colorCode),
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
            if (widget.patient.followUpEligibility.statusLabel == 'paid')
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
    setState(() {
      isLoading = true;
    });

    try {
      // Get the first eligible follow-up for this patient
      final eligibleFollowUp = widget.patient.eligibleFollowUps.first;
      
      await FollowUpApiService.createFollowUpAppointment(
        clinicPatientId: widget.patient.id,
        clinicId: widget.patient.clinicId,
        doctorId: eligibleFollowUp.doctorId,
        departmentId: eligibleFollowUp.departmentId,
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
      default:
        return Icon(Icons.info_outline, color: Colors.grey);
    }
  }
}
```

---

### **4. Book Regular Dialog**

```dart
class BookRegularDialog extends StatefulWidget {
  final Patient patient;
  final String token;
  final VoidCallback onSuccess;

  const BookRegularDialog({
    Key? key,
    required this.patient,
    required this.token,
    required this.onSuccess,
  }) : super(key: key);

  @override
  _BookRegularDialogState createState() => _BookRegularDialogState();
}

class _BookRegularDialogState extends State<BookRegularDialog> {
  DateTime selectedDate = DateTime.now();
  String selectedTime = '10:00';
  String consultationType = 'clinic_visit';
  String paymentMethod = 'pay_now';
  String? paymentMode;
  bool isLoading = false;

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text('Book Regular Appointment'),
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
            
            // Renewal Info
            Container(
              padding: EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.green[50],
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.green),
              ),
              child: Row(
                children: [
                  Icon(Icons.refresh, color: Colors.green),
                  SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      'This will grant a new 5-day free follow-up period',
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        color: Colors.green[800],
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
                  value: 'clinic_visit',
                  child: Text('Clinic Visit'),
                ),
                DropdownMenuItem(
                  value: 'video_consultation',
                  child: Text('Video Consultation'),
                ),
              ],
              onChanged: (value) {
                setState(() {
                  consultationType = value!;
                });
              },
            ),
            
            SizedBox(height: 16),
            
            // Payment Method
            DropdownButtonFormField<String>(
              value: paymentMethod,
              decoration: InputDecoration(
                labelText: 'Payment Method',
                border: OutlineInputBorder(),
              ),
              items: [
                DropdownMenuItem(
                  value: 'pay_now',
                  child: Text('Pay Now'),
                ),
                DropdownMenuItem(
                  value: 'pay_later',
                  child: Text('Pay Later'),
                ),
              ],
              onChanged: (value) {
                setState(() {
                  paymentMethod = value!;
                });
              },
            ),
            
            // Payment Mode (if pay_now)
            if (paymentMethod == 'pay_now')
              DropdownButtonFormField<String>(
                value: paymentMode,
                decoration: InputDecoration(
                  labelText: 'Payment Mode',
                  border: OutlineInputBorder(),
                ),
                items: [
                  DropdownMenuItem(
                    value: 'cash',
                    child: Text('Cash'),
                  ),
                  DropdownMenuItem(
                    value: 'card',
                    child: Text('Card'),
                  ),
                  DropdownMenuItem(
                    value: 'upi',
                    child: Text('UPI'),
                  ),
                ],
                onChanged: (value) {
                  setState(() {
                    paymentMode = value;
                  });
                },
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
          onPressed: isLoading ? null : _bookRegular,
          child: isLoading
              ? SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : Text('Book Appointment'),
        ),
      ],
    );
  }

  Future<void> _bookRegular() async {
    setState(() {
      isLoading = true;
    });

    try {
      await FollowUpApiService.createRegularAppointment(
        clinicPatientId: widget.patient.id,
        clinicId: widget.patient.clinicId,
        doctorId: 'doctor-id', // You need to implement doctor selection
        departmentId: 'department-id', // You need to implement department selection
        individualSlotId: 'slot-id', // You need to implement slot selection
        appointmentDate: selectedDate.toIso8601String().split('T')[0],
        appointmentTime: selectedTime,
        consultationType: consultationType,
        paymentMethod: paymentMethod,
        paymentMode: paymentMode,
        token: widget.token,
      );

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Regular appointment booked successfully!'),
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
}
```

---

## 🔄 **Complete Workflow**

### **1. Main App Screen**

```dart
class FollowUpApp extends StatefulWidget {
  @override
  _FollowUpAppState createState() => _FollowUpAppState();
}

class _FollowUpAppState extends State<FollowUpApp> {
  String? selectedDoctorId;
  String? selectedDepartmentId;
  String clinicId = 'your-clinic-id';
  String token = 'your-auth-token';

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Follow-Up Management'),
        backgroundColor: Colors.blue,
      ),
      body: Column(
        children: [
          // Doctor & Department Selection
          Container(
            padding: EdgeInsets.all(16),
            color: Colors.grey[100],
            child: Column(
              children: [
                // Doctor Selection
                DropdownButtonFormField<String>(
                  value: selectedDoctorId,
                  decoration: InputDecoration(
                    labelText: 'Select Doctor',
                    border: OutlineInputBorder(),
                  ),
                  items: [
                    DropdownMenuItem(
                      value: 'doctor-1',
                      child: Text('Dr. Smith'),
                    ),
                    DropdownMenuItem(
                      value: 'doctor-2',
                      child: Text('Dr. Johnson'),
                    ),
                  ],
                  onChanged: (value) {
                    setState(() {
                      selectedDoctorId = value;
                      selectedDepartmentId = null; // Reset department
                    });
                  },
                ),
                
                SizedBox(height: 16),
                
                // Department Selection
                DropdownButtonFormField<String>(
                  value: selectedDepartmentId,
                  decoration: InputDecoration(
                    labelText: 'Select Department',
                    border: OutlineInputBorder(),
                  ),
                  items: [
                    DropdownMenuItem(
                      value: 'dept-1',
                      child: Text('Cardiology'),
                    ),
                    DropdownMenuItem(
                      value: 'dept-2',
                      child: Text('Neurology'),
                    ),
                  ],
                  onChanged: (value) {
                    setState(() {
                      selectedDepartmentId = value;
                    });
                  },
                ),
              ],
            ),
          ),
          
          // Patient List
          Expanded(
            child: PatientListWidget(
              clinicId: clinicId,
              doctorId: selectedDoctorId,
              departmentId: selectedDepartmentId,
              token: token,
            ),
          ),
        ],
      ),
    );
  }
}
```

---

## 📊 **State Management**

### **Using Provider**

```dart
// Follow-Up Provider
class FollowUpProvider extends ChangeNotifier {
  List<Patient> patients = [];
  bool isLoading = false;
  String? errorMessage;
  String? selectedDoctorId;
  String? selectedDepartmentId;

  Future<void> loadPatients({
    required String clinicId,
    String? doctorId,
    String? departmentId,
    String? search,
    required String token,
  }) async {
    isLoading = true;
    errorMessage = null;
    notifyListeners();

    try {
      final response = await FollowUpApiService.getPatientsWithFollowUpStatus(
        clinicId: clinicId,
        doctorId: doctorId,
        departmentId: departmentId,
        search: search,
        token: token,
      );

      final List<dynamic> patientsJson = response['patients'];
      patients = patientsJson.map((json) => Patient.fromJson(json)).toList();
      isLoading = false;
      notifyListeners();
    } catch (e) {
      errorMessage = e.toString();
      isLoading = false;
      notifyListeners();
    }
  }

  void setSelectedDoctor(String? doctorId) {
    selectedDoctorId = doctorId;
    selectedDepartmentId = null; // Reset department
    notifyListeners();
  }

  void setSelectedDepartment(String? departmentId) {
    selectedDepartmentId = departmentId;
    notifyListeners();
  }

  Future<void> bookFollowUp({
    required Patient patient,
    required String appointmentDate,
    required String appointmentTime,
    required String consultationType,
    required String token,
  }) async {
    try {
      final eligibleFollowUp = patient.eligibleFollowUps.first;
      
      await FollowUpApiService.createFollowUpAppointment(
        clinicPatientId: patient.id,
        clinicId: patient.clinicId,
        doctorId: eligibleFollowUp.doctorId,
        departmentId: eligibleFollowUp.departmentId,
        individualSlotId: 'slot-id', // Implement slot selection
        appointmentDate: appointmentDate,
        appointmentTime: appointmentTime,
        consultationType: consultationType,
        token: token,
      );
      
      // Reload patients to reflect changes
      await loadPatients(
        clinicId: patient.clinicId,
        doctorId: selectedDoctorId,
        departmentId: selectedDepartmentId,
        token: token,
      );
    } catch (e) {
      throw Exception('Failed to book follow-up: $e');
    }
  }

  Future<void> bookRegular({
    required Patient patient,
    required String appointmentDate,
    required String appointmentTime,
    required String consultationType,
    required String paymentMethod,
    String? paymentMode,
    required String token,
  }) async {
    try {
      await FollowUpApiService.createRegularAppointment(
        clinicPatientId: patient.id,
        clinicId: patient.clinicId,
        doctorId: selectedDoctorId!,
        departmentId: selectedDepartmentId,
        individualSlotId: 'slot-id', // Implement slot selection
        appointmentDate: appointmentDate,
        appointmentTime: appointmentTime,
        consultationType: consultationType,
        paymentMethod: paymentMethod,
        paymentMode: paymentMode,
        token: token,
      );
      
      // Reload patients to reflect changes
      await loadPatients(
        clinicId: patient.clinicId,
        doctorId: selectedDoctorId,
        departmentId: selectedDepartmentId,
        token: token,
      );
    } catch (e) {
      throw Exception('Failed to book regular appointment: $e');
    }
  }
}
```

---

## 🧪 **Testing Guide**

### **1. Test Scenarios**

```dart
// Test Helper Class
class FollowUpTestHelper {
  static Future<void> testCompleteWorkflow() async {
    print('🧪 Testing Complete Follow-Up Workflow');
    
    // Test 1: Load patients without doctor selection
    print('Test 1: Loading patients without doctor selection...');
    await _testLoadPatientsWithoutDoctor();
    
    // Test 2: Load patients with doctor selection
    print('Test 2: Loading patients with doctor selection...');
    await _testLoadPatientsWithDoctor();
    
    // Test 3: Book regular appointment
    print('Test 3: Booking regular appointment...');
    await _testBookRegularAppointment();
    
    // Test 4: Book free follow-up
    print('Test 4: Booking free follow-up...');
    await _testBookFreeFollowUp();
    
    // Test 5: Book paid follow-up
    print('Test 5: Booking paid follow-up...');
    await _testBookPaidFollowUp();
    
    print('✅ All tests completed!');
  }

  static Future<void> _testLoadPatientsWithoutDoctor() async {
    try {
      final response = await FollowUpApiService.getPatientsWithFollowUpStatus(
        clinicId: 'test-clinic-id',
        token: 'test-token',
      );
      
      // Check if all patients have 'needs_selection' status
      final patients = response['patients'] as List;
      for (var patientJson in patients) {
        final statusLabel = patientJson['follow_up_eligibility']['status_label'];
        assert(statusLabel == 'needs_selection', 'Expected needs_selection, got $statusLabel');
      }
      
      print('✅ Test 1 passed: All patients show needs_selection status');
    } catch (e) {
      print('❌ Test 1 failed: $e');
    }
  }

  static Future<void> _testLoadPatientsWithDoctor() async {
    try {
      final response = await FollowUpApiService.getPatientsWithFollowUpStatus(
        clinicId: 'test-clinic-id',
        doctorId: 'test-doctor-id',
        departmentId: 'test-department-id',
        token: 'test-token',
      );
      
      // Check if patients have proper status labels
      final patients = response['patients'] as List;
      for (var patientJson in patients) {
        final statusLabel = patientJson['follow_up_eligibility']['status_label'];
        assert(['free', 'paid', 'none'].contains(statusLabel), 'Invalid status label: $statusLabel');
      }
      
      print('✅ Test 2 passed: Patients have valid status labels');
    } catch (e) {
      print('❌ Test 2 failed: $e');
    }
  }

  static Future<void> _testBookRegularAppointment() async {
    try {
      final response = await FollowUpApiService.createRegularAppointment(
        clinicPatientId: 'test-patient-id',
        clinicId: 'test-clinic-id',
        doctorId: 'test-doctor-id',
        departmentId: 'test-department-id',
        individualSlotId: 'test-slot-id',
        appointmentDate: '2025-10-23',
        appointmentTime: '10:00',
        consultationType: 'clinic_visit',
        paymentMethod: 'pay_now',
        paymentMode: 'cash',
        token: 'test-token',
      );
      
      assert(response['is_regular_appointment'] == true, 'Expected regular appointment');
      assert(response['followup_granted'] == true, 'Expected follow-up granted');
      
      print('✅ Test 3 passed: Regular appointment created with follow-up granted');
    } catch (e) {
      print('❌ Test 3 failed: $e');
    }
  }

  static Future<void> _testBookFreeFollowUp() async {
    try {
      final response = await FollowUpApiService.createFollowUpAppointment(
        clinicPatientId: 'test-patient-id',
        clinicId: 'test-clinic-id',
        doctorId: 'test-doctor-id',
        departmentId: 'test-department-id',
        individualSlotId: 'test-slot-id',
        appointmentDate: '2025-10-24',
        appointmentTime: '11:00',
        consultationType: 'follow-up-via-clinic',
        token: 'test-token',
      );
      
      assert(response['is_free_followup'] == true, 'Expected free follow-up');
      assert(response['followup_type'] == 'free', 'Expected free type');
      
      print('✅ Test 4 passed: Free follow-up created successfully');
    } catch (e) {
      print('❌ Test 4 failed: $e');
    }
  }

  static Future<void> _testBookPaidFollowUp() async {
    try {
      final response = await FollowUpApiService.createFollowUpAppointment(
        clinicPatientId: 'test-patient-id',
        clinicId: 'test-clinic-id',
        doctorId: 'test-doctor-id',
        departmentId: 'test-department-id',
        individualSlotId: 'test-slot-id',
        appointmentDate: '2025-10-30', // After 5 days
        appointmentTime: '12:00',
        consultationType: 'follow-up-via-clinic',
        token: 'test-token',
      );
      
      assert(response['is_free_followup'] == false, 'Expected paid follow-up');
      assert(response['followup_type'] == 'paid', 'Expected paid type');
      
      print('✅ Test 5 passed: Paid follow-up created successfully');
    } catch (e) {
      print('❌ Test 5 failed: $e');
    }
  }
}
```

---

## 🚨 **Error Handling**

### **Error Types & Solutions**

```dart
class FollowUpErrorHandler {
  static String handleError(dynamic error) {
    if (error.toString().contains('Free follow-up already used')) {
      return 'This free follow-up has already been used. Please book a paid follow-up or book a new regular appointment.';
    } else if (error.toString().contains('Not eligible for follow-up')) {
      return 'This patient is not eligible for a follow-up appointment.';
    } else if (error.toString().contains('No previous appointment found')) {
      return 'This patient has no previous appointment with the selected doctor and department.';
    } else if (error.toString().contains('Doctor not selected')) {
      return 'Please select a doctor to check follow-up eligibility.';
    } else if (error.toString().contains('Payment method required')) {
      return 'Payment method is required for this appointment.';
    } else if (error.toString().contains('Failed to check follow-up eligibility')) {
      return 'Unable to check follow-up eligibility. Please try again.';
    } else {
      return 'An unexpected error occurred. Please try again.';
    }
  }

  static void showErrorSnackBar(BuildContext context, String error) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(error),
        backgroundColor: Colors.red,
        duration: Duration(seconds: 5),
        action: SnackBarAction(
          label: 'Retry',
          textColor: Colors.white,
          onPressed: () {
            // Implement retry logic
          },
        ),
      ),
    );
  }

  static void showSuccessSnackBar(BuildContext context, String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: Colors.green,
        duration: Duration(seconds: 3),
      ),
    );
  }
}
```

---

## 📱 **Complete Implementation**

### **Main App with Provider**

```dart
void main() {
  runApp(
    ChangeNotifierProvider(
      create: (context) => FollowUpProvider(),
      child: MyApp(),
    ),
  );
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Follow-Up Management',
      theme: ThemeData(
        primarySwatch: Colors.blue,
        visualDensity: VisualDensity.adaptivePlatformDensity,
      ),
      home: FollowUpApp(),
    );
  }
}
```

---

## ✅ **Summary**

This complete documentation provides:

1. **✅ API Endpoints** - Complete REST API documentation
2. **✅ Flutter API Functions** - Ready-to-use service classes
3. **✅ Data Models** - Complete Dart models for all responses
4. **✅ UI Components** - Patient list, cards, dialogs
5. **✅ Complete Workflow** - End-to-end implementation
6. **✅ State Management** - Provider-based state management
7. **✅ Error Handling** - Comprehensive error handling
8. **✅ Testing Guide** - Complete testing scenarios

**Ready to implement!** 🚀✅

**Use this documentation to build your complete follow-up management system in Flutter!** 📱✨

