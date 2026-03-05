# Kwine Frontend Documentation

This document outlines the frontend features, tools used, and state management/folder structure for the Kwine application (Doctor & Me Client).

## 1. Overview
The Kwine frontend is built using **Flutter**, designed to provide a seamless experience for patients to book appointments, manage follow-ups, and view their medical history.

## 2. Tools & Technologies
- **Framework**: Flutter (Dart)
- **Networking**: `http` package (for API communication)
- **State Management**: `setState` (Local State) & Service-Repository Pattern
- **UI Components**: Material Design widgets (Scaffold, AppBar, Card, etc.)
- **Serialization**: `dart:convert` (JSON encoding/decoding)

---

## 3. Feature Breakdown & Structure

### 🏥 1. Appointment Booking
Allows patients to search for doctors, view available slots, and book new appointments.

- **Tools Used**: 
  - `http` (POST /api/appointments)
  - `showDatePicker` (Date selection)
  - `DropdownButton` (Payment mode)
  - `RadioListTile` (Consultation type)
- **State Folder & Files**:
  - **Screens**: `lib/screens/book_appointment_screen.dart` (UI & Local State)
  - **Service**: `lib/services/appointment_service.dart` (API Calls)
  - **Model**: `lib/models/appointment.dart` (Data Structure)
- **State Management**:
  - `_selectedDate`, `_selectedSlotId` managed via `setState`.
  - API loading state via `_isLoading` boolean.

### 🔄 2. Follow-Up Management
Handles logic for follow-up appointments, including eligibility checks (free vs paid) and validity tracking.

- **Tools Used**:
  - `http` (GET /api/appointments/followup-eligibility)
  - `AlertDialog` (Success/Failure feedback)
- **State Folder & Files**:
  - **Screens**: `lib/screens/patient_dashboard.dart` (Logic & UI)
  - **Service**: `lib/services/appointment_service.dart` (Check Eligibility)
  - **Model**: `lib/models/follow_up.dart` (Data Structure)
- **State Management**:
  - `followUpEligibility` object stores API response.
  - `checkFollowUpStatus` function updates UI based on eligibility.

### 👤 3. Patient Dashboard
Displays a list of patients linked to a specific clinic account (e.g., family members).

- **Tools Used**:
  - `ListView.builder` (Efficient list rendering)
  - `FutureBuilder` (Async data loading)
- **State Folder & Files**:
  - **Screens**: `lib/screens/patient_list_screen.dart`
  - **Service**: `lib/services/patient_service.dart`
  - **Model**: `lib/models/patient.dart`
- **State Management**:
  - `patients` list stored in state.
  - `fetchPatients()` method updates the list on load.

### 🔐 4. Authentication (Login/Register)
Secure access for users to manage their appointments.

- **Tools Used**:
  - `FlutterSecureStorage` (Token storage - Recommended)
  - `TextFormField` (Input validation)
- **State Folder & Files**:
  - **Screens**: `lib/screens/login_screen.dart`
  - **Service**: `lib/services/auth_service.dart`
- **State Management**:
  - `_token` stored locally.
  - Auth status checked on app startup.

---

## 4. Directory Structure (Recommended)

```
lib/
├── main.dart                 # App Entry Point
├── models/                   # Data Models
│   ├── appointment.dart      # Appointment Structure
│   ├── patient.dart          # Patient Structure
│   └── follow_up.dart        # Follow-Up Logic
├── services/                 # API Logic
│   ├── appointment_service.dart  # Booking & Slots
│   ├── auth_service.dart     # Login/Logout
│   └── patient_service.dart  # Patient Fetching
├── screens/                  # UI Pages
│   ├── login_screen.dart
│   ├── book_appointment_screen.dart
│   └── patient_dashboard.dart
└── widgets/                  # Reusable Components
    ├── custom_button.dart
    └── slot_chip.dart
```

## 5. Key API Integration Points
- **Base URL**: `http://localhost:8082/api` (or production URL)
- **Headers**: 
  - `Content-Type: application/json`
  - `Authorization: Bearer <token>`
