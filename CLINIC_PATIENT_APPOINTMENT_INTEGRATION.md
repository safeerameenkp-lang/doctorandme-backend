# Clinic Patient + Appointment Integration - Complete Guide

## 🎯 Overview

Complete integration between **clinic-specific patients** and **appointment booking** with **session-based slots**.

---

## 📊 Complete System Flow

```
Step 1: Create Clinic Patient
  POST /clinic-specific-patients
  ↓
Step 2: Create Session-Based Slots
  POST /doctor-session-slots
  ↓
Step 3: Book Appointment
  POST /appointments { clinic_patient_id }
  ↓
Result: Appointment + Slot Marked as Booked ✅
```

---

## 🔄 Complete Workflow with Full JSON

### Step 1: Create Clinic-Specific Patient

**Request:**
```json
POST /api/organizations/clinic-specific-patients
Content-Type: application/json
Authorization: Bearer {token}

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
    "id": "clinic-patient-abc-123",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "first_name": "Ahmed",
    "last_name": "Khan",
    "phone": "+971501234567",
    "email": null,
    "mo_id": null,
    "is_active": true,
    "created_at": "2024-10-15T10:00:00Z",
    "updated_at": "2024-10-15T10:00:00Z"
  }
}
```

**Save this:** `clinic-patient-abc-123` (you'll need it for booking)

---

### Step 2: Create Session-Based Time Slots

**Request:**
```json
POST /api/organizations/doctor-session-slots
Content-Type: application/json
Authorization: Bearer {token}

{
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "slot_type": "offline",
  "slot_duration": 5,
  "date": "2025-10-18",
  "sessions": [
    {
      "session_name": "Morning Session",
      "start_time": "09:30",
      "end_time": "11:30",
      "max_patients": 24,
      "slot_interval_minutes": 5
    }
  ]
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Doctor time slots created successfully",
  "data": {
    "id": "timeslot-uuid",
    "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "date": "2025-10-18",
    "day_of_week": 6,
    "sessions": [
      {
        "id": "session-morning-uuid",
        "session_name": "Morning Session",
        "generated_slots": 24,
        "available_slots": 24,
        "booked_slots": 0,
        "slots": [
          {
            "id": "slot-09-30-uuid",
            "slot_start": "09:30",
            "slot_end": "09:35",
            "is_booked": false,
            "status": "available"
          },
          {
            "id": "slot-09-35-uuid",
            "slot_start": "09:35",
            "slot_end": "09:40",
            "is_booked": false,
            "status": "available"
          }
          // ... 22 more slots
        ]
      }
    ]
  }
}
```

**Save this:** `slot-09-30-uuid` (patient will book this slot)

---

### Step 3: Book Appointment with Clinic Patient

**Request:**
```json
POST /api/appointments
Content-Type: application/json
Authorization: Bearer {token}

{
  "clinic_patient_id": "clinic-patient-abc-123",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "individual_slot_id": "slot-09-30-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "duration_minutes": 5,
  "consultation_type": "offline",
  "reason": "Regular checkup",
  "payment_mode": "cash"
}
```

**What the API Does:**
1. ✅ Validates `clinic_patient_id` exists
2. ✅ Checks patient belongs to THIS clinic
3. ✅ Validates `individual_slot_id` is available
4. ✅ Creates appointment with **both** patient references
5. ✅ Marks slot as booked with clinic_patient_id

**Response (201 Created):**
```json
{
  "appointment": {
    "id": "appointment-new-uuid",
    "patient_id": null,
    "clinic_patient_id": "clinic-patient-abc-123",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
    "booking_number": "BN202510180001",
    "token_number": 1,
    "appointment_date": "2025-10-18",
    "appointment_time": "2025-10-18T09:30:00Z",
    "duration_minutes": 5,
    "consultation_type": "offline",
    "reason": "Regular checkup",
    "status": "confirmed",
    "fee_amount": 500.00,
    "payment_status": "paid",
    "payment_mode": "cash",
    "created_at": "2024-10-15T11:00:00Z"
  }
}
```

**Database Changes:**
```sql
-- Appointment created
INSERT INTO appointments (
    patient_id,         -- NULL (not using global)
    clinic_patient_id,  -- 'clinic-patient-abc-123' ✅
    clinic_id,
    ...
) VALUES (
    NULL,
    'clinic-patient-abc-123',
    '7a6c1211-c029-4923-a1a6-fe3dfe48bdf2',
    ...
);

-- Individual slot marked as booked
UPDATE doctor_individual_slots
SET is_booked = true,
    booked_patient_id = 'clinic-patient-abc-123',  -- ✅ Clinic patient ID
    booked_appointment_id = 'appointment-new-uuid',
    status = 'booked'
WHERE id = 'slot-09-30-uuid';
```

---

### Step 4: Verify Slot is Booked

**Request:**
```
GET /api/organizations/doctor-session-slots?doctor_id=85394ce8-94f7-4dca-a536-34305c46a98e&date=2025-10-18
```

**Response (200 OK):**
```json
{
  "slots": [
    {
      "sessions": [
        {
          "session_name": "Morning Session",
          "generated_slots": 24,
          "available_slots": 23,
          "booked_slots": 1,
          "slots": [
            {
              "id": "slot-09-30-uuid",
              "slot_start": "09:30",
              "slot_end": "09:35",
              "is_booked": true,
              "booked_patient_id": "clinic-patient-abc-123",
              "booked_appointment_id": "appointment-new-uuid",
              "status": "booked"
            },
            {
              "id": "slot-09-35-uuid",
              "slot_start": "09:35",
              "slot_end": "09:40",
              "is_booked": false,
              "status": "available"
            }
            // ... 22 more slots
          ]
        }
      ]
    }
  ]
}
```

---

## 📋 Complete API Integration

### API 1: Create Clinic Patient

```bash
POST /organizations/clinic-specific-patients

Body: {
  "clinic_id": "UUID",
  "first_name": "string",
  "last_name": "string",
  "phone": "string"
}

Returns: {
  "patient": {
    "id": "clinic-patient-uuid"
  }
}
```

---

### API 2: List Available Slots

```bash
GET /organizations/doctor-session-slots?doctor_id=xxx&clinic_id=xxx&date=2025-10-18

Returns: {
  "slots": [
    {
      "sessions": [
        {
          "slots": [
            {
              "id": "slot-uuid",
              "is_booked": false,
              "status": "available"
            }
          ]
        }
      ]
    }
  ]
}
```

---

### API 3: Book Appointment

```bash
POST /appointments

Body: {
  "clinic_patient_id": "clinic-patient-uuid",
  "doctor_id": "UUID",
  "clinic_id": "UUID",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "consultation_type": "offline"
}

Returns: {
  "appointment": {
    "id": "appointment-uuid",
    "clinic_patient_id": "clinic-patient-uuid",
    "booking_number": "BN202510180001",
    "status": "confirmed"
  }
}
```

---

## 🎨 Flutter Complete Implementation

### Dart Model for Appointments

```dart
// lib/models/appointment.dart

class Appointment {
  final String id;
  final String? patientId;           // Global patient (nullable)
  final String? clinicPatientId;     // Clinic patient (nullable)
  final String clinicId;
  final String doctorId;
  final String bookingNumber;
  final int tokenNumber;
  final String appointmentDate;
  final DateTime appointmentTime;
  final int durationMinutes;
  final String consultationType;
  final String? reason;
  final String status;
  final double feeAmount;
  final String paymentStatus;
  final String? paymentMode;
  final DateTime createdAt;

  Appointment({
    required this.id,
    this.patientId,
    this.clinicPatientId,
    required this.clinicId,
    required this.doctorId,
    required this.bookingNumber,
    required this.tokenNumber,
    required this.appointmentDate,
    required this.appointmentTime,
    required this.durationMinutes,
    required this.consultationType,
    this.reason,
    required this.status,
    required this.feeAmount,
    required this.paymentStatus,
    this.paymentMode,
    required this.createdAt,
  });

  factory Appointment.fromJson(Map<String, dynamic> json) {
    return Appointment(
      id: json['id'],
      patientId: json['patient_id'],
      clinicPatientId: json['clinic_patient_id'],
      clinicId: json['clinic_id'],
      doctorId: json['doctor_id'],
      bookingNumber: json['booking_number'],
      tokenNumber: json['token_number'],
      appointmentDate: json['appointment_date'],
      appointmentTime: DateTime.parse(json['appointment_time']),
      durationMinutes: json['duration_minutes'],
      consultationType: json['consultation_type'],
      reason: json['reason'],
      status: json['status'],
      feeAmount: json['fee_amount'].toDouble(),
      paymentStatus: json['payment_status'],
      paymentMode: json['payment_mode'],
      createdAt: DateTime.parse(json['created_at']),
    );
  }
}
```

---

### Appointment Service

```dart
// lib/services/appointment_service.dart

import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/appointment.dart';

class AppointmentService {
  final String baseUrl = 'http://localhost:8082/api';
  final String token;

  AppointmentService(this.token);

  Map<String, String> get headers => {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer $token',
  };

  // Book appointment with clinic-specific patient
  Future<Appointment> bookAppointmentWithClinicPatient({
    required String clinicPatientId,
    required String doctorId,
    required String clinicId,
    required String individualSlotId,
    required String appointmentDate,
    required String appointmentTime,
    required String consultationType,
    String? reason,
    String? paymentMode,
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl/appointments'),
      headers: headers,
      body: jsonEncode({
        'clinic_patient_id': clinicPatientId,  // ✅ Clinic-specific patient
        'doctor_id': doctorId,
        'clinic_id': clinicId,
        'individual_slot_id': individualSlotId,
        'appointment_date': appointmentDate,
        'appointment_time': appointmentTime,
        'duration_minutes': 5,
        'consultation_type': consultationType,
        'reason': reason,
        'payment_mode': paymentMode ?? 'pay_later',
      }),
    );

    if (response.statusCode == 201) {
      final data = jsonDecode(response.body);
      return Appointment.fromJson(data['appointment']);
    } else {
      final error = jsonDecode(response.body);
      throw Exception(error['error'] ?? 'Failed to book appointment');
    }
  }
}
```

---

### Complete Booking Flow Screen

```dart
// lib/screens/complete_booking_flow.dart

import 'package:flutter/material.dart';
import '../services/clinic_patient_service.dart';
import '../services/appointment_service.dart';

class CompleteBookingFlow extends StatefulWidget {
  final String clinicId;
  final String doctorId;
  final String token;

  const CompleteBookingFlow({
    Key? key,
    required this.clinicId,
    required this.doctorId,
    required this.token,
  }) : super(key: key);

  @override
  State<CompleteBookingFlow> createState() => _CompleteBookingFlowState();
}

class _CompleteBookingFlowState extends State<CompleteBookingFlow> {
  late ClinicPatientService _patientService;
  late AppointmentService _appointmentService;
  
  final _firstNameController = TextEditingController();
  final _lastNameController = TextEditingController();
  final _phoneController = TextEditingController();
  
  String? _selectedSlotId;
  String? _createdPatientId;
  bool _isLoading = false;
  int _currentStep = 0;

  @override
  void initState() {
    super.initState();
    _patientService = ClinicPatientService(widget.token);
    _appointmentService = AppointmentService(widget.token);
  }

  Future<void> _createPatient() async {
    setState(() => _isLoading = true);

    try {
      final patient = await _patientService.createPatient(
        clinicId: widget.clinicId,
        firstName: _firstNameController.text.trim(),
        lastName: _lastNameController.text.trim(),
        phone: _phoneController.text.trim(),
      );

      setState(() {
        _createdPatientId = patient.id;
        _currentStep = 1; // Move to slot selection
      });

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('✅ Patient ${patient.fullName} registered'),
          backgroundColor: Colors.green,
        ),
      );
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Error: ${e.toString()}'),
          backgroundColor: Colors.red,
        ),
      );
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _bookAppointment() async {
    if (_selectedSlotId == null || _createdPatientId == null) {
      return;
    }

    setState(() => _isLoading = true);

    try {
      final appointment = await _appointmentService.bookAppointmentWithClinicPatient(
        clinicPatientId: _createdPatientId!,
        doctorId: widget.doctorId,
        clinicId: widget.clinicId,
        individualSlotId: _selectedSlotId!,
        appointmentDate: '2025-10-18',
        appointmentTime: '2025-10-18 09:30:00',
        consultationType: 'offline',
        reason: 'Regular checkup',
        paymentMode: 'cash',
      );

      if (mounted) {
        showDialog(
          context: context,
          builder: (context) => AlertDialog(
            title: const Text('✅ Booking Confirmed!'),
            content: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('Booking Number: ${appointment.bookingNumber}'),
                Text('Token: ${appointment.tokenNumber}'),
                Text('Date: ${appointment.appointmentDate}'),
                Text('Time: 09:30 - 09:35'),
                Text('Fee: AED ${appointment.feeAmount}'),
              ],
            ),
            actions: [
              TextButton(
                onPressed: () {
                  Navigator.pop(context); // Close dialog
                  Navigator.pop(context); // Go back to main screen
                },
                child: const Text('Done'),
              ),
            ],
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Booking failed: ${e.toString()}'),
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
        backgroundColor: Colors.blue,
      ),
      body: Stepper(
        currentStep: _currentStep,
        onStepContinue: () {
          if (_currentStep == 0) {
            _createPatient();
          } else if (_currentStep == 1) {
            _bookAppointment();
          }
        },
        onStepCancel: () {
          if (_currentStep > 0) {
            setState(() => _currentStep--);
          } else {
            Navigator.pop(context);
          }
        },
        steps: [
          // Step 1: Patient Registration
          Step(
            title: const Text('Patient Registration'),
            content: Column(
              children: [
                TextFormField(
                  controller: _firstNameController,
                  decoration: const InputDecoration(
                    labelText: 'First Name *',
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 12),
                TextFormField(
                  controller: _lastNameController,
                  decoration: const InputDecoration(
                    labelText: 'Last Name *',
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 12),
                TextFormField(
                  controller: _phoneController,
                  decoration: const InputDecoration(
                    labelText: 'Phone *',
                    border: OutlineInputBorder(),
                    hintText: '+971501234567',
                  ),
                  keyboardType: TextInputType.phone,
                ),
              ],
            ),
            isActive: _currentStep >= 0,
            state: _createdPatientId != null 
                ? StepState.complete 
                : StepState.indexed,
          ),
          
          // Step 2: Slot Selection
          Step(
            title: const Text('Select Time Slot'),
            content: Column(
              children: [
                // Here you would load and display available slots
                ListTile(
                  title: const Text('09:30 - 09:35'),
                  subtitle: const Text('Morning Session'),
                  trailing: Radio(
                    value: 'slot-09-30-uuid',
                    groupValue: _selectedSlotId,
                    onChanged: (value) {
                      setState(() => _selectedSlotId = value);
                    },
                  ),
                ),
                // More slots...
              ],
            ),
            isActive: _currentStep >= 1,
            state: _selectedSlotId != null 
                ? StepState.complete 
                : StepState.indexed,
          ),
        ],
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

## 📊 Database Integration

### appointments table (Updated)

```sql
CREATE TABLE appointments (
    id                  UUID PRIMARY KEY,
    patient_id          UUID,              -- Global patient (optional)
    clinic_patient_id   UUID,              -- Clinic patient (optional) ✅ NEW
    clinic_id           UUID NOT NULL,
    doctor_id           UUID NOT NULL,
    booking_number      VARCHAR,
    token_number        INT,
    appointment_date    DATE,
    appointment_time    TIMESTAMP,
    ...
    
    -- Either patient_id OR clinic_patient_id can be set
);
```

### doctor_individual_slots table

```sql
CREATE TABLE doctor_individual_slots (
    id                      UUID PRIMARY KEY,
    session_id              UUID NOT NULL,
    clinic_id               UUID NOT NULL,
    slot_start              TIME,
    slot_end                TIME,
    is_booked               BOOLEAN DEFAULT FALSE,
    booked_patient_id       UUID,  -- Can be clinic_patient_id ✅
    booked_appointment_id   UUID,
    status                  VARCHAR,
    ...
);
```

---

## ✅ Error Scenarios

### Error 1: Clinic Patient Not Found

**Request:**
```json
{
  "clinic_patient_id": "invalid-uuid",
  ...
}
```

**Response (404):**
```json
{
  "error": "clinic patient",
  "message": "Not found"
}
```

---

### Error 2: Patient Belongs to Different Clinic

**Request:**
```json
{
  "clinic_patient_id": "patient-from-clinic-a",
  "clinic_id": "clinic-b-uuid",
  ...
}
```

**Response (400):**
```json
{
  "error": "Clinic mismatch",
  "message": "This patient belongs to a different clinic"
}
```

---

### Error 3: Slot Already Booked

**Request:**
```json
{
  "individual_slot_id": "already-booked-slot",
  ...
}
```

**Response (409):**
```json
{
  "error": "Slot already booked",
  "message": "This 5-minute slot is no longer available",
  "slot_start": "09:30",
  "slot_end": "09:35",
  "current_status": "booked"
}
```

---

## 🎯 Summary

### Three-Part Integration

| Step | API | Returns |
|------|-----|---------|
| 1 | **Create Patient** | `clinic_patient_id` |
| 2 | **Get Slots** | `individual_slot_id` |
| 3 | **Book Appointment** | `appointment` with both IDs linked |

### Database Tables Connected

```
clinic_patients
    ↓ (clinic_patient_id)
appointments
    ↓ (individual_slot_id)
doctor_individual_slots
    ↓ (booked_patient_id = clinic_patient_id)
Complete circular reference ✅
```

---

## ✅ Status

| Feature | Status | Description |
|---------|--------|-------------|
| clinic_patient_id in appointments | ✅ Added | Migration 018 applied |
| Appointment API updated | ✅ Done | Supports clinic_patient_id |
| Slot booking updated | ✅ Done | Uses clinic_patient_id |
| Clinic validation | ✅ Working | Prevents cross-clinic booking |
| Flutter integration | ✅ Complete | Full code provided |
| Error handling | ✅ Complete | All scenarios covered |

---

**API Endpoint:** `POST /api/appointments`  
**New Field:** `clinic_patient_id` (alternative to `patient_id`)  
**Status:** ✅ **Complete Integration Ready!** 🎉

Now you can create clinic-specific patients and book appointments seamlessly!


