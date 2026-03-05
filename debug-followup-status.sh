#!/bin/bash

# Debug Follow-Up Status Script
# This will help us understand why follow-up shows as paid instead of free

echo "🔍 DEBUGGING FOLLOW-UP STATUS"
echo "=============================="

# Get the patient ID and doctor ID from user
echo "Please provide the following information:"
echo "1. Patient ID (clinic_patient_id):"
read PATIENT_ID

echo "2. Doctor ID:"
read DOCTOR_ID

echo "3. Department ID (optional, press Enter to skip):"
read DEPARTMENT_ID

echo "4. Clinic ID:"
read CLINIC_ID

echo ""
echo "🔍 Checking appointment history..."

# Query 1: Get all regular appointments for this patient+doctor+department
echo ""
echo "📋 REGULAR APPOINTMENTS:"
if [ -n "$DEPARTMENT_ID" ]; then
    psql -d drandme_db -c "
    SELECT 
        a.id,
        a.appointment_date,
        a.consultation_type,
        a.status,
        COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name,
        dept.name as department
    FROM appointments a
    JOIN doctors d ON d.id = a.doctor_id
    JOIN users u ON u.id = d.user_id
    LEFT JOIN departments dept ON dept.id = a.department_id
    WHERE a.clinic_patient_id = '$PATIENT_ID'
      AND a.clinic_id = '$CLINIC_ID'
      AND a.doctor_id = '$DOCTOR_ID'
      AND a.department_id = '$DEPARTMENT_ID'
      AND a.consultation_type IN ('clinic_visit', 'video_consultation')
      AND a.status IN ('completed', 'confirmed')
    ORDER BY a.appointment_date DESC, a.appointment_time DESC;
    "
else
    psql -d drandme_db -c "
    SELECT 
        a.id,
        a.appointment_date,
        a.consultation_type,
        a.status,
        COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name,
        dept.name as department
    FROM appointments a
    JOIN doctors d ON d.id = a.doctor_id
    JOIN users u ON u.id = d.user_id
    LEFT JOIN departments dept ON dept.id = a.department_id
    WHERE a.clinic_patient_id = '$PATIENT_ID'
      AND a.clinic_id = '$CLINIC_ID'
      AND a.doctor_id = '$DOCTOR_ID'
      AND a.consultation_type IN ('clinic_visit', 'video_consultation')
      AND a.status IN ('completed', 'confirmed')
    ORDER BY a.appointment_date DESC, a.appointment_time DESC;
    "
fi

echo ""
echo "🔄 FOLLOW-UP APPOINTMENTS:"
if [ -n "$DEPARTMENT_ID" ]; then
    psql -d drandme_db -c "
    SELECT 
        a.id,
        a.appointment_date,
        a.consultation_type,
        a.payment_status,
        a.status,
        COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name,
        dept.name as department
    FROM appointments a
    JOIN doctors d ON d.id = a.doctor_id
    JOIN users u ON u.id = d.user_id
    LEFT JOIN departments dept ON dept.id = a.department_id
    WHERE a.clinic_patient_id = '$PATIENT_ID'
      AND a.clinic_id = '$CLINIC_ID'
      AND a.doctor_id = '$DOCTOR_ID'
      AND a.department_id = '$DEPARTMENT_ID'
      AND a.consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
      AND a.status NOT IN ('cancelled', 'no_show')
    ORDER BY a.appointment_date DESC, a.appointment_time DESC;
    "
else
    psql -d drandme_db -c "
    SELECT 
        a.id,
        a.appointment_date,
        a.consultation_type,
        a.payment_status,
        a.status,
        COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name,
        dept.name as department
    FROM appointments a
    JOIN doctors d ON d.id = a.doctor_id
    JOIN users u ON u.id = d.user_id
    LEFT JOIN departments dept ON dept.id = a.department_id
    WHERE a.clinic_patient_id = '$PATIENT_ID'
      AND a.clinic_id = '$CLINIC_ID'
      AND a.doctor_id = '$DOCTOR_ID'
      AND a.consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
      AND a.status NOT IN ('cancelled', 'no_show')
    ORDER BY a.appointment_date DESC, a.appointment_time DESC;
    "
fi

echo ""
echo "🎯 LATEST REGULAR APPOINTMENT DATE:"
if [ -n "$DEPARTMENT_ID" ]; then
    LATEST_DATE=$(psql -d drandme_db -t -c "
    SELECT a.appointment_date
    FROM appointments a
    WHERE a.clinic_patient_id = '$PATIENT_ID'
      AND a.clinic_id = '$CLINIC_ID'
      AND a.doctor_id = '$DOCTOR_ID'
      AND a.department_id = '$DEPARTMENT_ID'
      AND a.consultation_type IN ('clinic_visit', 'video_consultation')
      AND a.status IN ('completed', 'confirmed')
    ORDER BY a.appointment_date DESC, a.appointment_time DESC
    LIMIT 1;
    " | xargs)
else
    LATEST_DATE=$(psql -d drandme_db -t -c "
    SELECT a.appointment_date
    FROM appointments a
    WHERE a.clinic_patient_id = '$PATIENT_ID'
      AND a.clinic_id = '$CLINIC_ID'
      AND a.doctor_id = '$DOCTOR_ID'
      AND a.consultation_type IN ('clinic_visit', 'video_consultation')
      AND a.status IN ('completed', 'confirmed')
    ORDER BY a.appointment_date DESC, a.appointment_time DESC
    LIMIT 1;
    " | xargs)
fi

echo "Latest regular appointment: $LATEST_DATE"

echo ""
echo "🔢 FREE FOLLOW-UP COUNT (from latest date onward):"
if [ -n "$DEPARTMENT_ID" ]; then
    psql -d drandme_db -c "
    SELECT COUNT(*) as free_follow_up_count
    FROM appointments
    WHERE clinic_patient_id = '$PATIENT_ID'
      AND clinic_id = '$CLINIC_ID'
      AND doctor_id = '$DOCTOR_ID'
      AND department_id = '$DEPARTMENT_ID'
      AND consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
      AND payment_status = 'waived'
      AND appointment_date >= '$LATEST_DATE'
      AND status NOT IN ('cancelled', 'no_show');
    "
else
    psql -d drandme_db -c "
    SELECT COUNT(*) as free_follow_up_count
    FROM appointments
    WHERE clinic_patient_id = '$PATIENT_ID'
      AND clinic_id = '$CLINIC_ID'
      AND doctor_id = '$DOCTOR_ID'
      AND consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
      AND payment_status = 'waived'
      AND appointment_date >= '$LATEST_DATE'
      AND status NOT IN ('cancelled', 'no_show');
    "
fi

echo ""
echo "📊 ANALYSIS:"
echo "- If free_follow_up_count = 0 → Should show FREE (GREEN)"
echo "- If free_follow_up_count > 0 → Should show PAID (ORANGE)"
echo ""
echo "✅ Debug complete!"


