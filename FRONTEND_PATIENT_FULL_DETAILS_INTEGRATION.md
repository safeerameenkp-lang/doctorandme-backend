# Frontend Integration Guide: Patient Full Details API

This guide provides the complete Flutter/Dart implementation for integrating the new **Clinic Patient Full Details API**. It includes the correct Dart data models, API service methods, and UI structure recommendations.

## 1. Dart Data Models

Create a new file `patient_full_details_model.dart` to map perfectly to the JSON structure returned by the backend.

```dart
class PatientFullDetailsResponse {
  final PatientInfo patientInfo;
  final double totalSpent;
  final int totalAppointments;
  final List<DoctorVisitSummary> doctorVisits;
  final List<VitalSignSummary> recentVitals;
  final List<TimelineEventSummary> visitTimeline;

  PatientFullDetailsResponse({
    required this.patientInfo,
    required this.totalSpent,
    required this.totalAppointments,
    required this.doctorVisits,
    required this.recentVitals,
    required this.visitTimeline,
  });

  factory PatientFullDetailsResponse.fromJson(Map<String, dynamic> json) {
    return PatientFullDetailsResponse(
      patientInfo: PatientInfo.fromJson(json['patient_info'] ?? {}),
      totalSpent: (json['total_spent'] ?? 0).toDouble(),
      totalAppointments: json['total_appointments'] ?? 0,
      doctorVisits: (json['doctor_visits'] as List? ?? [])
          .map((e) => DoctorVisitSummary.fromJson(e))
          .toList(),
      recentVitals: (json['recent_vitals'] as List? ?? [])
          .map((e) => VitalSignSummary.fromJson(e))
          .toList(),
      visitTimeline: (json['visit_timeline'] as List? ?? [])
          .map((e) => TimelineEventSummary.fromJson(e))
          .toList(),
    );
  }
}

class PatientInfo {
  final String id;
  final String firstName;
  final String lastName;
  final String phone;
  final String dateOfBirth;
  final int age;
  final String gender;
  final String bloodGroup;
  final String moid;
  final String medicalHistory;
  final String allergies;

  PatientInfo({
    required this.id,
    required this.firstName,
    required this.lastName,
    required this.phone,
    required this.dateOfBirth,
    required this.age,
    required this.gender,
    required this.bloodGroup,
    required this.moid,
    required this.medicalHistory,
    required this.allergies,
  });

  factory PatientInfo.fromJson(Map<String, dynamic> json) {
    return PatientInfo(
      id: json['id'] ?? '',
      firstName: json['first_name'] ?? '',
      lastName: json['last_name'] ?? '',
      phone: json['phone'] ?? '',
      dateOfBirth: json['date_of_birth'] ?? '',
      age: json['age'] ?? 0,
      gender: json['gender'] ?? '',
      bloodGroup: json['blood_group'] ?? '',
      moid: json['mo_id'] ?? '',
      medicalHistory: json['medical_history'] ?? '',
      allergies: json['allergies'] ?? '',
    );
  }

  String get fullName => '$firstName $lastName'.trim();
}

class DoctorVisitSummary {
  final String doctorId;
  final String doctorName;
  final String departmentName;
  final int totalVisits;
  final int normalVisits;
  final int walkinVisits;
  final double totalPaid;
  final String lastVisitDate;
  final List<DetailAppointment> appointments;

  DoctorVisitSummary({
    required this.doctorId,
    required this.doctorName,
    required this.departmentName,
    required this.totalVisits,
    required this.normalVisits,
    required this.walkinVisits,
    required this.totalPaid,
    required this.lastVisitDate,
    required this.appointments,
  });

  factory DoctorVisitSummary.fromJson(Map<String, dynamic> json) {
    return DoctorVisitSummary(
      doctorId: json['doctor_id'] ?? '',
      doctorName: json['doctor_name'] ?? '',
      departmentName: json['department_name'] ?? '',
      totalVisits: json['total_visits'] ?? 0,
      normalVisits: json['normal_visits'] ?? 0,
      walkinVisits: json['walkin_visits'] ?? 0,
      totalPaid: (json['total_paid'] ?? 0).toDouble(),
      lastVisitDate: json['last_visit_date'] ?? '',
      appointments: (json['appointments'] as List? ?? [])
          .map((e) => DetailAppointment.fromJson(e))
          .toList(),
    );
  }
}

class DetailAppointment {
  final String appointmentId;
  final String date;
  final String time;
  final String type;
  final String status;
  final double feeAmount;
  final String paymentStatus;
  final String diagnosis;
  final String followupStatus;
  final String followupValidTill;

  DetailAppointment({
    required this.appointmentId,
    required this.date,
    required this.time,
    required this.type,
    required this.status,
    required this.feeAmount,
    required this.paymentStatus,
    required this.diagnosis,
    required this.followupStatus,
    required this.followupValidTill,
  });

  factory DetailAppointment.fromJson(Map<String, dynamic> json) {
    return DetailAppointment(
      appointmentId: json['appointment_id'] ?? '',
      date: json['date'] ?? '',
      time: json['time'] ?? '',
      type: json['type'] ?? '',
      status: json['status'] ?? '',
      feeAmount: (json['fee_amount'] ?? 0).toDouble(),
      paymentStatus: json['payment_status'] ?? '',
      diagnosis: json['diagnosis'] ?? '',
      followupStatus: json['followup_status'] ?? '',
      followupValidTill: json['followup_valid_till'] ?? '',
    );
  }
}

class VitalSignSummary {
  final String recordedAt;
  final String bloodPressure;
  final int pulseRate;
  final double temperature;
  final double weightKg;
  final int spo2;

  VitalSignSummary({
    required this.recordedAt,
    required this.bloodPressure,
    required this.pulseRate,
    required this.temperature,
    required this.weightKg,
    required this.spo2,
  });

  factory VitalSignSummary.fromJson(Map<String, dynamic> json) {
    return VitalSignSummary(
      recordedAt: json['recorded_at'] ?? '',
      bloodPressure: json['blood_pressure'] ?? '',
      pulseRate: json['pulse_rate'] ?? 0,
      temperature: (json['temperature'] ?? 0).toDouble(),
      weightKg: (json['weight_kg'] ?? 0).toDouble(),
      spo2: json['spo2'] ?? 0,
    );
  }
}

class TimelineEventSummary {
  final String date;
  final String type;
  final String description;

  TimelineEventSummary({
    required this.date,
    required this.type,
    required this.description,
  });

  factory TimelineEventSummary.fromJson(Map<String, dynamic> json) {
    return TimelineEventSummary(
      date: json['date'] ?? '',
      type: json['type'] ?? '',
      description: json['description'] ?? '',
    );
  }
}
```

---

## 2. API Service Implementation

Add this function to your `PatientService` or `ApiService` to fetch the data.

```dart
import 'dart:convert';
import 'package:http/http.dart' as http;

class PatientService {
  final String baseUrl = "http://YOUR_API_GATEWAY:8000/api"; // Usually Kong

  Future<PatientFullDetailsResponse> getPatientFullDetails(String patientId, String token) async {
    final response = await http.get(
      Uri.parse('$baseUrl/clinic-specific-patients/$patientId/details'),
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
    );

    if (response.statusCode == 200) {
      final jsonResponse = jsonDecode(response.body);
      return PatientFullDetailsResponse.fromJson(jsonResponse);
    } else {
      throw Exception('Failed to load patient details');
    }
  }
}
```

---

## 3. UI Screen Structure (Patient Details Profile view)

For a clean UI, use a Tabbed layout or a very nice scrolling page. Here is how your `PatientDetailsScreen` should map the JSON to widgets:

### Overview Layout Recommendation:
1. **Header Section**: Profile Picture placeholder + `patientInfo.fullName` + `patientInfo.phone` + `MOID`.
2. **Top Stats Row**:
   - Total Appointments (`totalAppointments`)
   - Total Spent (`totalSpent`)
   - Age/Blood Group (`patientInfo.age` / `patientInfo.bloodGroup`)
3. **Tab Bar**:
   - **Tab 1: Medical Details:** Vitals (`recentVitals`) + Medical History.
   - **Tab 2: Doctors Visited:** Expandable list mapping to `doctorVisits`.
   - **Tab 3: Timeline:** List showing `visitTimeline`.

### Example Widget Structure for "Doctors Visited" (The most complex part)
Because you wanted to know "how many this patient each doctor, payment money, all normal or walk-in", map the `DoctorVisitSummary` like this:

```dart
Widget _buildDoctorsVisitedList(List<DoctorVisitSummary> doctorVisits) {
  if (doctorVisits.isEmpty) return const Center(child: Text('No doctor visits yet'));

  return ListView.builder(
    itemCount: doctorVisits.length,
    itemBuilder: (context, index) {
      final summary = doctorVisits[index];
      
      return ExpansionTile(
        leading: const CircleAvatar(child: Icon(Icons.person)),
        title: Text(summary.doctorName),
        subtitle: Text('Total Visits: ${summary.totalVisits} • Paid: AED ${summary.totalPaid.toStringAsFixed(2)}'),
        children: [
          // Stat breakdown row
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('Walk-in: ${summary.walkinVisits}', style: const TextStyle(fontWeight: FontWeight.bold)),
                Text('Normal: ${summary.normalVisits}', style: const TextStyle(fontWeight: FontWeight.bold)),
                Text('Dept: ${summary.departmentName}'),
              ],
            ),
          ),
          
          // List of appointments under this doctor
          ...summary.appointments.map((appt) => ListTile(
            dense: true,
            leading: Icon(
              appt.type == 'walk_in' ? Icons.directions_walk : Icons.event,
              color: Colors.blue,
            ),
            title: Text('${appt.date} at ${appt.time}'),
            subtitle: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('Diagnosis: ${appt.diagnosis.isEmpty ? 'N/A' : appt.diagnosis}'),
                Text('Follow-up: ${appt.followupStatus} (Valid till ${appt.followupValidTill})',
                  style: TextStyle(
                    color: appt.followupStatus == 'active' ? Colors.green : Colors.grey
                  ),
                ),
              ],
            ),
            trailing: Text(
              '${appt.status}\n₹${appt.feeAmount}',
              textAlign: TextAlign.right,
              style: TextStyle(
                color: appt.paymentStatus == 'completed' ? Colors.green : Colors.orange,
              ),
            ),
          )).toList(),
        ],
      );
    },
  );
}
```

### Example Widget Structure for "Timeline"
```dart
Widget _buildTimeline(List<TimelineEventSummary> timeline) {
  return ListView.builder(
    itemCount: timeline.length,
    itemBuilder: (context, index) {
      final event = timeline[index];
      return ListTile(
        leading: const Icon(Icons.timeline),
        title: Text(event.description),
        subtitle: Text(event.date),
      );
    },
  );
}
```

## Summary
By using the **`PatientFullDetailsResponse`** Dart model inside a `FutureBuilder` or calling it inside your state manager (Provider/Riverpod), your Flutter frontend will have absolutely all the structured data needed to paint out the entire history of a patient perfectly without needing any extra N+1 API calls.
