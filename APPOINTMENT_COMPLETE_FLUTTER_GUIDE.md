# Complete Appointment Booking - Flutter Integration Guide

## 📖 Table of Contents
1. [API Overview](#api-overview)
2. [Create Appointment - Full JSON Examples](#create-appointment)
3. [Flutter Models](#flutter-models)
4. [Flutter Service](#flutter-service)
5. [Flutter UI Screens](#flutter-ui-screens)
6. [Complete Workflow](#complete-workflow)

---

## 🎯 API Overview

**Endpoint:** `POST /api/appointments`

**Purpose:** Book an appointment with clinic-specific patient and session-based slot

**Required Fields:**
- Patient identification (clinic_patient_id)
- Doctor ID
- Clinic ID
- Individual slot ID (for 5-min slots)
- Appointment date & time
- Consultation type

---

## 📝 Create Appointment - Full JSON Examples

### Example 1: Minimal Appointment (Required Fields Only)

**Request:**
```json
POST /api/appointments
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "clinic_patient_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "individual_slot_id": "slot-09-30-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "consultation_type": "offline"
}
```

**Response (201 Created):**
```json
{
  "appointment": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "patient_id": null,
    "clinic_patient_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
    "department_id": null,
    "booking_number": "BN202510180001",
    "token_number": 1,
    "appointment_date": "2025-10-18",
    "appointment_time": "2025-10-18T09:30:00Z",
    "duration_minutes": 12,
    "consultation_type": "offline",
    "reason": null,
    "notes": null,
    "status": "confirmed",
    "fee_amount": 500.00,
    "payment_status": "pending",
    "payment_mode": null,
    "is_priority": false,
    "created_at": "2024-10-15T12:00:00Z"
  }
}
```

---

### Example 2: Full Appointment (All Fields)

**Request:**
```json
POST /api/appointments
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "clinic_patient_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "department_id": "dept-uuid-123",
  "individual_slot_id": "slot-09-35-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:35:00",
  "duration_minutes": 5,
  "consultation_type": "offline",
  "reason": "Follow-up consultation for diabetes management",
  "notes": "Patient requested morning appointment",
  "is_priority": false,
  "payment_mode": "cash"
}
```

**Response (201 Created):**
```json
{
  "appointment": {
    "id": "b2c3d4e5-f6a7-8901-bcde-f2345678901a",
    "patient_id": null,
    "clinic_patient_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
    "department_id": "dept-uuid-123",
    "booking_number": "BN202510180002",
    "token_number": 2,
    "appointment_date": "2025-10-18",
    "appointment_time": "2025-10-18T09:35:00Z",
    "duration_minutes": 5,
    "consultation_type": "offline",
    "reason": "Follow-up consultation for diabetes management",
    "notes": "Patient requested morning appointment",
    "status": "confirmed",
    "fee_amount": 500.00,
    "payment_status": "paid",
    "payment_mode": "cash",
    "is_priority": false,
    "created_at": "2024-10-15T12:05:00Z"
  }
}
```

---

### Example 3: Online Consultation Appointment

**Request:**
```json
POST /api/appointments
Content-Type: application/json
Authorization: Bearer {token}

{
  "clinic_patient_id": "c3d4e5f6-a7b8-9012-cdef-3456789012ab",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "individual_slot_id": "online-slot-10-00-uuid",
  "appointment_date": "2025-10-19",
  "appointment_time": "2025-10-19 10:00:00",
  "duration_minutes": 10,
  "consultation_type": "online",
  "reason": "Online video consultation",
  "payment_mode": "card"
}
```

**Response (201 Created):**
```json
{
  "appointment": {
    "id": "c4d5e6f7-a8b9-0123-def0-456789012bcd",
    "clinic_patient_id": "c3d4e5f6-a7b8-9012-cdef-3456789012ab",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
    "booking_number": "BN202510190001",
    "token_number": 1,
    "appointment_date": "2025-10-19",
    "appointment_time": "2025-10-19T10:00:00Z",
    "duration_minutes": 10,
    "consultation_type": "online",
    "reason": "Online video consultation",
    "status": "confirmed",
    "fee_amount": 300.00,
    "payment_status": "paid",
    "payment_mode": "card",
    "is_priority": false,
    "created_at": "2024-10-15T12:10:00Z"
  }
}
```

---

## 📱 Flutter Models

### Appointment Model

```dart
// lib/models/appointment.dart

class Appointment {
  final String id;
  final String? patientId;
  final String? clinicPatientId;
  final String clinicId;
  final String doctorId;
  final String? departmentId;
  final String bookingNumber;
  final int tokenNumber;
  final String appointmentDate;
  final DateTime appointmentTime;
  final int durationMinutes;
  final String consultationType;
  final String? reason;
  final String? notes;
  final String status;
  final double feeAmount;
  final String paymentStatus;
  final String? paymentMode;
  final bool isPriority;
  final DateTime createdAt;

  Appointment({
    required this.id,
    this.patientId,
    this.clinicPatientId,
    required this.clinicId,
    required this.doctorId,
    this.departmentId,
    required this.bookingNumber,
    required this.tokenNumber,
    required this.appointmentDate,
    required this.appointmentTime,
    required this.durationMinutes,
    required this.consultationType,
    this.reason,
    this.notes,
    required this.status,
    required this.feeAmount,
    required this.paymentStatus,
    this.paymentMode,
    required this.isPriority,
    required this.createdAt,
  });

  factory Appointment.fromJson(Map<String, dynamic> json) {
    return Appointment(
      id: json['id'],
      patientId: json['patient_id'],
      clinicPatientId: json['clinic_patient_id'],
      clinicId: json['clinic_id'],
      doctorId: json['doctor_id'],
      departmentId: json['department_id'],
      bookingNumber: json['booking_number'],
      tokenNumber: json['token_number'],
      appointmentDate: json['appointment_date'],
      appointmentTime: DateTime.parse(json['appointment_time']),
      durationMinutes: json['duration_minutes'],
      consultationType: json['consultation_type'],
      reason: json['reason'],
      notes: json['notes'],
      status: json['status'],
      feeAmount: (json['fee_amount'] as num).toDouble(),
      paymentStatus: json['payment_status'],
      paymentMode: json['payment_mode'],
      isPriority: json['is_priority'] ?? false,
      createdAt: DateTime.parse(json['created_at']),
    );
  }

  // Helper methods
  String get formattedDate => appointmentDate;
  String get formattedTime => '${appointmentTime.hour.toString().padLeft(2, '0')}:${appointmentTime.minute.toString().padLeft(2, '0')}';
  String get statusColor {
    switch (status) {
      case 'confirmed':
        return 'green';
      case 'completed':
        return 'blue';
      case 'cancelled':
        return 'red';
      default:
        return 'grey';
    }
  }
}
```

---

## 🔌 Flutter Service

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

  /// Create appointment with clinic-specific patient (minimal)
  Future<Appointment> createAppointment({
    required String clinicPatientId,
    required String doctorId,
    required String clinicId,
    required String individualSlotId,
    required String appointmentDate,
    required String appointmentTime,
    required String consultationType,
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl/appointments'),
      headers: headers,
      body: jsonEncode({
        'clinic_patient_id': clinicPatientId,
        'doctor_id': doctorId,
        'clinic_id': clinicId,
        'individual_slot_id': individualSlotId,
        'appointment_date': appointmentDate,
        'appointment_time': appointmentTime,
        'consultation_type': consultationType,
      }),
    );

    if (response.statusCode == 201) {
      final data = jsonDecode(response.body);
      return Appointment.fromJson(data['appointment']);
    } else {
      final error = jsonDecode(response.body);
      throw Exception(error['error'] ?? 'Failed to create appointment');
    }
  }

  /// Create appointment with all optional fields
  Future<Appointment> createAppointmentFull({
    required String clinicPatientId,
    required String doctorId,
    required String clinicId,
    required String individualSlotId,
    required String appointmentDate,
    required String appointmentTime,
    required String consultationType,
    String? departmentId,
    int? durationMinutes,
    String? reason,
    String? notes,
    bool? isPriority,
    String? paymentMode,
  }) async {
    final body = {
      'clinic_patient_id': clinicPatientId,
      'doctor_id': doctorId,
      'clinic_id': clinicId,
      'individual_slot_id': individualSlotId,
      'appointment_date': appointmentDate,
      'appointment_time': appointmentTime,
      'consultation_type': consultationType,
    };

    if (departmentId != null) body['department_id'] = departmentId;
    if (durationMinutes != null) body['duration_minutes'] = durationMinutes;
    if (reason != null) body['reason'] = reason;
    if (notes != null) body['notes'] = notes;
    if (isPriority != null) body['is_priority'] = isPriority;
    if (paymentMode != null) body['payment_mode'] = paymentMode;

    final response = await http.post(
      Uri.parse('$baseUrl/appointments'),
      headers: headers,
      body: jsonEncode(body),
    );

    if (response.statusCode == 201) {
      final data = jsonDecode(response.body);
      return Appointment.fromJson(data['appointment']);
    } else {
      final error = jsonDecode(response.body);
      throw Exception(error['error'] ?? 'Failed to create appointment');
    }
  }

  /// List appointments for a patient
  Future<List<Appointment>> listAppointmentsByPatient(String clinicPatientId) async {
    final response = await http.get(
      Uri.parse('$baseUrl/appointments?clinic_patient_id=$clinicPatientId'),
      headers: headers,
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      final List appointmentsJson = data['appointments'] ?? [];
      return appointmentsJson.map((json) => Appointment.fromJson(json)).toList();
    } else {
      throw Exception('Failed to load appointments');
    }
  }

  /// List appointments for a clinic
  Future<List<Appointment>> listAppointmentsByClinic(String clinicId, {String? date}) async {
    String url = '$baseUrl/appointments?clinic_id=$clinicId';
    if (date != null) {
      url += '&date=$date';
    }

    final response = await http.get(
      Uri.parse(url),
      headers: headers,
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      final List appointmentsJson = data['appointments'] ?? [];
      return appointmentsJson.map((json) => Appointment.fromJson(json)).toList();
    } else {
      throw Exception('Failed to load appointments');
    }
  }
}
```

---

## 🎨 Flutter UI Screens

### Screen 1: Appointment Booking Form

```dart
// lib/screens/book_appointment_screen.dart

import 'package:flutter/material.dart';
import '../services/appointment_service.dart';
import '../models/appointment.dart';

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
  late AppointmentService _appointmentService;
  
  DateTime _selectedDate = DateTime.now();
  String? _selectedSlotId;
  String? _selectedSlotTime;
  String _consultationType = 'offline';
  final _reasonController = TextEditingController();
  final _notesController = TextEditingController();
  String _paymentMode = 'pay_later';
  bool _isPriority = false;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _appointmentService = AppointmentService(widget.token);
  }

  Future<void> _bookAppointment() async {
    if (_selectedSlotId == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please select a time slot'),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    setState(() => _isLoading = true);

    try {
      final formattedDate = '${_selectedDate.year}-${_selectedDate.month.toString().padLeft(2, '0')}-${_selectedDate.day.toString().padLeft(2, '0')}';
      final appointmentTime = '$formattedDate $_selectedSlotTime:00';

      final appointment = await _appointmentService.createAppointmentFull(
        clinicPatientId: widget.clinicPatientId,
        doctorId: widget.doctorId,
        clinicId: widget.clinicId,
        individualSlotId: _selectedSlotId!,
        appointmentDate: formattedDate,
        appointmentTime: appointmentTime,
        consultationType: _consultationType,
        durationMinutes: 5,
        reason: _reasonController.text.trim().isEmpty ? null : _reasonController.text.trim(),
        notes: _notesController.text.trim().isEmpty ? null : _notesController.text.trim(),
        isPriority: _isPriority,
        paymentMode: _paymentMode,
      );

      if (mounted) {
        // Show success dialog
        showDialog(
          context: context,
          barrierDismissible: false,
          builder: (context) => AlertDialog(
            title: Row(
              children: const [
                Icon(Icons.check_circle, color: Colors.green, size: 32),
                SizedBox(width: 12),
                Text('Booking Confirmed!'),
              ],
            ),
            content: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildInfoRow('Booking Number', appointment.bookingNumber),
                _buildInfoRow('Token Number', '#${appointment.tokenNumber}'),
                _buildInfoRow('Date', appointment.formattedDate),
                _buildInfoRow('Time', appointment.formattedTime),
                _buildInfoRow('Type', appointment.consultationType.toUpperCase()),
                _buildInfoRow('Fee', 'AED ${appointment.feeAmount.toStringAsFixed(2)}'),
                _buildInfoRow('Status', appointment.status.toUpperCase()),
              ],
            ),
            actions: [
              TextButton(
                onPressed: () {
                  Navigator.pop(context); // Close dialog
                  Navigator.pop(context); // Go back to previous screen
                },
                child: const Text('Done'),
              ),
              ElevatedButton(
                onPressed: () {
                  // Print receipt or share booking details
                  _printReceipt(appointment);
                },
                child: const Text('Print Receipt'),
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

  Widget _buildInfoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            label,
            style: const TextStyle(
              fontSize: 14,
              color: Colors.grey,
            ),
          ),
          Text(
            value,
            style: const TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.bold,
            ),
          ),
        ],
      ),
    );
  }

  void _printReceipt(Appointment appointment) {
    // Implement receipt printing
    print('Print receipt for: ${appointment.bookingNumber}');
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Book Appointment'),
        backgroundColor: Colors.blue,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Date Selection
            Card(
              child: ListTile(
                leading: const Icon(Icons.calendar_today, color: Colors.blue),
                title: const Text('Appointment Date'),
                subtitle: Text(
                  '${_selectedDate.day}/${_selectedDate.month}/${_selectedDate.year}',
                  style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                ),
                trailing: const Icon(Icons.arrow_forward_ios, size: 16),
                onTap: () async {
                  final date = await showDatePicker(
                    context: context,
                    initialDate: _selectedDate,
                    firstDate: DateTime.now(),
                    lastDate: DateTime.now().add(const Duration(days: 90)),
                  );
                  if (date != null) {
                    setState(() => _selectedDate = date);
                  }
                },
              ),
            ),
            
            const SizedBox(height: 16),
            
            // Consultation Type
            const Text(
              'Consultation Type',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Expanded(
                  child: RadioListTile<String>(
                    title: const Text('In-Person'),
                    value: 'offline',
                    groupValue: _consultationType,
                    onChanged: (value) {
                      setState(() => _consultationType = value!);
                    },
                  ),
                ),
                Expanded(
                  child: RadioListTile<String>(
                    title: const Text('Online'),
                    value: 'online',
                    groupValue: _consultationType,
                    onChanged: (value) {
                      setState(() => _consultationType = value!);
                    },
                  ),
                ),
              ],
            ),
            
            const SizedBox(height: 16),
            
            // Time Slot Selection (simplified - you'd load from API)
            const Text(
              'Select Time Slot',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                _buildSlotChip('09:30', 'slot-09-30-uuid'),
                _buildSlotChip('09:35', 'slot-09-35-uuid'),
                _buildSlotChip('09:40', 'slot-09-40-uuid'),
                _buildSlotChip('09:45', 'slot-09-45-uuid'),
                _buildSlotChip('09:50', 'slot-09-50-uuid'),
              ],
            ),
            
            const SizedBox(height: 16),
            
            // Reason
            TextFormField(
              controller: _reasonController,
              decoration: const InputDecoration(
                labelText: 'Reason for Visit (Optional)',
                border: OutlineInputBorder(),
                prefixIcon: Icon(Icons.note),
                hintText: 'Regular checkup, Follow-up, etc.',
              ),
              maxLines: 2,
            ),
            
            const SizedBox(height: 16),
            
            // Notes
            TextFormField(
              controller: _notesController,
              decoration: const InputDecoration(
                labelText: 'Notes (Optional)',
                border: OutlineInputBorder(),
                prefixIcon: Icon(Icons.notes),
                hintText: 'Additional information...',
              ),
              maxLines: 2,
            ),
            
            const SizedBox(height: 16),
            
            // Payment Mode
            const Text(
              'Payment Method',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            DropdownButtonFormField<String>(
              value: _paymentMode,
              decoration: const InputDecoration(
                border: OutlineInputBorder(),
                prefixIcon: Icon(Icons.payment),
              ),
              items: const [
                DropdownMenuItem(value: 'pay_later', child: Text('Pay Later')),
                DropdownMenuItem(value: 'cash', child: Text('Cash')),
                DropdownMenuItem(value: 'card', child: Text('Card')),
                DropdownMenuItem(value: 'upi', child: Text('UPI')),
              ],
              onChanged: (value) {
                setState(() => _paymentMode = value!);
              },
            ),
            
            const SizedBox(height: 16),
            
            // Priority
            SwitchListTile(
              title: const Text('Priority Appointment'),
              subtitle: const Text('Skip regular queue'),
              value: _isPriority,
              onChanged: (value) {
                setState(() => _isPriority = value);
              },
            ),
            
            const SizedBox(height: 24),
            
            // Book Button
            ElevatedButton.icon(
              onPressed: _isLoading ? null : _bookAppointment,
              icon: _isLoading
                  ? const SizedBox(
                      height: 20,
                      width: 20,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                      ),
                    )
                  : const Icon(Icons.calendar_month),
              label: Text(
                _isLoading ? 'Booking...' : 'Confirm Booking',
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
    );
  }

  Widget _buildSlotChip(String time, String slotId) {
    final isSelected = _selectedSlotId == slotId;
    
    return ChoiceChip(
      label: Text(time),
      selected: isSelected,
      onSelected: (selected) {
        setState(() {
          _selectedSlotId = selected ? slotId : null;
          _selectedSlotTime = selected ? time : null;
        });
      },
      selectedColor: Colors.blue,
      labelStyle: TextStyle(
        color: isSelected ? Colors.white : Colors.black,
        fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
      ),
    );
  }

  @override
  void dispose() {
    _reasonController.dispose();
    _notesController.dispose();
    super.dispose();
  }
}
```

---

### Screen 2: Simple Appointment List

```dart
// lib/screens/appointment_list_screen.dart

import 'package:flutter/material.dart';
import '../services/appointment_service.dart';
import '../models/appointment.dart';

class AppointmentListScreen extends StatefulWidget {
  final String clinicId;
  final String token;

  const AppointmentListScreen({
    Key? key,
    required this.clinicId,
    required this.token,
  }) : super(key: key);

  @override
  State<AppointmentListScreen> createState() => _AppointmentListScreenState();
}

class _AppointmentListScreenState extends State<AppointmentListScreen> {
  late AppointmentService _appointmentService;
  List<Appointment> _appointments = [];
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _appointmentService = AppointmentService(widget.token);
    _loadAppointments();
  }

  Future<void> _loadAppointments() async {
    setState(() => _isLoading = true);

    try {
      final appointments = await _appointmentService.listAppointmentsByClinic(
        widget.clinicId,
      );
      setState(() => _appointments = appointments);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: ${e.toString()}')),
        );
      }
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Color _getStatusColor(String status) {
    switch (status) {
      case 'confirmed':
        return Colors.green;
      case 'completed':
        return Colors.blue;
      case 'cancelled':
        return Colors.red;
      default:
        return Colors.grey;
    }
  }

  IconData _getConsultationIcon(String type) {
    switch (type) {
      case 'online':
        return Icons.videocam;
      case 'offline':
        return Icons.person;
      default:
        return Icons.medical_services;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Appointments'),
        backgroundColor: Colors.blue,
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadAppointments,
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _appointments.isEmpty
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.calendar_today_outlined, size: 64, color: Colors.grey[400]),
                      const SizedBox(height: 16),
                      Text(
                        'No appointments yet',
                        style: TextStyle(fontSize: 16, color: Colors.grey[600]),
                      ),
                    ],
                  ),
                )
              : ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: _appointments.length,
                  itemBuilder: (context, index) {
                    final appointment = _appointments[index];
                    
                    return Card(
                      margin: const EdgeInsets.only(bottom: 12),
                      elevation: 2,
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            // Header
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Row(
                                  children: [
                                    Icon(
                                      _getConsultationIcon(appointment.consultationType),
                                      color: Colors.blue,
                                      size: 20,
                                    ),
                                    const SizedBox(width: 8),
                                    Text(
                                      appointment.bookingNumber,
                                      style: const TextStyle(
                                        fontSize: 16,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                  ],
                                ),
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 12,
                                    vertical: 4,
                                  ),
                                  decoration: BoxDecoration(
                                    color: _getStatusColor(appointment.status),
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                  child: Text(
                                    appointment.status.toUpperCase(),
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontSize: 12,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                            
                            const Divider(height: 24),
                            
                            // Details
                            Row(
                              children: [
                                const Icon(Icons.calendar_today, size: 16, color: Colors.grey),
                                const SizedBox(width: 8),
                                Text(appointment.formattedDate),
                                const SizedBox(width: 24),
                                const Icon(Icons.access_time, size: 16, color: Colors.grey),
                                const SizedBox(width: 8),
                                Text(appointment.formattedTime),
                              ],
                            ),
                            
                            const SizedBox(height: 8),
                            
                            Row(
                              children: [
                                const Icon(Icons.confirmation_number, size: 16, color: Colors.grey),
                                const SizedBox(width: 8),
                                Text('Token: #${appointment.tokenNumber}'),
                              ],
                            ),
                            
                            if (appointment.reason != null) ...[
                              const SizedBox(height: 8),
                              Row(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  const Icon(Icons.note, size: 16, color: Colors.grey),
                                  const SizedBox(width: 8),
                                  Expanded(child: Text(appointment.reason!)),
                                ],
                              ),
                            ],
                            
                            const SizedBox(height: 8),
                            
                            Row(
                              children: [
                                const Icon(Icons.payments, size: 16, color: Colors.grey),
                                const SizedBox(width: 8),
                                Text('AED ${appointment.feeAmount.toStringAsFixed(2)}'),
                                const SizedBox(width: 16),
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 8,
                                    vertical: 2,
                                  ),
                                  decoration: BoxDecoration(
                                    color: appointment.paymentStatus == 'paid'
                                        ? Colors.green.shade100
                                        : Colors.orange.shade100,
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  child: Text(
                                    appointment.paymentStatus.toUpperCase(),
                                    style: TextStyle(
                                      fontSize: 12,
                                      color: appointment.paymentStatus == 'paid'
                                          ? Colors.green.shade700
                                          : Colors.orange.shade700,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
    );
  }
}
```

---

## 🔄 Complete Workflow Example

### Scenario: Patient Registration + Appointment Booking

```dart
// lib/screens/patient_booking_workflow.dart

import 'package:flutter/material.dart';
import '../services/clinic_patient_service.dart';
import '../services/appointment_service.dart';

class PatientBookingWorkflow extends StatefulWidget {
  final String clinicId;
  final String doctorId;
  final String token;

  const PatientBookingWorkflow({
    Key? key,
    required this.clinicId,
    required this.doctorId,
    required this.token,
  }) : super(key: key);

  @override
  State<PatientBookingWorkflow> createState() => _PatientBookingWorkflowState();
}

class _PatientBookingWorkflowState extends State<PatientBookingWorkflow> {
  int _currentStep = 0;
  bool _isLoading = false;

  // Step 1: Patient details
  final _firstNameController = TextEditingController();
  final _lastNameController = TextEditingController();
  final _phoneController = TextEditingController();
  
  // Step 2: Appointment details
  DateTime _selectedDate = DateTime.now();
  String? _selectedSlotId;
  String? _selectedSlotTime;
  
  // State
  String? _createdPatientId;
  String? _createdAppointmentId;

  late ClinicPatientService _patientService;
  late AppointmentService _appointmentService;

  @override
  void initState() {
    super.initState();
    _patientService = ClinicPatientService(widget.token);
    _appointmentService = AppointmentService(widget.token);
  }

  Future<void> _executeStep() async {
    if (_currentStep == 0) {
      await _createPatient();
    } else if (_currentStep == 1) {
      await _bookAppointment();
    }
  }

  Future<void> _createPatient() async {
    if (_firstNameController.text.isEmpty ||
        _lastNameController.text.isEmpty ||
        _phoneController.text.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please fill all required fields')),
      );
      return;
    }

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
        _currentStep = 1;
      });

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('✅ ${patient.fullName} registered'),
          backgroundColor: Colors.green,
        ),
      );
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red),
      );
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _bookAppointment() async {
    if (_selectedSlotId == null || _createdPatientId == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select a time slot')),
      );
      return;
    }

    setState(() => _isLoading = true);

    try {
      final formattedDate = '${_selectedDate.year}-${_selectedDate.month.toString().padLeft(2, '0')}-${_selectedDate.day.toString().padLeft(2, '0')}';
      
      final appointment = await _appointmentService.createAppointment(
        clinicPatientId: _createdPatientId!,
        doctorId: widget.doctorId,
        clinicId: widget.clinicId,
        individualSlotId: _selectedSlotId!,
        appointmentDate: formattedDate,
        appointmentTime: '$formattedDate $_selectedSlotTime:00',
        consultationType: 'offline',
      );

      setState(() {
        _createdAppointmentId = appointment.id;
        _currentStep = 2;
      });

      // Show success
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (context) => AlertDialog(
          title: Row(
            children: const [
              Icon(Icons.check_circle, color: Colors.green, size: 32),
              SizedBox(width: 12),
              Text('Success!'),
            ],
          ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text('Appointment booked successfully!', 
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
              const SizedBox(height: 16),
              _buildDetail('Booking Number', appointment.bookingNumber),
              _buildDetail('Token', '#${appointment.tokenNumber}'),
              _buildDetail('Date', appointment.formattedDate),
              _buildDetail('Time', appointment.formattedTime),
              _buildDetail('Fee', 'AED ${appointment.feeAmount}'),
            ],
          ),
          actions: [
            ElevatedButton.icon(
              onPressed: () {
                Navigator.pop(context);
                Navigator.pop(context);
              },
              icon: const Icon(Icons.done),
              label: const Text('Done'),
            ),
          ],
        ),
      );
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Booking failed: $e'), backgroundColor: Colors.red),
      );
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Widget _buildDetail(String label, String value) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(color: Colors.grey)),
          Text(value, style: const TextStyle(fontWeight: FontWeight.bold)),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Book Appointment'),
      ),
      body: Stepper(
        currentStep: _currentStep,
        onStepContinue: _isLoading ? null : _executeStep,
        onStepCancel: () {
          if (_currentStep > 0) {
            setState(() => _currentStep--);
          } else {
            Navigator.pop(context);
          }
        },
        controlsBuilder: (context, details) {
          return Padding(
            padding: const EdgeInsets.only(top: 16),
            child: Row(
              children: [
                ElevatedButton(
                  onPressed: _isLoading ? null : details.onStepContinue,
                  child: Text(_currentStep == 1 ? 'Book Appointment' : 'Continue'),
                ),
                const SizedBox(width: 12),
                if (_currentStep > 0)
                  TextButton(
                    onPressed: details.onStepCancel,
                    child: const Text('Back'),
                  ),
              ],
            ),
          );
        },
        steps: [
          // Step 1: Patient Info
          Step(
            title: const Text('Patient Information'),
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
                  ),
                  keyboardType: TextInputType.phone,
                ),
              ],
            ),
            isActive: _currentStep >= 0,
            state: _createdPatientId != null ? StepState.complete : StepState.indexed,
          ),
          
          // Step 2: Slot Selection
          Step(
            title: const Text('Select Time Slot'),
            content: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('Date: ${_selectedDate.day}/${_selectedDate.month}/${_selectedDate.year}'),
                const SizedBox(height: 16),
                const Text('Available Slots:', style: TextStyle(fontWeight: FontWeight.bold)),
                const SizedBox(height: 8),
                Wrap(
                  spacing: 8,
                  children: [
                    _buildSlotButton('09:30', 'slot-09-30-uuid'),
                    _buildSlotButton('09:35', 'slot-09-35-uuid'),
                    _buildSlotButton('09:40', 'slot-09-40-uuid'),
                  ],
                ),
              ],
            ),
            isActive: _currentStep >= 1,
            state: _selectedSlotId != null ? StepState.complete : StepState.indexed,
          ),
        ],
      ),
    );
  }

  Widget _buildSlotButton(String time, String slotId) {
    final isSelected = _selectedSlotId == slotId;
    
    return ElevatedButton(
      onPressed: () {
        setState(() {
          _selectedSlotId = slotId;
          _selectedSlotTime = time;
        });
      },
      style: ElevatedButton.styleFrom(
        backgroundColor: isSelected ? Colors.blue : Colors.grey[200],
        foregroundColor: isSelected ? Colors.white : Colors.black,
      ),
      child: Text(time),
    );
  }
}
```

---

## 📋 Complete API Reference

### Required Fields
```json
{
  "clinic_patient_id": "UUID (required)",
  "doctor_id": "UUID (required)",
  "clinic_id": "UUID (required)",
  "individual_slot_id": "UUID (required)",
  "appointment_date": "YYYY-MM-DD (required)",
  "appointment_time": "YYYY-MM-DD HH:MM:SS (required)",
  "consultation_type": "offline|online|video|in_person|follow_up (required)"
}
```

### Optional Fields
```json
{
  "department_id": "UUID (optional)",
  "duration_minutes": "integer (optional, default: 12)",
  "reason": "string (optional)",
  "notes": "string (optional)",
  "is_priority": "boolean (optional, default: false)",
  "payment_mode": "pay_later|cash|card|upi (optional)"
}
```

---

## ✅ Complete Example Flow

### JSON Request (Full)
```json
POST /api/appointments

{
  "clinic_patient_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "individual_slot_id": "slot-09-30-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "duration_minutes": 5,
  "consultation_type": "offline",
  "reason": "Regular checkup",
  "notes": "Patient prefers morning appointments",
  "is_priority": false,
  "payment_mode": "cash"
}
```

### JSON Response (Full)
```json
{
  "appointment": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "patient_id": null,
    "clinic_patient_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
    "department_id": null,
    "booking_number": "BN202510180001",
    "token_number": 1,
    "appointment_date": "2025-10-18",
    "appointment_time": "2025-10-18T09:30:00Z",
    "duration_minutes": 5,
    "consultation_type": "offline",
    "reason": "Regular checkup",
    "notes": "Patient prefers morning appointments",
    "status": "confirmed",
    "fee_amount": 500.00,
    "payment_status": "paid",
    "payment_mode": "cash",
    "is_priority": false,
    "created_at": "2024-10-15T12:00:00Z"
  }
}
```

---

## ✅ Flutter Dependencies

### pubspec.yaml
```yaml
dependencies:
  flutter:
    sdk: flutter
  http: ^1.1.0
  intl: ^0.18.0  # For date formatting
```

---

## 🎯 Summary

### Appointment Creation Process

| Step | Action | Returns |
|------|--------|---------|
| 1 | **Create Patient** (`POST /clinic-specific-patients`) | `clinic_patient_id` |
| 2 | **Get Available Slots** (`GET /doctor-session-slots`) | List of `individual_slot_id` |
| 3 | **Book Appointment** (`POST /appointments`) | `appointment` with `booking_number` |

### Database Updates
```sql
-- Appointment created
INSERT INTO appointments (clinic_patient_id, ...) VALUES (...);

-- Slot marked as booked
UPDATE doctor_individual_slots 
SET is_booked = true, 
    booked_patient_id = clinic_patient_id,
    status = 'booked';
```

---

## ✅ Status

| Feature | Status |
|---------|--------|
| Full JSON examples | ✅ Provided |
| Flutter models | ✅ Complete |
| Flutter service | ✅ Complete |
| Flutter UI screens | ✅ 3 screens provided |
| Minimal creation | ✅ Supported |
| Full creation | ✅ Supported |
| Error handling | ✅ Included |

---

**API:** `POST /api/appointments`  
**Key Field:** `clinic_patient_id`  
**Status:** ✅ **Complete Flutter Integration Guide!** 🎉

