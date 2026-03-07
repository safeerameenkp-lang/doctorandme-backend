# Frontend API Integration Guide - Recent Updates

This document contains the latest JSON request and response structures for the appointment and vitals APIs updated on 2026-03-07.

---

## 📅 1. Appointment Details (Fixed)
**Endpoint:** `GET /api/v1/appointments/simple/:id`
**Fixed:** Previously failing with 500 status. Now correctly returns patient and doctor details.

### JSON Response:
```json
{
  "success": true,
  "appointment": {
    "id": "e3a32683-08ce-4bd8-9e36-8f7c5261442c",
    "token_number": "DD001-01",
    "mo_id": "MO12345",
    "patient_number": "9876543210",
    "clinic_patient_id": "6ef8eff3-2cd0-44ec-8e83-3fe0642001ce",
    "patient_name": "Amal Kumar",
    "doctor_name": "Dr. Sameer Khan",
    "department": "General Medicine",
    "consultation_type": "clinic_visit",
    "appointment_date_time": "15-01-2024 10:00 AM",
    "status": "confirmed",
    "fee_amount": 500.0,
    "payment_status": "paid",
    "fee_status": "₹500.00",
    "booking_number": "BN20240115001",
    "booking_mode": "slot",
    "created_at": "2024-01-14 10:00:00"
  }
}
```

---

## 🔍 2. Patient Appointment History (New)
**Endpoint:** `GET /api/v1/appointments/history/:patient_id`
**Usage:** Fetches all appointments for a patient (Global or Clinic-Specific).

### JSON Response:
```json
{
  "success": true,
  "patient_id": "6ef8eff3-2cd0-44ec-8e83-3fe0642001ce",
  "total_count": 4,
  "appointments": [
    {
      "id": "e3a32683-08ce-4bd8-9e36-8f7c5261442c",
      "booking_number": "BN20240115001",
      "appointment_date_time": "15-01-2024 10:00 AM",
      "patient_name": "Amal Kumar",
      "doctor_name": "Dr. Sameer Khan",
      "clinic_name": "City Health Clinic",
      "department": "General Medicine",
      "consultation_type": "clinic_visit",
      "status": "confirmed",
      "fee_amount": 500.00,
      "payment_status": "paid",
      "booking_mode": "slot",
      "created_at": "2024-01-14T10:00:00Z",
      "mo_id": "MO12345"
    }
  ]
}
```

---

## 📈 3. Patient Vitals History (Optimized)
**Endpoint:** `GET /api/v1/vitals/clinic-patient/:patient_id`
**Optimized:** Now correctly fetches records for patients registered at the clinic only.

### JSON Response:
```json
{
  "patient_id": "6ef8eff3-2cd0-44ec-8e83-3fe0642001ce",
  "count": 1,
  "vitals_history": [
    {
      "id": "vitals-uuid",
      "appointment_id": "e3a32683-08ce-4bd8-9e36-8f7c5261442c",
      "clinic_patient_id": "6ef8eff3-2cd0-44ec-8e83-3fe0642001ce",
      "systolic_bp": 120,
      "diastolic_bp": 80,
      "blood_pressure": "120/80",
      "temperature": 98.6,
      "pulse_rate": 72,
      "resp_bpm": 18,
      "spo2_percent": 98,
      "sugar_mgdl": 110.0,
      "height_cm": 175,
      "weight_kg": 70.0,
      "bmi": 22.8,
      "smoking_status": "never",
      "alcohol_use": "none",
      "notes": "Patient is stable",
      "recorded_at": "2024-01-15T10:15:00Z",
      "updated_at": "2024-01-15T10:15:00Z",
      "appointment": {
        "booking_number": "BN20240115001",
        "appointment_time": "2024-01-15T10:00:00Z",
        "status": "confirmed"
      },
      "patient": {
        "first_name": "Amal",
        "last_name": "Kumar"
      },
      "doctor": {
        "first_name": "Sameer",
        "last_name": "Khan"
      }
    }
  ]
}
```

---

## 💉 4. Record Vitals (Migration Fix)
**Endpoint:** `POST /api/v1/vitals`
**Verification:** Added `clinic_patient_id` to database.

### JSON Request:
```json
{
  "appointment_id": "e3a32683-08ce-4bd8-9e36-8f7c5261442c",
  "recorded_by": "c3aea59e-f2b8-4b9c-b7af-ebb920ebfe4a",
  "clinic_patient_id": "6ef8eff3-2cd0-44ec-8e83-3fe0642001ce",
  "temperature": 98.6,
  "blood_pressure": "120/80",
  "systolic_bp": 120,
  "diastolic_bp": 80,
  "pulse_rate": 72,
  "resp_bpm": 18,
  "spo2_percent": 98,
  "sugar_mgdl": 110.0,
  "height_cm": 175,
  "weight_kg": 70.0,
  "notes": "Optional patient notes"
}
```
