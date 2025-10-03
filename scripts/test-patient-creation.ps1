#!/bin/bash

# Test script for patient creation during appointment booking
# This script tests the new endpoint for creating patients with appointments

echo "Testing Patient Creation with Appointment Booking..."

# Base URL for the appointment service
BASE_URL="http://localhost:8080/api/appointments"

# Test data for creating a patient with appointment
TEST_DATA='{
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "email": "john.doe@example.com",
  "date_of_birth": "1990-01-15",
  "gender": "Male",
  "mo_id": "MO123456",
  "medical_history": "No significant medical history",
  "allergies": "None known",
  "blood_group": "O+",
  "clinic_id": "clinic-uuid-here",
  "doctor_id": "doctor-uuid-here",
  "appointment_time": "2024-01-20 10:00:00",
  "duration_minutes": 15,
  "consultation_type": "new",
  "is_priority": false,
  "payment_mode": "cash"
}'

echo "Test Data:"
echo "$TEST_DATA" | jq '.'

echo ""
echo "Testing POST /api/appointments/with-patient"
echo "Note: You need to replace clinic_id and doctor_id with actual UUIDs from your database"
echo ""

# Make the API call
curl -X POST "$BASE_URL/with-patient" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d "$TEST_DATA" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "Test completed!"
echo ""
echo "Expected Response Structure:"
echo '{
  "appointment": {
    "id": "appointment-uuid",
    "patient_id": "patient-uuid",
    "clinic_id": "clinic-uuid",
    "doctor_id": "doctor-uuid",
    "booking_number": "DOC001-20240120-001",
    "appointment_time": "2024-01-20T10:00:00Z",
    "duration_minutes": 15,
    "consultation_type": "new",
    "status": "booked",
    "fee_amount": 100.00,
    "payment_status": "paid",
    "payment_mode": "cash",
    "is_priority": false,
    "created_at": "2024-01-20T09:00:00Z"
  },
  "patient": {
    "id": "patient-uuid",
    "user_id": "user-uuid",
    "mo_id": "MO123456",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890",
    "email": "john.doe@example.com",
    "medical_history": "No significant medical history",
    "allergies": "None known",
    "blood_group": "O+"
  },
  "message": "Patient created and appointment booked successfully"
}'

