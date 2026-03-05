# Clinic-Specific Patient Management - Flutter Integration Guide

## 📖 Table of Contents
1. [Overview](#overview)
2. [API Endpoints](#api-endpoints)
3. [Complete JSON Examples](#complete-json-examples)
4. [Flutter Integration](#flutter-integration)
5. [UI/UX Examples](#uiux-examples)
6. [Error Handling](#error-handling)

---

## 🎯 Overview

**Clinic-Isolated Patient System** - Each clinic has their own patients with NO global conflicts.

### Key Features
- ✅ **Only name + phone required** - Quick patient registration
- ✅ **Clinic-isolated** - No conflicts between clinics
- ✅ **Same phone allowed** in different clinics
- ✅ **Complete CRUD** - Create, Read, Update, Delete
- ✅ **Search support** - By name, phone, or MO ID

---

## 🔌 API Endpoints

### Base URL
```
http://localhost:8081/api/organizations/clinic-specific-patients
```

### Available Endpoints

| Method | Endpoint | Purpose | Auth Required |
|--------|----------|---------|---------------|
| POST | `/clinic-specific-patients` | Create patient | ✅ Clinic Admin/Receptionist |
| GET | `/clinic-specific-patients?clinic_id=xxx` | List patients | ✅ Yes |
| GET | `/clinic-specific-patients/:id` | Get single patient | ✅ Yes |
| PUT | `/clinic-specific-patients/:id` | Update patient | ✅ Clinic Admin/Receptionist |
| DELETE | `/clinic-specific-patients/:id` | Delete patient | ✅ Clinic Admin |

---

## 📝 Complete JSON Examples

### Example 1: Minimal Patient Creation (Only Required Fields)

**Request:**
```json
POST /api/organizations/clinic-specific-patients
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "first_name": "Ahmed",
  "last_name": "Khan",
  "phone": "+971501234567"
}
```

**Response (201 Created):**
```json
{
  "message": "Patient created successfully for this clinic",
  "patient": {
    "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "first_name": "Ahmed",
    "last_name": "Khan",
    "phone": "+971501234567",
    "email": null,
    "date_of_birth": null,
    "gender": null,
    "mo_id": null,
    "medical_history": null,
    "allergies": null,
    "blood_group": null,
    "is_active": true,
    "created_at": "2024-10-15T11:30:00Z",
    "updated_at": "2024-10-15T11:30:00Z"
  }
}
```

---

### Example 2: Full Patient Creation (All Fields)

**Request:**
```json
POST /api/organizations/clinic-specific-patients
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "first_name": "Sara",
  "last_name": "Ali",
  "phone": "+971507654321",
  "email": "sara.ali@example.com",
  "date_of_birth": "1990-05-15",
  "gender": "female",
  "mo_id": "MO123456",
  "medical_history": "Hypertension, controlled with medication",
  "allergies": "Penicillin, Pollen",
  "blood_group": "A+"
}
```

**Response (201 Created):**
```json
{
  "message": "Patient created successfully for this clinic",
  "patient": {
    "id": "a1b2c3d4-e5f6-4789-a012-3456789abcde",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "first_name": "Sara",
    "last_name": "Ali",
    "phone": "+971507654321",
    "email": "sara.ali@example.com",
    "date_of_birth": "1990-05-15",
    "gender": "female",
    "mo_id": "MO123456",
    "medical_history": "Hypertension, controlled with medication",
    "allergies": "Penicillin, Pollen",
    "blood_group": "A+",
    "is_active": true,
    "created_at": "2024-10-15T11:35:00Z",
    "updated_at": "2024-10-15T11:35:00Z"
  }
}
```

---

### Example 3: List All Clinic Patients

**Request:**
```
GET /api/organizations/clinic-specific-patients?clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&only_active=true
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response (200 OK):**
```json
{
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "total": 3,
  "patients": [
    {
      "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "first_name": "Ahmed",
      "last_name": "Khan",
      "phone": "+971501234567",
      "email": null,
      "date_of_birth": null,
      "gender": null,
      "mo_id": null,
      "medical_history": null,
      "allergies": null,
      "blood_group": null,
      "is_active": true,
      "created_at": "2024-10-15T11:30:00Z",
      "updated_at": "2024-10-15T11:30:00Z"
    },
    {
      "id": "a1b2c3d4-e5f6-4789-a012-3456789abcde",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "first_name": "Sara",
      "last_name": "Ali",
      "phone": "+971507654321",
      "email": "sara.ali@example.com",
      "date_of_birth": "1990-05-15",
      "gender": "female",
      "mo_id": "MO123456",
      "medical_history": "Hypertension, controlled with medication",
      "allergies": "Penicillin, Pollen",
      "blood_group": "A+",
      "is_active": true,
      "created_at": "2024-10-15T11:35:00Z",
      "updated_at": "2024-10-15T11:35:00Z"
    },
    {
      "id": "b2c3d4e5-f6a7-4890-b123-456789abcdef",
      "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
      "first_name": "Mohammed",
      "last_name": "Hassan",
      "phone": "+971509876543",
      "email": "m.hassan@example.com",
      "mo_id": "MO789012",
      "blood_group": "O+",
      "is_active": true,
      "created_at": "2024-10-15T09:20:00Z",
      "updated_at": "2024-10-15T09:20:00Z"
    }
  ]
}
```

---

### Example 4: Search Patients

**By Name:**
```
GET /clinic-specific-patients?clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&search=Ahmed
```

**By Phone:**
```
GET /clinic-specific-patients?clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&search=+971501234567
```

**By MO ID:**
```
GET /clinic-specific-patients?clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&search=MO123456
```

**Response (200 OK):**
```json
{
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "total": 1,
  "patients": [
    {
      "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
      "first_name": "Ahmed",
      "last_name": "Khan",
      "phone": "+971501234567",
      "mo_id": null
    }
  ]
}
```

---

### Example 5: Get Single Patient

**Request:**
```
GET /api/organizations/clinic-specific-patients/f47ac10b-58cc-4372-a567-0e02b2c3d479
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "patient": {
    "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "first_name": "Ahmed",
    "last_name": "Khan",
    "phone": "+971501234567",
    "email": null,
    "date_of_birth": null,
    "gender": null,
    "mo_id": null,
    "medical_history": null,
    "allergies": null,
    "blood_group": null,
    "is_active": true,
    "created_at": "2024-10-15T11:30:00Z",
    "updated_at": "2024-10-15T11:30:00Z"
  }
}
```

---

### Example 6: Update Patient

**Request:**
```json
PUT /api/organizations/clinic-specific-patients/f47ac10b-58cc-4372-a567-0e02b2c3d479
Content-Type: application/json
Authorization: Bearer {token}

{
  "email": "ahmed.khan.updated@email.com",
  "mo_id": "MO999888",
  "blood_group": "O+",
  "medical_history": "Diabetes Type 2, diagnosed 2024",
  "allergies": "Penicillin"
}
```

**Response (200 OK):**
```json
{
  "message": "Patient updated successfully",
  "patient": {
    "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "first_name": "Ahmed",
    "last_name": "Khan",
    "phone": "+971501234567",
    "email": "ahmed.khan.updated@email.com",
    "mo_id": "MO999888",
    "blood_group": "O+",
    "medical_history": "Diabetes Type 2, diagnosed 2024",
    "allergies": "Penicillin",
    "is_active": true,
    "created_at": "2024-10-15T11:30:00Z",
    "updated_at": "2024-10-15T11:45:00Z"
  }
}
```

---

### Example 7: Delete Patient (Soft Delete)

**Request:**
```
DELETE /api/organizations/clinic-specific-patients/f47ac10b-58cc-4372-a567-0e02b2c3d479
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "message": "Patient deleted successfully"
}
```

---

## 🎯 Flutter Integration

### Step 1: Create API Service Class

```dart
// lib/services/clinic_patient_service.dart

import 'dart:convert';
import 'package:http/http.dart' as http;

class ClinicPatient {
  final String id;
  final String clinicId;
  final String firstName;
  final String lastName;
  final String phone;
  final String? email;
  final String? dateOfBirth;
  final String? gender;
  final String? moId;
  final String? medicalHistory;
  final String? allergies;
  final String? bloodGroup;
  final bool isActive;
  final DateTime createdAt;
  final DateTime updatedAt;

  ClinicPatient({
    required this.id,
    required this.clinicId,
    required this.firstName,
    required this.lastName,
    required this.phone,
    this.email,
    this.dateOfBirth,
    this.gender,
    this.moId,
    this.medicalHistory,
    this.allergies,
    this.bloodGroup,
    required this.isActive,
    required this.createdAt,
    required this.updatedAt,
  });

  factory ClinicPatient.fromJson(Map<String, dynamic> json) {
    return ClinicPatient(
      id: json['id'],
      clinicId: json['clinic_id'],
      firstName: json['first_name'],
      lastName: json['last_name'],
      phone: json['phone'],
      email: json['email'],
      dateOfBirth: json['date_of_birth'],
      gender: json['gender'],
      moId: json['mo_id'],
      medicalHistory: json['medical_history'],
      allergies: json['allergies'],
      bloodGroup: json['blood_group'],
      isActive: json['is_active'],
      createdAt: DateTime.parse(json['created_at']),
      updatedAt: DateTime.parse(json['updated_at']),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'clinic_id': clinicId,
      'first_name': firstName,
      'last_name': lastName,
      'phone': phone,
      'email': email,
      'date_of_birth': dateOfBirth,
      'gender': gender,
      'mo_id': moId,
      'medical_history': medicalHistory,
      'allergies': allergies,
      'blood_group': bloodGroup,
      'is_active': isActive,
    };
  }

  String get fullName => '$firstName $lastName';
}

class ClinicPatientService {
  final String baseUrl = 'http://localhost:8081/api/organizations';
  final String token;

  ClinicPatientService(this.token);

  Map<String, String> get headers => {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer $token',
  };

  // Create patient (minimal - only name and phone)
  Future<ClinicPatient> createPatient({
    required String clinicId,
    required String firstName,
    required String lastName,
    required String phone,
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl/clinic-specific-patients'),
      headers: headers,
      body: jsonEncode({
        'clinic_id': clinicId,
        'first_name': firstName,
        'last_name': lastName,
        'phone': phone,
      }),
    );

    if (response.statusCode == 201) {
      final data = jsonDecode(response.body);
      return ClinicPatient.fromJson(data['patient']);
    } else {
      final error = jsonDecode(response.body);
      throw Exception(error['error'] ?? 'Failed to create patient');
    }
  }

  // Create patient (full - with all optional fields)
  Future<ClinicPatient> createPatientFull({
    required String clinicId,
    required String firstName,
    required String lastName,
    required String phone,
    String? email,
    String? dateOfBirth,
    String? gender,
    String? moId,
    String? medicalHistory,
    String? allergies,
    String? bloodGroup,
  }) async {
    final body = {
      'clinic_id': clinicId,
      'first_name': firstName,
      'last_name': lastName,
      'phone': phone,
    };

    if (email != null) body['email'] = email;
    if (dateOfBirth != null) body['date_of_birth'] = dateOfBirth;
    if (gender != null) body['gender'] = gender;
    if (moId != null) body['mo_id'] = moId;
    if (medicalHistory != null) body['medical_history'] = medicalHistory;
    if (allergies != null) body['allergies'] = allergies;
    if (bloodGroup != null) body['blood_group'] = bloodGroup;

    final response = await http.post(
      Uri.parse('$baseUrl/clinic-specific-patients'),
      headers: headers,
      body: jsonEncode(body),
    );

    if (response.statusCode == 201) {
      final data = jsonDecode(response.body);
      return ClinicPatient.fromJson(data['patient']);
    } else {
      final error = jsonDecode(response.body);
      throw Exception(error['error'] ?? 'Failed to create patient');
    }
  }

  // List all patients for clinic
  Future<List<ClinicPatient>> listPatients(String clinicId) async {
    final response = await http.get(
      Uri.parse('$baseUrl/clinic-specific-patients?clinic_id=$clinicId&only_active=true'),
      headers: headers,
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      final List patientsJson = data['patients'] ?? [];
      return patientsJson.map((json) => ClinicPatient.fromJson(json)).toList();
    } else {
      throw Exception('Failed to load patients');
    }
  }

  // Search patients
  Future<List<ClinicPatient>> searchPatients(String clinicId, String query) async {
    final response = await http.get(
      Uri.parse('$baseUrl/clinic-specific-patients?clinic_id=$clinicId&search=$query'),
      headers: headers,
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      final List patientsJson = data['patients'] ?? [];
      return patientsJson.map((json) => ClinicPatient.fromJson(json)).toList();
    } else {
      throw Exception('Failed to search patients');
    }
  }

  // Get single patient
  Future<ClinicPatient> getPatient(String patientId) async {
    final response = await http.get(
      Uri.parse('$baseUrl/clinic-specific-patients/$patientId'),
      headers: headers,
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      return ClinicPatient.fromJson(data['patient']);
    } else {
      throw Exception('Failed to load patient');
    }
  }

  // Update patient
  Future<ClinicPatient> updatePatient(
    String patientId,
    Map<String, dynamic> updates,
  ) async {
    final response = await http.put(
      Uri.parse('$baseUrl/clinic-specific-patients/$patientId'),
      headers: headers,
      body: jsonEncode(updates),
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      return ClinicPatient.fromJson(data['patient']);
    } else {
      final error = jsonDecode(response.body);
      throw Exception(error['error'] ?? 'Failed to update patient');
    }
  }

  // Delete patient
  Future<void> deletePatient(String patientId) async {
    final response = await http.delete(
      Uri.parse('$baseUrl/clinic-specific-patients/$patientId'),
      headers: headers,
    );

    if (response.statusCode != 200) {
      final error = jsonDecode(response.body);
      throw Exception(error['error'] ?? 'Failed to delete patient');
    }
  }
}
```

---

### Step 2: Quick Patient Registration Screen

```dart
// lib/screens/quick_patient_registration.dart

import 'package:flutter/material.dart';
import '../services/clinic_patient_service.dart';

class QuickPatientRegistration extends StatefulWidget {
  final String clinicId;
  final String token;

  const QuickPatientRegistration({
    Key? key,
    required this.clinicId,
    required this.token,
  }) : super(key: key);

  @override
  State<QuickPatientRegistration> createState() => _QuickPatientRegistrationState();
}

class _QuickPatientRegistrationState extends State<QuickPatientRegistration> {
  final _formKey = GlobalKey<FormState>();
  final _firstNameController = TextEditingController();
  final _lastNameController = TextEditingController();
  final _phoneController = TextEditingController();
  
  bool _isLoading = false;
  late ClinicPatientService _patientService;

  @override
  void initState() {
    super.initState();
    _patientService = ClinicPatientService(widget.token);
  }

  Future<void> _createPatient() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _isLoading = true);

    try {
      final patient = await _patientService.createPatient(
        clinicId: widget.clinicId,
        firstName: _firstNameController.text.trim(),
        lastName: _lastNameController.text.trim(),
        phone: _phoneController.text.trim(),
      );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Patient ${patient.fullName} registered successfully!'),
            backgroundColor: Colors.green,
          ),
        );
        
        // Return patient data to previous screen
        Navigator.pop(context, patient);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Quick Patient Registration'),
        backgroundColor: Colors.blue,
      ),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text(
                'Only name and phone required for quick registration',
                style: TextStyle(
                  fontSize: 14,
                  color: Colors.grey,
                  fontStyle: FontStyle.italic,
                ),
              ),
              const SizedBox(height: 20),
              
              // First Name
              TextFormField(
                controller: _firstNameController,
                decoration: const InputDecoration(
                  labelText: 'First Name *',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.person),
                ),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'First name is required';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),
              
              // Last Name
              TextFormField(
                controller: _lastNameController,
                decoration: const InputDecoration(
                  labelText: 'Last Name *',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.person_outline),
                ),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'Last name is required';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),
              
              // Phone
              TextFormField(
                controller: _phoneController,
                decoration: const InputDecoration(
                  labelText: 'Phone Number *',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.phone),
                  hintText: '+971501234567',
                ),
                keyboardType: TextInputType.phone,
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'Phone number is required';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 24),
              
              // Submit Button
              ElevatedButton(
                onPressed: _isLoading ? null : _createPatient,
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  backgroundColor: Colors.blue,
                  foregroundColor: Colors.white,
                ),
                child: _isLoading
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                        ),
                      )
                    : const Text(
                        'Register Patient',
                        style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                      ),
              ),
              
              const SizedBox(height: 12),
              
              const Text(
                '* Required fields',
                style: TextStyle(fontSize: 12, color: Colors.grey),
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  void dispose() {
    _firstNameController.dispose();
    _lastNameController.dispose();
    _phoneController.dispose();
    super.dispose();
  }
}
```

---

### Step 3: Full Patient Registration Screen (With All Fields)

```dart
// lib/screens/full_patient_registration.dart

import 'package:flutter/material.dart';
import '../services/clinic_patient_service.dart';

class FullPatientRegistration extends StatefulWidget {
  final String clinicId;
  final String token;

  const FullPatientRegistration({
    Key? key,
    required this.clinicId,
    required this.token,
  }) : super(key: key);

  @override
  State<FullPatientRegistration> createState() => _FullPatientRegistrationState();
}

class _FullPatientRegistrationState extends State<FullPatientRegistration> {
  final _formKey = GlobalKey<FormState>();
  final _firstNameController = TextEditingController();
  final _lastNameController = TextEditingController();
  final _phoneController = TextEditingController();
  final _emailController = TextEditingController();
  final _moIdController = TextEditingController();
  final _medicalHistoryController = TextEditingController();
  final _allergiesController = TextEditingController();
  
  String? _selectedGender;
  String? _selectedBloodGroup;
  DateTime? _selectedDateOfBirth;
  bool _isLoading = false;
  late ClinicPatientService _patientService;

  @override
  void initState() {
    super.initState();
    _patientService = ClinicPatientService(widget.token);
  }

  Future<void> _createPatient() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _isLoading = true);

    try {
      final patient = await _patientService.createPatientFull(
        clinicId: widget.clinicId,
        firstName: _firstNameController.text.trim(),
        lastName: _lastNameController.text.trim(),
        phone: _phoneController.text.trim(),
        email: _emailController.text.trim().isEmpty ? null : _emailController.text.trim(),
        dateOfBirth: _selectedDateOfBirth?.toIso8601String().split('T')[0],
        gender: _selectedGender,
        moId: _moIdController.text.trim().isEmpty ? null : _moIdController.text.trim(),
        medicalHistory: _medicalHistoryController.text.trim().isEmpty 
            ? null 
            : _medicalHistoryController.text.trim(),
        allergies: _allergiesController.text.trim().isEmpty 
            ? null 
            : _allergiesController.text.trim(),
        bloodGroup: _selectedBloodGroup,
      );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Patient ${patient.fullName} registered successfully!'),
            backgroundColor: Colors.green,
          ),
        );
        Navigator.pop(context, patient);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Patient Registration'),
        backgroundColor: Colors.blue,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16.0),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Required Fields Section
              const Text(
                'Required Information',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 12),
              
              // First Name
              TextFormField(
                controller: _firstNameController,
                decoration: const InputDecoration(
                  labelText: 'First Name *',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.person),
                ),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'First name is required';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 12),
              
              // Last Name
              TextFormField(
                controller: _lastNameController,
                decoration: const InputDecoration(
                  labelText: 'Last Name *',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.person_outline),
                ),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'Last name is required';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 12),
              
              // Phone
              TextFormField(
                controller: _phoneController,
                decoration: const InputDecoration(
                  labelText: 'Phone Number *',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.phone),
                  hintText: '+971501234567',
                ),
                keyboardType: TextInputType.phone,
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'Phone number is required';
                  }
                  return null;
                },
              ),
              
              const SizedBox(height: 24),
              const Divider(),
              const SizedBox(height: 12),
              
              // Optional Fields Section
              const Text(
                'Optional Information',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 12),
              
              // Email
              TextFormField(
                controller: _emailController,
                decoration: const InputDecoration(
                  labelText: 'Email (Optional)',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.email),
                ),
                keyboardType: TextInputType.emailAddress,
              ),
              const SizedBox(height: 12),
              
              // Date of Birth
              ListTile(
                contentPadding: EdgeInsets.zero,
                leading: const Icon(Icons.cake),
                title: Text(
                  _selectedDateOfBirth == null
                      ? 'Date of Birth (Optional)'
                      : 'DOB: ${_selectedDateOfBirth!.toIso8601String().split('T')[0]}',
                ),
                trailing: const Icon(Icons.calendar_today),
                onTap: () async {
                  final date = await showDatePicker(
                    context: context,
                    initialDate: DateTime.now().subtract(const Duration(days: 365 * 30)),
                    firstDate: DateTime(1900),
                    lastDate: DateTime.now(),
                  );
                  if (date != null) {
                    setState(() => _selectedDateOfBirth = date);
                  }
                },
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(4),
                  side: const BorderSide(color: Colors.grey),
                ),
              ),
              const SizedBox(height: 12),
              
              // Gender
              DropdownButtonFormField<String>(
                value: _selectedGender,
                decoration: const InputDecoration(
                  labelText: 'Gender (Optional)',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.person_pin),
                ),
                items: const [
                  DropdownMenuItem(value: 'male', child: Text('Male')),
                  DropdownMenuItem(value: 'female', child: Text('Female')),
                  DropdownMenuItem(value: 'other', child: Text('Other')),
                ],
                onChanged: (value) => setState(() => _selectedGender = value),
              ),
              const SizedBox(height: 12),
              
              // MO ID
              TextFormField(
                controller: _moIdController,
                decoration: const InputDecoration(
                  labelText: 'MO ID (Optional)',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.badge),
                  hintText: 'MO123456',
                ),
              ),
              const SizedBox(height: 12),
              
              // Blood Group
              DropdownButtonFormField<String>(
                value: _selectedBloodGroup,
                decoration: const InputDecoration(
                  labelText: 'Blood Group (Optional)',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.bloodtype),
                ),
                items: const [
                  DropdownMenuItem(value: 'A+', child: Text('A+')),
                  DropdownMenuItem(value: 'A-', child: Text('A-')),
                  DropdownMenuItem(value: 'B+', child: Text('B+')),
                  DropdownMenuItem(value: 'B-', child: Text('B-')),
                  DropdownMenuItem(value: 'AB+', child: Text('AB+')),
                  DropdownMenuItem(value: 'AB-', child: Text('AB-')),
                  DropdownMenuItem(value: 'O+', child: Text('O+')),
                  DropdownMenuItem(value: 'O-', child: Text('O-')),
                ],
                onChanged: (value) => setState(() => _selectedBloodGroup = value),
              ),
              const SizedBox(height: 12),
              
              // Medical History
              TextFormField(
                controller: _medicalHistoryController,
                decoration: const InputDecoration(
                  labelText: 'Medical History (Optional)',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.medical_information),
                  hintText: 'Diabetes, Hypertension...',
                ),
                maxLines: 2,
              ),
              const SizedBox(height: 12),
              
              // Allergies
              TextFormField(
                controller: _allergiesController,
                decoration: const InputDecoration(
                  labelText: 'Allergies (Optional)',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.warning),
                  hintText: 'Penicillin, Pollen...',
                ),
                maxLines: 2,
              ),
              
              const SizedBox(height: 24),
              
              // Submit Button
              ElevatedButton.icon(
                onPressed: _isLoading ? null : _createPatient,
                icon: _isLoading
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                        ),
                      )
                    : const Icon(Icons.person_add),
                label: Text(
                  _isLoading ? 'Registering...' : 'Register Patient',
                  style: const TextStyle(fontSize: 16),
                ),
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  backgroundColor: Colors.blue,
                  foregroundColor: Colors.white,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  void dispose() {
    _firstNameController.dispose();
    _lastNameController.dispose();
    _phoneController.dispose();
    _emailController.dispose();
    _moIdController.dispose();
    _medicalHistoryController.dispose();
    _allergiesController.dispose();
    super.dispose();
  }
}
```

---

### Step 4: Patient List Screen

```dart
// lib/screens/patient_list_screen.dart

import 'package:flutter/material.dart';
import '../services/clinic_patient_service.dart';

class PatientListScreen extends StatefulWidget {
  final String clinicId;
  final String token;

  const PatientListScreen({
    Key? key,
    required this.clinicId,
    required this.token,
  }) : super(key: key);

  @override
  State<PatientListScreen> createState() => _PatientListScreenState();
}

class _PatientListScreenState extends State<PatientListScreen> {
  late ClinicPatientService _patientService;
  List<ClinicPatient> _patients = [];
  List<ClinicPatient> _filteredPatients = [];
  bool _isLoading = false;
  final _searchController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _patientService = ClinicPatientService(widget.token);
    _loadPatients();
  }

  Future<void> _loadPatients() async {
    setState(() => _isLoading = true);
    
    try {
      final patients = await _patientService.listPatients(widget.clinicId);
      setState(() {
        _patients = patients;
        _filteredPatients = patients;
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error loading patients: $e')),
        );
      }
    } finally {
      setState(() => _isLoading = false);
    }
  }

  void _searchPatients(String query) {
    if (query.isEmpty) {
      setState(() => _filteredPatients = _patients);
      return;
    }

    setState(() {
      _filteredPatients = _patients.where((patient) {
        final searchLower = query.toLowerCase();
        return patient.firstName.toLowerCase().contains(searchLower) ||
            patient.lastName.toLowerCase().contains(searchLower) ||
            patient.phone.contains(query) ||
            (patient.moId?.toLowerCase().contains(searchLower) ?? false);
      }).toList();
    });
  }

  Future<void> _addPatient() async {
    final result = await Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => QuickPatientRegistration(
          clinicId: widget.clinicId,
          token: widget.token,
        ),
      ),
    );

    if (result != null) {
      _loadPatients(); // Refresh list
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Clinic Patients'),
        backgroundColor: Colors.blue,
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadPatients,
          ),
        ],
      ),
      body: Column(
        children: [
          // Search Bar
          Padding(
            padding: const EdgeInsets.all(16.0),
            child: TextField(
              controller: _searchController,
              decoration: InputDecoration(
                hintText: 'Search by name, phone, or MO ID...',
                prefixIcon: const Icon(Icons.search),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
                suffixIcon: _searchController.text.isNotEmpty
                    ? IconButton(
                        icon: const Icon(Icons.clear),
                        onPressed: () {
                          _searchController.clear();
                          _searchPatients('');
                        },
                      )
                    : null,
              ),
              onChanged: _searchPatients,
            ),
          ),
          
          // Patient Count
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16.0),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'Total Patients: ${_filteredPatients.length}',
                  style: const TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                    color: Colors.grey,
                  ),
                ),
                if (_searchController.text.isNotEmpty)
                  Text(
                    'Found: ${_filteredPatients.length}',
                    style: const TextStyle(fontSize: 14, color: Colors.blue),
                  ),
              ],
            ),
          ),
          const SizedBox(height: 8),
          
          // Patient List
          Expanded(
            child: _isLoading
                ? const Center(child: CircularProgressIndicator())
                : _filteredPatients.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.people_outline, size: 64, color: Colors.grey[400]),
                            const SizedBox(height: 16),
                            Text(
                              _searchController.text.isEmpty
                                  ? 'No patients registered yet'
                                  : 'No patients found',
                              style: TextStyle(fontSize: 16, color: Colors.grey[600]),
                            ),
                          ],
                        ),
                      )
                    : ListView.builder(
                        itemCount: _filteredPatients.length,
                        itemBuilder: (context, index) {
                          final patient = _filteredPatients[index];
                          return Card(
                            margin: const EdgeInsets.symmetric(
                              horizontal: 16,
                              vertical: 8,
                            ),
                            child: ListTile(
                              leading: CircleAvatar(
                                backgroundColor: Colors.blue,
                                child: Text(
                                  patient.firstName[0].toUpperCase(),
                                  style: const TextStyle(color: Colors.white),
                                ),
                              ),
                              title: Text(
                                patient.fullName,
                                style: const TextStyle(fontWeight: FontWeight.bold),
                              ),
                              subtitle: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text('📞 ${patient.phone}'),
                                  if (patient.moId != null)
                                    Text('🆔 MO: ${patient.moId}'),
                                  if (patient.bloodGroup != null)
                                    Text('🩸 ${patient.bloodGroup}'),
                                ],
                              ),
                              trailing: const Icon(Icons.arrow_forward_ios, size: 16),
                              onTap: () {
                                // Navigate to patient details or booking
                                _navigateToPatientDetails(patient);
                              },
                            ),
                          );
                        },
                      ),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _addPatient,
        icon: const Icon(Icons.person_add),
        label: const Text('New Patient'),
        backgroundColor: Colors.blue,
      ),
    );
  }

  void _navigateToPatientDetails(ClinicPatient patient) {
    // Navigate to appointment booking or patient details
    showModalBottomSheet(
      context: context,
      builder: (context) => Container(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              patient.fullName,
              style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                const Icon(Icons.phone, size: 16, color: Colors.grey),
                const SizedBox(width: 8),
                Text(patient.phone),
              ],
            ),
            if (patient.email != null) ...[
              const SizedBox(height: 8),
              Row(
                children: [
                  const Icon(Icons.email, size: 16, color: Colors.grey),
                  const SizedBox(width: 8),
                  Text(patient.email!),
                ],
              ),
            ],
            if (patient.moId != null) ...[
              const SizedBox(height: 8),
              Row(
                children: [
                  const Icon(Icons.badge, size: 16, color: Colors.grey),
                  const SizedBox(width: 8),
                  Text('MO ID: ${patient.moId}'),
                ],
              ),
            ],
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: () {
                Navigator.pop(context);
                // Navigate to appointment booking with this patient
                // Navigator.push(context, MaterialPageRoute(
                //   builder: (context) => BookAppointmentScreen(
                //     clinicPatientId: patient.id,
                //     clinicId: widget.clinicId,
                //   ),
                // ));
              },
              icon: const Icon(Icons.calendar_today),
              label: const Text('Book Appointment'),
              style: ElevatedButton.styleFrom(
                minimumSize: const Size(double.infinity, 48),
                backgroundColor: Colors.blue,
                foregroundColor: Colors.white,
              ),
            ),
          ],
        ),
      ),
    );
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }
}
```

---

### Step 5: Book Appointment with Clinic Patient

```dart
// lib/screens/book_appointment_screen.dart

import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

class BookAppointmentScreen extends StatefulWidget {
  final String clinicPatientId;
  final String clinicId;
  final String doctorId;
  final String token;

  const BookAppointmentScreen({
    Key? key,
    required this.clinicPatientId,
    required this.clinicId,
    required this.doctorId,
    required this.token,
  }) : super(key: key);

  @override
  State<BookAppointmentScreen> createState() => _BookAppointmentScreenState();
}

class _BookAppointmentScreenState extends State<BookAppointmentScreen> {
  String? selectedSlotId;
  String? selectedDate;
  bool _isLoading = false;

  Future<void> _bookAppointment() async {
    if (selectedSlotId == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select a time slot')),
      );
      return;
    }

    setState(() => _isLoading = true);

    try {
      final response = await http.post(
        Uri.parse('http://localhost:8082/api/appointments'),
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer ${widget.token}',
        },
        body: jsonEncode({
          'clinic_patient_id': widget.clinicPatientId,  // ✅ Clinic-specific patient
          'doctor_id': widget.doctorId,
          'clinic_id': widget.clinicId,
          'individual_slot_id': selectedSlotId,
          'appointment_date': selectedDate,
          'appointment_time': '$selectedDate 09:30:00',
          'consultation_type': 'offline',
          'payment_mode': 'pay_later',
        }),
      );

      if (response.statusCode == 201) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Appointment booked successfully!'),
              backgroundColor: Colors.green,
            ),
          );
          Navigator.pop(context);
        }
      } else {
        final error = jsonDecode(response.body);
        throw Exception(error['error'] ?? 'Booking failed');
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Book Appointment'),
      ),
      body: Column(
        children: [
          // Slot selection UI here
          // Date picker, time slots, etc.
          
          const Spacer(),
          
          Padding(
            padding: const EdgeInsets.all(16.0),
            child: ElevatedButton(
              onPressed: _isLoading ? null : _bookAppointment,
              child: _isLoading
                  ? const CircularProgressIndicator()
                  : const Text('Confirm Booking'),
            ),
          ),
        ],
      ),
    );
  }
}
```

---

## 📋 Complete API Reference

### 1. Create Patient - Minimal (Name + Phone Only)

**Request:**
```json
POST /clinic-specific-patients
{
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+971501234567"
}
```

**Flutter Code:**
```dart
final patient = await _patientService.createPatient(
  clinicId: currentClinicId,
  firstName: 'John',
  lastName: 'Doe',
  phone: '+971501234567',
);
```

---

### 2. Create Patient - Full (With Optional Fields)

**Request:**
```json
POST /clinic-specific-patients
{
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "first_name": "Sara",
  "last_name": "Ali",
  "phone": "+971507654321",
  "email": "sara.ali@email.com",
  "date_of_birth": "1990-05-15",
  "gender": "female",
  "mo_id": "MO123456",
  "medical_history": "Diabetes",
  "allergies": "Penicillin",
  "blood_group": "A+"
}
```

**Flutter Code:**
```dart
final patient = await _patientService.createPatientFull(
  clinicId: currentClinicId,
  firstName: 'Sara',
  lastName: 'Ali',
  phone: '+971507654321',
  email: 'sara.ali@email.com',
  dateOfBirth: '1990-05-15',
  gender: 'female',
  moId: 'MO123456',
  medicalHistory: 'Diabetes',
  allergies: 'Penicillin',
  bloodGroup: 'A+',
);
```

---

### 3. List Patients

**Request:**
```
GET /clinic-specific-patients?clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2
```

**Flutter Code:**
```dart
final patients = await _patientService.listPatients(currentClinicId);

// Display in UI
setState(() {
  _patients = patients;
});
```

---

### 4. Search Patients

**Request:**
```
GET /clinic-specific-patients?clinic_id=xxx&search=Ahmed
```

**Flutter Code:**
```dart
final results = await _patientService.searchPatients(
  currentClinicId,
  searchQuery,
);

setState(() {
  _filteredPatients = results;
});
```

---

### 5. Update Patient

**Request:**
```json
PUT /clinic-specific-patients/f47ac10b-58cc-4372-a567-0e02b2c3d479
{
  "email": "updated@email.com",
  "blood_group": "O+",
  "medical_history": "Updated history"
}
```

**Flutter Code:**
```dart
final updatedPatient = await _patientService.updatePatient(
  patientId,
  {
    'email': 'updated@email.com',
    'blood_group': 'O+',
    'medical_history': 'Updated history',
  },
);
```

---

## 🎨 Complete Flutter App Example

### main.dart

```dart
import 'package:flutter/material.dart';
import 'screens/patient_list_screen.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Clinic Patient Management',
      theme: ThemeData(
        primarySwatch: Colors.blue,
        useMaterial3: true,
      ),
      home: const PatientListScreen(
        clinicId: '7a6c1211-c029-4923-a1a6-fe3dfe48bdf2',
        token: 'your-jwt-token-here',
      ),
    );
  }
}
```

---

## 💡 Best Practices

### 1. Quick Registration Flow

```dart
// Step 1: Quick form (name + phone only)
showDialog(
  context: context,
  builder: (context) => AlertDialog(
    title: const Text('Quick Registration'),
    content: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        TextField(
          controller: firstNameController,
          decoration: const InputDecoration(labelText: 'First Name *'),
        ),
        TextField(
          controller: lastNameController,
          decoration: const InputDecoration(labelText: 'Last Name *'),
        ),
        TextField(
          controller: phoneController,
          decoration: const InputDecoration(labelText: 'Phone *'),
          keyboardType: TextInputType.phone,
        ),
      ],
    ),
    actions: [
      TextButton(
        onPressed: () => Navigator.pop(context),
        child: const Text('Cancel'),
      ),
      ElevatedButton(
        onPressed: () async {
          // Create patient
          final patient = await createPatient(...);
          Navigator.pop(context);
          // Proceed to booking
        },
        child: const Text('Register & Continue'),
      ),
    ],
  ),
);
```

---

### 2. Search with Debouncing

```dart
import 'dart:async';

Timer? _debounce;

void _onSearchChanged(String query) {
  if (_debounce?.isActive ?? false) _debounce!.cancel();
  
  _debounce = Timer(const Duration(milliseconds: 500), () {
    _searchPatients(query);
  });
}

// In TextField
TextField(
  onChanged: _onSearchChanged,  // Waits 500ms before searching
  decoration: const InputDecoration(
    hintText: 'Search patients...',
    prefixIcon: Icon(Icons.search),
  ),
)
```

---

### 3. Error Handling

```dart
try {
  final patient = await _patientService.createPatient(...);
  // Success
} on Exception catch (e) {
  final errorMessage = e.toString();
  
  if (errorMessage.contains('Phone number exists')) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Patient Exists'),
        content: const Text(
          'A patient with this phone number already exists in your clinic.'
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('OK'),
          ),
        ],
      ),
    );
  } else {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('Error: $errorMessage')),
    );
  }
}
```

---

## 📊 Field Requirements

### Required Fields (✅ Must Provide)
| Field | Type | Example | Validation |
|-------|------|---------|------------|
| `clinic_id` | UUID | `7a6c1211-...` | Valid UUID |
| `first_name` | String | `Ahmed` | Max 100 chars |
| `last_name` | String | `Khan` | Max 100 chars |
| `phone` | String | `+971501234567` | Max 20 chars |

### Optional Fields (All Can Be Null)
| Field | Type | Example | Notes |
|-------|------|---------|-------|
| `email` | String? | `email@example.com` | Must be valid email if provided |
| `date_of_birth` | String? | `1990-05-15` | YYYY-MM-DD format |
| `gender` | String? | `male` / `female` | Max 20 chars |
| `mo_id` | String? | `MO123456` | Unique per clinic |
| `medical_history` | String? | `Diabetes` | Text field |
| `allergies` | String? | `Penicillin` | Text field |
| `blood_group` | String? | `O+` | A+, A-, B+, B-, AB+, AB-, O+, O- |

---

## ✅ Complete Workflow Example

### Scenario: Receptionist Registers Patient & Books Appointment

```dart
Future<void> registerAndBookWorkflow() async {
  // Step 1: Create patient (quick)
  final patient = await clinicPatientService.createPatient(
    clinicId: currentClinicId,
    firstName: 'Ahmed',
    lastName: 'Khan',
    phone: '+971501234567',
  );
  
  print('Patient created: ${patient.id}');
  
  // Step 2: Book appointment
  final appointment = await appointmentService.createAppointment(
    clinicPatientId: patient.id,  // ✅ Use clinic patient ID
    doctorId: selectedDoctorId,
    clinicId: currentClinicId,
    individualSlotId: selectedSlotId,
    appointmentDate: '2025-10-18',
    appointmentTime: '2025-10-18 09:30:00',
    consultationType: 'offline',
  );
  
  print('Appointment booked: ${appointment.bookingNumber}');
  
  // Step 3: Show confirmation
  showDialog(
    context: context,
    builder: (context) => AlertDialog(
      title: const Text('✅ Booking Confirmed'),
      content: Text(
        'Patient: ${patient.fullName}\n'
        'Booking: ${appointment.bookingNumber}\n'
        'Date: 2025-10-18\n'
        'Time: 09:30',
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('OK'),
        ),
      ],
    ),
  );
}
```

---

## 🚀 Summary

### Required Fields (Minimum)
```json
{
  "clinic_id": "UUID",
  "first_name": "string",
  "last_name": "string",
  "phone": "string"
}
```

### Optional Fields
```json
{
  "email": "string (optional)",
  "date_of_birth": "YYYY-MM-DD (optional)",
  "gender": "string (optional)",
  "mo_id": "string (optional)",
  "medical_history": "text (optional)",
  "allergies": "text (optional)",
  "blood_group": "string (optional)"
}
```

### Flutter Packages Needed
```yaml
dependencies:
  flutter:
    sdk: flutter
  http: ^1.1.0
```

---

## ✅ Status

| Component | Status |
|-----------|--------|
| API endpoints | ✅ Working |
| Minimal creation (name + phone) | ✅ Supported |
| Full creation (all fields) | ✅ Supported |
| Search functionality | ✅ Working |
| Flutter service class | ✅ Provided |
| Flutter UI examples | ✅ Provided |
| Complete workflow | ✅ Documented |

---

**API:** `POST /api/organizations/clinic-specific-patients`  
**Minimum Required:** Only `clinic_id`, `first_name`, `last_name`, `phone`  
**All other fields:** Optional ✅

**Status:** ✅ **Complete Flutter Integration Guide Ready!** 🎉

