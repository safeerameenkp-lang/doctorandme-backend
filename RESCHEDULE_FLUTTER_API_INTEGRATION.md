# Reschedule Appointment API - Flutter Integration Guide

## Overview
This guide provides complete Flutter integration examples for the reschedule appointment API, including HTTP requests, JSON examples, error handling, and UI state management.

## API Endpoint
```
POST /api/appointment-service/appointments/{appointment_id}/reschedule
```

## Authentication
All requests require JWT authentication via Authorization header:
```dart
headers: {
  'Content-Type': 'application/json',
  'Authorization': 'Bearer YOUR_JWT_TOKEN'
}
```

---

## 1. Flutter Service Class

```dart
import 'dart:convert';
import 'package:http/http.dart' as http;

class AppointmentService {
  static const String baseUrl = 'http://your-api-domain.com/api/appointment-service';
  static const String organizationUrl = 'http://your-api-domain.com/api/organization-service';
  
  // Reschedule appointment
  static Future<RescheduleResponse> rescheduleAppointment({
    required String appointmentId,
    required RescheduleRequest request,
    required String token,
  }) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/appointments/$appointmentId/reschedule'),
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer $token',
        },
        body: jsonEncode(request.toJson()),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return RescheduleResponse.fromJson(data);
      } else {
        final errorData = jsonDecode(response.body);
        throw AppointmentException(
          message: errorData['error'] ?? 'Unknown error',
          details: errorData['details'],
          statusCode: response.statusCode,
        );
      }
    } catch (e) {
      throw AppointmentException(
        message: 'Network error: ${e.toString()}',
        statusCode: 0,
      );
    }
  }

  // Get doctor session slots (for slot selection)
  static Future<List<DoctorSessionSlot>> getDoctorSessionSlots({
    required String doctorId,
    required String clinicId,
    required String date,
    String? slotType,
    required String token,
  }) async {
    try {
      String url = '$organizationUrl/doctor-session-slots?doctor_id=$doctorId&clinic_id=$clinicId&date=$date';
      if (slotType != null) {
        url += '&slot_type=$slotType';
      }

      final response = await http.get(
        Uri.parse(url),
        headers: {
          'Authorization': 'Bearer $token',
        },
      );

      if (response.statusCode == 200) {
        final List<dynamic> data = jsonDecode(response.body);
        return data.map((json) => DoctorSessionSlot.fromJson(json)).toList();
      } else {
        final errorData = jsonDecode(response.body);
        throw AppointmentException(
          message: errorData['error'] ?? 'Failed to fetch slots',
          statusCode: response.statusCode,
        );
      }
    } catch (e) {
      throw AppointmentException(
        message: 'Network error: ${e.toString()}',
        statusCode: 0,
      );
    }
  }
}
```

---

## 2. Data Models

```dart
// Reschedule Request Model
class RescheduleRequest {
  final String? departmentId;
  final String doctorId;
  final String clinicId;
  final String individualSlotId;
  final String appointmentDate;
  final String appointmentTime;
  final String? reason;
  final String? notes;

  RescheduleRequest({
    this.departmentId,
    required this.doctorId,
    required this.clinicId,
    required this.individualSlotId,
    required this.appointmentDate,
    required this.appointmentTime,
    this.reason,
    this.notes,
  });

  Map<String, dynamic> toJson() {
    return {
      if (departmentId != null) 'department_id': departmentId,
      'doctor_id': doctorId,
      'clinic_id': clinicId,
      'individual_slot_id': individualSlotId,
      'appointment_date': appointmentDate,
      'appointment_time': appointmentTime,
      if (reason != null) 'reason': reason,
      if (notes != null) 'notes': notes,
    };
  }

  factory RescheduleRequest.fromJson(Map<String, dynamic> json) {
    return RescheduleRequest(
      departmentId: json['department_id'],
      doctorId: json['doctor_id'],
      clinicId: json['clinic_id'],
      individualSlotId: json['individual_slot_id'],
      appointmentDate: json['appointment_date'],
      appointmentTime: json['appointment_time'],
      reason: json['reason'],
      notes: json['notes'],
    );
  }
}

// Reschedule Response Model
class RescheduleResponse {
  final String message;
  final Appointment appointment;
  final SlotReEnabled? slotReEnabled;

  RescheduleResponse({
    required this.message,
    required this.appointment,
    this.slotReEnabled,
  });

  factory RescheduleResponse.fromJson(Map<String, dynamic> json) {
    return RescheduleResponse(
      message: json['message'],
      appointment: Appointment.fromJson(json['appointment']),
      slotReEnabled: json['slot_re_enabled'] != null 
          ? SlotReEnabled.fromJson(json['slot_re_enabled']) 
          : null,
    );
  }
}

// Appointment Model
class Appointment {
  final String id;
  final String? clinicPatientId;
  final String clinicId;
  final String doctorId;
  final String? departmentId;
  final String bookingNumber;
  final int? tokenNumber;
  final String? appointmentDate;
  final DateTime appointmentTime;
  final int durationMinutes;
  final String consultationType;
  final String? reason;
  final String? notes;
  final String status;
  final double? feeAmount;
  final String paymentStatus;
  final String? paymentMode;
  final DateTime createdAt;

  Appointment({
    required this.id,
    this.clinicPatientId,
    required this.clinicId,
    required this.doctorId,
    this.departmentId,
    required this.bookingNumber,
    this.tokenNumber,
    this.appointmentDate,
    required this.appointmentTime,
    required this.durationMinutes,
    required this.consultationType,
    this.reason,
    this.notes,
    required this.status,
    this.feeAmount,
    required this.paymentStatus,
    this.paymentMode,
    required this.createdAt,
  });

  factory Appointment.fromJson(Map<String, dynamic> json) {
    return Appointment(
      id: json['id'],
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
      feeAmount: json['fee_amount']?.toDouble(),
      paymentStatus: json['payment_status'],
      paymentMode: json['payment_mode'],
      createdAt: DateTime.parse(json['created_at']),
    );
  }
}

// Slot Re-enabled Model
class SlotReEnabled {
  final String oldSlotId;
  final String message;

  SlotReEnabled({
    required this.oldSlotId,
    required this.message,
  });

  factory SlotReEnabled.fromJson(Map<String, dynamic> json) {
    return SlotReEnabled(
      oldSlotId: json['old_slot_id'],
      message: json['message'],
    );
  }
}

// Doctor Session Slot Model (for slot selection)
class DoctorSessionSlot {
  final String id;
  final String doctorId;
  final String clinicId;
  final String date;
  final int dayOfWeek;
  final String slotType;
  final bool isAvailable;
  final List<Session> sessions;

  DoctorSessionSlot({
    required this.id,
    required this.doctorId,
    required this.clinicId,
    required this.date,
    required this.dayOfWeek,
    required this.slotType,
    required this.isAvailable,
    required this.sessions,
  });

  factory DoctorSessionSlot.fromJson(Map<String, dynamic> json) {
    return DoctorSessionSlot(
      id: json['id'],
      doctorId: json['doctor_id'],
      clinicId: json['clinic_id'],
      date: json['date'],
      dayOfWeek: json['day_of_week'],
      slotType: json['slot_type'],
      isAvailable: json['is_available'],
      sessions: (json['sessions'] as List)
          .map((session) => Session.fromJson(session))
          .toList(),
    );
  }
}

class Session {
  final String id;
  final String sessionName;
  final String startTime;
  final String endTime;
  final int maxPatients;
  final int slotIntervalMinutes;
  final int generatedSlots;
  final int availableSlots;
  final int bookedSlots;
  final String? notes;
  final List<IndividualSlot> slots;

  Session({
    required this.id,
    required this.sessionName,
    required this.startTime,
    required this.endTime,
    required this.maxPatients,
    required this.slotIntervalMinutes,
    required this.generatedSlots,
    required this.availableSlots,
    required this.bookedSlots,
    this.notes,
    required this.slots,
  });

  factory Session.fromJson(Map<String, dynamic> json) {
    return Session(
      id: json['id'],
      sessionName: json['session_name'],
      startTime: json['start_time'],
      endTime: json['end_time'],
      maxPatients: json['max_patients'],
      slotIntervalMinutes: json['slot_interval_minutes'],
      generatedSlots: json['generated_slots'],
      availableSlots: json['available_slots'],
      bookedSlots: json['booked_slots'],
      notes: json['notes'],
      slots: (json['slots'] as List)
          .map((slot) => IndividualSlot.fromJson(slot))
          .toList(),
    );
  }
}

class IndividualSlot {
  final String id;
  final String slotStart;
  final String slotEnd;
  final bool isBooked;
  final bool isBookable;
  final int maxPatients;
  final int availableCount;
  final int bookedCount;
  final String? bookedPatientId;
  final String? bookedAppointmentId;
  final String status;
  final String displayMessage;

  IndividualSlot({
    required this.id,
    required this.slotStart,
    required this.slotEnd,
    required this.isBooked,
    required this.isBookable,
    required this.maxPatients,
    required this.availableCount,
    required this.bookedCount,
    this.bookedPatientId,
    this.bookedAppointmentId,
    required this.status,
    required this.displayMessage,
  });

  factory IndividualSlot.fromJson(Map<String, dynamic> json) {
    return IndividualSlot(
      id: json['id'],
      slotStart: json['slot_start'],
      slotEnd: json['slot_end'],
      isBooked: json['is_booked'],
      isBookable: json['is_bookable'],
      maxPatients: json['max_patients'],
      availableCount: json['available_count'],
      bookedCount: json['booked_count'],
      bookedPatientId: json['booked_patient_id'],
      bookedAppointmentId: json['booked_appointment_id'],
      status: json['status'],
      displayMessage: json['display_message'],
    );
  }
}

// Exception Model
class AppointmentException implements Exception {
  final String message;
  final String? details;
  final int statusCode;

  AppointmentException({
    required this.message,
    this.details,
    required this.statusCode,
  });

  @override
  String toString() {
    return 'AppointmentException: $message (Status: $statusCode)';
  }
}
```

---

## 3. Complete JSON Examples

### Request JSON Example
```json
{
  "department_id": "550e8400-e29b-41d4-a716-446655440001",
  "doctor_id": "550e8400-e29b-41d4-a716-446655440002",
  "clinic_id": "550e8400-e29b-41d4-a716-446655440003",
  "individual_slot_id": "550e8400-e29b-41d4-a716-446655440004",
  "appointment_date": "2024-07-20",
  "appointment_time": "2024-07-20 10:30:00",
  "reason": "Patient requested time change",
  "notes": "Rescheduled due to patient availability"
}
```

### Success Response JSON Example
```json
{
  "message": "Appointment rescheduled successfully",
  "appointment": {
    "id": "550e8400-e29b-41d4-a716-446655440005",
    "clinic_patient_id": "550e8400-e29b-41d4-a716-446655440006",
    "clinic_id": "550e8400-e29b-41d4-a716-446655440003",
    "doctor_id": "550e8400-e29b-41d4-a716-446655440002",
    "department_id": "550e8400-e29b-41d4-a716-446655440001",
    "booking_number": "BN20240720103000",
    "token_number": 5,
    "appointment_date": "2024-07-20",
    "appointment_time": "2024-07-20T10:30:00Z",
    "duration_minutes": 5,
    "consultation_type": "offline",
    "reason": "Patient requested time change",
    "notes": "Rescheduled due to patient availability",
    "status": "confirmed",
    "fee_amount": 500.0,
    "payment_status": "paid",
    "payment_mode": "cash",
    "created_at": "2024-07-19T10:00:00Z"
  },
  "slot_re_enabled": {
    "old_slot_id": "550e8400-e29b-41d4-a716-446655440007",
    "message": "Previous slot has been made available again"
  }
}
```

### Error Response JSON Examples

#### 400 Bad Request - Invalid Input
```json
{
  "error": "Invalid input",
  "details": "Key: 'RescheduleRequest.DoctorID' Error:Field validation for 'DoctorID' failed on the 'required' tag"
}
```

#### 404 Not Found - Appointment Not Found
```json
{
  "error": "Appointment not found or cannot be rescheduled"
}
```

#### 409 Conflict - Slot Not Available
```json
{
  "error": "Slot not available",
  "message": "This slot is fully booked. Please select another slot.",
  "details": {
    "max_patients": 3,
    "available_count": 0,
    "booked_count": 3
  }
}
```

#### 409 Conflict - Slot Just Got Booked
```json
{
  "error": "Slot just got booked",
  "message": "This slot was just booked by another patient. Please select another slot."
}
```

---

## 4. Flutter Widget Implementation

```dart
import 'package:flutter/material.dart';

class RescheduleAppointmentScreen extends StatefulWidget {
  final String appointmentId;
  final String token;

  const RescheduleAppointmentScreen({
    Key? key,
    required this.appointmentId,
    required this.token,
  }) : super(key: key);

  @override
  State<RescheduleAppointmentScreen> createState() => _RescheduleAppointmentScreenState();
}

class _RescheduleAppointmentScreenState extends State<RescheduleAppointmentScreen> {
  final _formKey = GlobalKey<FormState>();
  final _reasonController = TextEditingController();
  final _notesController = TextEditingController();
  
  String? _selectedDepartmentId;
  String? _selectedDoctorId;
  String? _selectedClinicId;
  String? _selectedSlotId;
  String? _selectedDate;
  String? _selectedTime;
  
  List<DoctorSessionSlot> _availableSlots = [];
  List<IndividualSlot> _morningSlots = [];
  List<IndividualSlot> _afternoonSlots = [];
  
  bool _isLoading = false;
  bool _isRescheduling = false;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _loadAvailableSlots();
  }

  Future<void> _loadAvailableSlots() async {
    if (_selectedDoctorId == null || _selectedClinicId == null || _selectedDate == null) {
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final slots = await AppointmentService.getDoctorSessionSlots(
        doctorId: _selectedDoctorId!,
        clinicId: _selectedClinicId!,
        date: _selectedDate!,
        token: widget.token,
      );

      setState(() {
        _availableSlots = slots;
        _categorizeSlots();
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _errorMessage = e.toString();
        _isLoading = false;
      });
    }
  }

  void _categorizeSlots() {
    _morningSlots.clear();
    _afternoonSlots.clear();

    for (final slot in _availableSlots) {
      for (final session in slot.sessions) {
        for (final individualSlot in session.slots) {
          final hour = int.parse(individualSlot.slotStart.split(':')[0]);
          if (hour < 12) {
            _morningSlots.add(individualSlot);
          } else {
            _afternoonSlots.add(individualSlot);
          }
        }
      }
    }
  }

  Future<void> _rescheduleAppointment() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    if (_selectedDoctorId == null || 
        _selectedClinicId == null || 
        _selectedSlotId == null || 
        _selectedDate == null || 
        _selectedTime == null) {
      setState(() {
        _errorMessage = 'Please select all required fields';
      });
      return;
    }

    setState(() {
      _isRescheduling = true;
      _errorMessage = null;
    });

    try {
      final request = RescheduleRequest(
        departmentId: _selectedDepartmentId,
        doctorId: _selectedDoctorId!,
        clinicId: _selectedClinicId!,
        individualSlotId: _selectedSlotId!,
        appointmentDate: _selectedDate!,
        appointmentTime: _selectedTime!,
        reason: _reasonController.text.isNotEmpty ? _reasonController.text : null,
        notes: _notesController.text.isNotEmpty ? _notesController.text : null,
      );

      final response = await AppointmentService.rescheduleAppointment(
        appointmentId: widget.appointmentId,
        request: request,
        token: widget.token,
      );

      // Show success message
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(response.message),
          backgroundColor: Colors.green,
        ),
      );

      // Show slot re-enabled info if applicable
      if (response.slotReEnabled != null) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(response.slotReEnabled!.message),
            backgroundColor: Colors.blue,
            duration: Duration(seconds: 3),
          ),
        );
      }

      // Navigate back or refresh
      Navigator.of(context).pop(true);
      
    } catch (e) {
      setState(() {
        _errorMessage = e.toString();
        _isRescheduling = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Reschedule Appointment'),
        backgroundColor: Colors.blue[600],
        foregroundColor: Colors.white,
      ),
      body: Form(
        key: _formKey,
        child: SingleChildScrollView(
          padding: EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Department Selection
              DropdownButtonFormField<String>(
                value: _selectedDepartmentId,
                decoration: InputDecoration(
                  labelText: 'Select Department',
                  border: OutlineInputBorder(),
                ),
                items: [
                  DropdownMenuItem(value: 'dept1', child: Text('Orthology')),
                  DropdownMenuItem(value: 'dept2', child: Text('Cardiology')),
                  // Add more departments
                ],
                onChanged: (value) {
                  setState(() {
                    _selectedDepartmentId = value;
                  });
                },
              ),
              
              SizedBox(height: 16),
              
              // Doctor Selection
              DropdownButtonFormField<String>(
                value: _selectedDoctorId,
                decoration: InputDecoration(
                  labelText: 'Select Doctor',
                  border: OutlineInputBorder(),
                ),
                items: [
                  DropdownMenuItem(value: 'doc1', child: Text('Dr. Arun Krishna')),
                  DropdownMenuItem(value: 'doc2', child: Text('Dr. Sarah Johnson')),
                  // Add more doctors
                ],
                onChanged: (value) {
                  setState(() {
                    _selectedDoctorId = value;
                    _selectedSlotId = null; // Reset slot selection
                  });
                  _loadAvailableSlots();
                },
                validator: (value) {
                  if (value == null) return 'Please select a doctor';
                  return null;
                },
              ),
              
              SizedBox(height: 16),
              
              // Reason/Notes
              TextFormField(
                controller: _reasonController,
                decoration: InputDecoration(
                  labelText: 'Reason or Add Notes',
                  hintText: 'Add Patient Note',
                  border: OutlineInputBorder(),
                ),
                maxLines: 3,
              ),
              
              SizedBox(height: 24),
              
              // Available Slots Section
              Text(
                'Available Slots',
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              
              SizedBox(height: 16),
              
              // Date Selection
              Row(
                children: [
                  Expanded(
                    child: TextFormField(
                      decoration: InputDecoration(
                        labelText: 'Date',
                        border: OutlineInputBorder(),
                        suffixIcon: Icon(Icons.calendar_today),
                      ),
                      readOnly: true,
                      onTap: () async {
                        final date = await showDatePicker(
                          context: context,
                          initialDate: DateTime.now(),
                          firstDate: DateTime.now(),
                          lastDate: DateTime.now().add(Duration(days: 30)),
                        );
                        if (date != null) {
                          setState(() {
                            _selectedDate = date.toIso8601String().split('T')[0];
                            _selectedSlotId = null; // Reset slot selection
                          });
                          _loadAvailableSlots();
                        }
                      },
                      validator: (value) {
                        if (value == null || value.isEmpty) return 'Please select a date';
                        return null;
                      },
                    ),
                  ),
                  SizedBox(width: 8),
                  IconButton(
                    onPressed: () {
                      // Navigate to next day
                      if (_selectedDate != null) {
                        final currentDate = DateTime.parse(_selectedDate!);
                        final nextDate = currentDate.add(Duration(days: 1));
                        setState(() {
                          _selectedDate = nextDate.toIso8601String().split('T')[0];
                          _selectedSlotId = null;
                        });
                        _loadAvailableSlots();
                      }
                    },
                    icon: Icon(Icons.arrow_forward),
                  ),
                ],
              ),
              
              SizedBox(height: 16),
              
              // Slot Legend
              Row(
                children: [
                  Container(
                    width: 16,
                    height: 16,
                    decoration: BoxDecoration(
                      color: Colors.grey[300],
                      shape: BoxShape.circle,
                    ),
                  ),
                  SizedBox(width: 8),
                  Text('Available Slots'),
                  SizedBox(width: 24),
                  Container(
                    width: 16,
                    height: 16,
                    decoration: BoxDecoration(
                      color: Colors.pink[200],
                      shape: BoxShape.circle,
                    ),
                  ),
                  SizedBox(width: 8),
                  Text('Booked Slots'),
                ],
              ),
              
              SizedBox(height: 16),
              
              if (_isLoading)
                Center(child: CircularProgressIndicator())
              else if (_errorMessage != null)
                Container(
                  padding: EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: Colors.red[50],
                    border: Border.all(color: Colors.red[200]!),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    _errorMessage!,
                    style: TextStyle(color: Colors.red[700]),
                  ),
                )
              else ...[
                // Morning Slots
                Text(
                  'Morning Slots',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                SizedBox(height: 8),
                Container(
                  height: 120,
                  child: ListView.builder(
                    scrollDirection: Axis.horizontal,
                    itemCount: _morningSlots.length,
                    itemBuilder: (context, index) {
                      final slot = _morningSlots[index];
                      final isSelected = _selectedSlotId == slot.id;
                      final isAvailable = slot.isBookable && slot.status == 'available';
                      
                      return GestureDetector(
                        onTap: isAvailable ? () {
                          setState(() {
                            _selectedSlotId = slot.id;
                            _selectedTime = '${_selectedDate} ${slot.slotStart}:00';
                          });
                        } : null,
                        child: Container(
                          width: 100,
                          margin: EdgeInsets.only(right: 8),
                          decoration: BoxDecoration(
                            color: isSelected 
                                ? Colors.black 
                                : isAvailable 
                                    ? Colors.green[100] 
                                    : Colors.pink[100],
                            border: Border.all(
                              color: isSelected 
                                  ? Colors.black 
                                  : isAvailable 
                                      ? Colors.green[300]! 
                                      : Colors.pink[300]!,
                            ),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Text(
                                slot.slotStart,
                                style: TextStyle(
                                  color: isSelected 
                                      ? Colors.white 
                                      : isAvailable 
                                          ? Colors.green[700] 
                                          : Colors.pink[700],
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                              SizedBox(height: 4),
                              Text(
                                slot.displayMessage,
                                style: TextStyle(
                                  color: isSelected 
                                      ? Colors.white 
                                      : isAvailable 
                                          ? Colors.green[700] 
                                          : Colors.pink[700],
                                  fontSize: 12,
                                ),
                                textAlign: TextAlign.center,
                              ),
                            ],
                          ),
                        ),
                      );
                    },
                  ),
                ),
                
                SizedBox(height: 16),
                
                // Afternoon Slots
                Text(
                  'Afternoon Slots',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                SizedBox(height: 8),
                Container(
                  height: 120,
                  child: ListView.builder(
                    scrollDirection: Axis.horizontal,
                    itemCount: _afternoonSlots.length,
                    itemBuilder: (context, index) {
                      final slot = _afternoonSlots[index];
                      final isSelected = _selectedSlotId == slot.id;
                      final isAvailable = slot.isBookable && slot.status == 'available';
                      
                      return GestureDetector(
                        onTap: isAvailable ? () {
                          setState(() {
                            _selectedSlotId = slot.id;
                            _selectedTime = '${_selectedDate} ${slot.slotStart}:00';
                          });
                        } : null,
                        child: Container(
                          width: 100,
                          margin: EdgeInsets.only(right: 8),
                          decoration: BoxDecoration(
                            color: isSelected 
                                ? Colors.black 
                                : isAvailable 
                                    ? Colors.green[100] 
                                    : Colors.pink[100],
                            border: Border.all(
                              color: isSelected 
                                  ? Colors.black 
                                  : isAvailable 
                                      ? Colors.green[300]! 
                                      : Colors.pink[300]!,
                            ),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Text(
                                slot.slotStart,
                                style: TextStyle(
                                  color: isSelected 
                                      ? Colors.white 
                                      : isAvailable 
                                          ? Colors.green[700] 
                                          : Colors.pink[700],
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                              SizedBox(height: 4),
                              Text(
                                slot.displayMessage,
                                style: TextStyle(
                                  color: isSelected 
                                      ? Colors.white 
                                      : isAvailable 
                                          ? Colors.green[700] 
                                          : Colors.pink[700],
                                  fontSize: 12,
                                ),
                                textAlign: TextAlign.center,
                              ),
                            ],
                          ),
                        ),
                      );
                    },
                  ),
                ),
              ],
              
              SizedBox(height: 24),
              
              // Add More Appointment Button
              Container(
                width: double.infinity,
                height: 50,
                decoration: BoxDecoration(
                  border: Border.all(color: Colors.grey[400]!, style: BorderStyle.solid),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: TextButton.icon(
                  onPressed: () {
                    // Handle add more appointment
                  },
                  icon: Icon(Icons.add, color: Colors.grey[600]),
                  label: Text(
                    'Add More Appointment',
                    style: TextStyle(color: Colors.grey[600]),
                  ),
                ),
              ),
              
              SizedBox(height: 24),
              
              // Action Buttons
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      onPressed: _isRescheduling ? null : () {
                        Navigator.of(context).pop();
                      },
                      child: Text('Cancel'),
                    ),
                  ),
                  SizedBox(width: 16),
                  Expanded(
                    child: ElevatedButton(
                      onPressed: _isRescheduling ? null : _rescheduleAppointment,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.black,
                        foregroundColor: Colors.white,
                      ),
                      child: _isRescheduling
                          ? SizedBox(
                              height: 20,
                              width: 20,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                                valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                              ),
                            )
                          : Text('Save'),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
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

## 5. Usage Example

```dart
// Navigate to reschedule screen
Navigator.push(
  context,
  MaterialPageRoute(
    builder: (context) => RescheduleAppointmentScreen(
      appointmentId: 'your-appointment-id',
      token: 'your-jwt-token',
    ),
  ),
).then((result) {
  if (result == true) {
    // Refresh appointment list or show success message
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('Appointment rescheduled successfully!')),
    );
  }
});
```

---

## 6. Error Handling Best Practices

```dart
class RescheduleErrorHandler {
  static String getErrorMessage(AppointmentException e) {
    switch (e.statusCode) {
      case 400:
        return 'Please check your input and try again.';
      case 404:
        return 'Appointment not found. It may have been cancelled.';
      case 409:
        if (e.message.contains('Slot just got booked')) {
          return 'This slot was just booked by another patient. Please select another slot.';
        } else if (e.message.contains('Slot not available')) {
          return 'This slot is fully booked. Please select another slot.';
        }
        return 'There was a conflict with your request. Please try again.';
      case 500:
        return 'Server error. Please try again later.';
      default:
        return 'An unexpected error occurred. Please try again.';
    }
  }

  static void showErrorSnackBar(BuildContext context, AppointmentException e) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(getErrorMessage(e)),
        backgroundColor: Colors.red,
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
}
```

---

## 7. Testing with Sample Data

```dart
// Test data for development
class TestData {
  static const String testAppointmentId = '550e8400-e29b-41d4-a716-446655440005';
  static const String testDoctorId = '550e8400-e29b-41d4-a716-446655440002';
  static const String testClinicId = '550e8400-e29b-41d4-a716-446655440003';
  static const String testSlotId = '550e8400-e29b-41d4-a716-446655440004';
  static const String testDate = '2024-07-20';
  static const String testTime = '2024-07-20 10:30:00';
  
  static RescheduleRequest getTestRequest() {
    return RescheduleRequest(
      departmentId: '550e8400-e29b-41d4-a716-446655440001',
      doctorId: testDoctorId,
      clinicId: testClinicId,
      individualSlotId: testSlotId,
      appointmentDate: testDate,
      appointmentTime: testTime,
      reason: 'Patient requested time change',
      notes: 'Rescheduled due to patient availability',
    );
  }
}
```

This comprehensive Flutter integration guide provides everything needed to implement the reschedule appointment functionality with proper error handling, UI components, and JSON examples.
